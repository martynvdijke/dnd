package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type CharacterFeature struct {
	ent.Schema
}

func (CharacterFeature) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("name"),
		field.String("description").Default(""),
		field.String("source").Default(""),
		field.Int("level_gained").Default(1),
	}
}

func (CharacterFeature) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("features").Field("character_id").Unique().Required(),
	}
}
