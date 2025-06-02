//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-lynx/lynx-layout/internal/biz"
	"github.com/go-lynx/lynx-layout/internal/data"
	"github.com/go-lynx/lynx-layout/internal/server"
	"github.com/go-lynx/lynx-layout/internal/service"
	kratos "github.com/go-lynx/lynx/app/kratos"
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
			kratos.NewKratos,
			kratos.ProvideKratosOptions,
		),
	)
}
