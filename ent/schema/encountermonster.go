package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type EncounterMonster struct {
	ent.Schema
}

func (EncounterMonster) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("encounter_id"),
		field.String("name"),
		field.Int("count").Default(1),
		field.String("cr").Default("0"),
		field.Int("xp").Default(0),
		field.Int("ac").Default(10),
		field.Int("hp").Default(1),
		field.Int("initiative_mod").Default(0),
		field.String("source").Default("homebrew"),
		field.String("notes").Default(""),
		field.Int64("compendium_monster_id").Optional().Nillable(),
	}
}

func (EncounterMonster) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("encounter", EncounterTemplate.Type).Ref("monsters").Field("encounter_id").Unique().Required(),
	}
}

func (EncounterMonster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("encounter_id"),
	}
}
