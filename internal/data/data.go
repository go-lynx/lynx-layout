package data

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	"github.com/go-lynx/lynx/app/log"
	lynxPgsql "github.com/go-lynx/lynx/plugins/db/pgsql"
	lynxRedis "github.com/go-lynx/lynx/plugins/nosql/redis"
	_ "github.com/go-lynx/lynx/plugins/tracer"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is a Google Wire provider set used to define dependency injection rules.
// It includes NewData, NewLoginRepo functions, and functions to get drivers and clients from database and Redis plugins.
var ProviderSet = wire.NewSet(
	NewData,
	NewLoginRepo,
	lynxPgsql.GetDriver,
	lynxRedis.GetRedis)

// Data struct encapsulates database client and Redis client for project data operations.
type Data struct {
	db  *ent.Client   // Database operation client
	rdb *redis.Client // Redis operation client
}

// NewData creates a new Data instance.
// Parameters: dri is the SQL driver, rdb is the Redis client, logger is the logger.
// Returns a Data instance pointer and any possible errors.
func NewData(dri *sql.Driver, rdb *redis.Client) (*Data, error) {
	// Create ent database client with debug mode enabled
	client := ent.NewClient(
		ent.Driver(dri),
		ent.Debug(),
	)
	// Auto create database table
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Errorf("failed creating database schema resources: %v", err)
		return nil, err
	}

	// Initialize Data instance
	d := &Data{
		db:  client,
		rdb: rdb,
	}
	return d, nil
}
