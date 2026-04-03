package integration

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	v1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type MockLoginHandler func(context.Context, *v1.LoginRequest) (*v1.LoginReply, error)

type MockLoginOption func(*MockLoginService)

type LoginCall struct {
	Request    *v1.LoginRequest
	Metadata   metadata.MD
	ReceivedAt time.Time
}

// MockLoginService provides a minimal reusable Login gRPC fake for integration tests.
type MockLoginService struct {
	v1.UnimplementedLoginServer

	mu      sync.Mutex
	calls   []LoginCall
	reply   *v1.LoginReply
	err     error
	handler MockLoginHandler
}

// MockGRPCHarness owns the server lifecycle, dialed connection, client, and request recorder.
type MockGRPCHarness struct {
	Address string
	Server  *grpc.Server
	Conn    *grpc.ClientConn
	Client  v1.LoginClient
	Service *MockLoginService

	listener net.Listener
}

func NewMockLoginService(opts ...MockLoginOption) *MockLoginService {
	svc := &MockLoginService{
		reply: &v1.LoginReply{
			Token: "mock-token",
			User: &v1.UserInfo{
				Num:      "MOCK-001",
				Account:  "",
				NickName: "Mock User",
			},
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(svc)
		}
	}
	return svc
}

func WithMockLoginReply(reply *v1.LoginReply) MockLoginOption {
	return func(svc *MockLoginService) {
		svc.reply = cloneLoginReply(reply)
	}
}

func WithMockLoginError(err error) MockLoginOption {
	return func(svc *MockLoginService) {
		svc.err = err
	}
}

func WithMockLoginHandler(handler MockLoginHandler) MockLoginOption {
	return func(svc *MockLoginService) {
		svc.handler = handler
	}
}

func (svc *MockLoginService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	svc.mu.Lock()
	svc.calls = append(svc.calls, LoginCall{
		Request:    cloneLoginRequest(req),
		Metadata:   metadataCopy(metadataFromIncomingContext(ctx)),
		ReceivedAt: time.Now(),
	})
	handler := svc.handler
	reply := cloneLoginReply(svc.reply)
	err := svc.err
	svc.mu.Unlock()

	if handler != nil {
		return handler(ctx, cloneLoginRequest(req))
	}
	if err != nil {
		return nil, err
	}
	if reply == nil {
		reply = &v1.LoginReply{
			Token: "mock-token",
			User: &v1.UserInfo{
				Num:      "MOCK-001",
				Account:  req.GetAccount(),
				NickName: "Mock User",
			},
		}
	}
	if reply.User != nil && reply.User.Account == "" {
		reply.User.Account = req.GetAccount()
	}
	return reply, nil
}

func (svc *MockLoginService) SetReply(reply *v1.LoginReply) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.reply = cloneLoginReply(reply)
}

func (svc *MockLoginService) SetError(err error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.err = err
}

func (svc *MockLoginService) SetHandler(handler MockLoginHandler) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.handler = handler
}

func (svc *MockLoginService) Calls() []LoginCall {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	calls := make([]LoginCall, len(svc.calls))
	for i, call := range svc.calls {
		calls[i] = LoginCall{
			Request:    cloneLoginRequest(call.Request),
			Metadata:   metadataCopy(call.Metadata),
			ReceivedAt: call.ReceivedAt,
		}
	}
	return calls
}

func (svc *MockLoginService) ResetCalls() {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.calls = nil
}

func (svc *MockLoginService) RequireCallCount(tb testing.TB, want int) {
	tb.Helper()
	require.Len(tb, svc.Calls(), want)
}

func (svc *MockLoginService) RequireLastLoginRequest(tb testing.TB, want *v1.LoginRequest) {
	tb.Helper()

	calls := svc.Calls()
	require.NotEmpty(tb, calls, "expected at least one recorded Login request")
	RequireLoginRequest(tb, calls[len(calls)-1].Request, want)
}

func (svc *MockLoginService) RequireLastMetadataValue(tb testing.TB, key, want string) {
	tb.Helper()

	calls := svc.Calls()
	require.NotEmpty(tb, calls, "expected at least one recorded Login request")
	got := calls[len(calls)-1].Metadata.Get(key)
	require.NotEmpty(tb, got, "expected metadata key %q to be present", key)
	require.Equal(tb, want, got[0])
}

func StartMockLoginGRPCServer(tb testing.TB, svc *MockLoginService) *MockGRPCHarness {
	tb.Helper()

	if svc == nil {
		svc = NewMockLoginService()
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(tb, err)

	server := grpc.NewServer()
	v1.RegisterLoginServer(server, svc)

	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- server.Serve(listener)
	}()

	conn, client, err := DialMockLoginGRPC(context.Background(), listener.Addr().String())
	require.NoError(tb, err)

	harness := &MockGRPCHarness{
		Address:  listener.Addr().String(),
		Server:   server,
		Conn:     conn,
		Client:   client,
		Service:  svc,
		listener: listener,
	}

	tb.Cleanup(func() {
		harness.Close()
		select {
		case serveErr := <-serveErrCh:
			if serveErr != nil {
				require.NoError(tb, serveErr)
			}
		case <-time.After(time.Second):
			tb.Fatalf("mock gRPC server did not stop within 1s")
		}
	})

	return harness
}

func DialMockLoginGRPC(ctx context.Context, address string, opts ...grpc.DialOption) (*grpc.ClientConn, v1.LoginClient, error) {
	if address == "" {
		return nil, nil, fmt.Errorf("mock gRPC address is required")
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}
	dialOpts = append(dialOpts, opts...)

	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, address, dialOpts...)
	if err != nil {
		return nil, nil, err
	}
	return conn, v1.NewLoginClient(conn), nil
}

func (h *MockGRPCHarness) Close() {
	if h == nil {
		return
	}
	if h.Conn != nil {
		_ = h.Conn.Close()
		h.Conn = nil
	}
	if h.Server != nil {
		h.Server.GracefulStop()
		h.Server = nil
	}
	if h.listener != nil {
		_ = h.listener.Close()
		h.listener = nil
	}
}

func RequireLoginRequest(tb testing.TB, got, want *v1.LoginRequest) {
	tb.Helper()
	require.Truef(tb, proto.Equal(got, want), "unexpected login request:\n got: %v\nwant: %v", got, want)
}

func RequireLoginReply(tb testing.TB, got, want *v1.LoginReply) {
	tb.Helper()
	require.Truef(tb, proto.Equal(got, want), "unexpected login reply:\n got: %v\nwant: %v", got, want)
}

func RequireGRPCStatusCode(tb testing.TB, err error, want codes.Code) {
	tb.Helper()
	require.Error(tb, err)
	require.Equal(tb, want, status.Code(err))
}

func metadataFromIncomingContext(ctx context.Context) metadata.MD {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return metadata.MD{}
	}
	return md
}

func metadataCopy(src metadata.MD) metadata.MD {
	if len(src) == 0 {
		return metadata.MD{}
	}
	dst := make(metadata.MD, len(src))
	for key, values := range src {
		copied := make([]string, len(values))
		copy(copied, values)
		dst[key] = copied
	}
	return dst
}

func cloneLoginRequest(req *v1.LoginRequest) *v1.LoginRequest {
	if req == nil {
		return nil
	}
	clone, ok := proto.Clone(req).(*v1.LoginRequest)
	if !ok {
		return nil
	}
	return clone
}

func cloneLoginReply(reply *v1.LoginReply) *v1.LoginReply {
	if reply == nil {
		return nil
	}
	clone, ok := proto.Clone(reply).(*v1.LoginReply)
	if !ok {
		return nil
	}
	return clone
}

func TestMockLoginGRPCHarnessRoundTrip(t *testing.T) {
	service := NewMockLoginService(WithMockLoginReply(&v1.LoginReply{
		Token: "roundtrip-token",
		User: &v1.UserInfo{
			Num:      "MOCK-ROUNDTRIP-001",
			Account:  "roundtrip-user",
			NickName: "Round Trip User",
		},
	}))
	harness := StartMockLoginGRPCServer(t, service)

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("x-route-target", "mock-login"))
	extraConn, extraClient, err := DialMockLoginGRPC(ctx, harness.Address)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = extraConn.Close()
	})

	req := &v1.LoginRequest{
		Account:  "roundtrip-user",
		Password: "roundtrip-password",
	}
	reply, err := extraClient.Login(ctx, req)
	require.NoError(t, err)

	RequireLoginReply(t, reply, &v1.LoginReply{
		Token: "roundtrip-token",
		User: &v1.UserInfo{
			Num:      "MOCK-ROUNDTRIP-001",
			Account:  "roundtrip-user",
			NickName: "Round Trip User",
		},
	})
	service.RequireCallCount(t, 1)
	service.RequireLastLoginRequest(t, req)
	service.RequireLastMetadataValue(t, "x-route-target", "mock-login")

	service.SetError(status.Error(codes.Unauthenticated, "mock login rejected"))
	_, err = harness.Client.Login(context.Background(), req)
	RequireGRPCStatusCode(t, err, codes.Unauthenticated)
}
