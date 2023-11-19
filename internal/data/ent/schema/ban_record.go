package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"time"
)

// BanRecord holds the schema definition for the BanRecord entity.
type BanRecord struct {
	ent.Schema
}

// Fields of the User.
func (BanRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("主键ID").
			Immutable().
			Positive(),
		field.Int64("user_id").
			Immutable().
			Comment("用户ID"),
		field.String("ban_code").
			MaxLen(120).
			Default("").
			Comment("封禁类型"),
		field.String("ban_note").
			MaxLen(240).
			Default("").
			Comment("封禁原因"),
		field.Time("release_at").
			Nillable().
			SchemaType(map[string]string{dialect.MySQL: "datetime(3)"}).
			Comment("解封时间"),
		field.String("release_note").
			MaxLen(240).
			Default("").
			Comment("解封原因"),
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
func (BanRecord) Edges() []ent.Edge {
	return nil
}

func (BanRecord) ID() ent.Field {
	return field.Int("id")
}

func (BanRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("用户封禁记录表"),
	}
}
