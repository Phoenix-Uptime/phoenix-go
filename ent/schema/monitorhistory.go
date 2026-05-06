package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type MonitorHistory struct {
	ent.Schema
}

func (MonitorHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (MonitorHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Int("monitor_id"),
		field.String("status").
			NotEmpty(),
		field.Int("response_time"),
		field.String("error_message").
			Optional(),
	}
}

func (MonitorHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("monitor", Monitor.Type).
			Ref("history").
			Field("monitor_id").
			Unique().
			Required(),
	}
}

func (MonitorHistory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("monitor_id"),
	}
}
