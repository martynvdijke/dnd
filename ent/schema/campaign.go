package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Campaign struct {
	ent.Schema
}

func (Campaign) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id"),
		field.String("name"),
		field.String("party_name").Default(""),
		field.String("description").Default(""),
		field.String("dm_notes").Default(""),
		field.String("created_at").Default(""),
	}
}

func (Campaign) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("campaigns").Field("user_id").Unique().Required(),
		edge.To("members", CampaignMember.Type),
		edge.To("calendar_events", CampaignCalendarEvent.Type),
		edge.To("timeline_events", CampaignTimelineEvent.Type),
		edge.To("wiki_pages", CampaignWikiPage.Type),
		edge.To("encounter_templates", EncounterTemplate.Type),
		edge.To("maps", CampaignMap.Type),
		edge.To("recaps", CampaignRecap.Type),
		edge.To("combat_entries", CombatEntry.Type),
		edge.To("shops", Shop.Type),
		edge.To("factions", Faction.Type),
		edge.To("combat_log_entries", CombatLogEntry.Type),
		edge.To("party_items", PartyItem.Type),
		edge.To("session_plans", SessionPlan.Type),
	}
}
