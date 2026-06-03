package server

import (
	"fmt"

	transporthttp "github.com/go-kratos/kratos/v2/transport/http"
	loginV1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/go-lynx/lynx-layout/internal/service"
)

type HTTPServerBase struct {
	Server *transporthttp.Server
}

var (
	registerLoginHTTPServer = loginV1.RegisterLoginHTTPServer
)

func NewHTTPServer(
	base HTTPServerBase,
	login *service.LoginService) (h *transporthttp.Server, err error) {
	if login == nil {
		return nil, fmt.Errorf("login HTTP service must not be nil")
	}
	if base.Server == nil {
		return nil, fmt.Errorf("HTTP server instance is nil")
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			h = nil
			err = fmt.Errorf("failed to initialize HTTP service: %v", recovered)
		}
	}()

	h = base.Server
	registerLoginHTTPServer(h, login)
	return h, nil
}
