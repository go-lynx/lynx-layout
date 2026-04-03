package data

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"google.golang.org/protobuf/types/known/structpb"
)

type staticSource struct {
	kv *config.KeyValue
}

type staticWatcher struct {
	stop chan struct{}
}

func (s *staticSource) Load() ([]*config.KeyValue, error) {
	return []*config.KeyValue{s.kv}, nil
}

func (s *staticSource) Watch() (config.Watcher, error) {
	return &staticWatcher{stop: make(chan struct{})}, nil
}

func (w *staticWatcher) Next() ([]*config.KeyValue, error) {
	<-w.stop
	return nil, context.Canceled
}

func (w *staticWatcher) Stop() error {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
	return nil
}

func TestValidateLoginAuthInput(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context
		user *bo.UserBO
	}{
		{
			name: "nil_context",
			user: &bo.UserBO{Id: 1, Account: "demo"},
		},
		{
			name: "nil_user",
			ctx:  context.Background(),
		},
		{
			name: "invalid_user_id",
			ctx:  context.Background(),
			user: &bo.UserBO{Id: 0, Account: "demo"},
		},
		{
			name: "empty_account",
			ctx:  context.Background(),
			user: &bo.UserBO{Id: 1, Account: "   "},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := validateLoginAuthInput(testCase.ctx, testCase.user); err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}

	if err := validateLoginAuthInput(context.Background(), &bo.UserBO{Id: 1, Account: "demo"}); err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}
}

func TestResolveLoginAuthConfig(t *testing.T) {
	cfg := config.New(
		config.WithSource(&staticSource{kv: &config.KeyValue{
			Key:    t.Name() + ".yaml",
			Format: "yaml",
			Value: []byte(`
lynx:
  layout:
    auth:
      grpc:
        service: "config-auth-service"
        method: "/layout.auth.v1.AuthService/IssueLoginToken"
        timeout: "7s"
`),
		}}),
	)
	if err := cfg.Load(); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	t.Cleanup(func() {
		_ = cfg.Close()
	})

	envValues := map[string]string{
		loginAuthServiceEnvKey: "env-auth-service",
		loginAuthTimeoutEnvKey: "3s",
	}

	authConfig, err := resolveLoginAuthConfig(cfg, func(key string) string {
		return envValues[key]
	})
	if err != nil {
		t.Fatalf("expected config to resolve, got error: %v", err)
	}

	if authConfig.ServiceName != "env-auth-service" {
		t.Fatalf("expected env service override, got %q", authConfig.ServiceName)
	}
	if authConfig.Method != "/layout.auth.v1.AuthService/IssueLoginToken" {
		t.Fatalf("unexpected grpc method: %q", authConfig.Method)
	}
	if authConfig.Timeout != 3*time.Second {
		t.Fatalf("expected env timeout override, got %s", authConfig.Timeout)
	}
}

func TestResolveLoginAuthConfigRejectsInvalidMethod(t *testing.T) {
	_, err := resolveLoginAuthConfig(nil, func(key string) string {
		switch key {
		case loginAuthServiceEnvKey:
			return "auth-service"
		case loginAuthMethodEnvKey:
			return "layout.auth.v1.AuthService/IssueLoginToken"
		default:
			return ""
		}
	})
	if err == nil {
		t.Fatalf("expected invalid grpc method error")
	}
}

func TestBuildAndExtractLoginAuthPayload(t *testing.T) {
	req, err := buildLoginAuthRequest(&bo.UserBO{
		Id:       7,
		Num:      "U-7",
		Account:  "demo",
		Nickname: "tester",
		Avatar:   "https://example.com/avatar.png",
		Stats:    1,
	})
	if err != nil {
		t.Fatalf("expected request payload, got error: %v", err)
	}

	if got := req.AsMap()["account"]; got != "demo" {
		t.Fatalf("unexpected account payload: %v", got)
	}

	token, err := extractLoginAuthToken(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			"token": structpb.NewStringValue("signed-token"),
		},
	})
	if err != nil {
		t.Fatalf("expected token extraction success, got error: %v", err)
	}
	if token != "signed-token" {
		t.Fatalf("unexpected token: %s", token)
	}

	if _, err := extractLoginAuthToken(&structpb.Struct{}); err == nil {
		t.Fatalf("expected missing token error")
	}
}
