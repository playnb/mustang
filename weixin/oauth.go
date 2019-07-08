package weixin

import (
	"encoding/json"
	//"github.com/playnb/mustang/auth"
	"github.com/playnb/mustang/log"
	//"golang.org/x/oauth2"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	//"sort"
	"crypto/sha1"
	"io"
)

type OAuthCallback func()

//构造微信验证码
func MakeSignature_js(timestamp, nonce, token, url string) string {
	sl := []string{"jsapi_ticket=" + token, "&noncestr=" + nonce, "&timestamp=" + timestamp, "&url=" + url}
	//sort.Strings(sl)
	s := sha1.New()
	//log.Debug(strings.Join(sl, ""))
	io.WriteString(s, strings.Join(sl, ""))
	return fmt.Sprintf("%x", s.Sum(nil))
}

func getOAuthUrl(state string, redirectUrl string) string {
	url := "https://open.weixin.qq.com/connect/oauth2/authorize?appid={APPID}&redirect_uri={REDIRECT_URI}&response_type=code&scope={SCOPE}&state={STATE}#wechat_redirect"
	url = strings.Replace(url, "{APPID}", wxProfile.AppID, -1)
	url = strings.Replace(url, "{REDIRECT_URI}", UrlEncode(redirectUrl), -1)
	url = strings.Replace(url, "{SCOPE}", "snsapi_userinfo", -1)
	url = strings.Replace(url, "{STATE}", state, -1)
	return url
}

var oauth_init bool = false

func InitWiXinOAuth() {
	if oauth_init {
		return
	}
	oauth_init = true
	wxProfile.mux.HandleFunc("/wx/authorize", handleWxAuthorize)
	wxProfile.mux.HandleFunc("/wx/authorize_redirect", handleWxAuthorizeRedirect)
}

func genWeiXinOAuthURL(state string) string {
	InitWiXinOAuth()
	url := getOAuthUrl(state, wxProfile.AuthRedirectURL+"/wx/authorize_redirect")
	return url
}

func handleWxAuthorize(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	state := strings.Join(r.Form["state"], "")
	log.Debug("Visit: " + r.RequestURI)
	url := genWeiXinOAuthURL(state)
	log.Debug("URL: %v\n", url)
	//redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)
}

//可以给网页的Accesstoken
type WebAccessToken struct {
	WeiXinError
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
}

func handleWxAuthorizeRedirect(w http.ResponseWriter, r *http.Request) {
	log.Debug("Visit: " + r.RequestURI)

	r.ParseForm()
	code := strings.Join(r.Form["code"], "")
	state := strings.Join(r.Form["state"], "")
	appName := state
	param := ""
	strs := strings.Split(state, "$")
	if len(strs) > 0 {
		appName = strs[0]
		if len(strs) > 1 {
			for i := 1; i < len(strs); i++ {
				param = param + "&" + strs[i]
			}
			param = "&" + param
		}
	}
	log.Trace("code:%s state:%s Request:%v", code, state, r)

	if len(code) == 0 {
		fmt.Fprintf(w, "亲! 不要拒绝授权我...")
		return
	}

	url := "https://api.weixin.qq.com/sns/oauth2/access_token?appid={APPID}&secret={SECRET}&code={CODE}&grant_type=authorization_code"
	url = strings.Replace(url, "{APPID}", wxProfile.AppID, -1)
	url = strings.Replace(url, "{SECRET}", wxProfile.AppSecret, -1)
	url = strings.Replace(url, "{CODE}", code, -1)
	res, err := http.Get(url)
	if err != nil {
		log.Error("微信OAuth Get state: %s  error: %s", state, err.Error())
		return
	}
	result, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	token := new(WebAccessToken)
	err = json.Unmarshal(result, token)
	if err != nil {
		log.Error("微信OAuth json state: %s  error: %s", state, err.Error())
		log.Error("%s", string(result))
		return
	}

	if !checkWeiXinReturn(&(token.WeiXinError)) {
		log.Error("微信OAuth err: %d  msg: %s", token.ErrCode, token.ErrMsg)
		return
	}

	///测试代码
	{
		log.Trace("-------------------SNS access_token %v", token)
		userInfo := GetUserInfo(token.OpenID, true)
		if userInfo != nil {
			log.Trace("-------------------userInfo:%v", userInfo)
		}
		/*
			url = "https://api.weixin.qq.com/sns/userinfo?access_token={{.AccessToken}}&openid={{.OpenID}}&lang=zh_CN"
			url = strings.Replace(url, "{{.AccessToken}}", token.AccessToken, -1)
			url = strings.Replace(url, "{{.OpenID}}", token.OpenID, -1)
			res, err := http.Get(url)
			if err == nil {
				result, _ := ioutil.ReadAll(res.Body)
				res.Body.CloseSession()
				log.Trace("-------------------SNS userinfo  %s", string(result))
			}
		*/
	}

	if wxProfile.UserAuthorized != nil {
		finish := wxProfile.UserAuthorized(state, token, w, r)
		if finish == true {
			return
		}
	}

	ts := strconv.Itoa(int(time.Now().UnixNano() / 1000 / 1000 / 1000))
	nonce := "wahaha"

	//fmt.Fprintf(w, token.AccessToken)
	rUrl, ok := wxProfile.authRedirectAppUrl[appName]
	if ok {
		if len(rUrl) > 1 {
			url = rUrl + "?access_token={AccessToken}&openid={OpenID}&timestamp={TimeStamp}&nonce={Nonce}" + param
			url = strings.Replace(url, "{AccessToken}", token.AccessToken, -1)
			url = strings.Replace(url, "{OpenID}", token.OpenID, -1)
			url = strings.Replace(url, "{TimeStamp}", ts, -1)
			url = strings.Replace(url, "{Nonce}", nonce, -1)
			//url = strings.Replace(url, "Signature", MakeSignature_js(ts, nonce, wxProfile.JsApiTicket, "liutp.vicp.net/app/demo"), -1)
			log.Debug(url)
			http.Redirect(w, r, url, http.StatusFound)
		}
	} else {
		log.Error("未知的App: %s", state)
	}
}
