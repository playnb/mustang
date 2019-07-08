package weixin

import (
	"cell/common/mustang/log"
	"crypto/sha1"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
//wxMsgToken = "test_wx" //测试微信的TOKEN
//wxAuthServiceUrl  = "http://liutp.vicp.net"                     //用于微信网页授权认证的域名
//wxAuthRedirectURL = wxAuthServiceUrl + "/wx/authorize_redirect" //用于微信网页授权认证的回调页
)

func UrlEncode(src string) string {
	return url.QueryEscape(src)
}
func UrlDecode(src string) string {
	ret, _ := url.QueryUnescape(src)
	return ret
}

type ProcessWeixinMessage func(*RecvWeiXinMsg, *SendWeiXinMsg) error                                         // 处理微信来的消息
type OnUserAuthorized func(state string, token *WebAccessToken, w http.ResponseWriter, r *http.Request) bool //有人授权时候调用

//配置类型
type WeiXinConfig struct {
	AppID     string
	AppSecret string
	OwnerID   string

	Master bool

	MyServiceDomain string
	AuthRedirectURL string

	ServiceToken string //原始表明身份的token

	//一些处理回调函数
	ProcessMsg     ProcessWeixinMessage
	UserAuthorized OnUserAuthorized
}

type WeiXinProfile struct {
	WeiXinConfig

	AccessToken string
	JsApiTicket string

	msgCollection map[string]bool
	mux           *http.ServeMux

	authRedirectAppUrl map[string]string
}

//核心对象
var wxProfile *WeiXinProfile

//Profile 返回核心对象
func Profile() *WeiXinProfile {
	return wxProfile
}

//RegistAuthRedirectUrl 注册处理redirecturl
func RegistAuthRedirectUrl(state string, redirect string) {
	url, ok := wxProfile.authRedirectAppUrl[state]
	if ok == true {
		log.Error(" 重复注册APP:%s的authRedirectAppUrl:%s (已经存在%s)", state, redirect, url)
	} else {
		wxProfile.authRedirectAppUrl[state] = redirect
	}
}

//InitWeiXin 初始化微信服务(要使用微信 就要先用这个函数初始化)
//func InitWeiXin(appID, appSecret, ownerID string, mux *http.ServeMux, processFunc ProcessWeixinMessage) error {
func InitWeiXin(config *WeiXinConfig, mux *http.ServeMux) error {
	if wxProfile == nil {
		log.Trace("APPID: %s ", config.AppID)
		wxProfile = &WeiXinProfile{}
		wxProfile.msgCollection = make(map[string]bool)
		wxProfile.authRedirectAppUrl = make(map[string]string)
		//wxProfile.AppID = appID
		//wxProfile.AppSecret = appSecret
		//wxProfile.OwnerID = ownerID
		wxProfile.WeiXinConfig = *config
		wxProfile.mux = mux
		if mux != nil {
			mux.HandleFunc("/wx/msg", handleWeiXinRoot)
		}
		InitWiXinOAuth()
		go RefreshAccessToken(true)
		go RefreshJsApiTicket(true)
		return nil
	}
	return errors.New("Already inited")
}

type WeiXinError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

//构造微信验证码
func makeSignature(timestamp, nonce, token string) string {
	sl := []string{token, timestamp, nonce}
	sort.Strings(sl)
	s := sha1.New()
	io.WriteString(s, strings.Join(sl, ""))
	return fmt.Sprintf("%x", s.Sum(nil))
}

//验证微信服务器请求
func checkWeiXinUrlVailed(r *http.Request) bool {
	timestamp := strings.Join(r.Form["timestamp"], "")
	nonce := strings.Join(r.Form["nonce"], "")
	signatureGen := makeSignature(timestamp, nonce, wxProfile.ServiceToken)
	signatureIn := strings.Join(r.Form["signature"], "")
	return signatureGen == signatureIn
}

func checkWeiXinReturn(retData *WeiXinError) bool {
	if retData.ErrCode != 0 {
		log.Error(retData.ErrMsg)
		return false
	}
	return true
}

//收到微信服务器的消息
type RecvWeiXinMsg struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   string   `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
	MsgId        string   `xml:"MsgId"`
}

//返回微信服务器的消息(MsgType,Content需要填写,其他默认已经填写好了)
type SendWeiXinMsg struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   string   `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
}

//解析微信服务器转发的用户消息
func parseRecvWeiXinMsg(content []byte) *RecvWeiXinMsg {
	msg := new(RecvWeiXinMsg)
	err := xml.Unmarshal(content, msg)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return msg
}

func handleWeiXinRoot(w http.ResponseWriter, r *http.Request) {
	log.Debug("\n--")
	log.Debug("Visit: " + r.RequestURI)
	r.ParseForm()
	if !checkWeiXinUrlVailed(r) {
		w.WriteHeader(505)
		log.Trace("微信认证失败")
		return
	}
	if r.Method == "GET" {
		//微信服务器验证请求
		echostr := strings.Join(r.Form["echostr"], "")
		fmt.Fprintf(w, echostr)
		log.Trace("微信认证成功")

	} else {
		log.Debug("其他请求 %v", r)
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		log.Debug("%s\n", result)
		revMsg := parseRecvWeiXinMsg(result)
		if revMsg == nil {
			return
		}
		if _, ok := wxProfile.msgCollection[revMsg.MsgId]; ok == true {
			log.Debug("过滤重复MsgID: %s", revMsg.MsgId)
			return
		}
		wxProfile.msgCollection[revMsg.MsgId] = true
		sendMsg := new(SendWeiXinMsg)
		sendMsg.FromUserName = wxProfile.OwnerID
		sendMsg.ToUserName = revMsg.FromUserName
		sendMsg.CreateTime = fmt.Sprintf("%d", time.Now().Second())

		if wxProfile.ProcessMsg == nil {
			sendMsg.MsgType = "text"
			if revMsg.Content == "我是谁" {
				info := GetUserInfo(revMsg.FromUserName, true)
				if info != nil {
					sendMsg.Content = "你是: " + info.NikeName
				} else {
					sendMsg.Content = "不认识你哦..."
				}
			} else {
				sendMsg.Content = "回你: " + revMsg.Content
			}
		} else {
			wxProfile.ProcessMsg(revMsg, sendMsg)
		}
		bytes, err := xml.Marshal(sendMsg)
		if err != nil {
			w.WriteHeader(500)
			log.Error(err.Error())
			return
		} else {
			w.WriteHeader(200)
			w.Write(bytes)
			log.Debug("微信返回: %s", string(bytes))
		}
	}
}

//GetHTTPClient 内网用代理服务器测试
func GetHTTPClient() *http.Client {
	return &http.Client{}
	/*
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse("https://192.168.219.90:808")
		}

		transport := &http.Transport{Proxy: proxy}

		return &http.Client{Transport: transport}
	*/
}
