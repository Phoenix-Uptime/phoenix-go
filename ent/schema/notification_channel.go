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
			Values("smtp", "telegram", "webhook", "ntfy"),
		field.Bool("is_active").
			Default(true),
		field.Bool("is_default").
			Default(false),
		field.String("smtp_server").
			Optional().
			Nillable(),
		field.Int("smtp_port").
			Optional().
			Nillable(),
		field.String("smtp_from_address").
			Optional().
			Nillable(),
		field.String("smtp_username").
			Optional().
			Nillable(),
		field.String("smtp_password").
			Optional().
			Nillable().
			Sensitive(),
		field.Bool("smtp_use_tls").
			Optional().
			Nillable(),
		field.String("telegram_bot_token").
			Optional().
			Nillable().
			Sensitive(),
		field.String("telegram_chat_id").
			Optional().
			Nillable(),
		field.String("webhook_url").
			Optional().
			Nillable().
			Sensitive(),
		field.String("webhook_method").
			Optional().
			Nillable(),
		field.String("ntfy_server_url").
			Optional().
			Nillable(),
		field.String("ntfy_topic").
			Optional().
			Nillable(),
		field.String("ntfy_token").
			Optional().
			Nillable().
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
