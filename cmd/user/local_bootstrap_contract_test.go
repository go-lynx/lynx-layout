package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type localBootstrapContractConfig struct {
	Lynx struct {
		Application struct {
			Name    string `yaml:"name"`
			Version string `yaml:"version"`
		} `yaml:"application"`
		Redis struct {
			Password string   `yaml:"password"`
			Addrs    []string `yaml:"addrs"`
		} `yaml:"redis"`
	} `yaml:"lynx"`
}

type localBootstrapComposeConfig struct {
	Services struct {
		Redis struct {
			Command []string `yaml:"command"`
		} `yaml:"redis"`
	} `yaml:"services"`
}

func TestLocalBootstrapMatchesComposeAndFixture(t *testing.T) {
	moduleRoot := userModuleRoot(t)

	bootstrapPath := filepath.Join(moduleRoot, "configs", "bootstrap.local.yaml")
	fixturePath := filepath.Join(moduleRoot, "tests", "integration", "testdata", "bootstrap.local.yaml")
	composePath := filepath.Join(moduleRoot, "deployments", "docker-compose.local.yml")
	readmePath := filepath.Join(moduleRoot, "README.md")

	bootstrap := loadYAMLFile[localBootstrapContractConfig](t, bootstrapPath)
	fixture := loadYAMLFile[localBootstrapContractConfig](t, fixturePath)
	compose := loadYAMLFile[localBootstrapComposeConfig](t, composePath)

	if bootstrap.Lynx.Application.Name == "" || bootstrap.Lynx.Application.Version == "" {
		t.Fatalf("bootstrap.local.yaml must define lynx.application.name/version")
	}
	if strings.TrimSpace(bootstrap.Lynx.Redis.Password) != "" {
		t.Fatalf("bootstrap.local.yaml redis password must stay empty for the default local compose path")
	}
	if strings.TrimSpace(fixture.Lynx.Redis.Password) != "" {
		t.Fatalf("integration fixture redis password must mirror bootstrap.local.yaml for the default local compose path")
	}
	if strings.Join(bootstrap.Lynx.Redis.Addrs, ",") != strings.Join(fixture.Lynx.Redis.Addrs, ",") {
		t.Fatalf("integration fixture redis addrs must mirror bootstrap.local.yaml")
	}
	if composeRedisRequiresPassword(compose.Services.Redis.Command) {
		t.Fatal("docker-compose.local.yml now requires a redis password; update bootstrap/readme contract")
	}

	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	readme := string(readmeContent)
	for _, expected := range []string{
		"boot.NewApplication(wireApp).Run()",
		"cmd/user/plugins.go",
		"cmd/user/providers.go",
		"go run ./cmd/user -conf ./configs/bootstrap.local.yaml",
		"无密码的 `redis://127.0.0.1:6379`",
	} {
		if !strings.Contains(readme, expected) {
			t.Fatalf("README is missing expected local bootstrap contract fragment %q", expected)
		}
	}
}

func userModuleRoot(tb testing.TB) string {
	tb.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("runtime.Caller unavailable")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func loadYAMLFile[T any](tb testing.TB, path string) T {
	tb.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read %s: %v", path, err)
	}

	var value T
	if err := yaml.Unmarshal(content, &value); err != nil {
		tb.Fatalf("unmarshal %s: %v", path, err)
	}
	return value
}

func composeRedisRequiresPassword(command []string) bool {
	for _, part := range command {
		normalized := strings.TrimSpace(strings.ToLower(part))
		if normalized == "" {
			continue
		}
		if normalized == "requirepass" || normalized == "--requirepass" || strings.HasPrefix(normalized, "requirepass ") || strings.HasPrefix(normalized, "--requirepass=") {
			return true
		}
	}
	return false
}
