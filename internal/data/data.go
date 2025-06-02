package data

import (
	"entgo.io/ent/dialect/sql"
	"github.com/go-lynx/lynx-layout/internal/data/ent"
	lynxMysql "github.com/go-lynx/plugins/db/mysql/v2"
	lynxRedis "github.com/go-lynx/plugins/db/redis/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet 是 Google Wire 的提供器集合，用于定义依赖注入的规则。
// 包含了 NewData、NewLoginRepo 函数，以及从数据库插件和 Redis 插件获取驱动和客户端的函数。
var ProviderSet = wire.NewSet(
	NewData,
	NewLoginRepo,
	lynxMysql.GetDriver,
	lynxRedis.GetRedis)

// Data 结构体封装了数据库客户端和 Redis 客户端，用于项目的数据操作。
type Data struct {
	db  *ent.Client   // 数据库操作客户端
	rdb *redis.Client // Redis 操作客户端
}

// NewData 创建一个新的 Data 实例。
// 参数 dri 是 SQL 驱动，rdb 是 Redis 客户端，logger 是日志记录器。
// 返回 Data 实例指针和可能出现的错误。
func NewData(dri *sql.Driver, rdb *redis.Client) (*Data, error) {
	// 创建 ent 数据库客户端，开启调试模式
	client := ent.NewClient(ent.Driver(dri), ent.Debug())

	// 初始化 Data 实例
	d := &Data{
		db:  client,
		rdb: rdb,
	}
	return d, nil
}
