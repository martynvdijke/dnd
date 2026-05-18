package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CharacterLocation struct {
	ent.Schema
}

func (CharacterLocation) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.Int64("location_id"),
		field.String("relationship").Default("visited"),
		field.String("notes").Default(""),
	}
}

func (CharacterLocation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("character_locations").Field("character_id").Unique().Required(),
		edge.From("location", Location.Type).Ref("character_locations").Field("location_id").Unique().Required(),
	}
}

func (CharacterLocation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
		index.Fields("location_id"),
	}
}
