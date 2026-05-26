package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type NPC struct {
	ent.Schema
}

func (NPC) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("npcs"),
	}
}

func (NPC) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id"),
		field.String("name"),
		field.String("race").Default(""),
		field.String("class").Default(""),
		field.String("description").Default(""),
		field.String("notes").Default(""),
		field.Int("str").Default(10),
		field.Int("dex").Default(10),
		field.Int("con").Default(10),
		field.Int("int").Default(10),
		field.Int("wis").Default(10),
		field.Int("cha").Default(10),
		field.Int("hp_max").Default(10),
		field.Int("hp_current").Default(10),
		field.Bool("is_alive").Default(true),
		field.Bool("is_full").Default(false),
		field.Int("ac").Default(10),
		field.Int("speed").Default(30),
		field.String("skills").Default(""),
		field.String("saves").Default(""),
		field.String("features").Default(""),
		field.String("actions").Default(""),
		field.String("backstory").Default(""),
		field.String("created_at").Default(""),
	}
}

func (NPC) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("npcs").Field("user_id").Unique().Required(),
		edge.To("character_npcs", CharacterNPC.Type),
	}
}

func (NPC) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
