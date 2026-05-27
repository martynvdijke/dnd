package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Character struct {
	ent.Schema
}

func (Character) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id"),
		field.String("name").Default(""),
		field.String("race").Default(""),
		field.String("class").Default(""),
		field.String("subclass").Default(""),
		field.Int("level").Default(1),
		field.Int("xp").Default(0),
		field.String("background").Default(""),
		field.String("alignment").Default(""),
		field.Int("str").Default(10),
		field.Int("dex").Default(10),
		field.Int("con").Default(10),
		field.Int("int").Default(10),
		field.Int("wis").Default(10),
		field.Int("cha").Default(10),
		field.Int("ac").Default(10),
		field.Int("initiative").Default(0),
		field.Int("speed").Default(30),
		field.Int("hp_max").Default(10),
		field.Int("hp_current").Default(10),
		field.Int("temp_hp").Default(0),
		field.String("hit_dice").Default("1d10"),
		field.Int("hit_dice_current").Default(1),
		field.Int("proficiency_bonus").Default(2),
		field.Int("inspiration").Default(0),
		field.Int("passive_perception").Default(10),
		field.String("personality_traits").Default(""),
		field.String("ideals").Default(""),
		field.String("bonds").Default(""),
		field.String("flaws").Default(""),
		field.String("appearance").Default(""),
		field.String("backstory").Default(""),
		field.String("portrait_url").Default(""),
		field.String("dm_notes").Default(""),
		field.Int("hp_auto_calc").Default(0),
		field.Int("death_saves_successes").Default(0),
		field.Int("death_saves_failures").Default(0),
		field.Int("exhaustion_level").Default(0),
		field.String("concentrating_on").Default(""),
		field.Int64("campaign_id").Optional(),
		field.String("created_at").Default(""),
		field.String("updated_at").Default(""),
	}
}

func (Character) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("characters").Field("user_id").Unique().Required(),
		edge.To("currency", CharacterCurrency.Type),
		edge.To("proficiencies", CharacterProficiency.Type),
		edge.To("features", CharacterFeature.Type),
		edge.To("spellcasting", CharacterSpellcasting.Type),
		edge.To("spells", Spell.Type),
		edge.To("inventory", InventoryItem.Type),
		edge.To("classes", CharacterClass.Type),
		edge.To("conditions", CharacterCondition.Type),
		edge.To("feats", CharacterFeat.Type),
		edge.To("companions", Companion.Type),
		edge.To("notes", CharacterNote.Type),
		edge.To("resources", CharacterResource.Type),
		edge.To("crafting", CharacterCrafting.Type),
		edge.To("sessions", Session.Type),
		edge.To("quests", Quest.Type),
		edge.To("journal", JournalEntry.Type),
		edge.To("rest_logs", RestLog.Type),
		edge.To("downtime_activities", DowntimeActivity.Type),
		edge.To("level_up_plans", LevelUpPlan.Type),
		edge.To("character_locations", CharacterLocation.Type),
		edge.To("character_npcs", CharacterNPC.Type),
		edge.To("faction_reputations", FactionReputation.Type),
		edge.To("combat_entries", CombatEntry.Type),
	}
}

func (Character) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
