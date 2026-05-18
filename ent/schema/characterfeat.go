package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CharacterFeat struct {
	ent.Schema
}

func (CharacterFeat) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("name"),
		field.String("description").Default(""),
		field.String("prerequisites").Default(""),
		field.String("source").Default(""),
		field.Int("level_gained").Default(1),
		field.String("created_at").Default(""),
	}
}

func (CharacterFeat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("feats").Field("character_id").Unique().Required(),
	}
}

func (CharacterFeat) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
