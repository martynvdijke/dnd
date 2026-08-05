package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type OneShotAdventure struct {
	ent.Schema
}

func (OneShotAdventure) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("oneshot_adventures"),
	}
}

func (OneShotAdventure) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id"),
		field.Int64("campaign_id").Optional(),
		field.String("title").Default(""),
		field.String("premise").Default(""),
		field.String("hook").Default(""),
		field.String("template").Default("custom"),
		field.Int("estimated_minutes").Default(180),
		field.String("difficulty").Default("medium"),
		field.String("notes").Default(""),
		field.Bool("is_mini_campaign").Default(false),
		field.Int("sort_order").Default(0),
		field.String("created_at").Default(""),
		field.String("updated_at").Default(""),
	}
}

func (OneShotAdventure) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("acts", OneShotAct.Type),
	}
}
