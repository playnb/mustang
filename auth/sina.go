package auth

import (
	//"github.com/playnb/mustang/log"
	"encoding/json"
	"golang.org/x/oauth2"
)

var sina *AuthService

func SinaOAuth2(accid uint64, cb AuthCallBackFunc) (string, error) {
	if sina == nil {
		sina = InitAuthService(new(sinaAuthSupply), default_http_service)
	}
	return DoOAuth2(sina, accid, cb)
}

type sinaAuthSupply struct {
}

func (s *sinaAuthSupply) GetClientID() string {
	return "3200218128"
}
func (s *sinaAuthSupply) GetClientSecret() string {
	return "ccef2a382ad2a99878ee47aaefd2bec0"
}
func (s *sinaAuthSupply) GetProfileInfoURL(token *oauth2.Token) string {
	return "https://api.weibo.com/2/users/show.json" + "?" + "access_token=" + token.AccessToken + "&" + "uid=" + token.Extra("uid").(string)
}
func (s *sinaAuthSupply) GetEndPoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  "https://api.weibo.com/oauth2/authorize",
		TokenURL: "https://api.weibo.com/oauth2/access_token",
	}
}
func (s *sinaAuthSupply) GetRedirectURL() string {
	return "/sina/authcallback"
}
func (s *sinaAuthSupply) GetName() string {
	return "新浪OAuth2接口"
}
func (s *sinaAuthSupply) ParseAuthUserData(data *AuthUserData, body []byte) error {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		//参考: http://open.weibo.com/wiki/2/users/show
		data.AuthAvatar, _ = jsonData["avatar_large"].(string)
		data.AuthName, _ = jsonData["name"].(string)
		data.AuthUID, _ = jsonData["idstr"].(string)
		return nil
	} else {
		return err
	}
}
