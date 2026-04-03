package server

import (
	"errors"
	"strings"
	"testing"

	transportgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	transporthttp "github.com/go-kratos/kratos/v2/transport/http"
	loginV1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/go-lynx/lynx-layout/internal/service"
	grpc "google.golang.org/grpc"
)

func TestNewGRPCServerRegistersService(t *testing.T) {
	originalGetter := grpcServerGetter
	originalRegister := registerLoginGRPCServer
	t.Cleanup(func() {
		grpcServerGetter = originalGetter
		registerLoginGRPCServer = originalRegister
	})

	expectedServer := transportgrpc.NewServer()
	registered := false

	grpcServerGetter = func(any) (*transportgrpc.Server, error) {
		return expectedServer, nil
	}
	registerLoginGRPCServer = func(registrar grpc.ServiceRegistrar, srv loginV1.LoginServer) {
		registered = true
	}

	got, err := NewGRPCServer(service.NewLoginService(nil))
	if err != nil {
		t.Fatalf("expected grpc server, got error: %v", err)
	}
	if got != expectedServer {
		t.Fatalf("expected original grpc server instance")
	}
	if !registered {
		t.Fatalf("expected grpc service registration to be invoked")
	}
}

func TestNewGRPCServerConvertsPanicToError(t *testing.T) {
	originalGetter := grpcServerGetter
	t.Cleanup(func() {
		grpcServerGetter = originalGetter
	})

	grpcServerGetter = func(any) (*transportgrpc.Server, error) {
		panic("grpc getter panic")
	}

	if _, err := NewGRPCServer(service.NewLoginService(nil)); err == nil || !strings.Contains(err.Error(), "grpc getter panic") {
		t.Fatalf("expected panic to be converted into error, got %v", err)
	}
}

func TestNewHTTPServerReturnsGetterError(t *testing.T) {
	originalGetter := httpServerGetter
	t.Cleanup(func() {
		httpServerGetter = originalGetter
	})

	expectedErr := errors.New("http getter failed")
	httpServerGetter = func() (*transporthttp.Server, error) {
		return nil, expectedErr
	}

	if _, err := NewHTTPServer(service.NewLoginService(nil)); !errors.Is(err, expectedErr) {
		t.Fatalf("expected getter error, got %v", err)
	}
}

func TestNewHTTPServerConvertsRegisterPanicToError(t *testing.T) {
	originalGetter := httpServerGetter
	originalRegister := registerLoginHTTPServer
	t.Cleanup(func() {
		httpServerGetter = originalGetter
		registerLoginHTTPServer = originalRegister
	})

	httpServerGetter = func() (*transporthttp.Server, error) {
		return transporthttp.NewServer(), nil
	}
	registerLoginHTTPServer = func(*transporthttp.Server, loginV1.LoginHTTPServer) {
		panic("http register panic")
	}

	if _, err := NewHTTPServer(service.NewLoginService(nil)); err == nil || !strings.Contains(err.Error(), "http register panic") {
		t.Fatalf("expected register panic to be converted into error, got %v", err)
	}
}
