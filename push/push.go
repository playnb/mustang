package push

import (
	"github.com/playnb/mustang/log"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"strings"
)

type Message struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

//保存推送Token的Redis
var _pushTokenRedis *redis.Client
//通过token查找AccidKey
var _keyPushTokenAccid func(string) string
//通过Accid查找Token的Key
var _keyAccidPushToken func(string) string

var _cfg *viper.Viper
/*
android
ios
huawei
*/

func Init(r *redis.Client, fTA func(string) string, fAT func(string) string, cfg *viper.Viper) {
	_keyPushTokenAccid = fTA
	_keyAccidPushToken = fAT
	_pushTokenRedis = r
	_cfg = cfg
}

//设置推送Token
func SetPushToken(accid string, platform string, token string) {
	platform = strings.ToLower(platform)
	t := platform + ":" + token
	oldID, err := _pushTokenRedis.Get(_keyPushTokenAccid(t)).Result()
	if err == nil && len(oldID) != 0 {
		_pushTokenRedis.Del(_keyAccidPushToken(oldID))
	}
	_pushTokenRedis.Set(_keyPushTokenAccid(t), accid, 0)
	_pushTokenRedis.Set(_keyAccidPushToken(accid), t, 0)
}

func GetPushToken(accid string) (string, string) {
	t, err := _pushTokenRedis.Get(_keyAccidPushToken(accid)).Result()
	if err == nil {
		ss := strings.SplitN(t, ":", 2)
		if len(ss) == 2 {
			return ss[0], ss[1]
		}
	}
	return "", ""
}

//推送消息
func PostMsg(accid string, msg *Message) {
	var err error
	device, token := GetPushToken(accid)
	switch device {
	case "android":
		err = umengAndroidPush(token, msg.Title, msg.Title, msg.Content)
	case "ios":
		err = umengIOSPush(token, msg.Title, msg.Title, msg.Content)
	case "huawei":
		err = huaweiPush(token, msg)
	default:
	}
	log.Dev("PushMsg[%d] %s", accid, err.Error())
}
