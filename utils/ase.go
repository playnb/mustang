package utils

/*
AES加密解密字符串
原始文件    https://github.com/polaris1119/myblog_article_code/blob/master/aes/aes.go
*/
import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	//"encoding/base64"
	"encoding/hex"
	"errors"
)

var aesKey = []byte("sfe023f_9fd&fwfl")

//设置秘钥
func SetAesKey(key string) {
	aesKey = []byte(key)
	aesKey = PKCS5Padding(aesKey, 16)
	aesKey = aesKey[:16]
}

func MakeAesKey(key string) []byte {
	b := []byte(key)
	b = PKCS5Padding(b, 16)
	b = aesKey[:16]
	return b
}

//加密字符串
func AesEncryptString(orgStr string, key string) (string, error) {
	result, err := AesEncrypt([]byte(orgStr), MakeAesKey(key))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(result), nil
}

//解密字符串
func AesDecryptString(cryptedStr string, key string) (string, error) {
	b, err := hex.DecodeString(cryptedStr)
	if err != nil {
		return "", err
	}
	origData, err := AesDecrypt(b, MakeAesKey(key))
	if err != nil {
		return "", err
	}
	return string(origData), nil
}

func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	slot := len(crypted) % blockSize;
	if (slot!=0){
		return nil, errors.New("crypto/cipher: input not full blocks")
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	if length > 0 {
		// 去掉最后一个字节 unpadding 次
		unpadding := int(origData[length-1])
		return origData[:(length - unpadding)]
	} else {
		return origData
	}
}
