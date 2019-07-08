package util

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"sort"
)

//TODO 这里每次都要New一把，姿势对不对？
func Md5String(src string) string {
	h := md5.New()
	h.Write([]byte(src))
	return hex.EncodeToString(h.Sum(nil))
}

type Charset string

const (
	UTF8 = Charset("UTF-8")
	GBK  = Charset("GB18030")
)

//GBK的byte转换成UTF-8的string
func ConvertGBKString(byte []byte) string {
	//转换编码的时候处理一下  遇到\0就截断
	//征途把聊天结尾部分有特殊处理
	i := 0
	for i = 0; i < len(byte); i++ {
		if byte[i] == 0 {
			break
		}
	}
	if i == 0 {
		return ""
	}
	var decodeBytes, _ = simplifiedchinese.GBK.NewDecoder().Bytes(byte[:i])
	return string(decodeBytes)
}

//UTF-8的string转换成GBK的byte
func ConvertUTF8Bytes(str string) []byte {
	reader := transform.NewReader(bytes.NewReader([]byte(str)), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil
	}
	return d
}

func UTF82GB2312(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

///////////////////////////////////////////
type byPinyin []string

func (s byPinyin) Len() int      { return len(s) }
func (s byPinyin) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byPinyin) Less(i, j int) bool {
	a := ConvertUTF8Bytes(s[i])
	b := ConvertUTF8Bytes(s[j])
	bLen := len(b)
	for idx, chr := range a {
		if idx > bLen-1 {
			return false
		}
		if chr != b[idx] {
			return chr < b[idx]
		}
	}
	return true
}

//按拼音顺序排序字符串
func StringSortByPinyin(ss []string) {
	sort.Sort(byPinyin(ss))
}

func LessStringByPinyin(sa, sb string) bool {
	a := ConvertUTF8Bytes(sa)
	b := ConvertUTF8Bytes(sb)
	bLen := len(b)
	for idx, chr := range a {
		if idx > bLen-1 {
			return false
		}
		if chr != b[idx] {
			return chr < b[idx]
		}
	}
	return true
}
