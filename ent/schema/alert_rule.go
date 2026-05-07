package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type AlertRule struct {
	ent.Schema
}

func (AlertRule) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),
		field.Enum("event").
			Values("down", "recovered", "degraded", "certificate_expiring"),
		field.Enum("scope").
			Values("all", "tags", "monitors").
			Default("all"),
		field.Bool("is_active").
			Default(true),
		field.Int("resend_interval_seconds").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (AlertRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("alert_rules").
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("tags", Tag.Type).
			StorageKey(edge.Table("alert_rule_tags"), edge.Columns("alert_rule_id", "tag_id")),
		edge.To("monitors", Monitor.Type).
			StorageKey(edge.Table("alert_rule_monitors"), edge.Columns("alert_rule_id", "monitor_id")),
		edge.To("notification_channels", NotificationChannel.Type).
			StorageKey(edge.Table("alert_rule_notification_channels"), edge.Columns("alert_rule_id", "notification_channel_id")),
		edge.To("deliveries", AlertDelivery.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (AlertRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "is_active"),
		index.Fields("event", "is_active"),
		index.Fields("scope"),
	}
}
