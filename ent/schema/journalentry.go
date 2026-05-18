package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type JournalEntry struct {
	ent.Schema
}

func (JournalEntry) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("journal"),
	}
}

func (JournalEntry) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("title").Default(""),
		field.String("entry").Default(""),
		field.String("entry_date").Default(""),
		field.String("created_at").Default(""),
	}
}

func (JournalEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("journal").Field("character_id").Unique().Required(),
	}
}

func (JournalEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("character_id"),
	}
}
