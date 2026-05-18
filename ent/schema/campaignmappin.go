package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CampaignMapPin struct {
	ent.Schema
}

func (CampaignMapPin) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("map_id"),
		field.String("name"),
		field.String("type").Default("poi"),
		field.Float("x").Default(0),
		field.Float("y").Default(0),
		field.String("icon").Default("fa-map-pin"),
		field.String("color").Default("#b8963e"),
		field.String("description").Default(""),
		field.String("linked_entity_type").Default(""),
		field.Int64("linked_entity_id").Optional(),
		field.Bool("is_hidden").Default(false),
		field.Int("sort_order").Default(0),
		field.String("created_at").Default(""),
	}
}

func (CampaignMapPin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("map", CampaignMap.Type).Ref("pins").Field("map_id").Unique().Required(),
	}
}

func (CampaignMapPin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("map_id"),
	}
}
