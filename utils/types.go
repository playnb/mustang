package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
)

//微信的登陆数据
type WinXinHttpParams struct {
	UserName    string `json:"Name"`
	ShareName   string `json:"SName"`
	Pic         string `json:"Pic"`
	OpenID      string `json:"OpenID"`
	Token       string `json:"Token"`
	ServerToken string `json:"SToken"`
}

func (t *WinXinHttpParams) Encode(key string, compress bool) (string, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	result, err := AesEncrypt(b, MakeAesKey(key))
	if err != nil {
		return "", err
	}
	var buf []byte
	if compress {
		buf = DoZlibCompress(result)
	} else {
		buf = result
	}
	str := base64.StdEncoding.EncodeToString(buf)
	str = base64.StdEncoding.EncodeToString([]byte(str))
	if compress {
		str = "Z" + str
	} else {
		str = "U" + str
	}
	return str, nil
}

func (t *WinXinHttpParams) Decode(data string, key string) error {
	compress := data[:1] == "Z"

	data = data[1:]
	data, _ = url.QueryUnescape(data)

	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	b, err = base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return err
	}
	var buf []byte
	if compress {
		buf, err = DoZlibUnCompress(b)
		if err != nil {
			return err
		}
	} else {
		buf = b
	}
	fmt.Println(buf)
	result, err := AesDecrypt(buf, MakeAesKey(key))
	fmt.Println(string(result))
	if err != nil {
		return err
	} else {
		err = json.Unmarshal(result, t)
		if err != nil {
			return err
		}
	}
	return nil
}
