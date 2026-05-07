package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Monitor struct {
	ent.Schema
}

func (Monitor) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),
		field.String("url").
			NotEmpty(),
		field.Int("interval").
			Default(60),
		field.Int("timeout").
			Default(10),
		field.Enum("status").
			Values("up", "down", "pending", "maintenance", "paused", "unknown").
			Default("unknown"),
		field.Enum("type").
			Values("http", "keyword", "json", "ping", "tcp", "smtp", "dns", "push", "grpc"),
		field.Bool("is_active").
			Default(true),
		field.String("method").
			Default("GET"),
		field.JSON("accepted_status_codes", []string{"200-299"}).
			Optional(),
		field.JSON("headers", map[string]string{}).
			Optional(),
		field.String("body").
			Optional().
			Nillable(),
		field.String("auth_username").
			Optional().
			Nillable(),
		field.String("auth_password").
			Optional().
			Nillable().
			Sensitive(),
		field.String("filters_contains").
			Optional().
			Nillable(),
		field.String("filters_not_contains").
			Optional().
			Nillable(),
		field.String("json_path").
			Optional().
			Nillable(),
		field.String("expected_value").
			Optional().
			Nillable(),
		field.Bool("ignore_tls_errors").
			Default(false),
		field.Int("max_redirects").
			Default(10),
		field.Int("retry").
			Default(3),
		field.Int("retry_after").
			Default(30),
		field.String("push_token").
			Optional().
			Nillable().
			Unique(),
		field.JSON("config", map[string]any{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Monitor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("monitors").
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("checks", MonitorCheck.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("stats", MonitorStat.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("tags", Tag.Type).
			StorageKey(edge.Table("monitor_tags"), edge.Columns("monitor_id", "tag_id")),
		edge.To("notification_channels", NotificationChannel.Type).
			StorageKey(edge.Table("monitor_notification_channels"), edge.Columns("monitor_id", "notification_channel_id")),
		edge.To("status_page_monitors", StatusPageMonitor.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("maintenance_windows", MaintenanceWindow.Type).
			Ref("monitors"),
		edge.To("incidents", Incident.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Monitor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "is_active"),
		index.Fields("user_id", "type"),
		index.Fields("status"),
	}
}
