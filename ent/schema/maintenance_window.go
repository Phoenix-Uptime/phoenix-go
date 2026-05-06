package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type MaintenanceWindow struct {
	ent.Schema
}

func (MaintenanceWindow) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Int("user_id"),
		field.String("title").
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),
		field.Bool("is_active").
			Default(true),
		field.Enum("strategy").
			Values("manual", "single", "recurring", "cron").
			Default("single"),
		field.Time("start_at").
			Optional().
			Nillable(),
		field.Time("end_at").
			Optional().
			Nillable(),
		field.String("cron").
			Optional().
			Nillable(),
		field.String("timezone").
			Optional().
			Nillable(),
		field.Int("duration_seconds").
			Optional().
			Nillable(),
	}
}

func (MaintenanceWindow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("maintenance_windows").
			Field("user_id").
			Unique().
			Required(),
		edge.To("monitors", Monitor.Type).
			StorageKey(edge.Table("monitor_maintenance_windows"), edge.Columns("maintenance_window_id", "monitor_id")),
	}
}

func (MaintenanceWindow) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "is_active"),
		index.Fields("strategy", "is_active"),
		index.Fields("start_at", "end_at"),
	}
}
