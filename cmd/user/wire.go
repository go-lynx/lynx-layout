//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	lynxkratos "github.com/go-lynx/lynx/kratos"

	"github.com/go-lynx/lynx-layout/internal/biz"
	"github.com/go-lynx/lynx-layout/internal/data"
	"github.com/go-lynx/lynx-layout/internal/server"
	"github.com/go-lynx/lynx-layout/internal/service"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp() (*kratos.App, error) {
	panic(
		wire.Build(
			server.ProviderSet,
			data.ProviderSet,
			biz.ProviderSet,
			service.ProviderSet,
			lynxkratos.ProvideKratosOptions,
			lynxkratos.NewKratos,
		),
	)
}
