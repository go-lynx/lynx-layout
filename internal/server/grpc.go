package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	loginV1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/go-lynx/lynx-layout/internal/service"
	lgrpc "github.com/go-lynx/lynx/plugin/grpc"
)

func NewGRPCServer(
	login *service.LoginService) *grpc.Server {
	g := lgrpc.GetGRPC()
	loginV1.RegisterLoginServer(g, login)
	return g
}
