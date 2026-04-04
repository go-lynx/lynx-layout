package main

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-lynx/lynx"
	"github.com/go-lynx/lynx-layout/internal/server"
	"github.com/go-lynx/lynx-layout/internal/service"
	lynxkratos "github.com/go-lynx/lynx/kratos"
	lynxlog "github.com/go-lynx/lynx/log"
)

type testConfigSource struct {
	kv *config.KeyValue
}

type testConfigWatcher struct {
	stop chan struct{}
}

func (s *testConfigSource) Load() ([]*config.KeyValue, error) {
	return []*config.KeyValue{s.kv}, nil
}

func (s *testConfigSource) Watch() (config.Watcher, error) {
	return &testConfigWatcher{stop: make(chan struct{})}, nil
}

func (w *testConfigWatcher) Next() ([]*config.KeyValue, error) {
	<-w.stop
	return nil, context.Canceled
}

func (w *testConfigWatcher) Stop() error {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
	return nil
}

func newLayoutTestConfig(t *testing.T, name string) config.Config {
	t.Helper()

	cfg := config.New(
		config.WithSource(&testConfigSource{kv: &config.KeyValue{
			Key:    t.Name() + ".yaml",
			Format: "yaml",
			Value: []byte("lynx:\n" +
				"  application:\n" +
				"    name: " + name + "\n" +
				"    version: v0.0.1\n"),
		}}),
	)
	if err := cfg.Load(); err != nil {
		t.Fatalf("failed to load layout test config: %v", err)
	}
	t.Cleanup(func() {
		_ = cfg.Close()
	})
	return cfg
}

func TestProvideLynxApp_UsesPublishedDefaultApp(t *testing.T) {
	lynx.ClearDefaultApp()
	t.Cleanup(lynx.ClearDefaultApp)

	cfg := newLayoutTestConfig(t, "layout-default-app")
	app, err := lynx.NewStandaloneApp(cfg)
	if err != nil {
		t.Fatalf("failed to create standalone app: %v", err)
	}
	t.Cleanup(func() {
		_ = app.Close()
	})

	lynx.SetDefaultApp(app)

	got, err := provideLynxApp()
	if err != nil {
		t.Fatalf("expected provideLynxApp to resolve published app: %v", err)
	}
	if got != app {
		t.Fatalf("expected published app pointer, got %p want %p", got, app)
	}
}

func TestProvideLynxApp_RejectsMissingDefaultApp(t *testing.T) {
	lynx.ClearDefaultApp()
	t.Cleanup(lynx.ClearDefaultApp)

	if _, err := provideLynxApp(); err == nil {
		t.Fatal("expected missing default app error")
	}
}

func TestProvideHTTPServerBase_RejectsMissingDefaultAppWithoutPanic(t *testing.T) {
	lynx.ClearDefaultApp()
	t.Cleanup(lynx.ClearDefaultApp)

	if _, err := provideHTTPServerBase(); err == nil {
		t.Fatal("expected http server base lookup to fail when default app is missing")
	}
}

func TestProvideServiceRegistrar_AllowsNilRegistrar(t *testing.T) {
	cfg := newLayoutTestConfig(t, "layout-nil-registrar")
	app, err := lynx.NewStandaloneApp(cfg)
	if err != nil {
		t.Fatalf("failed to create standalone app: %v", err)
	}
	t.Cleanup(func() {
		_ = app.Close()
	})

	registrar, err := provideServiceRegistrar(app)
	if err != nil {
		t.Fatalf("expected nil registrar path to be accepted: %v", err)
	}
	if registrar != nil {
		t.Fatalf("expected default control plane to expose nil registrar, got %#v", registrar)
	}
}

func TestProvideRuntimeConfig_UsesAppOwnedConfig(t *testing.T) {
	cfg := newLayoutTestConfig(t, "layout-runtime-config")
	app, err := lynx.NewStandaloneApp(cfg)
	if err != nil {
		t.Fatalf("failed to create standalone app: %v", err)
	}
	t.Cleanup(func() {
		_ = app.Close()
	})

	got, err := provideRuntimeConfig(app)
	if err != nil {
		t.Fatalf("expected runtime config from explicit app: %v", err)
	}
	if got != cfg {
		t.Fatalf("expected original config pointer, got %p want %p", got, cfg)
	}
}

func TestProvideRedisProvider_RejectsMissingProvider(t *testing.T) {
	lynx.ClearDefaultApp()
	t.Cleanup(lynx.ClearDefaultApp)

	provider, err := provideRedisProvider()
	if err == nil {
		t.Fatal("expected redis provider lookup to fail when no app is published")
	}
	if provider != nil {
		t.Fatal("expected redis provider to remain nil when lookup fails")
	}
}

func TestProvidersSmoke_LocalNoControlPlaneHTTPAndGRPCRegistration(t *testing.T) {
	lynx.ClearDefaultApp()
	t.Cleanup(lynx.ClearDefaultApp)

	cfg := config.New(
		config.WithSource(&testConfigSource{kv: &config.KeyValue{
			Key:    t.Name() + ".yaml",
			Format: "yaml",
			Value: []byte("" +
				"lynx:\n" +
				"  application:\n" +
				"    name: layout-smoke\n" +
				"    version: v0.0.1\n" +
				"  http:\n" +
				"    addr: \":18080\"\n" +
				"  grpc:\n" +
				"    service:\n" +
				"      addr: \":19090\"\n"),
		}}),
	)
	if err := cfg.Load(); err != nil {
		t.Fatalf("failed to load smoke config: %v", err)
	}
	t.Cleanup(func() {
		_ = cfg.Close()
	})

	app, err := lynx.NewStandaloneApp(cfg)
	if err != nil {
		t.Fatalf("failed to create standalone app: %v", err)
	}
	t.Cleanup(func() {
		_ = app.Close()
	})

	lynx.SetDefaultApp(app)

	if err := app.LoadPlugins(); err != nil {
		t.Fatalf("expected local no-control-plane plugin load to succeed: %v", err)
	}

	providedApp, err := provideLynxApp()
	if err != nil {
		t.Fatalf("expected published default app to resolve: %v", err)
	}
	if providedApp != app {
		t.Fatalf("expected resolved app pointer to match standalone app")
	}

	runtimeConfig, err := provideRuntimeConfig(app)
	if err != nil {
		t.Fatalf("expected runtime config to resolve: %v", err)
	}
	if runtimeConfig != cfg {
		t.Fatalf("expected runtime config pointer to match smoke config")
	}

	httpBase, err := provideHTTPServerBase()
	if err != nil {
		t.Fatalf("expected http server base: %v", err)
	}
	if httpBase.Server == nil {
		t.Fatal("expected non-nil http server")
	}

	grpcBase, err := provideGRPCServerBase(app)
	if err != nil {
		t.Fatalf("expected grpc server base: %v", err)
	}
	if grpcBase.Server == nil {
		t.Fatal("expected non-nil grpc server")
	}

	registrar, err := provideServiceRegistrar(app)
	if err != nil {
		t.Fatalf("expected nil registrar path to remain valid: %v", err)
	}
	if registrar != nil {
		t.Fatalf("expected no-control-plane startup to expose nil registrar, got %#v", registrar)
	}

	loginService := service.NewLoginService(nil, nil)
	httpServer, err := server.NewHTTPServer(httpBase, loginService)
	if err != nil {
		t.Fatalf("expected http server registration to succeed: %v", err)
	}
	grpcServer, err := server.NewGRPCServer(grpcBase, loginService)
	if err != nil {
		t.Fatalf("expected grpc server registration to succeed: %v", err)
	}

	if err := lynxlog.InitLogger(app.Name(), app.Host(), app.Version(), cfg); err != nil {
		t.Fatalf("expected logger init to succeed: %v", err)
	}

	kratosApp, err := lynxkratos.NewKratos(lynxkratos.ProvideKratosOptions(app, grpcServer, httpServer, registrar))
	if err != nil {
		t.Fatalf("expected kratos app creation to succeed with nil registrar: %v", err)
	}
	if kratosApp == nil {
		t.Fatal("expected kratos app instance")
	}
}
