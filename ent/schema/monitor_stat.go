package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type MonitorStat struct {
	ent.Schema
}

func (MonitorStat) Fields() []ent.Field {
	return []ent.Field{
		field.Int("monitor_id"),
		field.Enum("period").
			Values("hour", "day"),
		field.Time("period_start"),
		field.Int("total_checks").
			Default(0),
		field.Int("up_checks").
			Default(0),
		field.Int("down_checks").
			Default(0),
		field.Int("maintenance_checks").
			Default(0),
		field.Float("uptime_percentage").
			Default(0),
		field.Int("avg_response_time_ms").
			Default(0),
		field.Int("min_response_time_ms").
			Optional().
			Nillable(),
		field.Int("max_response_time_ms").
			Optional().
			Nillable(),
		field.Int("downtime_seconds").
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (MonitorStat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("monitor", Monitor.Type).
			Ref("stats").
			Field("monitor_id").
			Unique().
			Required(),
	}
}

func (MonitorStat) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("monitor_id"),
		index.Fields("period_start"),
		index.Fields("monitor_id", "period", "period_start").
			Unique(),
	}
}
