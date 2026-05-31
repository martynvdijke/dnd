package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OneShotAdventureEncounter struct {
	ent.Schema
}

func (OneShotAdventureEncounter) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("oneshot_adventure_encounters"),
	}
}

func (OneShotAdventureEncounter) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("adventure_id"),
		field.Int64("act_id").Optional(),
		field.Int64("encounter_id"),
	}
}

func (OneShotAdventureEncounter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("act", OneShotAct.Type).Ref("encounters").Field("act_id").Unique(),
	}
}

func (OneShotAdventureEncounter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("adventure_id"),
		index.Fields("act_id"),
	}
}
