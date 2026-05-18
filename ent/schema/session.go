package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Session struct {
	ent.Schema
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("session_date").Default(""),
		field.String("title").Default(""),
		field.String("notes").Default(""),
		field.Int("xp_earned").Default(0),
		field.Int("gold_earned").Default(0),
		field.String("important_events").Default(""),
		field.String("created_at").Default(""),
	}
}

func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("sessions").Field("character_id").Unique().Required(),
	}
}

func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
