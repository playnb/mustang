package push

import (
	"github.com/playnb/mustang/log"
	"encoding/json"
	"fmt"
	"github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	huaweiAppID     = _cfg.GetString("push.huawei.appid")  //"100664631"
	huaweiAppSecret = _cfg.GetString("push.huawei.secret") // "1ca76c5d4877850b9959bdc78ca1f948"
)

type huaweiTokenInfo struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	Error            int    `json:"error"`
	ErrorDescription int    `json:"error_description"`
}

type nspCTX struct {
	Ver   string `json:"ver"`
	AppId string `json:"appId"`
}

func (ctx *nspCTX) URL() string {
	s, e := jsoniter.MarshalToString(ctx)
	if e == nil {
		return ""
	}
	return url.QueryEscape(s)
}

var huaweiToken = &huaweiTokenInfo{}
var huaweiCtx = &nspCTX{}

func tokenRoutineHuaWei() {
	for {
		huaweiToken = getHuaWeiTokenInfo()

		seconds := 1000
		if huaweiToken.ExpiresIn > 0 {
			seconds = huaweiToken.ExpiresIn * 1000 / 2
		}

		// seconds = 2 // TODO: DELETE...

		timer1 := time.NewTimer(time.Duration(seconds) * time.Second)
		<-timer1.C
	}
}

func getHuaWeiTokenInfo() *huaweiTokenInfo {
	tokenInfo := &huaweiTokenInfo{}

	client := &http.Client{}
	var r http.Request
	r.ParseForm()
	r.Form.Add("grant_type", "client_credentials")
	r.Form.Add("client_secret", huaweiAppSecret)
	r.Form.Add("client_id", huaweiAppID)
	bodystr := strings.TrimSpace(r.Form.Encode())
	req, err := http.NewRequest("POST", "https://login.vmall.com/oauth2/token", strings.NewReader(bodystr))
	if err != nil {
		log.Error("getHuaWeiTokenInfo %s", err)
		return tokenInfo
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Error("getHuaWeiTokenInfo %s", err)
		return tokenInfo
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("getHuaWeiTokenInfo %s", err)
		return tokenInfo
	}

	err = json.Unmarshal(result, &tokenInfo)
	if err != nil {
		log.Error("getHuaWeiTokenInfo %s", err)
		return tokenInfo
	}
	return tokenInfo
}

func init() {
	huaweiCtx.Ver = "1"
	huaweiCtx.AppId = huaweiAppID

	go tokenRoutineHuaWei()
}

type DeviceInfo struct {
	CID      string
	Platform uint32
	PushType uint32
}

func huaweiPush(token string, message *Message) error {
	data, err := json.Marshal([1]string{token})
	if err != nil {
		log.Error("huaweiPush %s", err.Error())
		return err
	}
	return postMessageToHuaWei(message.Title, message.Content, string(data))
}

var payloadTpl = `{
  "hps": {
    "msg": {
      "type": 3,
      "body": {
        "content": "%v",
        "title": "%v"
      },
      "action": {
        "type": 1,
        "param": {
           "intent": "#Intent;compo=com.ztgame.ztas/com.ztgame.ztcommunity.MainActivity;"
        }
      }
    }
  }
}`

//
// 将数据发送至华为平台
//
func postMessageToHuaWei(title, content, tokenList string) error {
	url := "https://api.push.hicloud.com/pushsend.do?nsp_ctx=" + huaweiCtx.URL()
	payload := fmt.Sprintf(payloadTpl, content, title)

	now := time.Now()
	nsp_ts := time.Now().Unix()
	expireTime := fmt.Sprintf("%d-%d-%dT23:50", now.Year(), now.Month(), now.Day())

	client := &http.Client{}
	var r http.Request
	r.ParseForm()
	r.Form.Add("payload", payload)
	r.Form.Add("access_token", huaweiToken.AccessToken)
	r.Form.Add("nsp_svc", "openpush.message.api.send")
	r.Form.Add("nsp_ts", fmt.Sprintf("%v", nsp_ts))
	r.Form.Add("expire_time", expireTime)
	r.Form.Add("device_token_list", tokenList)
	bodystr := strings.TrimSpace(r.Form.Encode())
	req, err := http.NewRequest("POST", url, strings.NewReader(bodystr))
	if err != nil {
		log.Error("postMessageToHuaWei %d", err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Error("postMessageToHuaWei %d", err)
		return err
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("postMessageToHuaWei %d", err)
		return err
	}
	log.Dev("[PUSH]HuaWei %d", result)

	return nil
}
