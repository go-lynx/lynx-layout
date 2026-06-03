package data

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-lynx/lynx-layout/internal/bo"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	loginAuthServiceConfigKey = "lynx.layout.auth.grpc.service"
	loginAuthMethodConfigKey  = "lynx.layout.auth.grpc.method"
	loginAuthTimeoutConfigKey = "lynx.layout.auth.grpc.timeout"

	loginAuthServiceEnvKey = "LYNX_LAYOUT_AUTH_GRPC_SERVICE"
	loginAuthMethodEnvKey  = "LYNX_LAYOUT_AUTH_GRPC_METHOD"
	loginAuthTimeoutEnvKey = "LYNX_LAYOUT_AUTH_GRPC_TIMEOUT"

	defaultLoginAuthTimeout = 5 * time.Second
)

type GRPCClientConnectionGetter func(serviceName string) (*grpc.ClientConn, error)

type LoginAuthTokenIssuer func(ctx context.Context, user *bo.UserBO) (string, error)

func NewLoginAuthTokenIssuer(runtimeConfig config.Config, grpcClientConnectionGetter GRPCClientConnectionGetter) LoginAuthTokenIssuer {
	return func(ctx context.Context, user *bo.UserBO) (string, error) {
		return issueLoginAuthToken(ctx, user, runtimeConfig, grpcClientConnectionGetter)
	}
}

type loginAuthConfig struct {
	ServiceName string
	Method      string
	Timeout     time.Duration
}

func issueLoginAuthToken(ctx context.Context, user *bo.UserBO, runtimeConfig config.Config, grpcClientConnectionGetter GRPCClientConnectionGetter) (string, error) {
	if err := validateLoginAuthInput(ctx, user); err != nil {
		return "", err
	}
	if grpcClientConnectionGetter == nil {
		return "", fmt.Errorf("login auth gRPC connection getter must not be nil")
	}

	authConfig, err := loadLoginAuthConfig(runtimeConfig)
	if err != nil {
		return "", err
	}

	callCtx, cancel := withLoginAuthTimeout(ctx, authConfig.Timeout)
	if cancel != nil {
		defer cancel()
	}

	conn, err := grpcClientConnectionGetter(authConfig.ServiceName)
	if err != nil {
		return "", fmt.Errorf("failed to get login auth gRPC connection: %w", err)
	}
	if conn == nil {
		return "", fmt.Errorf("login auth gRPC connection is nil: service=%s", authConfig.ServiceName)
	}

	req, err := buildLoginAuthRequest(user)
	if err != nil {
		return "", err
	}

	resp := &structpb.Struct{}
	// The template uses a generic Struct payload until a dedicated auth proto is available.
	// Replace the method config and request/response types with the real proto once adopted.
	if err := conn.Invoke(callCtx, authConfig.Method, req, resp); err != nil {
		return "", fmt.Errorf("failed to invoke login auth gRPC method: %w", err)
	}

	token, err := extractLoginAuthToken(resp)
	if err != nil {
		return "", err
	}
	return token, nil
}

func validateLoginAuthInput(ctx context.Context, user *bo.UserBO) error {
	if ctx == nil {
		return fmt.Errorf("login auth context must not be nil")
	}
	if user == nil {
		return fmt.Errorf("login auth user must not be nil")
	}
	if user.Id <= 0 {
		return fmt.Errorf("login auth user ID is invalid: %d", user.Id)
	}
	if strings.TrimSpace(user.Account) == "" {
		return fmt.Errorf("login auth account must not be empty")
	}
	return nil
}

func loadLoginAuthConfig(runtimeConfig config.Config) (loginAuthConfig, error) {
	return resolveLoginAuthConfig(runtimeConfig, os.Getenv)
}

func resolveLoginAuthConfig(runtimeConfig config.Config, lookupEnv func(string) string) (loginAuthConfig, error) {
	authConfig := loginAuthConfig{
		Timeout: defaultLoginAuthTimeout,
	}

	if serviceName, ok := readLoginAuthStringConfig(runtimeConfig, loginAuthServiceConfigKey); ok {
		authConfig.ServiceName = serviceName
	}
	if method, ok := readLoginAuthStringConfig(runtimeConfig, loginAuthMethodConfigKey); ok {
		authConfig.Method = method
	}
	if timeout, ok, err := readLoginAuthDurationConfig(runtimeConfig, loginAuthTimeoutConfigKey); err != nil {
		return loginAuthConfig{}, err
	} else if ok {
		authConfig.Timeout = timeout
	}

	if envServiceName := readLoginAuthEnv(lookupEnv, loginAuthServiceEnvKey); envServiceName != "" {
		authConfig.ServiceName = envServiceName
	}
	if envMethod := readLoginAuthEnv(lookupEnv, loginAuthMethodEnvKey); envMethod != "" {
		authConfig.Method = envMethod
	}
	if envTimeout := readLoginAuthEnv(lookupEnv, loginAuthTimeoutEnvKey); envTimeout != "" {
		timeout, err := time.ParseDuration(envTimeout)
		if err != nil {
			return loginAuthConfig{}, fmt.Errorf("failed to parse env var %s: %w", loginAuthTimeoutEnvKey, err)
		}
		authConfig.Timeout = timeout
	}

	if authConfig.ServiceName == "" {
		return loginAuthConfig{}, fmt.Errorf("login auth gRPC service name not configured; set %s or %s", loginAuthServiceConfigKey, loginAuthServiceEnvKey)
	}
	if err := validateLoginAuthMethod(authConfig.Method); err != nil {
		return loginAuthConfig{}, err
	}
	if authConfig.Timeout <= 0 {
		return loginAuthConfig{}, fmt.Errorf("login auth timeout must be greater than 0: %s", authConfig.Timeout)
	}
	return authConfig, nil
}

func readLoginAuthStringConfig(runtimeConfig config.Config, key string) (string, bool) {
	if runtimeConfig == nil {
		return "", false
	}
	value, err := runtimeConfig.Value(key).String()
	if err != nil {
		return "", false
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	return value, true
}

func readLoginAuthDurationConfig(runtimeConfig config.Config, key string) (time.Duration, bool, error) {
	rawValue, ok := readLoginAuthStringConfig(runtimeConfig, key)
	if !ok {
		return 0, false, nil
	}
	timeout, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse config %s: %w", key, err)
	}
	return timeout, true, nil
}

func readLoginAuthEnv(lookupEnv func(string) string, key string) string {
	if lookupEnv == nil {
		return ""
	}
	return strings.TrimSpace(lookupEnv(key))
}

func validateLoginAuthMethod(method string) error {
	if method == "" {
		return fmt.Errorf("login auth gRPC method not configured; set %s or %s", loginAuthMethodConfigKey, loginAuthMethodEnvKey)
	}
	if !strings.HasPrefix(method, "/") {
		return fmt.Errorf("login auth gRPC method must start with /: %s", method)
	}
	if strings.Count(method, "/") != 2 {
		return fmt.Errorf("login auth gRPC method must be in the form /package.Service/Method: %s", method)
	}
	return nil
}

func withLoginAuthTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, nil
	}
	return context.WithTimeout(ctx, timeout)
}

func buildLoginAuthRequest(user *bo.UserBO) (*structpb.Struct, error) {
	payload, err := structpb.NewStruct(map[string]any{
		"user_id":  user.Id,
		"user_num": user.Num,
		"account":  user.Account,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"stats":    user.Stats,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build login auth request: %w", err)
	}
	return payload, nil
}

func extractLoginAuthToken(resp *structpb.Struct) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("login auth response must not be nil")
	}

	tokenValue, ok := resp.AsMap()["token"]
	if !ok {
		return "", fmt.Errorf("login auth response missing token field")
	}

	token, ok := tokenValue.(string)
	if !ok {
		return "", fmt.Errorf("login auth response token field has unexpected type: %T", tokenValue)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("login auth response token must not be empty")
	}
	return token, nil
}
