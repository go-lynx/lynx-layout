package integration

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	lynxmysqlpkg "github.com/go-lynx/lynx-mysql"
	lynxredis "github.com/go-lynx/lynx-redis"
	sqlinterfaces "github.com/go-lynx/lynx-sql-sdk/interfaces"
	"github.com/redis/go-redis/v9"
)

const (
	defaultSuiteProbeTimeout   = 3 * time.Second
	defaultSuiteCleanupTimeout = 5 * time.Second
	defaultSuiteGRPCAddress    = "localhost:9000"
)

type suiteDependency string

const (
	suiteDependencyRedis suiteDependency = "redis"
	suiteDependencyMySQL suiteDependency = "mysql"
	suiteDependencyGRPC  suiteDependency = "grpc"
)

type suiteDependencyState struct {
	Name      suiteDependency
	Ready     bool
	Reason    string
	Detail    string
	CheckedAt time.Time
}

type suiteCleanupHook struct {
	name string
	fn   func(context.Context) error
}

type integrationSuite struct {
	mu          sync.RWMutex
	deps        map[suiteDependency]suiteDependencyState
	cleanups    []suiteCleanupHook
	grpcAddress string
}

var packageSuite = newIntegrationSuite()

func newIntegrationSuite() *integrationSuite {
	return &integrationSuite{
		deps:        make(map[suiteDependency]suiteDependencyState, 3),
		grpcAddress: resolveSuiteGRPCAddress(),
	}
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultSuiteProbeTimeout)
	packageSuite.bootstrap(ctx)
	cancel()

	fmt.Fprintln(os.Stderr, packageSuite.Diagnostics())

	code := m.Run()

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), defaultSuiteCleanupTimeout)
	cleanupErr := packageSuite.runCleanups(cleanupCtx)
	cleanupCancel()
	if cleanupErr != nil {
		fmt.Fprintf(os.Stderr, "[integration-suite] cleanup failed: %v\n", cleanupErr)
		if code == 0 {
			code = 1
		}
	}

	os.Exit(code)
}

func (s *integrationSuite) bootstrap(ctx context.Context) {
	s.setDependency(s.probeRedis(ctx))
	s.setDependency(s.probeMySQL(ctx))
	s.setDependency(s.probeGRPC(ctx))
}

func (s *integrationSuite) probeRedis(ctx context.Context) suiteDependencyState {
	state := suiteDependencyState{
		Name:      suiteDependencyRedis,
		CheckedAt: time.Now(),
	}

	client, err := suiteGetRedisClient()
	if err != nil {
		state.Reason = "lynx-redis probe failed"
		state.Detail = err.Error()
		return state
	}
	if client == nil {
		state.Reason = "lynx-redis plugin not initialized"
		state.Detail = "GetUniversalRedis() returned nil"
		return state
	}

	probeCtx, cancel := context.WithTimeout(ctx, defaultSuiteProbeTimeout)
	defer cancel()
	if err := client.Ping(probeCtx).Err(); err != nil {
		state.Reason = "redis ping failed"
		state.Detail = err.Error()
		return state
	}

	state.Ready = true
	state.Detail = "redis ping succeeded"
	return state
}

func (s *integrationSuite) probeMySQL(ctx context.Context) suiteDependencyState {
	state := suiteDependencyState{
		Name:      suiteDependencyMySQL,
		CheckedAt: time.Now(),
	}

	provider, err := suiteGetMySQLProvider()
	if err != nil {
		state.Reason = "lynx-mysql probe failed"
		state.Detail = err.Error()
		return state
	}
	if provider == nil {
		state.Reason = "lynx-mysql provider not initialized"
		state.Detail = "GetProvider() returned nil"
		return state
	}

	probeCtx, cancel := context.WithTimeout(ctx, defaultSuiteProbeTimeout)
	defer cancel()

	db, err := provider.DB(probeCtx)
	if err != nil {
		state.Reason = "mysql DB acquisition failed"
		state.Detail = err.Error()
		return state
	}
	if db == nil {
		state.Reason = "mysql DB acquisition failed"
		state.Detail = "provider.DB returned nil"
		return state
	}
	if err := db.PingContext(probeCtx); err != nil {
		state.Reason = "mysql ping failed"
		state.Detail = err.Error()
		return state
	}

	state.Ready = true
	state.Detail = "mysql ping succeeded"
	return state
}

func (s *integrationSuite) probeGRPC(ctx context.Context) suiteDependencyState {
	state := suiteDependencyState{
		Name:      suiteDependencyGRPC,
		CheckedAt: time.Now(),
	}

	dialer := net.Dialer{Timeout: defaultSuiteProbeTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", s.grpcAddress)
	if err != nil {
		state.Reason = "grpc endpoint not reachable"
		state.Detail = fmt.Sprintf("addr=%s err=%v", s.grpcAddress, err)
		return state
	}
	_ = conn.Close()

	state.Ready = true
	state.Detail = fmt.Sprintf("tcp dial succeeded: %s", s.grpcAddress)
	return state
}

func (s *integrationSuite) setDependency(state suiteDependencyState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deps[state.Name] = state
}

func (s *integrationSuite) dependencyState(dep suiteDependency) suiteDependencyState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.deps[dep]
}

func (s *integrationSuite) registerCleanup(name string, fn func(context.Context) error) {
	if fn == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanups = append(s.cleanups, suiteCleanupHook{name: name, fn: fn})
}

func (s *integrationSuite) runCleanups(ctx context.Context) error {
	s.mu.Lock()
	hooks := append([]suiteCleanupHook(nil), s.cleanups...)
	s.cleanups = nil
	s.mu.Unlock()

	var errs []error
	for i := len(hooks) - 1; i >= 0; i-- {
		hook := hooks[i]
		if err := hook.fn(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", hook.name, err))
		}
	}

	return errors.Join(errs...)
}

func (s *integrationSuite) Diagnostics() string {
	states := []suiteDependencyState{
		s.dependencyState(suiteDependencyRedis),
		s.dependencyState(suiteDependencyMySQL),
		s.dependencyState(suiteDependencyGRPC),
	}

	sort.Slice(states, func(i, j int) bool {
		return states[i].Name < states[j].Name
	})

	lines := make([]string, 0, len(states)+2)
	lines = append(lines, "[integration-suite] dependency probe summary:")
	for _, state := range states {
		status := "not-ready"
		if state.Ready {
			status = "ready"
		}

		line := fmt.Sprintf("- %s: %s", state.Name, status)
		if state.Reason != "" {
			line += " (" + state.Reason + ")"
		}
		if state.Detail != "" {
			line += " [" + state.Detail + "]"
		}
		lines = append(lines, line)
	}
	lines = append(lines, "- grpc_addr: "+s.grpcAddress)
	return strings.Join(lines, "\n")
}

func resolveSuiteGRPCAddress() string {
	for _, key := range []string{
		"LYNX_LAYOUT_TEST_GRPC_ADDR",
		"LYNX_LAYOUT_GRPC_ADDR",
		"LYNX_GRPC_ADDR",
		"GRPC_ADDR",
	} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return defaultSuiteGRPCAddress
}

func suiteContext(tb testing.TB, timeout time.Duration) context.Context {
	tb.Helper()

	if timeout <= 0 {
		timeout = defaultSuiteProbeTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	tb.Cleanup(cancel)
	return ctx
}

func suiteRegisterCleanup(tb testing.TB, name string, fn func(context.Context) error) {
	tb.Helper()
	packageSuite.registerCleanup(name, fn)
}

func suiteLogDiagnostics(tb testing.TB) {
	tb.Helper()
	tb.Log(packageSuite.Diagnostics())
}

func suiteRequireRedis(tb testing.TB) redis.UniversalClient {
	tb.Helper()

	state := packageSuite.dependencyState(suiteDependencyRedis)
	if !state.Ready {
		tb.Skipf("redis dependency unavailable: %s [%s]\n%s", state.Reason, state.Detail, packageSuite.Diagnostics())
	}

	client, err := suiteGetRedisClient()
	if err != nil {
		tb.Skipf("redis dependency unavailable: %v\n%s", err, packageSuite.Diagnostics())
	}
	if client == nil {
		tb.Skipf("redis dependency unavailable: GetUniversalRedis() returned nil\n%s", packageSuite.Diagnostics())
	}

	return client
}

func suiteRequireMySQL(tb testing.TB) sqlinterfaces.DBProvider {
	tb.Helper()

	state := packageSuite.dependencyState(suiteDependencyMySQL)
	if !state.Ready {
		tb.Skipf("mysql dependency unavailable: %s [%s]\n%s", state.Reason, state.Detail, packageSuite.Diagnostics())
	}

	provider, err := suiteGetMySQLProvider()
	if err != nil {
		tb.Skipf("mysql dependency unavailable: %v\n%s", err, packageSuite.Diagnostics())
	}
	if provider == nil {
		tb.Skipf("mysql dependency unavailable: GetProvider() returned nil\n%s", packageSuite.Diagnostics())
	}

	return provider
}

func suiteRequireGRPC(tb testing.TB) string {
	tb.Helper()

	state := packageSuite.dependencyState(suiteDependencyGRPC)
	if !state.Ready {
		tb.Skipf("grpc dependency unavailable: %s [%s]\n%s", state.Reason, state.Detail, packageSuite.Diagnostics())
	}

	return packageSuite.grpcAddress
}

func suiteMissingDependencies(deps ...suiteDependency) []suiteDependencyState {
	states := make([]suiteDependencyState, 0, len(deps))
	for _, dep := range deps {
		state := packageSuite.dependencyState(dep)
		if !state.Ready {
			states = append(states, state)
		}
	}
	return states
}

func suiteSkipUnlessReady(tb testing.TB, deps ...suiteDependency) {
	tb.Helper()

	missing := suiteMissingDependencies(deps...)
	if len(missing) == 0 {
		return
	}

	parts := make([]string, 0, len(missing))
	for _, state := range missing {
		part := string(state.Name)
		if state.Reason != "" {
			part += ": " + state.Reason
		}
		if state.Detail != "" {
			part += " [" + state.Detail + "]"
		}
		parts = append(parts, part)
	}

	tb.Skipf("integration dependencies unavailable: %s\n%s", strings.Join(parts, "; "), packageSuite.Diagnostics())
}

func suiteGetRedisClient() (_ redis.UniversalClient, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("GetUniversalRedis panic: %v", recovered)
		}
	}()

	return lynxredis.GetUniversalRedis(), nil
}

func suiteGetMySQLProvider() (_ sqlinterfaces.DBProvider, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("GetProvider panic: %v", recovered)
		}
	}()

	return lynxmysqlpkg.GetProvider(), nil
}
