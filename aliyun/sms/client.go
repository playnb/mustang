package sms

import "cell/common/mustang/aliyun/common"

const (
	SmsEndPoint   = "https://dysmsapi.aliyuncs.com"
	SingleSendSms = "SendSms"
	SmsAPIVersion = "2017-05-25"
)

type Client struct {
	common.Client
}

func NewClient(accessKeyId, accessKeySecret string) *Client {
	client := new(Client)
	client.Init(SmsEndPoint, SmsAPIVersion, accessKeyId, accessKeySecret)
	return client
}
