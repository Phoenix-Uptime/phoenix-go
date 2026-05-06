package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type MonitorCheck struct {
	ent.Schema
}

func (MonitorCheck) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Int("monitor_id"),
		field.Enum("status").
			Values("up", "down", "pending", "maintenance").
			Default("pending"),
		field.Time("checked_at").
			Default(time.Now),
		field.Int("response_time_ms").
			Default(0),
		field.Int("status_code").
			Optional().
			Nillable(),
		field.String("message").
			Optional().
			Nillable(),
		field.String("error").
			Optional().
			Nillable(),
		field.Int("retry_count").
			Default(0),
		field.Int("duration_seconds").
			Default(0),
		field.Bool("important").
			Default(false),
		field.String("response").
			Optional().
			Nillable().
			Sensitive(),
	}
}

func (MonitorCheck) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("monitor", Monitor.Type).
			Ref("checks").
			Field("monitor_id").
			Unique().
			Required(),
	}
}

func (MonitorCheck) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("monitor_id"),
		index.Fields("monitor_id", "checked_at"),
		index.Fields("monitor_id", "important", "checked_at"),
		index.Fields("checked_at"),
	}
}
