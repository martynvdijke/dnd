package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ShopItem struct {
	ent.Schema
}

func (ShopItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("shop_id"),
		field.String("item_name"),
		field.String("category").Default("gear"),
		field.Float("price_gp").Default(0),
		field.Int("quantity_available").Default(-1),
		field.String("description").Default(""),
		field.Bool("is_magical").Default(false),
		field.Bool("attunement_required").Default(false),
		field.String("notes").Default(""),
	}
}

func (ShopItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("shop", Shop.Type).Ref("items").Field("shop_id").Unique().Required(),
	}
}

func (ShopItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("shop_id"),
	}
}
