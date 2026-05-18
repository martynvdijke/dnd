package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CampaignWikiPage struct {
	ent.Schema
}

func (CampaignWikiPage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id"),
		field.Int64("user_id"),
		field.Int64("parent_id").Optional(),
		field.String("title"),
		field.String("content").Default(""),
		field.String("visibility").Default("public"),
		field.String("tags").Default("[]"),
		field.Int("sort_order").Default(0),
		field.String("created_at").Default(""),
		field.String("updated_at").Default(""),
	}
}

func (CampaignWikiPage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("wiki_pages").Field("campaign_id").Unique().Required(),
		edge.From("user", User.Type).Ref("wiki_pages").Field("user_id").Unique().Required(),
		edge.To("children", CampaignWikiPage.Type).From("parent").Field("parent_id").Unique(),
	}
}

func (CampaignWikiPage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
		index.Fields("parent_id"),
	}
}
