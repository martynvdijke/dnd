package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CharacterNPC struct {
	ent.Schema
}

func (CharacterNPC) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.Int64("npc_id"),
		field.String("relationship").Default("acquaintance"),
		field.String("notes").Default(""),
		field.Int("interaction_count").Default(0),
		field.String("last_interacted").Default(""),
	}
}

func (CharacterNPC) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("character_npcs").Field("character_id").Unique().Required(),
		edge.From("npc", NPC.Type).Ref("character_npcs").Field("npc_id").Unique().Required(),
	}
}

func (CharacterNPC) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
		index.Fields("npc_id"),
	}
}
