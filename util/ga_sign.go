package util

import (
	"fmt"
	"encoding/base64"
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"crypto/sha1"
	"crypto"
	"sort"
	"strconv"
	"strings"
	"encoding/json"
)

type GAPayment struct {
	Account       string  `json:"account,omitempty"`        // account	否	varchar(100)	渠道方平台账号名，部分渠道无法取得账号名则为空可能
	Amount        float32 `json:"amount,omitempty"`         // amount		是	decimal(15,2)	总金额(单位人民币), 浮点数两位 decimal(15,2)
	Channel       uint64  `json:"channel,omitempty"`        // channel	是	int	渠道ID, 详情查看 渠道信息列表
	Extra         string  `json:"extra,omitempty"`          // extra		否	varchar(255)	游戏扩展数据, 创建订单传入的值, 原数据返回
	GameId        uint64  `json:"game_id,omitempty"`        // game_id	是	int	游戏ID
	OrderId       uint64  `json:"order_id,omitempty"`       // order_id	是	bigint	巨人移动订单号
	ProductId     string  `json:"product_id,omitempty"`     // product_id	否	varchar(45)	苹果商品编号或安卓渠道或游戏自定义商品编号(请验证商品ID与金额后发货)
	Time          uint64  `json:"time,omitempty"`           // time		是	int	巨人移动充值服发起请求,秒为单位的时间戳
	TransactionId string  `json:"transaction_id,omitempty"` // transaction_id	是	varchar(100)	第三方交易单号
	OpenId        string  `json:"openid,omitempty"`         // openid		是	varchar(128)	账号ID
	ZoneId        uint64  `json:"zone_id,omitempty"`        // zone_id	是	int	游戏区ID
	Version       string  `json:"version,omitempty"`        // version	是	varchar(5)	回调接口版本4.0
	IsTest        uint64  `json:"is_test,omitempty"`        //is_test		是	int	是否沙箱订单，0否 1是 (支付回调4.0版本新增)
	Sign          string  `json:"sign,omitempty"`           // sign		是	text	Base64加密的签名
	//接收到的参数,除sign以外,按字母排序连接后,进行RSA校验
}

func _getMapStr(param map[string]string, key string) string {
	if p, ok := param[key]; ok {
		return p
	}
	return ""
}

func CheckGAPaymentSign(params map[string]string, pubPem string) bool {
	data := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s",
		_getMapStr(params, "account"),
		_getMapStr(params, "amount"),
		_getMapStr(params, "channel"),
		_getMapStr(params, "extra"),
		_getMapStr(params, "game_id"),
		_getMapStr(params, "openid"),
		_getMapStr(params, "order_id"),
		_getMapStr(params, "product_id"),
		_getMapStr(params, "time"),
		_getMapStr(params, "transaction_id"),
		_getMapStr(params, "version"),
		_getMapStr(params, "zone_id"))
	sign, _ := base64.StdEncoding.DecodeString(_getMapStr(params, "sign"))
	//验签
	block, _ := pem.Decode([]byte(pubPem))
	if block == nil {
		//fmt.Println("public key error")
		return false
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		//fmt.Println("public key error")
		return false
	}
	pub := pubInterface.(*rsa.PublicKey)
	h := sha1.New()
	h.Write([]byte(data))
	digest := h.Sum(nil)
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA1, digest, sign)
	if err == nil {
		//签名验证到成功
		return true
	}
	return false
}

func _httpBuildQuery(params map[string]interface{}) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var query []string
	for _, k := range keys {
		v := params[k]
		str := ""
		switch v.(type) {
		case int:
			str = strconv.Itoa(v.(int))
		case string:
			str = v.(string)
		case float32:
		case float64:
			str = strconv.Itoa(int(v.(float64)))
		}
		query = append(query, fmt.Sprintf("%s=%s", k, str))
	}
	return strings.Join(query, "&")

}

func CheckGALoginSign(params string, pubPem string) bool {
	var jsonParams interface{}
	err := json.Unmarshal([]byte(params), &jsonParams)
	if err != nil {
		//fmt.Println("json unmarshal error")
		return false
	}

	data, ok := jsonParams.(map[string]interface{})
	if !ok {
		//fmt.Println("json data error")
		return false
	}

	if _, ok := data["sign"]; ok == false {
		return false
	}
	if _, ok := data["entity"]; ok == false {
		return false
	}

	sign, _ := base64.StdEncoding.DecodeString(data["sign"].(string))

	entity := data["entity"].(map[string]interface{})
	query := _httpBuildQuery(entity)
	//fmt.Println(query, data["sign"].(string))

	//验签
	block, _ := pem.Decode([]byte(pubPem))
	if block == nil {
		//fmt.Println("public key error")
		return false
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		//fmt.Println("public key error")
		return false
	}

	pub := pubInterface.(*rsa.PublicKey)
	h := sha1.New()
	h.Write([]byte(query))
	digest := h.Sum(nil)
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA1, digest, sign)
	if err == nil {
		//fmt.Println("签名验证到成功")
		return true
	}
	return false
}
