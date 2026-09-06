package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CampaignKnowledgeKnownBy struct {
	ent.Schema
}

func (CampaignKnowledgeKnownBy) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("knowledge_id"),
		field.Int64("character_id"),
		field.String("created_at").Default(""),
	}
}

func (CampaignKnowledgeKnownBy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("knowledge", CampaignKnowledge.Type).Ref("known_by").Field("knowledge_id").Unique().Required(),
		edge.From("character", Character.Type).Ref("knowledge_known_by").Field("character_id").Unique().Required(),
	}
}

func (CampaignKnowledgeKnownBy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("knowledge_id", "character_id").Unique(),
		index.Fields("character_id"),
	}
}
