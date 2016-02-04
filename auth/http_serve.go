package auth

import (
	"github.com/playnb/mustang/log"
	"net/http"
)

const (
	AuthServicePort = "6654"
	AuthServiceUrl  = "http://liutp.vicp.net"
)

var default_http_service *AuthHttpService

func InitAuthHttpService() {
	if default_http_service == nil {
		default_http_service = new(AuthHttpService)
		default_http_service.mux = http.NewServeMux()
		default_http_service.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Mustang authservice is running..."))
		})
		log.Trace("AuthService Listen On: " + AuthServicePort)
		go http.ListenAndServe(":"+AuthServicePort, default_http_service.mux)
	}
}

type AuthHttpService struct {
	mux *http.ServeMux
}

func (a *AuthHttpService) SetMux(mux *http.ServeMux) {
	a.mux = mux
}
func (a *AuthHttpService) Mux() *http.ServeMux {
	return a.mux
}
