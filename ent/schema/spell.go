package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Spell struct {
	ent.Schema
}

func (Spell) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("name"),
		field.Int("level").Default(0),
		field.String("school").Default(""),
		field.String("casting_time").Default(""),
		field.String("range").Default(""),
		field.String("components").Default(""),
		field.String("duration").Default(""),
		field.String("description").Default(""),
		field.Bool("prepared").Default(false),
		field.Bool("always_prepared").Default(false),
		field.String("source").Default(""),
		field.String("notes").Default(""),
		field.Int64("compendium_spell_id").Optional().Nillable(),
	}
}

func (Spell) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("spells").Field("character_id").Unique().Required(),
	}
}

func (Spell) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
