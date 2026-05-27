package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PartyItem struct {
	ent.Schema
}

func (PartyItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("campaign_id").Optional(),
		field.String("name"),
		field.Int("quantity").Default(1),
		field.String("notes").Default(""),
		field.String("created_at").Default(""),
	}
}

func (PartyItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("campaign", Campaign.Type).Ref("party_items").Field("campaign_id").Unique(),
	}
}

func (PartyItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
	}
}
