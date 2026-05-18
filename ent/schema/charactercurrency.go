package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type CharacterCurrency struct {
	ent.Schema
}

func (CharacterCurrency) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("character_currency"),
	}
}

func (CharacterCurrency) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("character_id"),
		field.Int("cp").Default(0),
		field.Int("sp").Default(0),
		field.Int("ep").Default(0),
		field.Int("gp").Default(0),
		field.Int("pp").Default(0),
	}
}

func (CharacterCurrency) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("currency").Field("character_id").Unique().Required(),
	}
}
