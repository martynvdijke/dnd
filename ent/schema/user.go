package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("username").Unique(),
		field.String("password"),
		field.String("display_name").Default(""),
		field.String("role").Default("user"),
		field.String("email").Default(""),
		field.String("created_at").Default(""),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("characters", Character.Type),
		edge.To("locations", Location.Type),
		edge.To("npcs", NPC.Type),
		edge.To("campaigns", Campaign.Type),
		edge.To("dice_rolls", DiceRoll.Type),
		edge.To("crafting_recipes", CraftingRecipe.Type),
		edge.To("shops", Shop.Type),
		edge.To("encounter_templates", EncounterTemplate.Type),
		edge.To("campaign_members", CampaignMember.Type),
		edge.To("wiki_pages", CampaignWikiPage.Type),
		edge.To("share_links", ShareLink.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").Unique(),
	}
}
