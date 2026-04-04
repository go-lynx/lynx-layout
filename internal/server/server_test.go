package server

import (
	"strings"
	"testing"

	transportgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	transporthttp "github.com/go-kratos/kratos/v2/transport/http"
	loginV1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/go-lynx/lynx-layout/internal/service"
	grpc "google.golang.org/grpc"
)

func TestNewGRPCServerRegistersService(t *testing.T) {
	originalRegister := registerLoginGRPCServer
	t.Cleanup(func() {
		registerLoginGRPCServer = originalRegister
	})

	expectedServer := transportgrpc.NewServer()
	registered := false

	registerLoginGRPCServer = func(registrar grpc.ServiceRegistrar, srv loginV1.LoginServer) {
		registered = true
	}

	got, err := NewGRPCServer(GRPCServerBase{Server: expectedServer}, service.NewLoginService(nil, nil))
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

func TestNewGRPCServerRejectsNilBaseServer(t *testing.T) {
	if _, err := NewGRPCServer(GRPCServerBase{}, service.NewLoginService(nil, nil)); err == nil || !strings.Contains(err.Error(), "gRPC 服务实例为空") {
		t.Fatalf("expected nil base server error, got %v", err)
	}
}

func TestNewHTTPServerRejectsNilBaseServer(t *testing.T) {
	if _, err := NewHTTPServer(HTTPServerBase{}, service.NewLoginService(nil, nil)); err == nil || !strings.Contains(err.Error(), "HTTP 服务实例为空") {
		t.Fatalf("expected nil base server error, got %v", err)
	}
}

func TestNewHTTPServerConvertsRegisterPanicToError(t *testing.T) {
	originalRegister := registerLoginHTTPServer
	t.Cleanup(func() {
		registerLoginHTTPServer = originalRegister
	})

	registerLoginHTTPServer = func(*transporthttp.Server, loginV1.LoginHTTPServer) {
		panic("http register panic")
	}

	if _, err := NewHTTPServer(HTTPServerBase{Server: transporthttp.NewServer()}, service.NewLoginService(nil, nil)); err == nil || !strings.Contains(err.Error(), "http register panic") {
		t.Fatalf("expected register panic to be converted into error, got %v", err)
	}
}
