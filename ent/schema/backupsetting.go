package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type BackupSetting struct {
	ent.Schema
}

func (BackupSetting) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.Bool("enabled").Default(true),
		field.Int("interval_hours").Default(168),
		field.Int("interval_days").Default(7),
		field.Int("keep_count").Default(7),
		field.String("last_backup").Default(""),
	}
}
