package util

import (
	"crypto/aes"
	"crypto/cipher"
	"bytes"
	"encoding/base64"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const IdipSecruetKey = "c990fcb94e811bc3bdf1278e7d6217ea"

func IdipDecrypt(cryptText string) string {
	decodeBytes, err := base64.StdEncoding.DecodeString(cryptText);
	d, err := aesDecrypt([]byte(decodeBytes), IdipSecruetKey[:16], IdipSecruetKey[16:]);
	if err != nil {
		return ""
	}
	return string(d)
}

//AesEncrypt 加密
func aesEncrypt(src, key string, iv string) ([]byte, error) {
	bKey := []byte(key)
	bIv := []byte(iv)
	origData := []byte(src)
	block, err := aes.NewCipher(bKey)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, bIv[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//AesDecrypt 解密
func aesDecrypt(crypted []byte, key string, iv string) (string, error) {
	sKey := []byte(key)
	sIv := []byte(iv)
	block, err := aes.NewCipher(sKey)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, sIv[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs5UnPadding(origData)
	return string(origData), nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
