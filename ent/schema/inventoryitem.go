package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type InventoryItem struct {
	ent.Schema
}

func (InventoryItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("inventory"),
	}
}

func (InventoryItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("name"),
		field.Int("quantity").Default(1),
		field.Float("weight").Default(0),
		field.String("category").Default("gear"),
		field.String("damage_dice").Default(""),
		field.String("damage_type").Default(""),
		field.String("weapon_properties").Default(""),
		field.Int("ac_bonus").Default(0),
		field.String("armor_type").Default(""),
		field.String("description").Default(""),
		field.Bool("is_equipped").Default(false),
		field.Bool("is_magical").Default(false),
		field.Bool("attunement").Default(false),
		field.Bool("is_identified").Default(false),
		field.String("notes").Default(""),
	}
}

func (InventoryItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("inventory").Field("character_id").Unique().Required(),
	}
}

func (InventoryItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
