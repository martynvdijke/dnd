package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CampaignRecap struct {
	ent.Schema
}

func (CampaignRecap) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id"),
		field.String("title"),
		field.String("content").Default(""),
		field.String("session_start_date").Optional(),
		field.String("session_end_date").Optional(),
		field.Int("word_count").Default(0),
		field.Bool("is_edited").Default(false),
		field.Bool("is_sent").Default(false),
		field.String("created_at").Default(""),
	}
}

func (CampaignRecap) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("recaps").Field("campaign_id").Unique().Required(),
	}
}

func (CampaignRecap) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
	}
}
