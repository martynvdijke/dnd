package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Quest struct {
	ent.Schema
}

func (Quest) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("name"),
		field.String("description").Default(""),
		field.String("status").Default("active"),
		field.String("objectives").Default(""),
		field.String("rewards").Default(""),
		field.String("notes").Default(""),
		field.String("created_at").Default(""),
		field.String("updated_at").Default(""),
	}
}

func (Quest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("quests").Field("character_id").Unique().Required(),
	}
}

func (Quest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
