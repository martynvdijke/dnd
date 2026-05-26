package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Upload struct {
	ent.Schema
}

func (Upload) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("hash").Unique().Optional(),
		field.String("ext").Optional(),
		field.String("url").Optional(),
		field.String("resized_url").Optional(),
		field.String("thumbnail_url").Optional(),
		field.String("owner_type").Default(""),
		field.Int64("owner_id").Default(0),
		field.String("created_at").Default(""),
	}
}
