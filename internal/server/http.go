package server

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	lynx "github.com/go-lynx/lynx-http"
	loginV1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/go-lynx/lynx-layout/internal/service"
)

func NewHTTPServer(
	login *service.LoginService) *http.Server {
	h, err := lynx.GetHttpServer()
	if err != nil {
		panic(err)
	}
	loginV1.RegisterLoginHTTPServer(h, login)
	return h
}
