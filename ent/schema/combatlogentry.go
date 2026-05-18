package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CombatLogEntry struct {
	ent.Schema
}

func (CombatLogEntry) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id").Optional(),
		field.Int64("combat_entry_id").Optional(),
		field.String("actor_name"),
		field.String("action"),
		field.String("target_name").Default(""),
		field.Int("damage").Default(0),
		field.String("damage_type").Default(""),
		field.Int("healing").Default(0),
		field.String("condition_applied").Default(""),
		field.String("roll_expression").Default(""),
		field.Int("roll_total").Default(0),
		field.Bool("is_critical").Default(false),
		field.String("description").Default(""),
		field.String("created_at").Default(""),
	}
}

func (CombatLogEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("combat_log_entries").Field("campaign_id").Unique(),
	}
}

func (CombatLogEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
		index.Fields("created_at"),
	}
}
