package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type CompendiumEquipment struct {
	ent.Schema
}

func (CompendiumEquipment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("compendium_equipment"),
	}
}

func (CompendiumEquipment) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name"),
		field.String("category").Default(""),
		field.String("cost").Default("{}"),
		field.Float("weight").Default(0),
		field.String("description").Default(""),
		field.String("source_page").Default(""),
		field.String("system").Default("dnd5e"),
		field.String("source").Default("srd"),
	}
}
