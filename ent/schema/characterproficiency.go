package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type CharacterProficiency struct {
	ent.Schema
}

func (CharacterProficiency) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("type"),
		field.String("name"),
	}
}

func (CharacterProficiency) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("proficiencies").Field("character_id").Unique().Required(),
	}
}
