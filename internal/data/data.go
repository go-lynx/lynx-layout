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
	db EntClientProvider
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
// The Redis provider is kept in the constructor only to preserve the current Wire contract owned by cmd/user.
// internal/data itself does not retain or call the provider at runtime.
func NewData(dbProvider EntClientProvider, _ lynxredis.Provider) (*Data, error) {
	if dbProvider == nil {
		return nil, fmt.Errorf("ent client provider is nil")
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
	return d.db()
}
