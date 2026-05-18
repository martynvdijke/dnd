package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ShareLink struct {
	ent.Schema
}

func (ShareLink) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("token").Unique(),
		field.String("entity_type"),
		field.Int64("entity_id"),
		field.Int64("created_by"),
		field.String("created_at").Default(""),
		field.String("expires_at").Optional(),
	}
}

func (ShareLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("share_links").Field("created_by").Unique().Required(),
	}
}

func (ShareLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token").Unique(),
		index.Fields("entity_type", "entity_id"),
	}
}
