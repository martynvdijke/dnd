package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type UploadLink struct {
	ent.Schema
}

func (UploadLink) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("upload_id"),
		field.String("entity_type"),
		field.Int64("entity_id"),
		field.String("field_name").Default(""),
		field.String("created_at").Default(""),
	}
}

func (UploadLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("upload_id", "entity_type", "entity_id").Unique(),
		index.Fields("entity_type", "entity_id"),
	}
}
