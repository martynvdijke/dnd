package handlers

import (
	"github.com/gin-gonic/gin"
)

// RegisterCharacterRoutes registers character CRUD and sub-resource routes.
func RegisterCharacterRoutes(r *gin.RouterGroup) {
	// Characters
	r.GET("/characters", ListCharacters)
	r.POST("/characters", CreateCharacter)
	r.GET("/characters/:id", GetCharacter)
	r.PUT("/characters/:id", UpdateCharacter)
	r.PUT("/characters/:id/dm-notes", UpdateCharacterDMNotes)
	r.DELETE("/characters/:id", DeleteCharacter)
	r.GET("/characters/:id/export", ExportCharacter)
	r.GET("/characters/:id/print", PrintCharacter)
	r.POST("/characters/import", ImportJSON)
	r.GET("/characters/compare", CompareCharacters)

	// Currency
	r.PUT("/characters/:id/currency", UpdateCurrency)

	// Spellcasting
	r.PUT("/characters/:id/spellcasting", UpdateSpellcasting)

	// Inventory sub-resource
	r.POST("/characters/:id/inventory", CreateInventory)
	r.PUT("/inventory/:iid", UpdateInventory)
	r.DELETE("/inventory/:iid", DeleteInventory)
	r.POST("/characters/:id/inventory/link", LinkCompendiumEquipment)
	r.DELETE("/characters/:id/inventory/:itemId/link", UnlinkCompendiumEquipment)
	r.POST("/characters/:id/inventory/:itemId/unlink", UnlinkCompendiumEquipment)

	// Race/class/background compendium linking
	r.POST("/characters/:id/race/link", linkCharacterRace)
	r.DELETE("/characters/:id/race/link", unlinkCharacterRace)
	r.POST("/characters/:id/class/link", linkCharacterClass)
	r.DELETE("/characters/:id/class/link", unlinkCharacterClass)
	r.POST("/characters/:id/background/link", linkCharacterBackground)
	r.DELETE("/characters/:id/background/link", unlinkCharacterBackground)

	// Spells sub-resource
	r.POST("/characters/:id/spells", CreateSpell)
	r.PUT("/spells/:sid", UpdateSpell)
	r.DELETE("/spells/:sid", DeleteSpell)
	r.POST("/characters/:id/spells/link", LinkCompendiumSpell)
	r.DELETE("/characters/:id/spells/:spellId/link", UnlinkCompendiumSpell)
	r.POST("/characters/:id/spells/:spellId/unlink", UnlinkCompendiumSpell)

	// Features sub-resource
	r.POST("/characters/:id/features", CreateFeature)
	r.PUT("/features/:fid", UpdateFeature)
	r.DELETE("/features/:fid", DeleteFeature)

	// Proficiencies
	r.POST("/proficiencies", CreateProficiency)
	r.DELETE("/proficiencies/:pid", DeleteProficiency)

	// Locations
	r.GET("/locations/search", SearchLocations)
	r.GET("/locations", ListLocations)
	r.POST("/locations", CreateLocation)
	r.PUT("/locations/:id", UpdateLocation)
	r.DELETE("/locations/:id", DeleteLocation)
	r.GET("/characters/:id/locations", GetCharacterLocations)
	r.POST("/characters/:id/locations", LinkLocation)
	r.DELETE("/locations/link/:lid", UnlinkLocation)

	// NPCs
	r.GET("/npcs/search", SearchNPCs)
	r.GET("/npcs", ListNPCs)
	r.POST("/npcs", CreateNPC)
	r.PUT("/npcs/:id", UpdateNPC)
	r.DELETE("/npcs/:id", DeleteNPC)
	r.GET("/characters/:id/npcs", GetCharacterNPCs)
	r.POST("/characters/:id/npcs", LinkNPC)
	r.DELETE("/npcs/link/:nid", UnlinkNPC)
	r.POST("/npcs/link/:nid/interact", LogNPCInteraction)

	// Sessions
	r.GET("/characters/:id/sessions", ListSessions)
	r.POST("/characters/:id/sessions", CreateSession)
	r.PUT("/sessions/:sid", UpdateSession)
	r.DELETE("/sessions/:sid", DeleteSession)

	// Quests
	r.GET("/characters/:id/quests", ListQuests)
	r.POST("/characters/:id/quests", CreateQuest)
	r.PUT("/quests/:qid", UpdateQuest)
	r.DELETE("/quests/:qid", DeleteQuest)

	// Journal
	r.GET("/characters/:id/journal", ListJournal)
	r.POST("/characters/:id/journal", CreateJournalEntry)
	r.PUT("/journal/:jid", UpdateJournalEntry)
	r.DELETE("/journal/:jid", DeleteJournalEntry)

	// Graph & Stats
	r.GET("/characters/:id/graph", GetGraphData)
	r.GET("/characters/:id/stats", GetCharacterStats)

	// Rest & Level Up
	r.POST("/characters/:id/rest", DoRest)
	r.POST("/characters/:id/levelup", LevelUp)

	// Exhaustion
	r.PATCH("/characters/:id/exhaustion", UpdateExhaustion)

	// Batch Spell Prep
	r.PUT("/characters/:id/spells/prepare", BatchPrepareSpells)

	// Multi-class
	r.POST("/characters/:id/classes", CreateCharacterClass)
	r.PUT("/classes/:ccid", UpdateCharacterClass)
	r.DELETE("/classes/:ccid", DeleteCharacterClass)

	// Conditions
	r.GET("/conditions", ListConditions)
	r.POST("/conditions", CreateCondition)
	r.PUT("/conditions/:id", UpdateCondition)
	r.DELETE("/conditions/:id", DeleteCondition)
	r.POST("/conditions/tick", TickConditions)
	r.GET("/conditions/types", GetConditionTypes)
	r.GET("/conditions/summary", GetActiveConditionSummary)

	// Concentration
	r.POST("/characters/:id/check-concentration", CheckConcentration)

	// Feats
	r.GET("/feats", ListFeats)
	r.POST("/feats", CreateFeat)
	r.PUT("/feats/:id", UpdateFeat)
	r.DELETE("/feats/:id", DeleteFeat)

	// Companions
	r.GET("/companions", ListCompanions)
	r.POST("/companions", CreateCompanion)
	r.PUT("/companions/:id", UpdateCompanion)
	r.DELETE("/companions/:id", DeleteCompanion)

	// Notes
	r.GET("/notes", ListCharacterNotes)
	r.POST("/notes", CreateCharacterNote)
	r.PUT("/notes/:id", UpdateCharacterNote)
	r.DELETE("/notes/:id", DeleteCharacterNote)

	// HP Auto-Calc
	r.POST("/characters/:id/calc-hp", CalculateHP)

	// Crafting
	r.GET("/crafting/recipes", ListCraftingRecipes)
	r.POST("/crafting/recipes", CreateCraftingRecipe)
	r.DELETE("/crafting/recipes/:id", DeleteCraftingRecipe)
	r.GET("/characters/:id/crafting", ListCharacterCrafting)
	r.POST("/characters/:id/crafting", CreateCharacterCrafting)
	r.PUT("/crafting/:id", UpdateCharacterCrafting)
	r.DELETE("/crafting/:id", DeleteCharacterCrafting)

	// Character Resources
	r.GET("/characters/:id/resources", ListCharacterResources)
	r.POST("/characters/:id/resources", CreateCharacterResource)
	r.PUT("/resources/:id", UpdateCharacterResource)
	r.DELETE("/resources/:id", DeleteCharacterResource)
	r.POST("/characters/:id/recover-resources", RecoverResourcesOnRest)

	// Downtime Activities
	r.GET("/characters/:id/downtime", ListDowntimeActivities)
	r.POST("/characters/:id/downtime", CreateDowntimeActivity)
	r.PUT("/downtime/:id", UpdateDowntimeActivity)
	r.DELETE("/downtime/:id", DeleteDowntimeActivity)
	r.POST("/downtime/:id/advance", AdvanceDowntimeDay)
	r.GET("/downtime/types", GetDowntimeTypes)

	// Level Up Planner
	r.GET("/characters/:id/level-plan", GetLevelUpPlan)
	r.POST("/characters/:id/level-plan", SaveLevelUpPlan)
	r.DELETE("/characters/:id/level-plan", DeleteLevelUpPlan)
	r.GET("/characters/:id/level-suggestions", GetLevelUpSuggestions)
}

// RegisterDMCharacterRoutes registers DM-only character routes.
func RegisterDMCharacterRoutes(r *gin.RouterGroup) {
	r.GET("/characters/all", ListAllCharacters)
}
