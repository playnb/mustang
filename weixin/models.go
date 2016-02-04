package weixin

//这里都是数据模型

import (
	"github.com/astaxie/beego/orm"
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/sqldb"
)

func init() {
	// 需要在init中注册定义的model
	orm.RegisterModel(new(WeiXinAccessToken), new(WeiXinJsApiTicket))
}

//AccessToken
type WeiXinAccessToken struct {
	AppID       string `orm:"pk"`
	AccessToken string
	ExpireIn    int64
	Remark      string
}

func (token *WeiXinAccessToken) Load(appID string) bool {
	token.AppID = appID
	err := sqldb.Ormer.Read(token)
	if err == nil {
		return true
	} else if err == orm.ErrNoRows {
		log.Error("查询不到")
		return false
	} else if err == orm.ErrMissPK {
		log.Error("找不到主键")
		return false
	}
	return false
}

func (token *WeiXinAccessToken) Save() {
	var err error
	if sqldb.Ormer.QueryTable(token).Filter("AppID", token.AppID).Exist() {
		_, err = sqldb.Ormer.Update(token)
	} else {
		_, err = sqldb.Ormer.Insert(token)
	}
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Trace("保存微信AccessToken")
	}
}

//JsApiTicket
type WeiXinJsApiTicket struct {
	AppID       string `orm:"pk"`
	JsApiTicket string
	ExpireIn    int64
	Remark      string
}

func (token *WeiXinJsApiTicket) Load(appID string) bool {
	token.AppID = appID
	err := sqldb.Ormer.Read(token)
	if err == nil {
		return true
	} else if err == orm.ErrNoRows {
		log.Error("查询不到")
		return false
	} else if err == orm.ErrMissPK {
		log.Error("找不到主键")
		return false
	}
	return false
}

func (token *WeiXinJsApiTicket) Save() {
	var err error
	if sqldb.Ormer.QueryTable(token).Filter("AppID", token.AppID).Exist() {
		_, err = sqldb.Ormer.Update(token)
	} else {
		_, err = sqldb.Ormer.Insert(token)
	}
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Trace("保存微信AccessToken")
	}
}
