package oss

import (
	"cell/common/mustang/log"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"strings"
)

type aliyunOss struct {
	client *oss.Client
	bucket *oss.Bucket

	bucketName      string
	accessKeyId     string
	accessKeySecret string
	endPoint        string
	cdnUrl          string
}

func (ao *aliyunOss) init() {
	ao.endPoint = _cfg.GetString("oss.aliyun.EndPoint")
	ao.accessKeyId = _cfg.GetString("oss.aliyun.AccessKeyId")
	ao.accessKeySecret = _cfg.GetString("oss.aliyun.AccessKeySecret")
	ao.bucketName = _cfg.GetString("oss.aliyun.BucketName")
	ao.cdnUrl = _cfg.GetString("oss.aliyun.CDN")

	ao.connect()
}

func (ao *aliyunOss) String() string {
	return ao.bucketName + "." + ao.endPoint
}

func (ao *aliyunOss) connect() error {
	var err error
	ao.client, err = oss.New("https://"+ao.endPoint, ao.accessKeyId, ao.accessKeySecret)
	if err != nil {
		log.Error("Error: %s | %v", ao, err)
		return err
	}

	ao.bucket, err = ao.client.Bucket(ao.bucketName)
	if err != nil {
		log.Error("Error: %s | %v", ao, err)
		return err
	}

	return nil
}

func (ao *aliyunOss) uploadString(path string, content string) (string, error) {
	var err error
	err = ao.bucket.PutObject(path, strings.NewReader(content))
	if err != nil {
		log.Error("Error: %s | %v", ao, err)
		return "", err
	}
	if len(ao.cdnUrl) > 0 {
		return ao.cdnUrl + "/" + path, nil

	}
	return "https://" + ao.bucketName + "." + ao.endPoint + "/" + path, nil
}

func (ao *aliyunOss) UploadString(path string, content string) (string, error) {
	path = strings.Trim(path, "/")
	path = strings.Trim(path, "\\")

	var url string
	var err error
	url, err = ao.uploadString(path, content)
	if err != nil {
		ao.connect()
		return ao.uploadString(path, content)
	}
	return url, err
}

func (ao *aliyunOss) DownloadString(path string) string {
	path = strings.Trim(path, "/")
	path = strings.Trim(path, "\\")

	body, err := ao.bucket.GetObject(path)
	if err != nil {
		log.Error("Error: %s | %v", ao, err)
		return ""
	}
	defer body.Close()
	data, err := ioutil.ReadAll(body)
	if err != nil {
		log.Error("Error: %s | %v", ao, err)
		return ""
	}
	return string(data)
}
