package server

import (
	"fmt"

	transporthttp "github.com/go-kratos/kratos/v2/transport/http"
	lynx "github.com/go-lynx/lynx-http"
	loginV1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/go-lynx/lynx-layout/internal/service"
)

var (
	httpServerGetter        = lynx.GetHttpServer
	registerLoginHTTPServer = loginV1.RegisterLoginHTTPServer
)

func NewHTTPServer(
	login *service.LoginService) (h *transporthttp.Server, err error) {
	if login == nil {
		return nil, fmt.Errorf("login HTTP 服务不能为空")
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			h = nil
			err = fmt.Errorf("初始化 HTTP 服务失败: %v", recovered)
		}
	}()

	h, err = httpServerGetter()
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, fmt.Errorf("HTTP 服务实例为空")
	}

	registerLoginHTTPServer(h, login)
	return h, nil
}
