package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type DowntimeActivity struct {
	ent.Schema
}

func (DowntimeActivity) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("activity_type"),
		field.String("name"),
		field.String("description").Default(""),
		field.Int("dc").Default(10),
		field.Int("days_required").Default(10),
		field.Int("days_completed").Default(0),
		field.Float("cost_per_day").Default(0),
		field.Float("total_cost").Default(0),
		field.String("reward").Default(""),
		field.String("status").Default("in-progress"),
		field.String("notes").Default(""),
		field.String("created_at").Default(""),
		field.String("updated_at").Default(""),
	}
}

func (DowntimeActivity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("downtime_activities").Field("character_id").Unique().Required(),
	}
}

func (DowntimeActivity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
