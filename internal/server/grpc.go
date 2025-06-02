package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	loginV1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/go-lynx/lynx-layout/internal/service"
	lynx "github.com/go-lynx/plugins/service/grpc/v2"
)

func NewGRPCServer(
	login *service.LoginService) *grpc.Server {
	g := lynx.GetGrpcServer()
	loginV1.RegisterLoginServer(g, login)
	return g
}
