package util

import (
	"strings"
	"fmt"
	"crypto/md5"
	"encoding/base64"
	"github.com/json-iterator/go"
	"sort"
	"math/rand"
	"strconv"
)

type gakf_SignData struct {
	Account  string `json:"account"`
	Channel  string `json:"channel"`
	Game     string `json:"room"`
	Openid   string `json:"openid"`
	Rolename string `json:"rolename"`
	Time     string `json:"time"`
	Zoneid   string `json:"zoneid"`
}

type gakf_KeyData struct {
	Key  string `json:"key"`
	Sign string `json:"sign"`
}

func gakf_EncryptStr(data string, key string) string {
	str := string(gakf_Encrypt([]byte(data), []byte(key)))
	str = strings.Replace(str, "Z", "Z0", -1)
	str = strings.Replace(str, "+", "Z1", -1)
	str = strings.Replace(str, "/", "Z2", -1)
	str = strings.Replace(str, "=", "Z3", -1)
	return str
}

func gakf_Encrypt(data []byte, k []byte) []byte {
	key := []byte(fmt.Sprintf("%x", md5.Sum(k)))
	dataLen := len(data)
	keyLen := len(key)

	x := 0
	char := make([]byte, dataLen)
	for i := 0; i < dataLen; i++ {
		if x == keyLen {
			x = 0
		}
		char[i] = key[x]
		x++
	}

	str := make([]byte, dataLen)
	for i := 0; i < dataLen; i++ {
		str[i] = byte(int(data[i]) + int(char[i])%256)
	}

	return []byte(base64.StdEncoding.EncodeToString(str))
}

func gakf_SignArgs(arg map[string]string, privateKey string, randKey string) map[string]string {
	//privateKey := "faae44f61777953207850d60c84be8b3"
	//randKey = "e3b1333ed9c54d3dce3e25c8db6c4a36"
	back := make(map[string]string)
	for k, v := range arg {
		str := base64.StdEncoding.EncodeToString([]byte(v))
		back[k] = gakf_EncryptStr(str, randKey)
	}

	jMd5, _ := jsoniter.MarshalToString(back)
	{
		sd := &gakf_SignData{}
		jsoniter.UnmarshalFromString(jMd5, sd)
		jMd5, _ = jsoniter.MarshalToString(sd)
		//fmt.Println(jMd5)
	}
	sign := fmt.Sprintf("%x", md5.Sum([]byte(jMd5))) + privateKey
	sign = fmt.Sprintf("%x", md5.Sum([]byte(sign)))

	b2 := make(map[string]string)
	b2["key"] = randKey
	b2["sign"] = sign
	jMd5, _ = jsoniter.MarshalToString(b2)
	{
		kd := &gakf_KeyData{}
		jsoniter.UnmarshalFromString(jMd5, kd)
		jMd5, _ = jsoniter.MarshalToString(kd)
		//fmt.Println(jMd5)
	}
	//fmt.Println(jMd5)
	back["key"] = gakf_EncryptStr(jMd5, privateKey)

	return back
}

func GAKFBuildSignUrl(arg map[string]string, privateKey string) string {
	randKey := fmt.Sprintf("%x", md5.Sum([]byte(strconv.FormatUint(rand.Uint64(), 10))))
	//randKey = "e3657677165600c17bba5bf6079d7c70"
	url := "https://kfim.ztgame.com/m_web_order/games/login"
	back := gakf_SignArgs(arg, privateKey, randKey)

	i := 0
	var keys []string
	for k := range back {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if i == 0 {
			url += "?"
		} else {
			url += "&"
		}
		i++
		url += k + "=" + back[k]
	}
	//fmt.Println(arg)
	//fmt.Println("randKey=" + randKey)
	//fmt.Println("===============")
	//fmt.Println(back)
	//fmt.Println(url)
	return url
}
