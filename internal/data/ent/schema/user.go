package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"time"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("主键ID").
			Immutable().
			Positive(),
		field.String("num").
			Unique().
			MaxLen(32).
			Comment("编号"),
		field.String("account").
			MaxLen(32).
			Default("").
			Comment("账号"),
		field.String("password").
			MaxLen(100).
			Default("").
			Comment("密码"),
		field.String("phone").
			MaxLen(11).
			Default("").
			Comment("手机号"),
		field.String("nickname").
			MaxLen(20).
			MinLen(1).
			Comment("昵称"),
		field.String("avatar").
			MaxLen(240).
			Default("").
			Comment("个人肖像"),
		field.Int32("stats").
			Default(1).
			Comment("账号状态 1:正常;2:封禁"),
		field.String("note").
			MaxLen(120).
			Default("").
			Comment("个人简介"),
		field.Int32("register_source").
			Comment("注册来源 1:web端;2:app端"),
		field.Time("last_login_at").
			SchemaType(map[string]string{dialect.MySQL: "datetime(3)"}).
			Comment("最近登录时间"),
		field.Time("created_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.MySQL: "datetime(3)"}).
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.MySQL: "datetime(3)"}).
			Comment("修改时间"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

func (User) ID() ent.Field {
	return field.Int("id")
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("用户表"),
	}
}
