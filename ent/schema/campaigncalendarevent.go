package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CampaignCalendarEvent struct {
	ent.Schema
}

func (CampaignCalendarEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id"),
		field.String("title"),
		field.String("description").Default(""),
		field.String("event_date"),
		field.String("event_type").Default("other"),
		field.String("color").Default("#b8963e"),
		field.String("created_at").Default(""),
	}
}

func (CampaignCalendarEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("calendar_events").Field("campaign_id").Unique().Required(),
	}
}

func (CampaignCalendarEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
	}
}
