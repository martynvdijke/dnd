package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type SessionPlan struct {
	ent.Schema
}

func (SessionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id").Optional(),
		field.String("title"),
		field.String("session_date").Default(""),
		field.String("status").Default("planned"),
		field.String("dm_notes").Default(""),
		field.String("planned_encounters").Default("[]"),
		field.String("npc_ids").Default("[]"),
		field.String("player_goals").Default("[]"),
		field.Int("expected_duration").Default(0),
		field.String("created_at").Default(""),
		field.String("updated_at").Default(""),
	}
}

func (SessionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("session_plans").Field("campaign_id").Unique(),
	}
}

func (SessionPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
	}
}
