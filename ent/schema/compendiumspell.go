package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type CompendiumSpell struct {
	ent.Schema
}

func (CompendiumSpell) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name"),
		field.Int("level").Default(0),
		field.String("school").Default(""),
		field.String("casting_time").Default(""),
		field.String("range").Default(""),
		field.String("components").Default(""),
		field.String("duration").Default(""),
		field.String("description").Default(""),
		field.String("higher_levels").Default(""),
		field.String("classes").Default("[]"),
		field.String("source_page").Default(""),
		field.String("system").Default("dnd5e"),
		field.String("source").Default("srd"),
	}
}
