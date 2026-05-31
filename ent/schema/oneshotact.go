package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OneShotAct struct {
	ent.Schema
}

func (OneShotAct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("oneshot_acts"),
	}
}

func (OneShotAct) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("adventure_id"),
		field.Int64("parent_act_id").Optional(),
		field.Int("number").Default(1),
		field.Int("sort_order").Default(0),
		field.String("title").Default(""),
		field.String("description").Default(""),
		field.Int("estimated_minutes").Default(30),
		field.String("notes").Default(""),
	}
}

func (OneShotAct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("adventure", OneShotAdventure.Type).Ref("acts").Field("adventure_id").Unique().Required(),
		edge.To("scenes", OneShotScene.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("children", OneShotAct.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("parent", OneShotAct.Type).Ref("children").Field("parent_act_id").Unique(),
		edge.To("items", OneShotItem.Type),
		edge.To("encounters", OneShotAdventureEncounter.Type),
	}
}

func (OneShotAct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("adventure_id"),
		index.Fields("parent_act_id"),
	}
}
