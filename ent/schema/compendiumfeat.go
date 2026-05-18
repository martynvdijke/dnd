package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type CompendiumFeat struct {
	ent.Schema
}

func (CompendiumFeat) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name").Unique(),
		field.String("description").Default(""),
		field.String("prerequisites").Default("[]"),
		field.String("source_page").Default(""),
		field.String("system").Default("dnd5e"),
		field.String("source").Default("srd"),
	}
}
