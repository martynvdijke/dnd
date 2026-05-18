package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type DiceRoll struct {
	ent.Schema
}

func (DiceRoll) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id"),
		field.Int64("character_id").Optional(),
		field.String("expression"),
		field.String("result"),
		field.Int("total"),
		field.String("timestamp").Default(""),
	}
}

func (DiceRoll) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("dice_rolls").Field("user_id").Unique().Required(),
	}
}

func (DiceRoll) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("character_id"),
	}
}
