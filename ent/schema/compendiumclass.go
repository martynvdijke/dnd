package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type CompendiumClass struct {
	ent.Schema
}

func (CompendiumClass) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name").Unique(),
		field.String("description").Default(""),
		field.Int("hit_die").Default(8),
		field.String("primary_ability").Default(""),
		field.String("saving_throws").Default("[]"),
		field.String("proficiencies").Default("{}"),
		field.String("spellcasting_ability").Default(""),
		field.String("source_page").Default(""),
		field.String("system").Default("dnd5e"),
		field.String("source").Default("srd"),
	}
}
