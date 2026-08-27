// Package data provides the data-access layer of the user service.
// It wires together the ent ORM client (backed by MySQL via lynx-mysql) and
// exposes repository implementations consumed by the business-logic layer.
package data

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	lynxredis "github.com/go-lynx/lynx-redis"
	"github.com/go-lynx/lynx-sql-sdk/interfaces"
	"github.com/go-lynx/lynx/log"
	"github.com/google/wire"
)

// ProviderSet is a Google Wire provider set used to define dependency injection rules.
// It includes NewData, NewLoginRepo functions, and functions to get drivers and providers from database and Redis plugins.
var ProviderSet = wire.NewSet(
	NewData,
	NewLoginRepo,
	NewLoginAuthTokenIssuer,
)

type EntClientProvider func() (*ent.Client, error)

type Data struct {
	db     EntClientProvider
	memory *memoryStore // non-nil when running without the mysql client plugin
}

// InMemory reports whether the data layer is backed by the in-memory store instead of ent/MySQL.
func (d *Data) InMemory() bool {
	return d != nil && d.memory != nil
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
// When dbProvider is nil (the mysql client plugin is not loaded, e.g. `lynx.mysql.enabled: false`
// in bootstrap.local.yaml) the data layer falls back to a mutex-guarded in-memory user/token store
// so a freshly generated project can start and serve HTTP/gRPC with no external dependencies.
// The Redis provider is kept in the constructor only to preserve the current Wire contract owned by cmd/user.
// internal/data itself does not retain or call the provider at runtime; it may be nil as well.
func NewData(dbProvider EntClientProvider, redisProvider lynxredis.Provider) (*Data, error) {
	if dbProvider == nil {
		store, err := newSeededMemoryStore()
		if err != nil {
			return nil, err
		}
		log.Infof("mysql client plugin not loaded: data layer is running with in-memory storage (demo account %q)", memoryDemoAccount)
		if redisProvider == nil {
			log.Infof("redis client plugin not loaded: login tokens are kept in memory")
		}
		return &Data{memory: store}, nil
	}

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
		// Keep only the stable DB provider. Redis wiring is validated by bootstrap, but we do not
		// retain a replaceable redis handle or provider in the data layer singleton.
		db: dbProvider,
	}
	return d, nil
}

func (d *Data) entClient() (*ent.Client, error) {
	if d == nil || d.db == nil {
		return nil, fmt.Errorf("ent client provider is not configured (running in-memory?)")
	}
	return d.db()
}
