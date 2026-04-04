package data

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-lynx/lynx-layout/internal/data/ent"
	lynxredis "github.com/go-lynx/lynx-redis"
	"github.com/go-lynx/lynx-sql-sdk/interfaces"
	"github.com/redis/go-redis/v9"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

// staticDBProvider wraps an already-opened *sql.DB as an interfaces.DBProvider.
type staticDBProvider struct {
	db      *sql.DB
	dialect string
}

func (p *staticDBProvider) DB(_ context.Context) (*sql.DB, error) { return p.db, nil }
func (p *staticDBProvider) ValidatedConn(ctx context.Context) (*sql.Conn, error) {
	return p.db.Conn(ctx)
}
func (p *staticDBProvider) Dialect() string { return p.dialect }

var _ interfaces.DBProvider = (*staticDBProvider)(nil)

type staticRedisProvider struct {
	client redis.UniversalClient
}

func (p staticRedisProvider) UniversalClient(context.Context) (redis.UniversalClient, error) {
	return p.client, nil
}

func (p staticRedisProvider) SingleClient(context.Context) (*redis.Client, error) {
	client, ok := p.client.(*redis.Client)
	if !ok {
		return nil, fmt.Errorf("redis client is not standalone")
	}
	return client, nil
}

func (p staticRedisProvider) Mode(context.Context) (string, error) {
	return "standalone", nil
}

var _ lynxredis.Provider = staticRedisProvider{}

// isMySQLAvailable returns true when a MySQL server is reachable on localhost:3306.
func isMySQLAvailable() bool {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/lynx_test?charset=utf8mb4&parseTime=True")
	if err != nil {
		return false
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return db.PingContext(ctx) == nil
}

// isRedisAvailable returns true when a Redis server is reachable on localhost:6379.
func isRedisAvailable() bool {
	c := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return c.Ping(ctx).Err() == nil
}

// openTestMySQL opens a pooled MySQL connection to the shared test database.
func openTestMySQL(t *testing.T) *sql.DB {
	t.Helper()
	const dsn = "root:@tcp(localhost:3306)/lynx_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("sql.Open mysql: %v", err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Second)
	t.Cleanup(func() { db.Close() })
	return db
}

// openTestEntClient opens an ent.Client directly against the MySQL test DSN and
// runs auto-migration so all tables exist before seeding.
func openTestEntClient(t *testing.T) *ent.Client {
	t.Helper()
	const dsn = "root:@tcp(localhost:3306)/lynx_test?charset=utf8mb4&parseTime=True&loc=Local"
	c, err := ent.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("ent.Open: %v", err)
	}
	if err := c.Schema.Create(context.Background()); err != nil {
		c.Close()
		t.Fatalf("ent schema create: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

// newTestRedisClient returns a Redis client pointing at localhost:6379.
func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	c := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	t.Cleanup(func() { c.Close() })
	return c
}

// ─── tests ────────────────────────────────────────────────────────────────────

// TestCrossPlugin_DataLayerInitializesWithBothPlugins verifies that NewData
// succeeds when a MySQL provider and a Redis provider facade are both supplied,
// which exercises the MySQL-plugin + Redis-plugin wiring path.
func TestCrossPlugin_DataLayerInitializesWithBothPlugins(t *testing.T) {
	if !isMySQLAvailable() {
		t.Skip("MySQL not available")
	}
	if !isRedisAvailable() {
		t.Skip("Redis not available")
	}

	db := openTestMySQL(t)
	rdb := newTestRedisClient(t)

	provider := &staticDBProvider{db: db, dialect: "mysql"}
	d, err := NewData(NewEntClientProviderFromDB(provider), staticRedisProvider{client: rdb})
	if err != nil {
		t.Fatalf("NewData (MySQL+Redis): %v", err)
	}
	if d == nil {
		t.Fatal("NewData returned nil")
	}
}

// TestCrossPlugin_MySQLWriteVisibleViaRedis seeds a user in MySQL through the
// loginRepo, then caches the result in Redis and reads it back.  This confirms
// that both plugins can round-trip data reliably in the same request flow.
func TestCrossPlugin_MySQLWriteVisibleViaRedis(t *testing.T) {
	if !isMySQLAvailable() {
		t.Skip("MySQL not available")
	}
	if !isRedisAvailable() {
		t.Skip("Redis not available")
	}

	ctx := context.Background()
	db := openTestMySQL(t)
	rdb := newTestRedisClient(t)
	entClient := openTestEntClient(t)

	testAccount := fmt.Sprintf("integ_%d", time.Now().UnixNano())
	created, err := entClient.User.Create().
		SetAccount(testAccount).
		SetPassword("hashed_pw").
		SetNickname("Integration Tester").
		SetAvatar("").
		SetNum("IT-1").
		SetStats(1).
		Save(ctx)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { entClient.User.DeleteOneID(created.ID).Exec(ctx) }) //nolint:errcheck

	provider := &staticDBProvider{db: db, dialect: "mysql"}
	d, err := NewData(NewEntClientProviderFromDB(provider), staticRedisProvider{client: rdb})
	if err != nil {
		t.Fatalf("NewData: %v", err)
	}

	repo := NewLoginRepo(d, nil).(*loginRepo)

	// Phase 1: read user from MySQL.
	user, err := repo.FindUserByAccount(ctx, testAccount)
	if err != nil {
		t.Fatalf("FindUserByAccount: %v", err)
	}
	if user.Account != testAccount {
		t.Fatalf("account mismatch: got %q want %q", user.Account, testAccount)
	}

	// Phase 2: write user-id to Redis (simulates cache-aside population).
	cacheKey := "user:account:" + testAccount
	if err := rdb.Set(ctx, cacheKey, fmt.Sprintf("%d", user.Id), time.Minute).Err(); err != nil {
		t.Fatalf("Redis SET: %v", err)
	}

	// Phase 3: read back from Redis (cache-hit path).
	cachedID, err := rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		t.Fatalf("Redis GET: %v", err)
	}
	if cachedID != fmt.Sprintf("%d", user.Id) {
		t.Fatalf("cached id mismatch: got %q want %d", cachedID, user.Id)
	}

	rdb.Del(ctx, cacheKey)
}

// TestCrossPlugin_UpdateUserLastLoginTime verifies that MySQL writes from
// loginRepo succeed while the Redis plugin is simultaneously active.
func TestCrossPlugin_UpdateUserLastLoginTime(t *testing.T) {
	if !isMySQLAvailable() {
		t.Skip("MySQL not available")
	}
	if !isRedisAvailable() {
		t.Skip("Redis not available")
	}

	ctx := context.Background()
	db := openTestMySQL(t)
	rdb := newTestRedisClient(t)
	entClient := openTestEntClient(t)

	testAccount := fmt.Sprintf("update_%d", time.Now().UnixNano())
	created, err := entClient.User.Create().
		SetAccount(testAccount).
		SetPassword("pw").
		SetNickname("Updater").
		SetAvatar("").
		SetNum("U-2").
		SetStats(0).
		Save(ctx)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { entClient.User.DeleteOneID(created.ID).Exec(ctx) }) //nolint:errcheck

	provider := &staticDBProvider{db: db, dialect: "mysql"}
	d, err := NewData(NewEntClientProviderFromDB(provider), staticRedisProvider{client: rdb})
	if err != nil {
		t.Fatalf("NewData: %v", err)
	}

	repo := NewLoginRepo(d, nil).(*loginRepo)
	user, err := repo.FindUserByAccount(ctx, testAccount)
	if err != nil {
		t.Fatalf("FindUserByAccount: %v", err)
	}

	if err := repo.UpdateUserLastLoginTime(ctx, user); err != nil {
		t.Fatalf("UpdateUserLastLoginTime: %v", err)
	}
}

// TestCrossPlugin_ConcurrentMySQLAndRedisOperations runs concurrent reads
// against MySQL and writes/reads against Redis to validate that both plugins
// remain stable under parallel load when used together.
func TestCrossPlugin_ConcurrentMySQLAndRedisOperations(t *testing.T) {
	if !isMySQLAvailable() {
		t.Skip("MySQL not available")
	}
	if !isRedisAvailable() {
		t.Skip("Redis not available")
	}

	const workers = 5
	ctx := context.Background()
	db := openTestMySQL(t)
	rdb := newTestRedisClient(t)
	entClient := openTestEntClient(t)

	testAccount := fmt.Sprintf("concurrent_%d", time.Now().UnixNano())
	created, err := entClient.User.Create().
		SetAccount(testAccount).
		SetPassword("pw").
		SetNickname("Concurrent User").
		SetAvatar("").
		SetNum("C-1").
		SetStats(0).
		Save(ctx)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() {
		entClient.User.DeleteOneID(created.ID).Exec(ctx) //nolint:errcheck
		for i := 0; i < workers; i++ {
			rdb.Del(ctx, fmt.Sprintf("concurrent_key_%d", i))
		}
	})

	provider := &staticDBProvider{db: db, dialect: "mysql"}
	d, err := NewData(NewEntClientProviderFromDB(provider), staticRedisProvider{client: rdb})
	if err != nil {
		t.Fatalf("NewData: %v", err)
	}
	repo := NewLoginRepo(d, nil).(*loginRepo)

	var wg sync.WaitGroup
	errs := make([]error, workers)

	for i := 0; i < workers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			user, err := repo.FindUserByAccount(ctx, testAccount)
			if err != nil {
				errs[i] = fmt.Errorf("worker %d FindUserByAccount: %w", i, err)
				return
			}
			key := fmt.Sprintf("concurrent_key_%d", i)
			val := fmt.Sprintf("worker_%d_user_%d", i, user.Id)
			if err := rdb.Set(ctx, key, val, 30*time.Second).Err(); err != nil {
				errs[i] = fmt.Errorf("worker %d Redis SET: %w", i, err)
				return
			}
			got, err := rdb.Get(ctx, key).Result()
			if err != nil {
				errs[i] = fmt.Errorf("worker %d Redis GET: %w", i, err)
				return
			}
			if got != val {
				errs[i] = fmt.Errorf("worker %d value mismatch: got %q want %q", i, got, val)
			}
		}()
	}

	wg.Wait()
	for _, e := range errs {
		if e != nil {
			t.Error(e)
		}
	}
}

// TestCrossPlugin_RedisUnavailableDoesNotBlockMySQLInit verifies that the data
// layer initialises and MySQL operations still succeed when Redis is absent (nil
// client).  This ensures MySQL and Redis plugins remain independently resilient.
func TestCrossPlugin_RedisUnavailableDoesNotBlockMySQLInit(t *testing.T) {
	if !isMySQLAvailable() {
		t.Skip("MySQL not available")
	}

	ctx := context.Background()
	db := openTestMySQL(t)
	entClient := openTestEntClient(t)

	testAccount := fmt.Sprintf("no_redis_%d", time.Now().UnixNano())
	created, err := entClient.User.Create().
		SetAccount(testAccount).
		SetPassword("pw").
		SetNickname("NoRedis").
		SetAvatar("").
		SetNum("NR-1").
		SetStats(0).
		Save(ctx)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { entClient.User.DeleteOneID(created.ID).Exec(ctx) }) //nolint:errcheck

	// Pass nil Redis provider facade — NewData must not panic or fail.
	provider := &staticDBProvider{db: db, dialect: "mysql"}
	d, err := NewData(NewEntClientProviderFromDB(provider), nil)
	if err != nil {
		t.Fatalf("NewData with nil Redis: %v", err)
	}

	repo := NewLoginRepo(d, nil).(*loginRepo)
	user, err := repo.FindUserByAccount(ctx, testAccount)
	if err != nil {
		t.Fatalf("FindUserByAccount without Redis: %v", err)
	}
	if user.Account != testAccount {
		t.Fatalf("account mismatch: got %q want %q", user.Account, testAccount)
	}
}
