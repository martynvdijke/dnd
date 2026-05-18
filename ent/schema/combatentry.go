package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CombatEntry struct {
	ent.Schema
}

func (CombatEntry) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id").Optional(),
		field.Int64("campaign_id").Optional(),
		field.String("name"),
		field.String("type").Default("character"),
		field.Int("initiative_roll").Default(0),
		field.Int("initiative_mod").Default(0),
		field.Int("hp_max").Default(1),
		field.Int("hp_current").Default(1),
		field.Int("ac").Default(10),
		field.Bool("is_active").Default(true),
		field.Int("turn_order").Default(0),
		field.String("condition_ids").Default("[]"),
		field.String("notes").Default(""),
		field.String("created_at").Default(""),
	}
}

func (CombatEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("combat_entries").Field("character_id").Unique(),
		edge.From("campaign", Campaign.Type).Ref("combat_entries").Field("campaign_id").Unique(),
	}
}

func (CombatEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
		index.Fields("character_id"),
	}
}
