package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	v1 "github.com/go-lynx/lynx-layout/api/login/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlowGRPCRouting verifies gRPC service registration and request routing using MockGRPCHarness.
func TestFlowGRPCRouting(t *testing.T) {
	// 1. Setup Mock gRPC Harness with custom handler to reflect request account
	svc := NewMockLoginService(WithMockLoginHandler(func(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
		return &v1.LoginReply{
			Token: "flow-token",
			User: &v1.UserInfo{
				Account:  req.Account,
				NickName: "Flow User",
			},
		}, nil
	}))
	harness := StartMockLoginGRPCServer(t, svc)
	defer harness.Close()

	ctx := suiteContext(t, 5*time.Second)

	// 2. Perform Login Request
	req := &v1.LoginRequest{
		Account:  "flow-routing-account",
		Password: "flow-routing-password",
	}
	
	reply, err := harness.Client.Login(ctx, req)
	require.NoError(t, err, "gRPC Login request should succeed through the routing harness")

	// 3. Verify Mock Interaction
	// This confirms the request reached the intended service and routing is active.
	assert.NotNil(t, reply, "Login reply should not be nil")
	assert.Equal(t, req.Account, reply.User.Account, "Mock service should return the requested account")
	
	harness.Service.RequireCallCount(t, 1)
	harness.Service.RequireLastLoginRequest(t, req)
}

// TestFlowCachePersistence verifies the Redis cache + MySQL persistence business chain using storageFixture.
func TestFlowCachePersistence(t *testing.T) {
	ctx := suiteContext(t, 10*time.Second)

	// 1. Setup Storage Helper (Wraps storageFixture)
	helper, err := newFlowStorageHelper(ctx)
	if err != nil {
		if isStorageFixtureSkipped(err) || isStorageDependencyUnavailable(err) || strings.Contains(err.Error(), "lynx not initialized") {
			t.Skipf("skipping TestFlowCachePersistence: %v", err)
			return
		}
		require.NoError(t, err, "Failed to initialize storage fixture")
	}
	defer helper.Close()

	testAccount := "flow-persistence-user-" + time.Now().Format("150405.000")
	testPassword := "flow-persistence-password"
	cacheKey := "flow:cache:user:" + testAccount
	cacheValue := "FlowUserNickname"

	// 2. Persistent Assertion: Seed MySQL User
	// This confirms MySQL persistence capability within the integrated flow.
	u, err := helper.SeedUser(ctx, testAccount, testPassword)
	require.NoError(t, err, "Failed to persist user in MySQL via storage fixture")
	assert.Equal(t, testAccount, u.Account, "Stored MySQL account should match")

	// 3. Cache Assertion: Seed and Assert Redis Cache
	// This confirms Redis cache capability within the integrated flow.
	err = helper.SeedCache(ctx, cacheKey, cacheValue)
	require.NoError(t, err, "Failed to seed Redis cache")

	err = helper.AssertCacheHit(ctx, cacheKey, cacheValue)
	require.NoError(t, err, "Redis cache hit assertion failed for seeded key")

	// 4. Cleanup Verification
	err = helper.CleanupCache(ctx, cacheKey)
	require.NoError(t, err, "Failed to cleanup Redis keys")

	err = helper.CleanupUser(ctx, testAccount)
	require.NoError(t, err, "Failed to cleanup MySQL user")
}
