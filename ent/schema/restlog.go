package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type RestLog struct {
	ent.Schema
}

func (RestLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("rest_type"),
		field.Int("hp_healed").Default(0),
		field.String("slots_recovered").Default("[]"),
		field.Int("hit_dice_spent").Default(0),
		field.String("notes").Default(""),
		field.String("timestamp").Default(""),
	}
}

func (RestLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("rest_logs").Field("character_id").Unique().Required(),
	}
}
