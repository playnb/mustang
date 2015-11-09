package auth

import (
	//"TestGo/mustang/log"
	"encoding/json"
	"golang.org/x/oauth2"
)

var sina *AuthService

func SinaOAuth2(accid uint64, cb AuthCallBackFunc) (string, error) {
	if sina == nil {
		sina = initAuthService(new(sinaAuthSupply))
	}
	return doOAuth2(sina, accid, cb)
}

type sinaAuthSupply struct {
}

func (s *sinaAuthSupply) getClientID() string {
	return "3200218128"
}
func (s *sinaAuthSupply) getClientSecret() string {
	return "ccef2a382ad2a99878ee47aaefd2bec0"
}
func (s *sinaAuthSupply) getProfileInfoURL(token *oauth2.Token) string {
	return "https://api.weibo.com/2/users/show.json" + "?" + "access_token=" + token.AccessToken + "&" + "uid=" + token.Extra("uid").(string)
}
func (s *sinaAuthSupply) getEndPoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  "https://api.weibo.com/oauth2/authorize",
		TokenURL: "https://api.weibo.com/oauth2/access_token",
	}
}
func (s *sinaAuthSupply) getRedirectURL() string {
	return "/sina/authcallback"
}
func (s *sinaAuthSupply) getName() string {
	return "新浪OAuth2接口"
}
func (s *sinaAuthSupply) parseAuthUserData(data *AuthUserData, body []byte) error {
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
