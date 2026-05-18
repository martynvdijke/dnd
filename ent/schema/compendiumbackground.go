package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type CompendiumBackground struct {
	ent.Schema
}

func (CompendiumBackground) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name").Unique(),
		field.String("description").Default(""),
		field.String("feature_name").Default(""),
		field.String("feature_description").Default(""),
		field.String("proficiencies").Default("{}"),
		field.String("source_page").Default(""),
		field.String("system").Default("dnd5e"),
		field.String("source").Default("srd"),
	}
}
