package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type FactionReputation struct {
	ent.Schema
}

func (FactionReputation) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.Int64("faction_id"),
		field.Int("standing").Default(0),
		field.String("rank").Default(""),
		field.String("notes").Default(""),
	}
}

func (FactionReputation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("faction_reputations").Field("character_id").Unique().Required(),
		edge.From("faction", Faction.Type).Ref("reputations").Field("faction_id").Unique().Required(),
	}
}

func (FactionReputation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
		index.Fields("faction_id"),
	}
}
