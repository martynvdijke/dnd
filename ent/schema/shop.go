package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Shop struct {
	ent.Schema
}

func (Shop) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id"),
		field.Int64("campaign_id").Optional(),
		field.Int64("oneshot_adventure_id").Optional(),
		field.Int64("act_id").Optional(),
		field.String("name"),
		field.String("description").Default(""),
		field.Float("markup_percent").Default(100),
		field.Float("markup_buy_percent").Default(50),
		field.String("created_at").Default(""),
	}
}

func (Shop) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id"),
		index.Fields("oneshot_adventure_id"),
		index.Fields("act_id"),
	}
}

func (Shop) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("shops").Field("user_id").Unique().Required(),
		edge.From("campaign", Campaign.Type).Ref("shops").Field("campaign_id").Unique(),
		edge.To("items", ShopItem.Type),
		edge.To("transactions", ShopTransaction.Type),
	}
}
