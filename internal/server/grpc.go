package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	lynx "github.com/go-lynx/lynx-grpc"
	loginV1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/go-lynx/lynx-layout/internal/service"
)

func NewGRPCServer(
	login *service.LoginService) *grpc.Server {
	g, err := lynx.GetGrpcServer(nil)
	if err != nil {
		panic(err)
	}
	loginV1.RegisterLoginServer(g, login)
	return g
}
