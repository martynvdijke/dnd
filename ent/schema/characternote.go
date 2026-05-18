package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CharacterNote struct {
	ent.Schema
}

func (CharacterNote) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("title"),
		field.String("content").Default(""),
		field.String("visibility").Default("player"),
		field.String("category").Default("general"),
		field.String("created_at").Default(""),
		field.String("updated_at").Default(""),
	}
}

func (CharacterNote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("notes").Field("character_id").Unique().Required(),
	}
}

func (CharacterNote) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
