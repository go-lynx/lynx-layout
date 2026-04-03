package integration

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	integrationConfigOverrideEnv        = "LYNX_LAYOUT_TEST_CONFIG"
	integrationGRPCAddrOverrideEnv      = "LYNX_LAYOUT_TEST_GRPC_ADDR"
	integrationMySQLSourceOverrideEnv   = "LYNX_LAYOUT_TEST_MYSQL_SOURCE"
	integrationRedisAddrOverrideEnv     = "LYNX_LAYOUT_TEST_REDIS_ADDR"
	integrationRedisAddrsOverrideEnv    = "LYNX_LAYOUT_TEST_REDIS_ADDRS"
	integrationRedisPasswordOverrideEnv = "LYNX_LAYOUT_TEST_REDIS_PASSWORD"
	integrationRedisDBOverrideEnv       = "LYNX_LAYOUT_TEST_REDIS_DB"
)

type integrationConfigFixture struct {
	path   string
	config integrationBootstrapLocalConfig
}

type integrationDependencyTargets struct {
	ConfigPath     string
	GRPCAddress    string
	MySQLSource    string
	RedisAddress   string
	RedisAddresses []string
	RedisPassword  string
	RedisDB        int
}

type integrationBootstrapLocalConfig struct {
	Lynx struct {
		GRPC struct {
			Service struct {
				Network string `yaml:"network"`
				Addr    string `yaml:"addr"`
				Timeout string `yaml:"timeout"`
			} `yaml:"service"`
		} `yaml:"grpc"`
		MySQL struct {
			Driver string `yaml:"driver"`
			Source string `yaml:"source"`
		} `yaml:"mysql"`
		Redis struct {
			Network      string   `yaml:"network"`
			Addr         string   `yaml:"addr"`
			Addrs        []string `yaml:"addrs"`
			Password     string   `yaml:"password"`
			DB           int      `yaml:"db"`
			DialTimeout  string   `yaml:"dial_timeout"`
			ReadTimeout  string   `yaml:"read_timeout"`
			WriteTimeout string   `yaml:"write_timeout"`
		} `yaml:"redis"`
	} `yaml:"lynx"`
}

func loadIntegrationConfigFixture(tb testing.TB) integrationConfigFixture {
	tb.Helper()

	configPath, err := resolveIntegrationConfigPath()
	if err != nil {
		tb.Fatalf("resolve integration config path: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		tb.Fatalf("read integration config %q: %v", configPath, err)
	}

	var cfg integrationBootstrapLocalConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		tb.Fatalf("unmarshal integration config %q: %v", configPath, err)
	}

	return integrationConfigFixture{
		path:   configPath,
		config: cfg,
	}
}

func resolveIntegrationDependencyTargets(tb testing.TB) integrationDependencyTargets {
	tb.Helper()

	fixture := loadIntegrationConfigFixture(tb)

	redisAddresses := uniqueAddresses(
		[]string{normalizeAddress(os.Getenv(integrationRedisAddrOverrideEnv))},
		splitAndNormalizeAddresses(os.Getenv(integrationRedisAddrsOverrideEnv)),
		normalizeAddresses(fixture.config.Lynx.Redis.Addrs),
		[]string{normalizeAddress(fixture.config.Lynx.Redis.Addr)},
	)

	redisDB := fixture.config.Lynx.Redis.DB
	if raw := strings.TrimSpace(os.Getenv(integrationRedisDBOverrideEnv)); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			tb.Fatalf("parse %s=%q: %v", integrationRedisDBOverrideEnv, raw, err)
		}
		redisDB = parsed
	}

	targets := integrationDependencyTargets{
		ConfigPath:     fixture.path,
		GRPCAddress:    firstNonEmptyString(normalizeAddress(os.Getenv(integrationGRPCAddrOverrideEnv)), normalizeAddress(fixture.config.Lynx.GRPC.Service.Addr), "127.0.0.1:9000"),
		MySQLSource:    firstNonEmptyString(strings.TrimSpace(os.Getenv(integrationMySQLSourceOverrideEnv)), strings.TrimSpace(fixture.config.Lynx.MySQL.Source)),
		RedisAddresses: redisAddresses,
		RedisPassword:  firstNonEmptyString(strings.TrimSpace(os.Getenv(integrationRedisPasswordOverrideEnv)), strings.TrimSpace(fixture.config.Lynx.Redis.Password)),
		RedisDB:        redisDB,
	}
	if len(redisAddresses) > 0 {
		targets.RedisAddress = redisAddresses[0]
	}

	return targets
}

func integrationGRPCAddress(tb testing.TB) string {
	tb.Helper()
	return resolveIntegrationDependencyTargets(tb).GRPCAddress
}

func integrationMySQLSource(tb testing.TB) string {
	tb.Helper()
	return resolveIntegrationDependencyTargets(tb).MySQLSource
}

func integrationRedisTarget(tb testing.TB) (string, string, int) {
	tb.Helper()

	targets := resolveIntegrationDependencyTargets(tb)
	return targets.RedisAddress, targets.RedisPassword, targets.RedisDB
}

func resolveIntegrationConfigPath() (string, error) {
	moduleRoot, fixturesDir, err := integrationTestPaths()
	if err != nil {
		return "", err
	}

	if override := strings.TrimSpace(os.Getenv(integrationConfigOverrideEnv)); override != "" {
		return resolveIntegrationOverridePath(moduleRoot, override)
	}

	candidates := []string{
		// Prefer a test-local copy that mirrors bootstrap.local field layout so
		// integration helpers can resolve stable endpoints without depending on a
		// developer's full local runtime config.
		filepath.Join(fixturesDir, "testdata", "bootstrap.local.yaml"),
		filepath.Join(moduleRoot, "configs", "bootstrap.local.yaml"),
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("stat %q: %w", candidate, err)
		}
	}

	return "", fmt.Errorf("integration config not found in %v", candidates)
}

func integrationTestPaths() (string, string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", "", fmt.Errorf("runtime caller not available")
	}

	fixturesDir := filepath.Dir(file)
	moduleRoot := filepath.Clean(filepath.Join(fixturesDir, "..", ".."))
	return moduleRoot, fixturesDir, nil
}

func resolveIntegrationOverridePath(moduleRoot string, raw string) (string, error) {
	path := strings.TrimSpace(raw)
	if path == "" {
		return "", fmt.Errorf("empty override path")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(moduleRoot, path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("abs override path %q: %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("stat override path %q: %w", absPath, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("override path %q is a directory", absPath)
	}
	return absPath, nil
}

func normalizeAddress(raw string) string {
	address := strings.TrimSpace(raw)
	if address == "" {
		return ""
	}
	if strings.Contains(address, "://") {
		return address
	}
	if strings.HasPrefix(address, ":") {
		return "127.0.0.1" + address
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return address
	}
	if host == "" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

func normalizeAddresses(values []string) []string {
	addresses := make([]string, 0, len(values))
	for _, value := range values {
		if normalized := normalizeAddress(value); normalized != "" {
			addresses = append(addresses, normalized)
		}
	}
	return addresses
}

func splitAndNormalizeAddresses(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	addresses := make([]string, 0, len(parts))
	for _, part := range parts {
		if normalized := normalizeAddress(part); normalized != "" {
			addresses = append(addresses, normalized)
		}
	}
	return addresses
}

func uniqueAddresses(groups ...[]string) []string {
	seen := make(map[string]struct{})
	var addresses []string

	for _, group := range groups {
		for _, address := range group {
			normalized := normalizeAddress(address)
			if normalized == "" {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			addresses = append(addresses, normalized)
		}
	}

	return addresses
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func TestLoadIntegrationConfigFixturePrefersTestdata(t *testing.T) {
	clearIntegrationOverrideEnvs(t)

	_, fixturesDir, err := integrationTestPaths()
	if err != nil {
		t.Fatalf("integrationTestPaths: %v", err)
	}

	wantPath := filepath.Join(fixturesDir, "testdata", "bootstrap.local.yaml")

	fixture := loadIntegrationConfigFixture(t)
	if fixture.path != wantPath {
		t.Fatalf("fixture path = %q, want %q", fixture.path, wantPath)
	}
	if fixture.config.Lynx.GRPC.Service.Addr != "127.0.0.1:9000" {
		t.Fatalf("grpc addr = %q, want %q", fixture.config.Lynx.GRPC.Service.Addr, "127.0.0.1:9000")
	}
	if fixture.config.Lynx.MySQL.Source == "" {
		t.Fatal("mysql source should not be empty")
	}
	if !reflect.DeepEqual(fixture.config.Lynx.Redis.Addrs, []string{"127.0.0.1:6379"}) {
		t.Fatalf("redis addrs = %#v, want %#v", fixture.config.Lynx.Redis.Addrs, []string{"127.0.0.1:6379"})
	}
}

func TestResolveIntegrationDependencyTargetsHonorsOverrides(t *testing.T) {
	clearIntegrationOverrideEnvs(t)

	overridePath := writeIntegrationConfigFixtureFile(t, `
lynx:
  grpc:
    service:
      network: tcp
      addr: 10.0.0.1:9001
      timeout: 10s
  mysql:
    driver: mysql
    source: "mysql-from-config"
  redis:
    network: tcp
    addr: ":6378"
    addrs:
      - 10.0.0.2:6379
    password: config-password
    db: 3
`)

	t.Setenv(integrationConfigOverrideEnv, overridePath)
	t.Setenv(integrationGRPCAddrOverrideEnv, ":9100")
	t.Setenv(integrationMySQLSourceOverrideEnv, "mysql-from-env")
	t.Setenv(integrationRedisAddrOverrideEnv, ":6381")
	t.Setenv(integrationRedisAddrsOverrideEnv, "localhost:6382, :6383")
	t.Setenv(integrationRedisPasswordOverrideEnv, "env-password")
	t.Setenv(integrationRedisDBOverrideEnv, "7")

	targets := resolveIntegrationDependencyTargets(t)

	if targets.ConfigPath != overridePath {
		t.Fatalf("config path = %q, want %q", targets.ConfigPath, overridePath)
	}
	if targets.GRPCAddress != "127.0.0.1:9100" {
		t.Fatalf("grpc address = %q, want %q", targets.GRPCAddress, "127.0.0.1:9100")
	}
	if got := integrationGRPCAddress(t); got != "127.0.0.1:9100" {
		t.Fatalf("integrationGRPCAddress() = %q, want %q", got, "127.0.0.1:9100")
	}
	if targets.MySQLSource != "mysql-from-env" {
		t.Fatalf("mysql source = %q, want %q", targets.MySQLSource, "mysql-from-env")
	}
	if got := integrationMySQLSource(t); got != "mysql-from-env" {
		t.Fatalf("integrationMySQLSource() = %q, want %q", got, "mysql-from-env")
	}

	wantRedisAddresses := []string{
		"127.0.0.1:6381",
		"localhost:6382",
		"127.0.0.1:6383",
		"10.0.0.2:6379",
		"127.0.0.1:6378",
	}
	if !reflect.DeepEqual(targets.RedisAddresses, wantRedisAddresses) {
		t.Fatalf("redis addresses = %#v, want %#v", targets.RedisAddresses, wantRedisAddresses)
	}
	if targets.RedisAddress != "127.0.0.1:6381" {
		t.Fatalf("redis address = %q, want %q", targets.RedisAddress, "127.0.0.1:6381")
	}
	if targets.RedisPassword != "env-password" {
		t.Fatalf("redis password = %q, want %q", targets.RedisPassword, "env-password")
	}
	if targets.RedisDB != 7 {
		t.Fatalf("redis db = %d, want %d", targets.RedisDB, 7)
	}

	redisAddr, redisPassword, redisDB := integrationRedisTarget(t)
	if redisAddr != "127.0.0.1:6381" || redisPassword != "env-password" || redisDB != 7 {
		t.Fatalf(
			"integrationRedisTarget() = (%q, %q, %d), want (%q, %q, %d)",
			redisAddr,
			redisPassword,
			redisDB,
			"127.0.0.1:6381",
			"env-password",
			7,
		)
	}
}

func TestAddressNormalizationHelpers(t *testing.T) {
	addressCases := []struct {
		name string
		raw  string
		want string
	}{
		{name: "hostless port", raw: ":9000", want: "127.0.0.1:9000"},
		{name: "empty host via join", raw: net.JoinHostPort("", "6379"), want: "127.0.0.1:6379"},
		{name: "explicit host preserved", raw: "localhost:9100", want: "localhost:9100"},
		{name: "uri preserved", raw: "redis://cache:6379/0", want: "redis://cache:6379/0"},
		{name: "blank input", raw: "   ", want: ""},
	}

	for _, tc := range addressCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeAddress(tc.raw); got != tc.want {
				t.Fatalf("normalizeAddress(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}

	gotSplit := splitAndNormalizeAddresses(" :9000, localhost:9001, ,:9002 ")
	wantSplit := []string{"127.0.0.1:9000", "localhost:9001", "127.0.0.1:9002"}
	if !reflect.DeepEqual(gotSplit, wantSplit) {
		t.Fatalf("splitAndNormalizeAddresses() = %#v, want %#v", gotSplit, wantSplit)
	}

	gotUnique := uniqueAddresses([]string{":9000", "127.0.0.1:9000"}, splitAndNormalizeAddresses("localhost:9001,:9000"))
	wantUnique := []string{"127.0.0.1:9000", "localhost:9001"}
	if !reflect.DeepEqual(gotUnique, wantUnique) {
		t.Fatalf("uniqueAddresses() = %#v, want %#v", gotUnique, wantUnique)
	}
}

func clearIntegrationOverrideEnvs(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		integrationConfigOverrideEnv,
		integrationGRPCAddrOverrideEnv,
		integrationMySQLSourceOverrideEnv,
		integrationRedisAddrOverrideEnv,
		integrationRedisAddrsOverrideEnv,
		integrationRedisPasswordOverrideEnv,
		integrationRedisDBOverrideEnv,
	} {
		t.Setenv(key, "")
	}
}

func writeIntegrationConfigFixtureFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "bootstrap.local.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o600); err != nil {
		t.Fatalf("write config fixture %q: %v", path, err)
	}
	return path
}
