package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type CompendiumRace struct {
	ent.Schema
}

func (CompendiumRace) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name").Unique(),
		field.String("description").Default(""),
		field.Int("speed").Default(30),
		field.String("size").Default("Medium"),
		field.String("ability_bonuses").Default("{}"),
		field.String("traits").Default("{}"),
		field.String("languages").Default(""),
		field.String("source_page").Default(""),
		field.String("system").Default("dnd5e"),
		field.String("source").Default("srd"),
	}
}
