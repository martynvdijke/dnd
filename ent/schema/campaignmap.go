package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CampaignMap struct {
	ent.Schema
}

func (CampaignMap) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id"),
		field.String("name"),
		field.String("image_url").Default(""),
		field.Int("width").Default(1000),
		field.Int("height").Default(800),
		field.Int("grid_size").Default(50),
		field.Bool("is_active").Default(false),
		field.String("fog_of_war").Default("[]"),
		field.String("created_at").Default(""),
	}
}

func (CampaignMap) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("maps").Field("campaign_id").Unique().Required(),
		edge.To("pins", CampaignMapPin.Type),
	}
}

func (CampaignMap) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
	}
}
