package integration

import (
	"context"
	"database/sql"
	sqldriver "database/sql/driver"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	entdialect "entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	"github.com/go-lynx/lynx-layout/internal/data/ent/user"
	lynxmysqlpkg "github.com/go-lynx/lynx-mysql"
	lynxredis "github.com/go-lynx/lynx-redis"
	sqlinterfaces "github.com/go-lynx/lynx-sql-sdk/interfaces"
	"github.com/redis/go-redis/v9"
)

// storageDependencyUnavailableError 用于区分依赖不可用和真正的断言失败。
type storageDependencyUnavailableError struct {
	dependency string
	cause      error
}

func (e *storageDependencyUnavailableError) Error() string {
	return fmt.Sprintf("storage dependency unavailable: %s: %v", e.dependency, e.cause)
}

func (e *storageDependencyUnavailableError) Unwrap() error {
	return e.cause
}

func newStorageDependencyUnavailableError(dependency string, err error) error {
	if err == nil {
		err = errors.New("unknown dependency error")
	}

	return &storageDependencyUnavailableError{
		dependency: dependency,
		cause:      err,
	}
}

func isStorageDependencyUnavailable(err error) bool {
	var target *storageDependencyUnavailableError
	return errors.As(err, &target)
}

// storageFixtureSkipError 用于表达表结构或前置条件不足，调用方可转成 skip。
type storageFixtureSkipError struct {
	reason string
}

func (e *storageFixtureSkipError) Error() string {
	return fmt.Sprintf("storage fixture skipped: %s", e.reason)
}

func newStorageFixtureSkipError(format string, args ...any) error {
	return &storageFixtureSkipError{reason: fmt.Sprintf(format, args...)}
}

func isStorageFixtureSkipped(err error) bool {
	var target *storageFixtureSkipError
	return errors.As(err, &target)
}

// sharedPoolSafeEntDriver 避免测试夹具关闭插件维护的共享连接池。
type sharedPoolSafeEntDriver struct {
	*entsql.Driver
}

func (d sharedPoolSafeEntDriver) Close() error {
	return nil
}

var _ entdialect.Driver = sharedPoolSafeEntDriver{}

type storageFixture struct {
	mysqlProvider sqlinterfaces.DBProvider
	mysqlDB       *sql.DB
	mysqlDialect  string
	entClient     *ent.Client
	redisClient   redis.UniversalClient
}

type storageUserSeed struct {
	Account        string
	Password       string
	Nickname       string
	Num            string
	RegisterSource int32
	LastLoginAt    time.Time
}

type storageCacheSeed struct {
	Key   string
	Value string
	TTL   time.Duration
}

func newStorageFixture(ctx context.Context) (*storageFixture, error) {
	mysqlProvider, mysqlErr := resolveStorageMySQLProvider()
	if mysqlErr != nil {
		return nil, mysqlErr
	}

	db, err := mysqlProvider.DB(ctx)
	if err != nil {
		return nil, classifyMySQLError("acquire mysql DB", err)
	}
	if db == nil {
		return nil, newStorageDependencyUnavailableError("mysql", errors.New("provider.DB returned nil"))
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, classifyMySQLError("ping mysql", err)
	}

	dialect := normalizeStorageDialect(mysqlProvider.Dialect())
	if dialect == "" {
		return nil, newStorageFixtureSkipError("mysql provider dialect is empty")
	}
	if !isStorageMySQLDialect(dialect) {
		return nil, newStorageFixtureSkipError("mysql provider dialect %q is unsupported by storage fixture", mysqlProvider.Dialect())
	}

	redisClient, redisErr := resolveStorageRedisClient(ctx)
	if redisErr != nil {
		return nil, redisErr
	}

	return &storageFixture{
		mysqlProvider: mysqlProvider,
		mysqlDB:       db,
		mysqlDialect:  dialect,
		entClient:     ent.NewClient(ent.Driver(sharedPoolSafeEntDriver{Driver: entsql.OpenDB(dialect, db)})),
		redisClient:   redisClient,
	}, nil
}

func (f *storageFixture) Close() error {
	if f == nil || f.entClient == nil {
		return nil
	}

	return f.entClient.Close()
}

func (f *storageFixture) SeedMySQLUser(ctx context.Context, seed storageUserSeed) (*ent.User, error) {
	if err := f.validateMySQLFixture(ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(seed.Account) == "" {
		return nil, errors.New("seed mysql user: account is required")
	}
	if strings.TrimSpace(seed.Password) == "" {
		return nil, fmt.Errorf("seed mysql user %q: password is required", seed.Account)
	}

	normalizedSeed := seed.withDefaults()
	created, err := f.entClient.User.Create().
		SetAccount(normalizedSeed.Account).
		SetPassword(normalizedSeed.Password).
		SetNickname(normalizedSeed.Nickname).
		SetNum(normalizedSeed.Num).
		SetRegisterSource(normalizedSeed.RegisterSource).
		SetLastLoginAt(normalizedSeed.LastLoginAt).
		Save(ctx)
	if err != nil {
		return nil, classifyMySQLError(fmt.Sprintf("seed mysql user %q", normalizedSeed.Account), err)
	}

	return created, nil
}

func (f *storageFixture) CleanupMySQLUser(ctx context.Context, account string) error {
	if err := f.validateMySQLFixture(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(account) == "" {
		return errors.New("cleanup mysql user: account is required")
	}

	_, err := f.entClient.User.Delete().Where(user.Account(account)).Exec(ctx)
	if err != nil {
		return classifyMySQLError(fmt.Sprintf("cleanup mysql user %q", account), err)
	}

	return nil
}

func (f *storageFixture) SeedRedisCache(ctx context.Context, seed storageCacheSeed) error {
	if err := f.validateRedisFixture(); err != nil {
		return err
	}
	if strings.TrimSpace(seed.Key) == "" {
		return errors.New("seed redis cache: key is required")
	}

	ttl := seed.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	if err := f.redisClient.Set(ctx, seed.Key, seed.Value, ttl).Err(); err != nil {
		return classifyRedisError(fmt.Sprintf("seed redis cache %q", seed.Key), err)
	}

	return nil
}

func (f *storageFixture) CleanupRedisKeys(ctx context.Context, keys ...string) error {
	if err := f.validateRedisFixture(); err != nil {
		return err
	}

	filteredKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		if strings.TrimSpace(key) != "" {
			filteredKeys = append(filteredKeys, key)
		}
	}
	if len(filteredKeys) == 0 {
		return nil
	}

	if err := f.redisClient.Del(ctx, filteredKeys...).Err(); err != nil {
		return classifyRedisError(fmt.Sprintf("cleanup redis keys %v", filteredKeys), err)
	}

	return nil
}

func (f *storageFixture) AssertRedisCacheHit(ctx context.Context, key string, want string) error {
	if err := f.validateRedisFixture(); err != nil {
		return err
	}
	if strings.TrimSpace(key) == "" {
		return errors.New("assert redis cache hit: key is required")
	}

	got, err := f.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("assert redis cache hit %q: key not found", key)
		}
		return classifyRedisError(fmt.Sprintf("assert redis cache hit %q", key), err)
	}
	if got != want {
		return fmt.Errorf("assert redis cache hit %q: got %q want %q", key, got, want)
	}

	return nil
}

func (f *storageFixture) AssertRedisCacheMiss(ctx context.Context, key string) error {
	if err := f.validateRedisFixture(); err != nil {
		return err
	}
	if strings.TrimSpace(key) == "" {
		return errors.New("assert redis cache miss: key is required")
	}

	got, err := f.redisClient.Get(ctx, key).Result()
	if err == nil {
		return fmt.Errorf("assert redis cache miss %q: got unexpected value %q", key, got)
	}
	if errors.Is(err, redis.Nil) {
		return nil
	}

	return classifyRedisError(fmt.Sprintf("assert redis cache miss %q", key), err)
}

func (f *storageFixture) validateMySQLFixture(ctx context.Context) error {
	if f == nil || f.mysqlProvider == nil || f.mysqlDB == nil || f.entClient == nil {
		return newStorageDependencyUnavailableError("mysql", errors.New("mysql fixture not initialized"))
	}
	if f.mysqlDialect == "" {
		return newStorageFixtureSkipError("mysql fixture dialect is empty")
	}
	if !isStorageMySQLDialect(f.mysqlDialect) {
		return newStorageFixtureSkipError("mysql fixture dialect %q is unsupported", f.mysqlDialect)
	}

	if err := f.mysqlDB.PingContext(ctx); err != nil {
		return classifyMySQLError("re-ping mysql", err)
	}

	return f.validateMySQLUserTableSchema(ctx)
}

func (f *storageFixture) validateRedisFixture() error {
	if f == nil || f.redisClient == nil {
		return newStorageDependencyUnavailableError("redis", errors.New("redis fixture not initialized"))
	}

	return nil
}

func (f *storageFixture) validateMySQLUserTableSchema(ctx context.Context) error {
	databaseName, err := queryCurrentMySQLDatabase(ctx, f.mysqlDB)
	if err != nil {
		return classifyMySQLError("resolve current mysql database", err)
	}
	if databaseName == "" {
		return newStorageFixtureSkipError("mysql connection has no selected database; cannot validate %q schema", user.Table)
	}

	return ensureMySQLTableColumns(ctx, f.mysqlDB, databaseName, user.Table, user.Columns)
}

func (s storageUserSeed) withDefaults() storageUserSeed {
	now := time.Now()

	if strings.TrimSpace(s.Nickname) == "" {
		s.Nickname = "fixture-user"
	}
	if strings.TrimSpace(s.Num) == "" {
		s.Num = fmt.Sprintf("FIX-%d", now.UnixNano())
	}
	if s.RegisterSource == 0 {
		s.RegisterSource = 1
	}
	if s.LastLoginAt.IsZero() {
		s.LastLoginAt = now
	}

	return s
}

func resolveStorageMySQLProvider() (sqlinterfaces.DBProvider, error) {
	provider, err := storageGetMySQLProvider()
	if err != nil {
		return nil, newStorageDependencyUnavailableError("mysql", err)
	}
	if provider == nil {
		return nil, newStorageDependencyUnavailableError("mysql", errors.New("GetProvider() returned nil"))
	}

	return provider, nil
}

func resolveStorageRedisClient(ctx context.Context) (redis.UniversalClient, error) {
	client, err := storageGetRedisClient()
	if err != nil {
		return nil, newStorageDependencyUnavailableError("redis", err)
	}
	if client == nil {
		return nil, newStorageDependencyUnavailableError("redis", errors.New("GetUniversalRedis() returned nil"))
	}
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, classifyRedisError("ping redis", err)
	}

	return client, nil
}

func classifyMySQLError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if isStorageConnectivityError(err) {
		return newStorageDependencyUnavailableError("mysql", fmt.Errorf("%s: %w", operation, err))
	}

	return fmt.Errorf("%s: %w", operation, err)
}

func classifyRedisError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if isStorageConnectivityError(err) {
		return newStorageDependencyUnavailableError("redis", fmt.Errorf("%s: %w", operation, err))
	}

	return fmt.Errorf("%s: %w", operation, err)
}

func normalizeStorageDialect(dialect string) string {
	return strings.ToLower(strings.TrimSpace(dialect))
}

func isStorageMySQLDialect(dialect string) bool {
	return strings.HasPrefix(normalizeStorageDialect(dialect), "mysql")
}

func queryCurrentMySQLDatabase(ctx context.Context, db *sql.DB) (string, error) {
	var databaseName sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&databaseName); err != nil {
		return "", err
	}
	if !databaseName.Valid {
		return "", nil
	}

	return strings.TrimSpace(databaseName.String), nil
}

func ensureMySQLTableColumns(ctx context.Context, db *sql.DB, databaseName string, tableName string, requiredColumns []string) error {
	tableExists, err := mysqlTableExists(ctx, db, databaseName, tableName)
	if err != nil {
		return classifyMySQLError(fmt.Sprintf("inspect mysql table %q", tableName), err)
	}
	if !tableExists {
		return newStorageFixtureSkipError("mysql table %q not found in database %q", tableName, databaseName)
	}

	actualColumns, err := mysqlTableColumns(ctx, db, databaseName, tableName)
	if err != nil {
		return classifyMySQLError(fmt.Sprintf("inspect mysql columns for %q", tableName), err)
	}

	missingColumns := diffMissingColumns(requiredColumns, actualColumns)
	if len(missingColumns) > 0 {
		return newStorageFixtureSkipError(
			"mysql table %q in database %q is missing required columns: %s",
			tableName,
			databaseName,
			strings.Join(missingColumns, ", "),
		)
	}

	return nil
}

func mysqlTableExists(ctx context.Context, db *sql.DB, databaseName string, tableName string) (bool, error) {
	var count int
	err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?`,
		databaseName,
		tableName,
	).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func mysqlTableColumns(ctx context.Context, db *sql.DB, databaseName string, tableName string) (map[string]struct{}, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT column_name FROM information_schema.columns WHERE table_schema = ? AND table_name = ?`,
		databaseName,
		tableName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make(map[string]struct{})
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		columns[columnName] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

func diffMissingColumns(requiredColumns []string, actualColumns map[string]struct{}) []string {
	missingColumns := make([]string, 0)
	for _, columnName := range requiredColumns {
		if _, ok := actualColumns[columnName]; !ok {
			missingColumns = append(missingColumns, columnName)
		}
	}

	return missingColumns
}

func storageGetRedisClient() (_ redis.UniversalClient, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("GetUniversalRedis panic: %v", recovered)
		}
	}()

	return lynxredis.GetUniversalRedis(), nil
}

func storageGetMySQLProvider() (_ sqlinterfaces.DBProvider, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("GetProvider panic: %v", recovered)
		}
	}()

	return lynxmysqlpkg.GetProvider(), nil
}

func isStorageConnectivityError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	if errors.Is(err, sql.ErrConnDone) || errors.Is(err, sqldriver.ErrBadConn) {
		return true
	}

	var networkError net.Error
	if errors.As(err, &networkError) {
		return true
	}

	message := strings.ToLower(err.Error())
	for _, fragment := range []string{
		"broken pipe",
		"closed network connection",
		"connection refused",
		"connection reset",
		"connection timed out",
		"eof",
		"i/o timeout",
		"no such host",
		"server has gone away",
	} {
		if strings.Contains(message, fragment) {
			return true
		}
	}

	return false
}

func TestStorageFixture_HelperBoundaries(t *testing.T) {
	ctx := context.Background()
	fixture := &storageFixture{}

	if _, err := fixture.SeedMySQLUser(ctx, storageUserSeed{Account: "demo", Password: "secret"}); !isStorageDependencyUnavailable(err) {
		t.Fatalf("SeedMySQLUser should report dependency unavailable, got %v", err)
	}

	if err := fixture.CleanupMySQLUser(ctx, "demo"); !isStorageDependencyUnavailable(err) {
		t.Fatalf("CleanupMySQLUser should report dependency unavailable, got %v", err)
	}

	if err := fixture.SeedRedisCache(ctx, storageCacheSeed{Key: "k", Value: "v"}); !isStorageDependencyUnavailable(err) {
		t.Fatalf("SeedRedisCache should report dependency unavailable, got %v", err)
	}

	if err := fixture.CleanupRedisKeys(ctx, "k"); !isStorageDependencyUnavailable(err) {
		t.Fatalf("CleanupRedisKeys should report dependency unavailable, got %v", err)
	}

	if err := fixture.AssertRedisCacheHit(ctx, "k", "v"); !isStorageDependencyUnavailable(err) {
		t.Fatalf("AssertRedisCacheHit should report dependency unavailable, got %v", err)
	}

	if err := fixture.AssertRedisCacheMiss(ctx, "k"); !isStorageDependencyUnavailable(err) {
		t.Fatalf("AssertRedisCacheMiss should report dependency unavailable, got %v", err)
	}
}

func TestStorageFixture_DefaultsAndMarkers(t *testing.T) {
	seed := storageUserSeed{
		Account:  "demo",
		Password: "secret",
	}

	normalized := seed.withDefaults()
	if normalized.Nickname == "" {
		t.Fatal("withDefaults should populate nickname")
	}
	if normalized.Num == "" {
		t.Fatal("withDefaults should populate num")
	}
	if normalized.RegisterSource == 0 {
		t.Fatal("withDefaults should populate register source")
	}
	if normalized.LastLoginAt.IsZero() {
		t.Fatal("withDefaults should populate last login time")
	}

	if !isStorageDependencyUnavailable(newStorageDependencyUnavailableError("redis", errors.New("down"))) {
		t.Fatal("dependency marker should be detectable")
	}
	if !isStorageFixtureSkipped(newStorageFixtureSkipError("schema missing")) {
		t.Fatal("skip marker should be detectable")
	}
}
