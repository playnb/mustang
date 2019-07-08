package weixin

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/redis"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

//微信用户信息
type WeiXinUserInfo struct {
	WeiXinError
	Subscribe     int    `json:"subscribe"`
	OpenID        string `json:"openid"`
	NikeName      string `json:"nickname"`
	Sex           int    `json:"sex"`
	Language      string `json:"language"`
	City          string `json:"city"`
	Province      string `json:"province"`
	Country       string `json:"country"`
	HeadImgUrl    string `json:"headimgurl"`
	SubscribeTime int    `json:"subscribe_time"`
	unionid       string `json:"unionid"`
	remark        string `json:"remark"`
	groupid       int    `json:"groupid"`
}

//获取微信用户信息
//TODO: 加入缓存的处理,一般数据直接从redis里面读取
func GetUserInfo(openID string, cache bool) *WeiXinUserInfo {
	return getUserInfo("", openID, false, cache)
}

func GetUserInfoBySNS(token string, openID string, cache bool) *WeiXinUserInfo {
	return getUserInfo(token, openID, true, cache)
}

func getUserInfo(token string, openID string, sns bool, cache bool) *WeiXinUserInfo {
	key := "WX:UserInfo:" + openID
	if cache {
		rc := redis.GetRedis()
		defer redis.PutRedis(rc)

		ret, err := rc.Get(key).Result()
		if err == nil {
			userinfo := &WeiXinUserInfo{}
			err = json.Unmarshal([]byte(ret), userinfo)
			if err == nil {
				return userinfo
			} else {
				log.Error(err.Error())
			}
		} else {
			log.Error(err.Error())
		}
	}

	if wxProfile == nil {
		log.Error("还没有获取AccessToken")
		return nil
	}
	if len(wxProfile.AccessToken) == 0 {
		log.Error("还没有获取AccessToken")
		return nil
	}

	url := ""
	if sns {
		url = "https://api.weixin.qq.com/sns/userinfo?access_token={{.AccessToken}}&openid={{.OpenID}}&lang=zh_CN"
		url = strings.Replace(url, "{{.AccessToken}}", token, -1)
		url = strings.Replace(url, "{{.OpenID}}", openID, -1)
	} else {
		url = "https://api.weixin.qq.com/cgi-bin/user/info?access_token={{.ACCESS_TOKEN}}&openid={{.OPENID}}&lang=zh_CN"
		url = strings.Replace(url, "{{.ACCESS_TOKEN}}", wxProfile.AccessToken, -1)
		url = strings.Replace(url, "{{.OPENID}}", openID, -1)
	}
	res, err := http.Get(url)
	if err != nil {
		log.Trace("请求URL: %s", url)
		log.Error("获取微信用户信息失败 %v", err)
		return nil
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Error("解析微信用户信息失败 %v", err)
		return nil
	}
	log.Debug(string(result))
	info := &WeiXinUserInfo{}
	if err := json.Unmarshal(result, info); err == nil {
		rc := redis.GetRedis()
		defer redis.PutRedis(rc)

		err = rc.Set(key, string(result), 0).Err()
		if err != nil {
			log.Error("Redis Error: %s", err.Error())
		}
		if checkWeiXinReturn(&(info.WeiXinError)) {
			return info
		}
	} else {
		log.Error(err.Error())
	}
	return nil
}
