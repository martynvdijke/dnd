package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Companion struct {
	ent.Schema
}

func (Companion) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("name"),
		field.String("type").Default("companion"),
		field.String("race").Default(""),
		field.Int("hp_max").Default(1),
		field.Int("hp_current").Default(1),
		field.Int("ac").Default(10),
		field.Int("str").Default(10),
		field.Int("dex").Default(10),
		field.Int("con").Default(10),
		field.Int("int").Default(10),
		field.Int("wis").Default(10),
		field.Int("cha").Default(10),
		field.Int("speed").Default(30),
		field.String("abilities").Default(""),
		field.String("notes").Default(""),
		field.String("portrait_url").Default(""),
		field.Bool("is_alive").Default(true),
		field.String("created_at").Default(""),
	}
}

func (Companion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("companions").Field("character_id").Unique().Required(),
	}
}

func (Companion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
