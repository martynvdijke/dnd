package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type OTelSetting struct {
	ent.Schema
}

func (OTelSetting) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("endpoint").Optional(),
		field.Bool("enabled").Default(false),
	}
}
