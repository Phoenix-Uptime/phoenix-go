package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StatusMessage struct {
	ent.Schema
}

func (StatusMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Int("status_page_id"),
		field.Int("incident_id").
			Optional().
			Nillable(),
		field.Int("parent_id").
			Optional().
			Nillable(),
		field.Enum("type").
			Values("issue", "investigating", "identified", "monitoring", "resolved", "maintenance"),
		field.String("title").
			Optional().
			Nillable(),
		field.String("content").
			NotEmpty(),
	}
}

func (StatusMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("status_page", StatusPage.Type).
			Ref("messages").
			Field("status_page_id").
			Unique().
			Required(),
		edge.From("incident", Incident.Type).
			Ref("messages").
			Field("incident_id").
			Unique(),
		edge.To("sub_messages", StatusMessage.Type).
			From("parent").
			Field("parent_id").
			Unique(),
	}
}

func (StatusMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status_page_id"),
		index.Fields("incident_id"),
		index.Fields("parent_id"),
	}
}
