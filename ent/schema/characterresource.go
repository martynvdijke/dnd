package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CharacterResource struct {
	ent.Schema
}

func (CharacterResource) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("name"),
		field.Int("current").Default(0),
		field.Int("max").Default(0),
		field.Bool("short_rest_recovery").Default(false),
		field.Bool("long_rest_recovery").Default(true),
		field.String("icon").Default("fa-bolt"),
		field.Int("sort_order").Default(0),
		field.String("created_at").Default(""),
	}
}

func (CharacterResource) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("resources").Field("character_id").Unique().Required(),
	}
}

func (CharacterResource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
