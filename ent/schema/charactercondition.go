package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CharacterCondition struct {
	ent.Schema
}

func (CharacterCondition) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("name"),
		field.String("type").Default("other"),
		field.String("source").Default(""),
		field.Int("duration").Default(0),
		field.String("duration_type").Default("round"),
		field.String("saving_throw").Default(""),
		field.Int("save_dc").Default(0),
		field.String("description").Default(""),
		field.String("started_at").Default(""),
	}
}

func (CharacterCondition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("conditions").Field("character_id").Unique().Required(),
	}
}

func (CharacterCondition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
