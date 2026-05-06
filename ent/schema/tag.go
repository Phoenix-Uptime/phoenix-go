package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Tag struct {
	ent.Schema
}

func (Tag) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Unique(),
		field.String("description").
			Optional(),
		field.String("color").
			Optional(),
	}
}

func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("monitors", Monitor.Type).
			Ref("tags"),
		edge.To("status_pages", StatusPage.Type),
	}
}
