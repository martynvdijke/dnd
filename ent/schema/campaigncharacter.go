package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CampaignCharacter links a character to a campaign. Membership is many-to-many;
// the pair is unique so repeated attach requests are idempotent.
type CampaignCharacter struct {
	ent.Schema
}

func (CampaignCharacter) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id"),
		field.Int64("character_id"),
	}
}

func (CampaignCharacter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("character_links").Field("campaign_id").Unique().Required(),
		edge.From("character", Character.Type).Ref("campaign_links").Field("character_id").Unique().Required(),
	}
}

func (CampaignCharacter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id", "character_id").Unique(),
		index.Fields("character_id"),
	}
}
