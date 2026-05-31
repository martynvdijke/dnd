package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OneShotItem struct {
	ent.Schema
}

func (OneShotItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("oneshot_items"),
	}
}

func (OneShotItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("adventure_id"),
		field.Int64("act_id").Optional(),
		field.String("name").Default(""),
		field.String("description").Default(""),
		field.String("category").Default("gear"),
		field.Int("quantity").Default(1),
		field.Float("weight").Default(0),
		field.Float("price_gp").Default(0),
		field.Bool("is_magical").Default(false),
		field.Bool("attunement").Default(false),
		field.String("notes").Default(""),
		field.String("created_at").Default(""),
	}
}

func (OneShotItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("act", OneShotAct.Type).Ref("items").Field("act_id").Unique(),
	}
}

func (OneShotItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("adventure_id"),
		index.Fields("act_id"),
	}
}
