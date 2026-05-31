package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OneShotScene struct {
	ent.Schema
}

func (OneShotScene) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("oneshot_scenes"),
	}
}

func (OneShotScene) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("act_id"),
		field.Int("number").Default(1),
		field.Int("sort_order").Default(0),
		field.String("title").Default(""),
		field.String("description").Default(""),
		field.String("scene_type").Default("roleplay"),
		field.Int64("location_id").Optional(),
		field.Int64("encounter_id").Optional(),
		field.Int("estimated_minutes").Default(15),
		field.String("notes").Default(""),
	}
}

func (OneShotScene) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("act", OneShotAct.Type).Ref("scenes").Field("act_id").Unique().Required(),
	}
}

func (OneShotScene) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("act_id"),
	}
}
