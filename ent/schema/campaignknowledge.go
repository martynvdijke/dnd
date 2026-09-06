package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CampaignKnowledge struct {
	ent.Schema
}

func (CampaignKnowledge) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id"),
		field.String("title"),
		field.String("content").Default(""),
		field.String("source").Default(""),
		field.String("status").Default("rumor"),
		field.Bool("shared").Default(false),
		field.String("status_history").Default("[]"),
		field.String("created_at").Default(""),
		field.String("updated_at").Default(""),
	}
}

func (CampaignKnowledge) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("knowledge_entries").Field("campaign_id").Unique().Required(),
	}
}

func (CampaignKnowledge) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
		index.Fields("status"),
	}
}
