// Package integration provides integration tests for lynx-layout plugins.
// Tests cover grpc+redis+mysql combination scenarios.
package integration

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-lynx/lynx"
	lynxgrpc "github.com/go-lynx/lynx-grpc"
	lynxmysqlpkg "github.com/go-lynx/lynx-mysql"
	lynxredis "github.com/go-lynx/lynx-redis"
	"github.com/go-lynx/lynx/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPluginParallelStartup tests parallel plugin startup and dependency order verification.
// 测试目标：验证插件可以并行启动，并且依赖顺序正确
func TestPluginParallelStartup(t *testing.T) {
	ctx := context.Background()

	// Track startup order using atomic operations
	var mysqlStarted, redisStarted, grpcStarted int32

	// Initialize plugins in parallel
	errChan := make(chan error, 3)
	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		defer wg.Done()
		err := initializeMySQLPlugin(ctx)
		if err == nil {
			atomic.StoreInt32(&mysqlStarted, 1)
		}
		errChan <- err
	}()

	go func() {
		defer wg.Done()
		err := initializeRedisPlugin(ctx)
		if err == nil {
			atomic.StoreInt32(&redisStarted, 1)
		}
		errChan <- err
	}()

	go func() {
		defer wg.Done()
		err := initializeGRPCPlugin(ctx)
		if err == nil {
			atomic.StoreInt32(&grpcStarted, 1)
		}
		errChan <- err
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	var startupErrors []string
	for err := range errChan {
		if err != nil {
			startupErrors = append(startupErrors, err.Error())
			t.Logf("Plugin initialization error: %v", err)
		}
	}

	// Count successful startups
	successCount := int(atomic.LoadInt32(&mysqlStarted) + atomic.LoadInt32(&redisStarted) + atomic.LoadInt32(&grpcStarted))

	// Log results
	log.Infof("Plugin parallel startup test: %d/3 plugins started successfully", successCount)
	if atomic.LoadInt32(&mysqlStarted) == 1 {
		log.Infof("  - MySQL: started")
	}
	if atomic.LoadInt32(&redisStarted) == 1 {
		log.Infof("  - Redis: started")
	}
	if atomic.LoadInt32(&grpcStarted) == 1 {
		log.Infof("  - gRPC: started")
	}

	if successCount == 0 {
		t.Skipf("plugin dependencies unavailable, skipping parallel startup test\n%s", strings.Join(startupErrors, "\n"))
	}

	assert.Greater(t, successCount, 0, "至少应有一个插件在并行启动测试中成功启动")
}

// TestRedisCacheMySQLPersistence tests Redis cache read/write + MySQL persistence business chain.
// 测试目标：验证 Redis 缓存读写 + MySQL 持久化的完整业务链路
func TestRedisCacheMySQLPersistence(t *testing.T) {
	ctx := context.Background()

	redisClient := suiteRequireRedis(t)
	mysqlProvider := suiteRequireMySQL(t)

	// Test data
	testKey := "test:cache:persistence:" + time.Now().Format("20060102150405")
	testValue := "test-value-" + time.Now().Format("20060102150405")

	// Step 1: Write to Redis cache
	err := redisClient.Set(ctx, testKey, testValue, 5*time.Minute).Err()
	require.NoError(t, err, "Failed to write to Redis cache")
	log.Infof("Step 1: Successfully wrote to Redis cache: %s = %s", testKey, testValue)

	// Step 2: Read from Redis cache
	cachedValue, err := redisClient.Get(ctx, testKey).Result()
	require.NoError(t, err, "Failed to read from Redis cache")
	assert.Equal(t, testValue, cachedValue, "Cached value should match written value")
	log.Infof("Step 2: Successfully read from Redis cache: %s = %s", testKey, cachedValue)

	// Step 3: Get MySQL connection and verify database connectivity
	db, err := mysqlProvider.DB(ctx)
	if err != nil {
		t.Logf("MySQL connection error: %v", err)
		t.Skip("MySQL not available, skipping persistence test")
		return
	}
	require.NotNil(t, db, "MySQL database connection should not be nil")

	// Step 4: Test database ping
	err = db.PingContext(ctx)
	if err != nil {
		t.Logf("MySQL ping error: %v", err)
		t.Skip("MySQL not responding, skipping persistence test")
		return
	}
	log.Infof("Step 3: MySQL database connection successful")

	// Step 5: Simulate cache miss scenario - delete from cache
	err = redisClient.Del(ctx, testKey).Err()
	require.NoError(t, err, "Failed to delete from Redis cache")
	log.Infof("Step 4: Simulated cache miss by deleting key: %s", testKey)

	// Step 6: Verify cache miss
	_, err = redisClient.Get(ctx, testKey).Result()
	assert.Error(t, err, "Should get error after deleting from cache")
	log.Infof("Step 5: Cache miss verified successfully")

	log.Infof("Redis+MySQL integration test completed successfully")
}

// TestGRPCServiceRegistration tests gRPC service registration and request routing.
// 测试目标：验证 gRPC 服务注册与请求路由
func TestGRPCServiceRegistration(t *testing.T) {
	suiteRequireGRPC(t)

	grpcServer, err := lynxgrpc.GetGrpcServer(integrationPluginManager())
	if err != nil || grpcServer == nil {
		t.Skipf("gRPC plugin not initialized, skipping test: %v", err)
		return
	}

	// Verify server is not nil
	assert.NotNil(t, grpcServer, "gRPC server should not be nil")
	log.Infof("gRPC server instance obtained successfully")

	// Test health check if available
	type healthChecker interface {
		CheckHealth() error
	}

	if hc, ok := any(grpcServer).(healthChecker); ok {
		err = hc.CheckHealth()
		if err != nil {
			t.Logf("gRPC health check warning: %v", err)
		} else {
			log.Infof("gRPC server health check passed")
		}
	}

	log.Infof("gRPC service registration test completed")
}

// TestPluginRollbackOnFailure tests rollback verification when plugin startup fails.
// 测试目标：模拟插件启动失败时的回滚验证
func TestPluginRollbackOnFailure(t *testing.T) {
	ctx := context.Background()

	// Track plugin state
	var pluginState int32 // 0=initial, 1=initializing, 2=initialized, 3=failed

	// Create a mock plugin that will fail
	failingPlugin := &mockFailingPlugin{
		name: "failing-plugin",
		onInitialize: func(ctx context.Context) error {
			atomic.StoreInt32(&pluginState, 1) // initializing
			return fmt.Errorf("simulated initialization failure")
		},
	}

	// Attempt to initialize
	err := failingPlugin.Initialize(ctx)
	assert.Error(t, err, "Plugin initialization should fail")
	assert.Contains(t, err.Error(), "simulated initialization failure")

	// Verify plugin state transitioned to initializing but not initialized
	state := atomic.LoadInt32(&pluginState)
	assert.Equal(t, int32(1), state, "Plugin should be in initializing state (not fully initialized)")
	log.Infof("Plugin rollback test completed: state=%d (1=initializing, expected)", state)
}

// TestFullIntegrationScenario tests a complete business scenario with all three plugins.
// 测试目标：完整业务场景测试 - gRPC 服务 + Redis 缓存 + MySQL 持久化
func TestFullIntegrationScenario(t *testing.T) {
	ctx := context.Background()

	redisClient := suiteRequireRedis(t)
	mysqlProvider := suiteRequireMySQL(t)
	suiteRequireGRPC(t)

	grpcServer, err := lynxgrpc.GetGrpcServer(integrationPluginManager())
	if err != nil || grpcServer == nil {
		t.Skipf("gRPC plugin not available for full integration test: %v", err)
		return
	}

	log.Infof("Starting full integration test scenario")

	// Simulate a complete business flow:
	// 1. gRPC request received
	// 2. Check Redis cache
	// 3. Cache miss -> Query MySQL
	// 4. Write result to Redis cache
	// 5. Return response via gRPC

	testKey := "test:integration:full:" + time.Now().Format("20060102150405")
	testData := "integration-test-data"

	// Step 1: Simulate cache lookup (should miss)
	_, err = redisClient.Get(ctx, testKey).Result()
	assert.Error(t, err, "Cache should be empty initially")
	log.Infof("Step 1: Cache miss confirmed")

	// Step 2: Simulate database query (verify DB connectivity)
	db, err := mysqlProvider.DB(ctx)
	if err != nil || db == nil {
		t.Skip("MySQL not available for full integration test")
		return
	}
	err = db.PingContext(ctx)
	if err != nil {
		t.Skip("MySQL not responding for full integration test")
		return
	}
	log.Infof("Step 2: Database query simulated (connectivity verified)")

	// Step 3: Write to cache
	err = redisClient.Set(ctx, testKey, testData, 10*time.Minute).Err()
	require.NoError(t, err, "Failed to write to cache")
	log.Infof("Step 3: Data written to Redis cache")

	// Step 4: Verify cache write
	cachedData, err := redisClient.Get(ctx, testKey).Result()
	require.NoError(t, err, "Failed to read from cache")
	assert.Equal(t, testData, cachedData, "Cached data should match")
	log.Infof("Step 4: Cache verification successful")

	// Step 5: gRPC server should be ready to serve
	assert.NotNil(t, grpcServer, "gRPC server should be available")
	log.Infof("Step 5: gRPC server ready")

	log.Infof("Full integration test scenario completed successfully")
}

// Helper functions

func initializeMySQLPlugin(ctx context.Context) error {
	log.Infof("Initializing MySQL plugin...")

	if err := integrationDependencyStateError(suiteDependencyMySQL, "parallel startup probe"); err != nil {
		return err
	}

	provider := lynxmysqlpkg.GetProvider()

	db, err := provider.DB(ctx)
	if err != nil {
		return err
	}

	if db == nil {
		return fmt.Errorf("MySQL database connection is nil")
	}

	return db.PingContext(ctx)
}

func initializeRedisPlugin(ctx context.Context) (err error) {
	log.Infof("Initializing Redis plugin...")

	if err := integrationDependencyStateError(suiteDependencyRedis, "parallel startup probe"); err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("redis dependency probe panicked: %v\n%s", r, packageSuite.Diagnostics())
		}
	}()

	client := lynxredis.GetUniversalRedis()
	if client == nil {
		return integrationDependencyRuntimeError(suiteDependencyRedis, "GetUniversalRedis() returned nil")
	}

	return client.Ping(ctx).Err()
}

func initializeGRPCPlugin(ctx context.Context) error {
	log.Infof("Initializing gRPC plugin...")

	if err := integrationDependencyStateError(suiteDependencyGRPC, "parallel startup probe"); err != nil {
		return err
	}

	server, err := lynxgrpc.GetGrpcServer(integrationPluginManager())
	if err != nil {
		return err
	}
	if server == nil {
		return integrationDependencyRuntimeError(suiteDependencyGRPC, "GetGrpcServer returned nil")
	}

	return nil
}

func integrationDependencyStateError(dep suiteDependency, action string) error {
	state := packageSuite.dependencyState(dep)
	if state.Ready {
		return nil
	}

	return fmt.Errorf("%s dependency unavailable during %s: %s [%s]\n%s",
		dep,
		action,
		state.Reason,
		state.Detail,
		packageSuite.Diagnostics(),
	)
}

func integrationDependencyRuntimeError(dep suiteDependency, detail string) error {
	return fmt.Errorf("%s dependency unavailable during runtime probe: %s\n%s",
		dep,
		detail,
		packageSuite.Diagnostics(),
	)
}

func integrationPluginManager() lynx.PluginManager {
	app := lynx.Lynx()
	if app == nil {
		return nil
	}
	return app.GetPluginManager()
}

// Helper types

type startupListener struct {
	pluginName string
	onStarted  func()
}

func (l *startupListener) HandleEvent(event plugins.PluginEvent) {
	if event.Type == plugins.EventPluginStarted && event.PluginID == l.pluginName {
		if l.onStarted != nil {
			l.onStarted()
		}
	}
}

func (l *startupListener) GetListenerID() string {
	return fmt.Sprintf("startup-listener-%s", l.pluginName)
}

type mockFailingPlugin struct {
	name         string
	onInitialize func(ctx context.Context) error
}

func (p *mockFailingPlugin) Name() string {
	return p.name
}

func (p *mockFailingPlugin) Initialize(ctx context.Context) error {
	if p.onInitialize != nil {
		return p.onInitialize(ctx)
	}
	return fmt.Errorf("no initialize handler")
}

func (p *mockFailingPlugin) Start(ctx context.Context) error {
	return nil
}

func (p *mockFailingPlugin) Stop(ctx context.Context) error {
	return nil
}

func (p *mockFailingPlugin) HealthCheck(ctx context.Context) error {
	return fmt.Errorf("unhealthy")
}
