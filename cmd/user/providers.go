package main

import (
	"context"
	"fmt"
	"sync"
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
	"github.com/go-lynx/lynx/log"
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

// Plugin names as registered by lynx-mysql / lynx-redis. When a plugin is absent from the
// plugin manager (section missing or `enabled: false` in bootstrap), the providers below return
// nil instead of failing so internal/data can fall back to in-memory storage.
const (
	mysqlClientPluginName = "mysql.client"
	redisClientPluginName = "redis.client"
)

func pluginLoaded(app *lynx.LynxApp, name string) bool {
	if app == nil {
		return false
	}
	manager := app.GetPluginManager()
	if manager == nil {
		return false
	}
	return manager.GetPlugin(name) != nil
}

// provideDBProvider returns the stable MySQL provider, or nil when the mysql client plugin is not loaded.
func provideDBProvider(app *lynx.LynxApp) (interfaces.DBProvider, error) {
	if app == nil {
		return nil, fmt.Errorf("lynx app is nil")
	}
	if !pluginLoaded(app, mysqlClientPluginName) {
		log.Infof("plugin %s not loaded; database provider disabled", mysqlClientPluginName)
		return nil, nil
	}
	provider := lynxmysql.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("mysql provider is nil")
	}
	return provider, nil
}

func provideEntClientProvider(provider interfaces.DBProvider) data.EntClientProvider {
	if provider == nil {
		return nil
	}
	return data.NewEntClientProviderFromDB(provider)
}

// provideRedisProvider keeps template-side wiring on the stable redis provider facade.
// It returns nil when the redis client plugin is not loaded.
// Transport readiness/health aliases remain runtime/plugin concerns and are not consumed here.
func provideRedisProvider(app *lynx.LynxApp) (lynxredis.Provider, error) {
	if app == nil {
		return nil, fmt.Errorf("lynx app is nil")
	}
	if !pluginLoaded(app, redisClientPluginName) {
		log.Infof("plugin %s not loaded; redis provider disabled", redisClientPluginName)
		return nil, nil
	}
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

// provideLoginLockRunner uses the redis distributed lock when redis is loaded and otherwise
// falls back to an in-process, per-key mutex (single instance semantics only).
func provideLoginLockRunner(app *lynx.LynxApp) service.LockRunner {
	if !pluginLoaded(app, redisClientPluginName) {
		log.Infof("plugin %s not loaded; login lock runs with an in-process mutex", redisClientPluginName)
		return newLocalLockRunner()
	}
	return func(ctx context.Context, key string, expiration time.Duration, fn func() error) error {
		return redislock.Lock(ctx, key, expiration, fn)
	}
}

// newLocalLockRunner returns a LockRunner backed by per-key sync.Mutex values.
func newLocalLockRunner() service.LockRunner {
	var locks sync.Map
	return func(ctx context.Context, key string, _ time.Duration, fn func() error) error {
		if fn == nil {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		mu, _ := locks.LoadOrStore(key, &sync.Mutex{})
		m := mu.(*sync.Mutex)
		m.Lock()
		defer m.Unlock()
		return fn()
	}
}
