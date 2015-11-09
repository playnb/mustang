package auth

import (
	"TestGo/mustang/log"
	//"encoding/json"
	//"errors"
	"errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

type AuthUserData struct {
	Accid       uint64 //验证用户的唯一ID
	AcessToken  string
	AuthUID     string
	AuthName    string
	AuthAvatar  string
	ServiceName string
	Raw         string
	Response    http.ResponseWriter
	Request     *http.Request
}

type AuthCallBackFunc func(*AuthUserData, uint64, error) // 声明了一个函数类型

type AuthService struct {
	oauthCfg *oauth2.Config
	callback map[uint64]AuthCallBackFunc
}

type IAuthSupply interface {
	getRedirectURL() string
	getName() string
	getClientID() string
	getClientSecret() string
	getEndPoint() oauth2.Endpoint
	getProfileInfoURL(token *oauth2.Token) string
	parseAuthUserData(data *AuthUserData, body []byte) error
}

func doOAuth2(service *AuthService, accid uint64, cb AuthCallBackFunc) (string, error) {
	if service == nil {
		err := errors.New("SinaOAuth2没有初始化")
		cb(nil, accid, err)
		return "", err
	}
	if _, ok := service.callback[accid]; ok {
		err := errors.New("accid冲突")
		cb(nil, accid, err)
		return "", err
	}
	service.callback[accid] = cb
	log.Debug("授权回调地址" + service.oauthCfg.RedirectURL)
	return service.oauthCfg.AuthCodeURL(strconv.FormatUint(accid, 10)), nil
}

func initAuthService(supply IAuthSupply) *AuthService {
	if (http_service==nil){
		log.Error("AuthHttpService没有初始化, 调用InitAuthHttpService函数初始化")
		return nil
	}else{
		service := new(AuthService)
		service = new(AuthService)
		service.callback = make(map[uint64]AuthCallBackFunc)
		service.oauthCfg = &oauth2.Config{
			ClientID:     supply.getClientID(),
			ClientSecret: supply.getClientSecret(),
			Endpoint:     supply.getEndPoint(),
			RedirectURL:  AuthServiceUrl + supply.getRedirectURL(),
			Scopes:       nil,
		}

		http_service.mux.HandleFunc(supply.getRedirectURL(), func(w http.ResponseWriter, r *http.Request) {
			data := new(AuthUserData)
			data.Response = w
			data.Request = r

			log.Trace("OAuth2接口授权返回: "+ r.RequestURI+ "来自" +r.RemoteAddr)
			code := r.FormValue("code")
			state, err := strconv.ParseInt(r.FormValue("state"), 10, 64)
			if err != nil {
				log.Error("OAuth2接口授权返回accid错误")
				return
			}
			accid := uint64(state)
			defer delete(service.callback, accid)

			cb, ok := service.callback[uint64(accid)]
			if !ok {
				log.Error("没有找到授权回调函数: " + strconv.FormatUint(accid, 10))
				return
			}
			var (
				ctx    context.Context
				cancel context.CancelFunc
			)
			ctx, cancel = context.WithCancel(context.Background())
			defer cancel()
			token, err := service.oauthCfg.Exchange(ctx, code)

			if err != nil {
				log.Error(err.Error())
				cb(nil, accid, err)
			} else {
				req, err := service.oauthCfg.Client(ctx, token).Get(supply.getProfileInfoURL(token))
				if err != nil {
					log.Error("OAuth2接口获取用户信息错误:%s accid:%d", err.Error(), accid)
					cb(nil, accid, err)
				} else {
					body, _ := ioutil.ReadAll(io.LimitReader(req.Body, 1<<20))
					data.Accid = accid
					data.AcessToken = token.AccessToken
					data.Raw = string(body)

					if err := supply.parseAuthUserData(data, body); err == nil {
						cb(data, accid, nil)
					} else {
						cb(nil, accid, err)
					}
				}
			}
		})
		return service
	}
}
