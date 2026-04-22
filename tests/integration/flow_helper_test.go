package integration

import (
	"context"

	"github.com/go-lynx/lynx-layout/internal/data/ent"
)

// flowStorageHelper wraps the storageFixture to provide simplified access for flow tests.
type flowStorageHelper struct {
	fixture *storageFixture
}

func newFlowStorageHelper(ctx context.Context) (*flowStorageHelper, error) {
	fix, err := newStorageFixture(ctx)
	if err != nil {
		return nil, err
	}
	return &flowStorageHelper{fixture: fix}, nil
}

func (h *flowStorageHelper) Close() error {
	if h == nil || h.fixture == nil {
		return nil
	}
	return h.fixture.Close()
}

func (h *flowStorageHelper) SeedUser(ctx context.Context, account, password string) (*ent.User, error) {
	return h.fixture.SeedMySQLUser(ctx, storageUserSeed{
		Account:  account,
		Password: password,
		Nickname: "FlowTestUser",
	})
}

func (h *flowStorageHelper) CleanupUser(ctx context.Context, account string) error {
	return h.fixture.CleanupMySQLUser(ctx, account)
}

func (h *flowStorageHelper) SeedCache(ctx context.Context, key, value string) error {
	return h.fixture.SeedRedisCache(ctx, storageCacheSeed{
		Key:   key,
		Value: value,
	})
}

func (h *flowStorageHelper) CleanupCache(ctx context.Context, key string) error {
	return h.fixture.CleanupRedisKeys(ctx, key)
}

func (h *flowStorageHelper) AssertCacheHit(ctx context.Context, key, want string) error {
	return h.fixture.AssertRedisCacheHit(ctx, key, want)
}
