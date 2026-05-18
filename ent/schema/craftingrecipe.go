package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CraftingRecipe struct {
	ent.Schema
}

func (CraftingRecipe) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id").Optional(),
		field.String("name"),
		field.String("description").Default(""),
		field.String("category").Default("other"),
		field.Int("difficulty_dc").Default(10),
		field.Float("crafting_time_hours").Default(1),
		field.String("required_tools").Default("[]"),
		field.String("required_materials").Default("[]"),
		field.String("result_item_name"),
		field.String("result_item_category").Default("other"),
		field.Int("result_quantity").Default(1),
		field.String("result_description").Default(""),
		field.String("notes").Default(""),
		field.String("created_at").Default(""),
	}
}

func (CraftingRecipe) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("crafting_recipes").Field("user_id").Unique(),
		edge.To("character_crafting", CharacterCrafting.Type),
	}
}

func (CraftingRecipe) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
