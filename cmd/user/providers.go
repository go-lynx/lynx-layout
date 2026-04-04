package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-lynx/lynx"
	lynxgrpc "github.com/go-lynx/lynx-grpc"
	lynxhttp "github.com/go-lynx/lynx-http"
	"github.com/go-lynx/lynx-layout/internal/data"
	"github.com/go-lynx/lynx-layout/internal/server"
	"github.com/go-lynx/lynx-layout/internal/service"
	lynxmysql "github.com/go-lynx/lynx-mysql"
	lynxredis "github.com/go-lynx/lynx-redis"
	redislock "github.com/go-lynx/lynx-redis-lock"
	"github.com/go-lynx/lynx-sql-sdk/interfaces"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

var bootstrapProviderSet = wire.NewSet(
	provideLynxApp,
	provideRuntimeConfig,
	provideServiceRegistrar,
	provideDBProvider,
	provideEntClientProvider,
	provideRedisProvider,
	provideHTTPServerBase,
	provideGRPCServerBase,
	provideGRPCClientConnectionGetter,
	provideLoginLockRunner,
)

func provideLynxApp() (*lynx.LynxApp, error) {
	app := lynx.Lynx()
	if app == nil {
		return nil, fmt.Errorf("lynx app is nil")
	}
	return app, nil
}

func provideRuntimeConfig(app *lynx.LynxApp) (config.Config, error) {
	if app == nil {
		return nil, fmt.Errorf("lynx app is nil")
	}
	runtimeConfig := app.GetGlobalConfig()
	if runtimeConfig == nil {
		return nil, fmt.Errorf("runtime config is nil")
	}
	return runtimeConfig, nil
}

func provideServiceRegistrar(app *lynx.LynxApp) (registry.Registrar, error) {
	if app == nil {
		return nil, fmt.Errorf("lynx app is nil")
	}
	return app.GetServiceRegistry()
}

func provideDBProvider() (interfaces.DBProvider, error) {
	provider := lynxmysql.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("mysql provider is nil")
	}
	return provider, nil
}

func provideEntClientProvider(provider interfaces.DBProvider) data.EntClientProvider {
	return data.NewEntClientProviderFromDB(provider)
}

// provideRedisProvider keeps template-side wiring on the stable redis provider facade.
// Transport readiness/health aliases remain runtime/plugin concerns and are not consumed here.
func provideRedisProvider() (lynxredis.Provider, error) {
	provider := lynxredis.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("redis provider is nil")
	}
	if _, err := provider.UniversalClient(context.Background()); err != nil {
		return nil, fmt.Errorf("resolve redis client from provider: %w", err)
	}
	return provider, nil
}

func provideHTTPServerBase() (base server.HTTPServerBase, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			base = server.HTTPServerBase{}
			err = fmt.Errorf("http server lookup panicked: %v", recovered)
		}
	}()
	httpServer, err := lynxhttp.GetHttpServer()
	if err != nil {
		return server.HTTPServerBase{}, err
	}
	if httpServer == nil {
		return server.HTTPServerBase{}, fmt.Errorf("http server is nil")
	}
	return server.HTTPServerBase{Server: httpServer}, nil
}

func provideGRPCServerBase(app *lynx.LynxApp) (server.GRPCServerBase, error) {
	if app == nil {
		return server.GRPCServerBase{}, fmt.Errorf("lynx app is nil")
	}
	grpcServer, err := lynxgrpc.GetGrpcServer(app.GetPluginManager())
	if err != nil {
		return server.GRPCServerBase{}, err
	}
	if grpcServer == nil {
		return server.GRPCServerBase{}, fmt.Errorf("grpc server is nil")
	}
	return server.GRPCServerBase{Server: grpcServer}, nil
}

func provideGRPCClientConnectionGetter(app *lynx.LynxApp) data.GRPCClientConnectionGetter {
	return func(serviceName string) (*grpc.ClientConn, error) {
		if app == nil {
			return nil, fmt.Errorf("lynx app is nil")
		}
		return lynxgrpc.GetGrpcClientConnection(serviceName, app.GetPluginManager())
	}
}

func provideLoginLockRunner() service.LockRunner {
	return func(ctx context.Context, key string, expiration time.Duration, fn func() error) error {
		return redislock.Lock(ctx, key, expiration, fn)
	}
}
