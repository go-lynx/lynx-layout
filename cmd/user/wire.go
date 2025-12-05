//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-lynx/lynx"
	"github.com/go-lynx/lynx-layout/internal/biz"
	"github.com/go-lynx/lynx-layout/internal/data"
	"github.com/go-lynx/lynx-layout/internal/server"
	"github.com/go-lynx/lynx-layout/internal/service"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(log.Logger) (*kratos.App, error) {
	panic(
		wire.Build(
			server.ProviderSet,
			data.ProviderSet,
			biz.ProviderSet,
			service.ProviderSet,
			provideKratosApp,
		),
	)
}

// provideKratosApp creates a kratos app from servers and registry
func provideKratosApp(
	grpcServer *grpc.Server,
	httpServer *http.Server,
	registrar registry.Registrar,
) (*kratos.App, error) {
	var serverList []transport.Server
	if grpcServer != nil {
		serverList = append(serverList, grpcServer)
	}
	if httpServer != nil {
		serverList = append(serverList, httpServer)
	}

	opts := []kratos.Option{
		kratos.ID(lynx.GetHost()),
		kratos.Name(lynx.GetName()),
		kratos.Version(lynx.GetVersion()),
		kratos.Metadata(map[string]string{
			"host":    lynx.GetHost(),
			"version": lynx.GetVersion(),
		}),
		kratos.Logger(log.DefaultLogger),
	}

	if registrar != nil {
		opts = append(opts, kratos.Registrar(registrar))
	}

	if len(serverList) > 0 {
		opts = append(opts, kratos.Server(serverList...))
	}

	return kratos.New(opts...), nil
}
