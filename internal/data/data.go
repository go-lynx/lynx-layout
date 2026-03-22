package data

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	lynxMysql "github.com/go-lynx/lynx-mysql"
	lynxRedis "github.com/go-lynx/lynx-redis"
	"github.com/go-lynx/lynx-sql-sdk/interfaces"
	_ "github.com/go-lynx/lynx-tracer"
	"github.com/go-lynx/lynx/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is a Google Wire provider set used to define dependency injection rules.
// It includes NewData, NewLoginRepo functions, and functions to get drivers and clients from database and Redis plugins.
var ProviderSet = wire.NewSet(
	NewData,
	NewLoginRepo,
	NewEntClientProvider,
	lynxRedis.GetRedis,
)

type EntClientProvider func() (*ent.Client, error)

type Data struct {
	db  EntClientProvider
	rdb *redis.Client // Redis operation client
}

// NewEntClientProvider returns a provider that always builds an ent client from the current database pool.
// Do not close the returned client in request-scoped code; the underlying pool is owned by the plugin.
func NewEntClientProvider() EntClientProvider {
	return NewEntClientProviderFromDB(lynxMysql.GetProvider())
}

// NewEntClientProviderFromDB creates an ent client provider from a stable SQL DB provider.
func NewEntClientProviderFromDB(provider interfaces.DBProvider) EntClientProvider {
	driverProvider := NewEntDriverProvider(provider)
	return func() (*ent.Client, error) {
		driver, err := driverProvider(context.Background())
		if err != nil {
			return nil, err
		}
		return ent.NewClient(
			ent.Driver(driver),
			ent.Debug(),
		), nil
	}
}

// NewEntDriverProvider resolves the current ent SQL driver from a stable DB provider.
func NewEntDriverProvider(provider interfaces.DBProvider) func(ctx context.Context) (*sql.Driver, error) {
	return func(ctx context.Context) (*sql.Driver, error) {
		db, err := provider.DB(ctx)
		if err != nil {
			return nil, err
		}
		if db == nil {
			return nil, fmt.Errorf("database connection is nil")
		}
		return sql.OpenDB(provider.Dialect(), db), nil
	}
}

// NewData creates a new Data instance.
func NewData(dbProvider EntClientProvider, rdb *redis.Client) (*Data, error) {
	client, err := dbProvider()
	if err != nil {
		return nil, err
	}
	// Auto create database table
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Errorf("failed creating database schema resources: %v", err)
		return nil, err
	}

	// Initialize Data instance
	d := &Data{
		db:  dbProvider,
		rdb: rdb,
	}
	return d, nil
}

func (d *Data) entClient() (*ent.Client, error) {
	return d.db()
}
