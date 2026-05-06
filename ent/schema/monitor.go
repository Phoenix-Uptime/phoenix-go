package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Monitor struct {
	ent.Schema
}

func (Monitor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (Monitor) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.String("name").
			NotEmpty(),
		field.String("url").
			NotEmpty(),
		field.Int("interval").
			Default(60),
		field.Enum("status").
			Values("up", "down", "paused", "unknown").
			Default("unknown"),
		field.Enum("type").
			Values("url", "ping", "smtp"),
		field.String("method").
			Optional().
			Nillable(),
		field.String("filters_contains").
			Optional(),
		field.String("filters_not_contains").
			Optional(),
		field.Int("retry").
			Default(3),
		field.Int("retry_after").
			Default(30),
		field.JSON("alert_types", []string{}).
			Optional(),
	}
}

func (Monitor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("monitors").
			Field("user_id").
			Unique().
			Required(),
		edge.To("history", MonitorHistory.Type),
		edge.To("tags", Tag.Type).
			StorageKey(edge.Table("monitor_tags"), edge.Columns("monitor_id", "tag_id")),
	}
}

func (Monitor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
