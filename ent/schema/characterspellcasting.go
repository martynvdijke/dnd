package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type CharacterSpellcasting struct {
	ent.Schema
}

func (CharacterSpellcasting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("character_spellcasting"),
	}
}

func (CharacterSpellcasting) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("character_id"),
		field.String("ability").Default(""),
		field.Int("save_dc").Default(10),
		field.Int("attack_bonus").Default(0),
		field.Int("slots_1_max").Default(0),
		field.Int("slots_1_used").Default(0),
		field.Int("slots_2_max").Default(0),
		field.Int("slots_2_used").Default(0),
		field.Int("slots_3_max").Default(0),
		field.Int("slots_3_used").Default(0),
		field.Int("slots_4_max").Default(0),
		field.Int("slots_4_used").Default(0),
		field.Int("slots_5_max").Default(0),
		field.Int("slots_5_used").Default(0),
		field.Int("slots_6_max").Default(0),
		field.Int("slots_6_used").Default(0),
		field.Int("slots_7_max").Default(0),
		field.Int("slots_7_used").Default(0),
		field.Int("slots_8_max").Default(0),
		field.Int("slots_8_used").Default(0),
		field.Int("slots_9_max").Default(0),
		field.Int("slots_9_used").Default(0),
	}
}

func (CharacterSpellcasting) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("character", Character.Type).Ref("spellcasting").Field("character_id").Unique().Required(),
	}
}
