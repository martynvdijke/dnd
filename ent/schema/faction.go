package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Faction struct {
	ent.Schema
}

func (Faction) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id").Optional(),
		field.String("name"),
		field.String("description").Default(""),
		field.String("type").Default("organization"),
		field.String("headquarters").Default(""),
		field.String("created_at").Default(""),
	}
}

func (Faction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("factions").Field("campaign_id").Unique(),
		edge.To("reputations", FactionReputation.Type),
	}
}

func (Faction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
	}
}
