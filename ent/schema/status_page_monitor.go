package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StatusPageMonitor struct {
	ent.Schema
}

func (StatusPageMonitor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (StatusPageMonitor) Fields() []ent.Field {
	return []ent.Field{
		field.Int("status_page_id"),
		field.Int("monitor_id"),
		field.String("display_name").
			Optional().
			Nillable(),
		field.Int("weight").
			Default(1000),
		field.Bool("send_url").
			Default(false),
	}
}

func (StatusPageMonitor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("status_page", StatusPage.Type).
			Ref("status_page_monitors").
			Field("status_page_id").
			Unique().
			Required(),
		edge.From("monitor", Monitor.Type).
			Ref("status_page_monitors").
			Field("monitor_id").
			Unique().
			Required(),
	}
}

func (StatusPageMonitor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status_page_id"),
		index.Fields("monitor_id"),
		index.Fields("status_page_id", "monitor_id").
			Unique(),
		index.Fields("status_page_id", "weight"),
	}
}
