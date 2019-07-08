package util

import (
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"golang.org/x/net/context"
	"github.com/playnb/mustang/log"
	"io"
)

var _accessKey = "your access key"
var _secretKey = "your secret key"

func InitUpload(access, secret string) {
	_accessKey = access
	_secretKey = secret
}

func UploadData(data io.Reader, dataLen int64, saveName string, bucket string) bool {
	putPolicy := storage.PutPolicy{
		Scope: bucket + ":" + saveName,
		//ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)","name":"$(x:name)"}`,
	}
	mac := qbox.NewMac(_accessKey, _secretKey)
	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{}
	cfg.Zone = &storage.ZoneHuadong
	cfg.UseHTTPS = false
	cfg.UseCdnDomains = false

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}
	err := formUploader.Put(context.Background(), &ret, upToken, saveName, data, dataLen, &putExtra)
	if err != nil {
		log.Error("UploadData %s:%s", saveName, err.Error())
		return false
	}
	return true
}
