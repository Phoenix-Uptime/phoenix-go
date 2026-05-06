package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Incident struct {
	ent.Schema
}

func (Incident) Fields() []ent.Field {
	return []ent.Field{
		field.Int("monitor_id"),
		field.Int("status_page_id").
			Optional().
			Nillable(),
		field.Int("resolved_by_id").
			Optional().
			Nillable(),
		field.String("title").
			NotEmpty(),
		field.String("content").
			Optional().
			Nillable(),
		field.Enum("status").
			Values("open", "acknowledged", "resolved").
			Default("open"),
		field.Enum("severity").
			Values("info", "warning", "critical").
			Default("warning"),
		field.Time("started_at").
			Default(time.Now),
		field.Time("ended_at").
			Optional().
			Nillable(),
		field.Bool("is_pinned").
			Default(true),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Incident) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("monitor", Monitor.Type).
			Ref("incidents").
			Field("monitor_id").
			Unique().
			Required(),
		edge.From("status_page", StatusPage.Type).
			Ref("incidents").
			Field("status_page_id").
			Unique(),
		edge.From("resolved_by", User.Type).
			Ref("resolved_incidents").
			Field("resolved_by_id").
			Unique(),
		edge.To("messages", StatusMessage.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Incident) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("monitor_id"),
		index.Fields("monitor_id", "status"),
		index.Fields("status_page_id", "status"),
		index.Fields("status", "started_at"),
		index.Fields("started_at"),
	}
}
