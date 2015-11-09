package auth

import (
	"TestGo/mustang/log"
	"net/http"
)

const (
	AuthServicePort = "6654"
	AuthServiceUrl  = "http://liutp.vicp.net"
)

var http_service *AuthHttpService

func InitAuthHttpService() {
	if http_service == nil {
		http_service = new(AuthHttpService)
		http_service.mux = http.NewServeMux()
		http_service.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Mustang authservice is running..."))
		})
		log.Trace("AuthService Listen On: " + AuthServicePort)
		go http.ListenAndServe(":"+AuthServicePort, http_service.mux)
	}
}

type AuthHttpService struct {
	mux *http.ServeMux
}
