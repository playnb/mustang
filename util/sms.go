package util

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"cell/common/mustang/log"
)

func _getHttpClient() *http.Client {
	if log.LtpLog == true {
		proxyUrl := "http://192.168.177.140:808"

		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxyUrl)
		}
		transport := &http.Transport{Proxy: proxy}
		client := &http.Client{Transport: transport}

		return client
	} else {
		return &http.Client{}
	}
}

// SendSmsReply 发送短信返回
type SendSmsReply struct {
	Code    string `json:"Code,omitempty"`
	Message string `json:"Message,omitempty"`
}

func replace(in string) string {
	rep := strings.NewReplacer("+", "%20", "*", "%2A", "%7E", "~")
	return rep.Replace(url.QueryEscape(in))
}

var _SmsAccessKeyID, _SmsAccessSecret string

func InitSms(accessKeyID, accessSecret string) {
	_SmsAccessKeyID = accessKeyID
	_SmsAccessSecret = accessSecret
}

// SendSms 发送短信
func SendSms(phoneNumbers, signName, templateParam, templateCode string) error {
	paras := map[string]string{
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   fmt.Sprintf("%d", rand.Int63()),
		"AccessKeyId":      _SmsAccessKeyID,
		"SignatureVersion": "1.0",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Format":           "JSON",
		"Action":           "SendSms",
		"Version":          "2017-05-25",
		"RegionId":         "cn-hangzhou",
		"PhoneNumbers":     phoneNumbers,
		"SignName":         signName,
		"TemplateParam":    templateParam,
		"TemplateCode":     templateCode,
	}

	var keys []string

	for k := range paras {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var sortQueryString string

	for _, v := range keys {
		sortQueryString = fmt.Sprintf("%s&%s=%s", sortQueryString, replace(v), replace(paras[v]))
	}

	stringToSign := fmt.Sprintf("GET&%s&%s", replace("/"), replace(sortQueryString[1:]))

	mac := hmac.New(sha1.New, []byte(fmt.Sprintf("%s&", _SmsAccessSecret)))
	mac.Write([]byte(stringToSign))
	sign := replace(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	str := fmt.Sprintf("http://dysmsapi.aliyuncs.com/?Signature=%s%s", sign, sortQueryString)

	resp, err := _getHttpClient().Get(str)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	ssr := &SendSmsReply{}

	if err := json.Unmarshal(body, ssr); err != nil {
		return err
	}

	if ssr.Code == "SignatureNonceUsed" {
		return SendSms(phoneNumbers, signName, templateParam, templateCode)
	} else if ssr.Code != "OK" {
		return errors.New(ssr.Code)
	}

	return nil
}

type SmsServerStatus struct {
	Status string `json:"status"`
}

func SendServerStatusSms(phoneNumber string, status *SmsServerStatus) error {
	str, _ := json.Marshal(status)
	return SendSms(phoneNumber, "刘天鹏通知你服务器状态", string(str), "SMS_129747897")
}
