package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			NotEmpty().
			Unique(),
		field.String("email").
			NotEmpty().
			Unique(),
		field.String("password").
			Sensitive().
			NotEmpty(),
		field.String("api_key").
			NotEmpty().
			Unique(),
		field.String("smtp_smtp_server").
			Optional(),
		field.Int("smtp_smtp_port").
			Optional(),
		field.String("smtp_from_address").
			Optional(),
		field.String("smtp_username").
			Optional(),
		field.String("smtp_password").
			Optional().
			Sensitive(),
		field.Bool("smtp_use_tls").
			Optional(),
		field.String("telegram_bot_token").
			Optional().
			Sensitive(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("monitors", Monitor.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username"),
		index.Fields("email"),
		index.Fields("api_key"),
	}
}
