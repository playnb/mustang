package push

import (
	"bytes"
	"cell/common/mustang/log"
	"cell/common/mustang/util"
	"fmt"
	"github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"strconv"
)

type BoolString string

const (
	TRUE  = BoolString("true")
	FALSE = BoolString("false")
)

var (
	umengPushURL = "https://msgapi.umeng.com/api/send" //?sign=

	//企业版
	umengIOSAppkey = _cfg.GetString("push.umeng.ios.appkey") //"5caca2a661f5647a7c0004c0"
	umengIOSSecret = _cfg.GetString("push.umeng.ios.secret") //"ejvvmoozqcgvoriwsdsoyxax0jw0xddt"
	//正式版
	/*
		umengIOSAppkey = "5c86356161f564479e001220"
		umengIOSSecret = "e9ma1vfszkgycblo2ohacnajgddpm1rp"
	*/
	//踩坑版
	/*
		umengIOSAppkey = "5cb3ebea0cafb265c7000887"
		umengIOSSecret = "tyazemrevpxawfdvmhhdh808w4fkkeuj"
	*/
	umengMiPushActivity = _cfg.GetString("push.umeng.android.activity") //com.ztgame.ztcommunity.MainActivity"

	umengAndroidAppkey = _cfg.GetString("push.umeng.android.appkey") //"5c86342f61f5642657001062"
	umengAndroidSecret = _cfg.GetString("push.umeng.android.secret") //"0zckg9aksfbpfi31wj6xv8ymh8jhitfs"
)

type umengAndroidPushBody struct {
	// 当display_type=message时，body的内容只需填写custom字段。
	// 当display_type=notification时，body包含如下参数:
	Ticker      string     `json:"ticker"`       // 必填，通知栏提示文字
	Title       string     `json:"title"`        // 必填，通知标题
	Text        string     `json:"text"`         // 必填，通知文字描述
	Icon        string     `json:"icon"`         // 可选，状态栏图标ID，R.drawable.[smallIcon]，
	LargeIcon   string     `json:"largeIcon"`    // 可选，通知栏拉开后左侧图标ID，R.drawable.[largeIcon]，
	Img         string     `json:"img"`          // 可选，通知栏大图标的URL链接。该字段的优先级大于largeIcon(HTTPS)
	Sound       string     `json:"sound"`        // 可选，通知声音，R.raw.[sound]
	BuilderID   int        `json:"builder_id"`   // 可选，默认为0，用于标识该通知采用的样式。使用该参数时，开发者必须在SDK里面实现自定义通知栏样式。
	PlayVibrate BoolString `json:"play_vibrate"` // 可选，收到通知是否震动，默认为"true"
	PlayLights  BoolString `json:"play_lights"`  // 可选，收到通知是否闪灯，默认为"true"
	PlaySound   BoolString `json:"play_sound"`   // 可选，收到通知是否发出声音，默认为"true"
	AfterOpen   string     `json:"after_open"`   //"go_app": 打开应用  "go_url": 跳转到URL  "go_activity": 打开特定的activity  "go_custom": 用户自定义内容
	Url         string     `json:"url"`          //通知栏点击后跳转的URL，要求以http或者https开头
	Activity    string     `json:"activity"`     //通知栏点击后打开的Activity
	Custom      struct{}   `json:"custom"`       //用户自定义内容，可以为字符串或者JSON格式。
}
type umengAndroidPushPayload struct {
	DisplayType string               `json:"display_type"` // 必填，消息类型: notification(通知)、message(消息)
	Body        umengAndroidPushBody `json:"body"`
}
type umengAndroidPushRequest struct {
	AppKey       string                  `json:"appkey"`
	Timestamp    string                  `json:"timestamp"`
	Type         string                  `json:"type"` //消息发送类型unicast	listcast ...
	DeviceTokens string                  `json:"device_tokens"`
	Payload      umengAndroidPushPayload `json:"payload"`
	Description  string                  `json:"description"`
	MiPush       bool                    `json:"mipush"`      //可选，默认为false。当为true时，表示MIUI、EMUI、Flyme系统设备离线转为系统下发
	MiActivity   string                  `json:"mi_activity"` // 可选，mipush值为true时生效，表示走系统通道时打开指定页面acitivity的完整包路径
}

type umengAPNsAlert struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Body     string `json:"body"`
}

type umengAPNs struct {
	Alert umengAPNsAlert `json:"alert"`
	Badge int            `json:"badge"`
	Sound string         `json:"sound"`
	//ContentAvailable int            `json:"content-available"` //1 代表静默推送
	//Category         string         `json:"category"`
}

type umengIOSPushPayload struct {
	Aps  umengAPNs `json:"aps"`
	Key1 string    `json:"key1"`
	Key2 string    `json:"key2"`
}

type umengIOSPushRequest struct {
	AppKey       string              `json:"appkey"`
	Timestamp    string              `json:"timestamp"`
	Type         string              `json:"type"` //消息发送类型unicast	listcast ...
	DeviceTokens string              `json:"device_tokens"`
	Payload      umengIOSPushPayload `json:"payload"`
	Description  string              `json:"description"` // 可选，发送消息描述，建议填写
}

type umengRetData struct {
	MsgId     string `json:"msg_id"`     // 单播类消息(type为unicast、listcast、customizedcast且不带file_id)返回：
	TaskId    string `json:"task_id"`    //任务类消息(type为broadcast、groupcast、filecast、customizedcast且file_id不为空)返回：
	ErrorCode string `json:"error_code"` // 错误码，详见附录I
	ErrorMsg  string `json:"error_msg"`  // 错误信息
}
type umengRetCode struct {
	Ret  string       `json:"ret"`
	Data umengRetData `json:"data"`
}

func umengUrlPush(secret string, postBody string) error {
	sign := "POST" + umengPushURL + postBody + secret
	sign = util.Md5String(sign)

	url := umengPushURL + "?sign=" + sign
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(postBody)))
	if err != nil {
		log.Error("PUSH %s", err.Error())
		return err
	}
	var client = &http.Client{}
	rsp, _ := client.Do(req)
	if rsp.StatusCode != http.StatusOK {
		log.Error("PUSH %s", rsp.Status)
		return err
	}
	body, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		log.Error("PUSH %s", err.Error())
		return err
	}

	ret := &umengRetCode{}
	err = jsoniter.Unmarshal(body, ret)
	if err != nil {
		log.Error("PUSH %s", err.Error())
		return err
	}
	log.Dev("[PUSH] Req:%s Rsp:%v", postBody, ret)
	if ret.Ret != "SUCCESS" {
		return fmt.Errorf("%s", err.Error())
	}
	return nil
}

func umengAndroidPush(token string, Title string, Subtitle string, Body string) error {
	request := &umengAndroidPushRequest{
		AppKey:       umengAndroidAppkey,
		Timestamp:    strconv.FormatInt(util.AppTimestamp(), 10),
		Type:         "unicast",
		DeviceTokens: token,
		Description:  "androidPush",
		MiPush:       true,
		MiActivity:   umengMiPushActivity,
		Payload: umengAndroidPushPayload{
			DisplayType: "notification",
			Body: umengAndroidPushBody{
				Ticker:      Title,
				Title:       Subtitle,
				Text:        Body,
				PlayLights:  TRUE,
				PlaySound:   TRUE,
				PlayVibrate: TRUE,
				AfterOpen:   "go_app",
			},
		},
	}

	js, err := jsoniter.MarshalToString(request)
	if err == nil {
		return umengUrlPush(umengAndroidSecret, js)
	}
	return err
}

func umengIOSPush(token string, Title string, Subtitle string, Body string) error {
	request := &umengIOSPushRequest{
		AppKey:       umengIOSAppkey,
		Timestamp:    strconv.FormatInt(util.AppTimestamp(), 10),
		Type:         "unicast",
		DeviceTokens: token,
		Description:  "iosPush",
		Payload: umengIOSPushPayload{
			Aps: umengAPNs{
				Badge: 1,
				Sound: "default",
				Alert: umengAPNsAlert{
					Title: Title,
					//Subtitle: Subtitle,
					Body: Body,
				},
			},
		},
	}
	js, err := jsoniter.MarshalToString(request)
	if err == nil {
		return umengUrlPush(umengIOSSecret, js)
	}
	return err
}
