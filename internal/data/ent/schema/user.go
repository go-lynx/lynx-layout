package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("Primary key ID").
			Immutable().
			Positive(),
		field.String("num").
			Unique().
			MaxLen(32).
			Comment("User number"),
		field.String("account").
			MaxLen(32).
			Default("").
			Comment("Account"),
		field.String("password").
			MaxLen(100).
			Default("").
			Comment("Password"),
		field.String("phone").
			MaxLen(11).
			Default("").
			Comment("Phone number"),
		field.String("nickname").
			MaxLen(20).
			MinLen(1).
			Comment("Nickname"),
		field.String("avatar").
			MaxLen(240).
			Default("").
			Comment("Avatar"),
		field.Int32("stats").
			Default(1).
			Comment("Account status 1:normal;2:banned"),
		field.String("note").
			MaxLen(120).
			Default("").
			Comment("Personal introduction"),
		field.Int32("register_source").
			Comment("Registration source 1:web;2:app"),
		field.Time("last_login_at").
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime(3)",
				dialect.Postgres: "timestamptz",
			}).
			Comment("Last login time"),
		field.Time("created_at").
			Default(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime(3)",
				dialect.Postgres: "timestamptz",
			}).
			Comment("Created time"),
		field.Time("updated_at").
			Default(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL:    "datetime(3)",
				dialect.Postgres: "timestamptz",
			}).
			Comment("Updated time"),
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
		schema.Comment("User table"),
	}
}
