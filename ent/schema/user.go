package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			NotEmpty().
			Unique(),
		field.String("email").
			NotEmpty().
			Unique(),
		field.String("password").
			Sensitive().
			NotEmpty(),
		field.String("api_key").
			NotEmpty().
			Unique(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("monitors", Monitor.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("tags", Tag.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("api_keys", APIKey.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("notification_channels", NotificationChannel.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("status_pages", StatusPage.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("maintenance_windows", MaintenanceWindow.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("resolved_incidents", Incident.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username"),
		index.Fields("email"),
		index.Fields("api_key"),
	}
}
