package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StatusPage struct {
	ent.Schema
}

func (StatusPage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (StatusPage) Fields() []ent.Field {
	return []ent.Field{
		field.Int("tag_id"),
		field.String("name").
			NotEmpty(),
		field.Bool("is_public").
			Default(true),
	}
}

func (StatusPage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tag", Tag.Type).
			Ref("status_pages").
			Field("tag_id").
			Unique().
			Required(),
		edge.To("messages", StatusMessage.Type),
	}
}

func (StatusPage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tag_id"),
	}
}
