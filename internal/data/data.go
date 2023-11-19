package data

import (
	"context"
	"entgo.io/ent/dialect/sql"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	"github.com/go-lynx/lynx/plugin/mysql"
	br "github.com/go-lynx/lynx/plugin/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewData,
	NewLoginRepo,
	mysql.GetDB,
	br.GetRedis)

type Data struct {
	db  *ent.Client
	rdb *redis.Client
}

func NewData(dri *sql.Driver, rdb *redis.Client, logger log.Logger) (*Data, error) {
	client := ent.NewClient(ent.Driver(dri), ent.Debug())

	// auto create database table
	if err := client.Schema.Create(context.Background()); err != nil {
		dfLog := log.NewHelper(logger)
		dfLog.Errorf("failed creating database schema resources: %v", err)
		return nil, err
	}

	d := &Data{
		db:  client,
		rdb: rdb,
	}
	return d, nil
}
