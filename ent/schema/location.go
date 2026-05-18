package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Location struct {
	ent.Schema
}

func (Location) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id"),
		field.String("name"),
		field.String("type").Default("region"),
		field.String("description").Default(""),
		field.Int64("parent_id").Optional(),
		field.Float("latitude").Optional(),
		field.Float("longitude").Optional(),
		field.String("created_at").Default(""),
	}
}

func (Location) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("locations").Field("user_id").Unique().Required(),
		edge.To("children", Location.Type).From("parent").Field("parent_id").Unique(),
		edge.To("character_locations", CharacterLocation.Type),
	}
}

func (Location) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("parent_id"),
	}
}
