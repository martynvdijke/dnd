package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CharacterCrafting struct {
	ent.Schema
}

func (CharacterCrafting) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.Int64("recipe_id").Optional(),
		field.String("name"),
		field.Float("progress_hours").Default(0),
		field.Float("total_hours_required").Default(1),
		field.Int("dc").Default(10),
		field.String("status").Default("in-progress"),
		field.String("materials_allocated").Default("[]"),
		field.String("notes").Default(""),
		field.String("started_at").Default(""),
	}
}

func (CharacterCrafting) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("crafting").Field("character_id").Unique().Required(),
		edge.From("recipe", CraftingRecipe.Type).Ref("character_crafting").Field("recipe_id").Unique(),
	}
}

func (CharacterCrafting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
