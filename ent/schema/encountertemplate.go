package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type EncounterTemplate struct {
	ent.Schema
}

func (EncounterTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id").Optional(),
		field.Int64("user_id"),
		field.String("name"),
		field.String("description").Default(""),
		field.String("environment").Default(""),
		field.String("difficulty").Default("medium"),
		field.Int("xp_budget").Default(0),
		field.Int("total_xp").Default(0),
		field.String("notes").Default(""),
		field.String("created_at").Default(""),
	}
}

func (EncounterTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("encounter_templates").Field("user_id").Unique().Required(),
		edge.From("campaign", Campaign.Type).Ref("encounter_templates").Field("campaign_id").Unique(),
		edge.To("monsters", EncounterMonster.Type),
	}
}

func (EncounterTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
	}
}
