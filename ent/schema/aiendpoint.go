package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type AIEndpoint struct {
	ent.Schema
}

func (AIEndpoint) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.String("name").Unique(),
		field.Enum("type").Values("text", "image"),
		field.String("base_url"),
		field.String("encrypted_api_key"),
		field.String("model"),
		field.Strings("tags").Optional(),
		field.Bool("enabled").Default(true),
		field.Float("temperature").Optional().Nillable(),
		field.Int("max_tokens").Optional().Nillable(),
		field.String("image_size").Optional().Nillable(),
		field.String("created_at"),
		field.String("updated_at"),
	}
}
