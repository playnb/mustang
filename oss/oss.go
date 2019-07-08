package oss

import "github.com/spf13/viper"

type OSS interface {
	UploadString(path string, content string) (string, error)
	DownloadString(path string) string
	init()
}

var Aliyun OSS

var _cfg *viper.Viper

func Init(cfg *viper.Viper) {
	_cfg = cfg
	Aliyun = &aliyunOss{}
	Aliyun.init()
}
