package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StatusPage struct {
	ent.Schema
}

func (StatusPage) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Int("user_id"),
		field.String("slug").
			NotEmpty().
			Unique(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),
		field.Bool("is_public").
			Default(true),
		field.String("password").
			Optional().
			Nillable().
			Sensitive(),
		field.String("theme").
			Default("default"),
		field.String("custom_css").
			Optional().
			Nillable(),
		field.String("footer_text").
			Optional().
			Nillable(),
		field.Bool("show_tags").
			Default(false),
		field.Bool("show_charts").
			Default(true),
		field.Bool("show_uptime_percentage").
			Default(true),
		field.Bool("show_powered_by").
			Default(true),
		field.Int("auto_refresh_interval").
			Default(300),
	}
}

func (StatusPage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("status_pages").
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("status_page_monitors", StatusPageMonitor.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("messages", StatusMessage.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("incidents", Incident.Type),
	}
}

func (StatusPage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "slug").
			Unique(),
	}
}
