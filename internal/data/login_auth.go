package data

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-lynx/lynx"
	lynxgrpc "github.com/go-lynx/lynx-grpc"
	"github.com/go-lynx/lynx-layout/internal/bo"
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

var grpcClientConnectionGetter = lynxgrpc.GetGrpcClientConnection

type loginAuthConfig struct {
	ServiceName string
	Method      string
	Timeout     time.Duration
}

func (r *loginRepo) issueLoginAuthToken(ctx context.Context, user *bo.UserBO) (string, error) {
	if err := validateLoginAuthInput(ctx, user); err != nil {
		return "", err
	}

	authConfig, err := loadLoginAuthConfig()
	if err != nil {
		return "", err
	}

	callCtx, cancel := withLoginAuthTimeout(ctx, authConfig.Timeout)
	if cancel != nil {
		defer cancel()
	}

	conn, err := grpcClientConnectionGetter(authConfig.ServiceName, nil)
	if err != nil {
		return "", fmt.Errorf("获取登录鉴权 gRPC 连接失败: %w", err)
	}
	if conn == nil {
		return "", fmt.Errorf("登录鉴权 gRPC 连接为空: service=%s", authConfig.ServiceName)
	}

	req, err := buildLoginAuthRequest(user)
	if err != nil {
		return "", err
	}

	resp := &structpb.Struct{}
	// 当前模板还没有沉淀统一的鉴权 proto，这里先用 Struct 打通标准 gRPC 调用链。
	// 接入真实鉴权服务后，只需要把 method 配置和请求/响应消息体替换为正式 proto 类型。
	if err := conn.Invoke(callCtx, authConfig.Method, req, resp); err != nil {
		return "", fmt.Errorf("调用登录鉴权 gRPC 方法失败: %w", err)
	}

	token, err := extractLoginAuthToken(resp)
	if err != nil {
		return "", err
	}
	return token, nil
}

func validateLoginAuthInput(ctx context.Context, user *bo.UserBO) error {
	if ctx == nil {
		return fmt.Errorf("登录鉴权上下文不能为空")
	}
	if user == nil {
		return fmt.Errorf("登录鉴权用户信息不能为空")
	}
	if user.Id <= 0 {
		return fmt.Errorf("登录鉴权用户ID非法: %d", user.Id)
	}
	if strings.TrimSpace(user.Account) == "" {
		return fmt.Errorf("登录鉴权账号不能为空")
	}
	return nil
}

func loadLoginAuthConfig() (loginAuthConfig, error) {
	return resolveLoginAuthConfig(currentLoginAuthRuntimeConfig(), os.Getenv)
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
			return loginAuthConfig{}, fmt.Errorf("解析环境变量 %s 失败: %w", loginAuthTimeoutEnvKey, err)
		}
		authConfig.Timeout = timeout
	}

	if authConfig.ServiceName == "" {
		return loginAuthConfig{}, fmt.Errorf("未配置登录鉴权 gRPC 服务名，请设置 %s 或 %s", loginAuthServiceConfigKey, loginAuthServiceEnvKey)
	}
	if err := validateLoginAuthMethod(authConfig.Method); err != nil {
		return loginAuthConfig{}, err
	}
	if authConfig.Timeout <= 0 {
		return loginAuthConfig{}, fmt.Errorf("登录鉴权超时时间必须大于 0: %s", authConfig.Timeout)
	}
	return authConfig, nil
}

func currentLoginAuthRuntimeConfig() config.Config {
	app := lynx.Lynx()
	if app == nil {
		return nil
	}
	pluginManager := app.GetPluginManager()
	if pluginManager == nil {
		return nil
	}
	runtime := pluginManager.GetRuntime()
	if runtime == nil {
		return nil
	}
	return runtime.GetConfig()
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
		return 0, false, fmt.Errorf("解析配置 %s 失败: %w", key, err)
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
		return fmt.Errorf("未配置登录鉴权 gRPC 方法，请设置 %s 或 %s", loginAuthMethodConfigKey, loginAuthMethodEnvKey)
	}
	if !strings.HasPrefix(method, "/") {
		return fmt.Errorf("登录鉴权 gRPC 方法格式非法，必须以 / 开头: %s", method)
	}
	if strings.Count(method, "/") != 2 {
		return fmt.Errorf("登录鉴权 gRPC 方法格式非法，必须是 /package.Service/Method: %s", method)
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
		return nil, fmt.Errorf("构建登录鉴权请求失败: %w", err)
	}
	return payload, nil
}

func extractLoginAuthToken(resp *structpb.Struct) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("登录鉴权响应不能为空")
	}

	tokenValue, ok := resp.AsMap()["token"]
	if !ok {
		return "", fmt.Errorf("登录鉴权响应缺少 token 字段")
	}

	token, ok := tokenValue.(string)
	if !ok {
		return "", fmt.Errorf("登录鉴权响应 token 字段类型非法: %T", tokenValue)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("登录鉴权响应 token 不能为空")
	}
	return token, nil
}
