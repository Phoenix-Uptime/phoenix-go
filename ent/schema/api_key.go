package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type APIKey struct {
	ent.Schema
}

func (APIKey) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.String("name").
			NotEmpty(),
		field.String("key").
			NotEmpty().
			Unique().
			Sensitive(),
		field.Bool("is_active").
			Default(true),
		field.Time("expires_at").
			Optional().
			Nillable(),
		field.Time("last_used_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (APIKey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("api_keys").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (APIKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "is_active"),
	}
}
