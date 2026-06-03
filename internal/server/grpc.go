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
		return nil, fmt.Errorf("login gRPC service must not be nil")
	}
	if base.Server == nil {
		return nil, fmt.Errorf("gRPC server instance is nil")
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			g = nil
			err = fmt.Errorf("failed to initialize gRPC service: %v", recovered)
		}
	}()

	g = base.Server
	registerLoginGRPCServer(g, login)
	return g, nil
}
