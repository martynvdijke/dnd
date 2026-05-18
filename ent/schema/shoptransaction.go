package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ShopTransaction struct {
	ent.Schema
}

func (ShopTransaction) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("shop_id"),
		field.Int64("character_id"),
		field.String("item_name"),
		field.Int("quantity").Default(1),
		field.Float("price_gp").Default(0),
		field.String("transaction_type"),
		field.String("timestamp").Default(""),
	}
}

func (ShopTransaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("shop", Shop.Type).Ref("transactions").Field("shop_id").Unique().Required(),
	}
}

func (ShopTransaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("shop_id"),
		index.Fields("character_id"),
	}
}
