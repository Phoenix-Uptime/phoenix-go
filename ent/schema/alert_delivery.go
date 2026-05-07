package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type AlertDelivery struct {
	ent.Schema
}

func (AlertDelivery) Fields() []ent.Field {
	return []ent.Field{
		field.Int("alert_rule_id"),
		field.Int("notification_channel_id"),
		field.Int("monitor_id").
			Optional().
			Nillable(),
		field.Int("incident_id").
			Optional().
			Nillable(),
		field.Enum("event").
			Values("down", "recovered", "degraded", "certificate_expiring"),
		field.Enum("status").
			Values("pending", "sent", "failed", "skipped").
			Default("pending"),
		field.String("error").
			Optional().
			Nillable(),
		field.Int("retry_count").
			Default(0),
		field.Time("sent_at").
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

func (AlertDelivery) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("alert_rule", AlertRule.Type).
			Ref("deliveries").
			Field("alert_rule_id").
			Unique().
			Required(),
		edge.From("notification_channel", NotificationChannel.Type).
			Ref("alert_deliveries").
			Field("notification_channel_id").
			Unique().
			Required(),
		edge.From("monitor", Monitor.Type).
			Ref("alert_deliveries").
			Field("monitor_id").
			Unique(),
		edge.From("incident", Incident.Type).
			Ref("alert_deliveries").
			Field("incident_id").
			Unique(),
	}
}

func (AlertDelivery) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("alert_rule_id"),
		index.Fields("notification_channel_id"),
		index.Fields("monitor_id"),
		index.Fields("incident_id"),
		index.Fields("status", "created_at"),
		index.Fields("event", "created_at"),
	}
}
