package weixin

import (
	"github.com/playnb/mustang/log"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//获取微信服务器的Token
func RefreshAccessToken(loop bool) {
	type wxJsonToken struct {
		WeiXinError
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	wxToken := new(WeiXinAccessToken)
	if !wxToken.Load(wxProfile.AppID) {
		wxToken.ExpireIn = 0
		wxProfile.AccessToken = ""
	} else {
		wxProfile.AccessToken = wxToken.AccessToken
		log.Trace("加载数据库微信Accesstoken: %s", wxProfile.AccessToken)
	}

	for loop {
		httpClient := GetHTTPClient()
		newSec := time.Now().UnixNano() / time.Second.Nanoseconds()
		for newSec < wxToken.ExpireIn {
			time.Sleep(time.Second)
			newSec = time.Now().UnixNano() / time.Second.Nanoseconds()
		}
		wxProfile.AccessToken = ""
		log.Trace("微信Accesstoken过期,重新获取Token")
		url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid={APPID}&secret={APPSECRET}"
		url = strings.Replace(url, "{APPID}", wxProfile.AppID, -1)
		url = strings.Replace(url, "{APPSECRET}", wxProfile.AppSecret, -1)
		//log.Debug(url)

		res, err := httpClient.Get(url)
		if err != nil {
			log.Error("获取Token失败 %v", err)
		}
		result, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Error("解析Token失败 %v", err)
		}
		log.Debug(string(result))
		var jsonData wxJsonToken
		if err := json.Unmarshal(result, &jsonData); err == nil {
			wxProfile.AccessToken = jsonData.AccessToken
			log.Trace("收到Token: %s", wxProfile.AccessToken)

			wxToken.AccessToken = jsonData.AccessToken
			wxToken.ExpireIn = newSec + int64(jsonData.ExpiresIn) - 60*10
			wxToken.Save()
		} else {
			log.Error(err.Error())
		}
	}
}

//刷新JsApiTicket
func RefreshJsApiTicket(loop bool) {
	type wxJsonToken struct {
		WeiXinError
		JsApiTicket string `json:"ticket"`
		ExpiresIn   int    `json:"expires_in"`
	}

	wxTicket := new(WeiXinJsApiTicket)
	if !wxTicket.Load(wxProfile.AppID) {
		wxTicket.ExpireIn = 0
	} else {
		wxProfile.JsApiTicket = wxTicket.JsApiTicket
		log.Trace("加载数据库微信JsApiTicket: %s", wxProfile.JsApiTicket)
	}

	for loop {
		newSec := time.Now().UnixNano() / time.Second.Nanoseconds()
		for newSec < wxTicket.ExpireIn || len(wxProfile.AccessToken) == 0 {
			time.Sleep(time.Second)
			newSec = time.Now().UnixNano() / time.Second.Nanoseconds()
		}
		log.Trace("微信JsApiTicket过期,重新获取JsApiTicket")
		url := "https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=ACCESS_TOKEN&type=jsapi"
		url = strings.Replace(url, "ACCESS_TOKEN", wxProfile.AccessToken, -1)
		//log.Debug(url)
		res, err := http.Get(url)
		if err != nil {
			log.Error("获取JsApiTicket失败 %v", err)
		}
		result, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Error("解析JsApiTicket失败 %v", err)
		}
		log.Debug(string(result))
		var jsonData wxJsonToken
		if err := json.Unmarshal(result, &jsonData); err == nil {
			wxProfile.JsApiTicket = wxTicket.JsApiTicket
			log.Trace("收到JsApiTicket: %s", wxProfile.AccessToken)

			wxTicket.JsApiTicket = jsonData.JsApiTicket
			wxTicket.ExpireIn = newSec + int64(jsonData.ExpiresIn) - 60*10
			wxTicket.Save()
		} else {
			log.Error(err.Error())
		}
	}
}
