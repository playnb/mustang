package sms

import (
	"cell/common/mustang/aliyun/common"
	"net/http"
)

type SendSmsArgs struct {
	SignName     string
	TemplateCode string
	PhoneNumbers string
	ParamString  string
}

//please set the signature and template in the console of Aliyun before you call this API
func (this *Client) SendSms(args *SendSmsArgs, response *common.Response) error {
	return this.InvokeByAnyMethod(http.MethodPost, SingleSendSms, "", args, response)
}

/*

	{
		response := &common.Response{}
		err := sms.NewClient("GRDPqG8j5kx2XNR1", "tEidhIEtnK3uNjYWOPhRZjVMHo98a9").SendSms(&sms.SendSmsArgs{
			SignName:     "刘天鹏通知你服务器状态",
			TemplateCode: "SMS_85805025",
			PhoneNumbers: "18621784806",
			ParamString:  `{"server": "123","data":"111111"}`,
		},
			response)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(response)
	}
	return

*/
