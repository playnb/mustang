package weixin

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/redis"
	"encoding/json"
)

//这里都是数据模型

func init() {
	// 需要在init中注册定义的model
}

//WeiXinAccessToken  微信的token_access结构
type WeiXinAccessToken struct {
	AppID       string `json:"app_id"`
	AccessToken string `json:"access_token"`
	ExpireIn    int64  `json:"expire_in"`
	Remark      string `json:"remark"`
}

//Load 从数据库加载token
func (token *WeiXinAccessToken) Load(appID string) bool {
	data, err := redis.GetPoolClient().Get("WX:Access:Token").Bytes()
	if err == nil {
		err = json.Unmarshal(data, token)
		if err == nil {
			return true
		}
	}
	log.Trace("加载 WeiXinAccessToken 失败")
	return false
}

//Save 保存token到数据库
func (token *WeiXinAccessToken) Save() {
	data, err := json.Marshal(token)
	if err == nil {
		redis.GetPoolClient().Set("WX:Access:Token", data, 0)
	}
	return
}

//WeiXinJsApiTicket js凭证
type WeiXinJsApiTicket struct {
	AppID       string `json:"app_id"`
	JsApiTicket string `json:"js_api_ticket"`
	ExpireIn    int64  `json:"expire_in"`
	Remark      string `json:"remark"`
}

func (token *WeiXinJsApiTicket) Load(appID string) bool {
	data, err := redis.GetPoolClient().Get("WX:Access:Token").Bytes()
	if err == nil {
		err = json.Unmarshal(data, token)
		if err == nil {
			return true
		}
	}
	log.Trace("加载 WeiXinJsApiTicket 失败")
	return false
}

func (token *WeiXinJsApiTicket) Save() {
	data, err := json.Marshal(token)
	if err == nil {
		redis.GetPoolClient().Set("WX:Access:Token", data, 0)
	}
	return
}
