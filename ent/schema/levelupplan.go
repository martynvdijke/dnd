package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type LevelUpPlan struct {
	ent.Schema
}

func (LevelUpPlan) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.Int("target_level").Default(20),
		field.String("plan_data").Default("[]"),
		field.String("notes").Default(""),
		field.String("created_at").Default(""),
		field.String("updated_at").Default(""),
	}
}

func (LevelUpPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("level_up_plans").Field("character_id").Unique().Required(),
	}
}

func (LevelUpPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
