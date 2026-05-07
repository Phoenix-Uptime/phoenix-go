package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type NotificationChannel struct {
	ent.Schema
}

func (NotificationChannel) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.String("name").
			NotEmpty(),
		field.Enum("type").
			Values("smtp", "telegram", "webhook", "discord", "slack", "pagerduty", "pushover", "twilio"),
		field.Bool("is_active").
			Default(true),
		field.Bool("is_default").
			Default(false),
		field.JSON("config", map[string]any{}).
			Optional().
			Sensitive(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (NotificationChannel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("notification_channels").
			Field("user_id").
			Unique().
			Required(),
		edge.From("monitors", Monitor.Type).
			Ref("notification_channels"),
	}
}

func (NotificationChannel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "type"),
		index.Fields("user_id", "is_default"),
	}
}
