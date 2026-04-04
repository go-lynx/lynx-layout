package server

import (
	"fmt"

	transportgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	loginV1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/go-lynx/lynx-layout/internal/service"
	grpc "google.golang.org/grpc"
)

type GRPCServerBase struct {
	Server *transportgrpc.Server
}

var (
	registerLoginGRPCServer = func(registrar grpc.ServiceRegistrar, srv loginV1.LoginServer) {
		loginV1.RegisterLoginServer(registrar, srv)
	}
)

func NewGRPCServer(
	base GRPCServerBase,
	login *service.LoginService) (g *transportgrpc.Server, err error) {
	if login == nil {
		return nil, fmt.Errorf("login gRPC 服务不能为空")
	}
	if base.Server == nil {
		return nil, fmt.Errorf("gRPC 服务实例为空")
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			g = nil
			err = fmt.Errorf("初始化 gRPC 服务失败: %v", recovered)
		}
	}()

	g = base.Server
	registerLoginGRPCServer(g, login)
	return g, nil
}
