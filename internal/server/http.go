package server

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	loginV1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/go-lynx/lynx-layout/internal/service"
	lhttp "github.com/go-lynx/lynx/plugin/http"
)

func NewHTTPServer(
	login *service.LoginService) *http.Server {
	h := lhttp.GetHTTP()
	loginV1.RegisterLoginHTTPServer(h, login)
	return h
}
