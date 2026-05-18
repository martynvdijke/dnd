package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CharacterClass struct {
	ent.Schema
}

func (CharacterClass) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("class"),
		field.String("subclass").Default(""),
		field.Int("level").Default(1),
		field.String("hit_dice").Default("d10"),
		field.String("created_at").Default(""),
	}
}

func (CharacterClass) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("classes").Field("character_id").Unique().Required(),
	}
}

func (CharacterClass) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
