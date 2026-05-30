package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OneShotActNPC struct {
	ent.Schema
}

func (OneShotActNPC) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("oneshot_act_npcs"),
	}
}

func (OneShotActNPC) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("act_id"),
		field.Int64("npc_id").Optional(),
		field.String("name").Default(""),
		field.String("role").Default(""),
		field.String("notes").Default(""),
		field.Bool("is_inline").Default(true),
		field.String("created_at").Default(""),
	}
}

func (OneShotActNPC) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("act_id"),
		index.Fields("npc_id"),
	}
}
