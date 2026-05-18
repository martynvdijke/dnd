package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type CampaignMember struct {
	ent.Schema
}

func (CampaignMember) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id"),
		field.Int64("user_id"),
		field.String("role").Default("player"),
	}
}

func (CampaignMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("members").Field("campaign_id").Unique().Required(),
		edge.From("user", User.Type).Ref("campaign_members").Field("user_id").Unique().Required(),
	}
}
