package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CampaignTimelineEvent struct {
	ent.Schema
}

func (CampaignTimelineEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id"),
		field.String("title"),
		field.String("description").Default(""),
		field.String("event_date"),
		field.String("event_type").Default("other"),
		field.Int("importance").Default(1),
		field.String("icon").Default("fa-star"),
		field.String("linked_entity_type").Default(""),
		field.Int64("linked_entity_id").Optional(),
		field.String("created_at").Default(""),
	}
}

func (CampaignTimelineEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("timeline_events").Field("campaign_id").Unique().Required(),
	}
}

func (CampaignTimelineEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
	}
}
