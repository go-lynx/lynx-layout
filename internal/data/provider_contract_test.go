package data

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"reflect"
	"testing"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	lynxredis "github.com/go-lynx/lynx-redis"
	"github.com/go-lynx/lynx-sql-sdk/interfaces"
	"github.com/redis/go-redis/v9"
)

type countingDBProvider struct {
	db      *sql.DB
	dialect string
	calls   int
}

func (p *countingDBProvider) DB(context.Context) (*sql.DB, error) {
	p.calls++
	return p.db, nil
}

func (p *countingDBProvider) ValidatedConn(ctx context.Context) (*sql.Conn, error) {
	p.calls++
	return p.db.Conn(ctx)
}

func (p *countingDBProvider) Dialect() string { return p.dialect }

var _ interfaces.DBProvider = (*countingDBProvider)(nil)

type noopConnector struct{}

func (noopConnector) Connect(context.Context) (driver.Conn, error) {
	return noopConn{}, nil
}

func (noopConnector) Driver() driver.Driver {
	return noopDriver{}
}

type noopDriver struct{}

func (noopDriver) Open(string) (driver.Conn, error) {
	return noopConn{}, nil
}

type noopConn struct{}

func (noopConn) Prepare(string) (driver.Stmt, error) { return noopStmt{}, nil }
func (noopConn) Close() error                        { return nil }
func (noopConn) Begin() (driver.Tx, error)           { return noopTx{}, nil }

type noopStmt struct{}

func (noopStmt) Close() error                               { return nil }
func (noopStmt) NumInput() int                              { return -1 }
func (noopStmt) Exec([]driver.Value) (driver.Result, error) { return noopResult(0), nil }
func (noopStmt) Query([]driver.Value) (driver.Rows, error)  { return noopRows{}, nil }

type noopTx struct{}

func (noopTx) Commit() error   { return nil }
func (noopTx) Rollback() error { return nil }

type noopResult int64

func (r noopResult) LastInsertId() (int64, error) { return int64(r), nil }
func (r noopResult) RowsAffected() (int64, error) { return int64(r), nil }

type noopRows struct{}

func (noopRows) Columns() []string         { return nil }
func (noopRows) Close() error              { return nil }
func (noopRows) Next([]driver.Value) error { return io.EOF }

func newNoopSQLDB() *sql.DB {
	return sql.OpenDB(noopConnector{})
}

func newNoopEntClientProvider() EntClientProvider {
	return func() (*ent.Client, error) {
		return ent.NewClient(ent.Driver(entsql.OpenDB("mysql", newNoopSQLDB()))), nil
	}
}

func TestNewEntDriverProvider_ResolvesDBProviderEachCall(t *testing.T) {
	db := newNoopSQLDB()
	defer db.Close()

	provider := &countingDBProvider{
		db:      db,
		dialect: "mysql",
	}

	driverProvider := NewEntDriverProvider(provider)
	first, err := driverProvider(context.Background())
	if err != nil {
		t.Fatalf("first driver resolve failed: %v", err)
	}
	second, err := driverProvider(context.Background())
	if err != nil {
		t.Fatalf("second driver resolve failed: %v", err)
	}

	if first == nil || second == nil {
		t.Fatal("expected non-nil ent SQL drivers")
	}
	if provider.calls != 2 {
		t.Fatalf("expected provider to be resolved twice, got %d", provider.calls)
	}
}

func TestNewEntClientProviderFromDB_ResolvesDBProviderEachCall(t *testing.T) {
	db := newNoopSQLDB()
	defer db.Close()

	provider := &countingDBProvider{
		db:      db,
		dialect: "mysql",
	}

	clientProvider := NewEntClientProviderFromDB(provider)
	first, err := clientProvider()
	if err != nil {
		t.Fatalf("first ent client resolve failed: %v", err)
	}
	second, err := clientProvider()
	if err != nil {
		t.Fatalf("second ent client resolve failed: %v", err)
	}

	if first == nil || second == nil {
		t.Fatal("expected non-nil ent clients")
	}
	if provider.calls != 2 {
		t.Fatalf("expected provider to be resolved twice, got %d", provider.calls)
	}
}

func TestData_DoesNotRetainRedisRuntimeHandles(t *testing.T) {
	redisClientType := reflect.TypeOf((*redis.UniversalClient)(nil)).Elem()
	redisProviderType := reflect.TypeOf((*lynxredis.Provider)(nil)).Elem()
	dataType := reflect.TypeOf(Data{})

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		if field.Type == redisClientType || field.Type == redisProviderType {
			t.Fatalf("unexpected redis runtime handle retained in Data: %s", field.Name)
		}
	}
}
