package handlers

import "github.com/gin-gonic/gin"

// RegisterDMOneShotRoutes registers DM-only one-shot adventure routes.
func RegisterDMOneShotRoutes(r *gin.RouterGroup) {
	// One-Shot Adventures
	r.GET("/oneshot-adventures", ListOneShotAdventures)
	r.POST("/oneshot-adventures", CreateOneShotAdventure)
	r.GET("/oneshot-adventures/:id", GetOneShotAdventure)
	r.PUT("/oneshot-adventures/:id", UpdateOneShotAdventure)
	r.DELETE("/oneshot-adventures/:id", DeleteOneShotAdventure)
	r.POST("/oneshot-adventures/generate", GenerateOneShotFromTemplate)
	r.POST("/oneshot-adventures/:id/acts", CreateOneShotAct)
	r.GET("/oneshot-adventures/:id/npcs", GetOneShotNPCs)
	r.POST("/oneshot-adventures/:id/npcs", LinkOneShotNPC)
	r.DELETE("/oneshot-adventures/:id/npcs/:nid", UnlinkOneShotNPC)
	r.GET("/oneshot-adventures/:id/locations", GetOneShotLocations)
	r.POST("/oneshot-adventures/:id/locations", LinkOneShotLocation)
	r.DELETE("/oneshot-adventures/:id/locations/:lid", UnlinkOneShotLocation)
	r.GET("/oneshot-adventures/:id/encounters", GetOneShotEncounters)
	r.POST("/oneshot-adventures/:id/encounters", LinkOneShotEncounter)
	r.DELETE("/oneshot-adventures/:id/encounters/:eid", UnlinkOneShotEncounter)

	// Acts & Scenes
	r.PUT("/oneshot-acts/:id", UpdateOneShotAct)
	r.DELETE("/oneshot-acts/:id", DeleteOneShotAct)
	r.POST("/oneshot-acts/:id/scenes", CreateOneShotScene)
	r.PUT("/oneshot-scenes/:id", UpdateOneShotScene)
	r.DELETE("/oneshot-scenes/:id", DeleteOneShotScene)

	// Act-level NPCs
	r.GET("/oneshot-acts/:id/npcs", ListActNPCs)
	r.POST("/oneshot-acts/:id/npcs", CreateActNPC)
	r.DELETE("/oneshot-acts/:id/npcs/:nid", DeleteActNPC)

	// Act-level Notes
	r.GET("/oneshot-acts/:id/notes", ListActNotes)
	r.POST("/oneshot-acts/:id/notes", CreateActNote)

	// Act-level Details (HTMX)
	r.GET("/htmx/oneshot-acts/:id/details", HtmxActDetails)

	// Session Pacing
	r.POST("/oneshot-adventures/:id/pacing/start", StartPacingSession)
	r.GET("/oneshot-adventures/:id/pacing", GetPacingSession)
	r.GET("/session-pacing/:id", GetPacingSession)
	r.POST("/session-pacing/:id/pause", PausePacingSession)
	r.POST("/session-pacing/:id/resume", ResumePacingSession)
	r.POST("/session-pacing/:id/next-scene", AdvanceToNextScene)
	r.POST("/session-pacing/:id/complete", CompletePacingSession)
	r.POST("/session-pacing/:id/tick", UpdatePacingTimers)

	// Clue/Mystery Tracker
	r.GET("/oneshot-adventures/:id/clues", ListClues)
	r.POST("/oneshot-adventures/:id/clues", CreateClue)
	r.GET("/clues/:id", GetClue)
	r.PUT("/clues/:id", UpdateClue)
	r.DELETE("/clues/:id", DeleteClue)
	r.POST("/clues/:id/reveal", RevealClue)
	r.POST("/clues/:id/hide", HideClue)
	r.POST("/clues/:id/dependencies", AddClueDependency)
	r.DELETE("/clues/:id/dependencies/:did", RemoveClueDependency)
	r.POST("/clues/:id/npcs", LinkClueNPC)
	r.DELETE("/clues/:id/npcs/:nid", UnlinkClueNPC)
	r.POST("/clues/:id/locations", LinkClueLocation)
	r.DELETE("/clues/:id/locations/:lid", UnlinkClueLocation)

	// Prep Checklist
	r.GET("/oneshot-adventures/:id/checklist", ListPrepChecklist)
	r.POST("/oneshot-adventures/:id/checklist", CreatePrepChecklistItem)
	r.PUT("/prep-checklist/:cid", UpdatePrepChecklistItem)
	r.DELETE("/prep-checklist/:cid", DeletePrepChecklistItem)

	// DM Screen / Notes
	r.GET("/oneshot-adventures/:id/dm-screen", HtmxDmScreen)
	r.GET("/oneshot-adventures/:id/notes", ListDmNotes)
	r.POST("/oneshot-adventures/:id/notes", CreateDmNote)
	r.PUT("/dm-notes/:nid", UpdateDmNote)
	r.DELETE("/oneshot-adventures/:id/notes/:nid", DeleteDmNote)

	// Parties
	r.GET("/parties", ListParties)
	r.POST("/parties", CreateParty)
	r.GET("/parties/:id", GetParty)
	r.PUT("/parties/:id", UpdateParty)
	r.DELETE("/parties/:id", DeleteParty)
	r.GET("/parties/:id/factions", ListPartyFactions)
	r.POST("/parties/:id/factions", CreatePartyFaction)
	r.GET("/parties/:id/uploads", ListPartyUploads)

	// One-Shot Items
	r.GET("/oneshot-adventures/:id/items", ListOneShotItems)
	r.POST("/oneshot-adventures/:id/items", CreateOneShotItem)
	r.PUT("/oneshot-items/:id", UpdateOneShotItem)
	r.DELETE("/oneshot-items/:id", DeleteOneShotItem)
	r.GET("/oneshot-items/:id/uploads", ListItemUploads)
	r.GET("/oneshot-items/:id/npcs", ListNPCsForItem)

	// NPC-Item Links
	r.POST("/oneshot-adventures/:id/npc-item-links", CreateNPCItemLink)
	r.GET("/oneshot-adventures/:id/npcs/:nid/items", ListItemsForNPC)
	r.DELETE("/npc-item-links/:id", DeleteNPCItemLink)

	// One-Shot Shops
	r.GET("/oneshot-adventures/:id/shops", ListOneShotShops)
	r.POST("/oneshot-adventures/:id/shops", CreateOneShotShop)
	r.POST("/oneshot-adventures/:id/shops/:shopId/items", CreateOneShotShopItem)
	r.DELETE("/oneshot-adventures/:id/shops/:shopId", DeleteOneShotShop)

	// One-Shot Monsters
	r.GET("/oneshot-acts/:id/monsters", ListActMonsters)
	r.POST("/oneshot-acts/:id/monsters", CreateActMonster)
	r.PUT("/oneshot-monsters/:id", UpdateOneShotMonster)
	r.DELETE("/oneshot-monsters/:id", DeleteOneShotMonster)
	r.GET("/oneshot-scenes/:id/monsters", ListSceneMonsters)
	r.POST("/oneshot-scenes/:id/monsters", CreateSceneMonster)
	r.POST("/oneshot-adventures/:id/acts/:aid/monsters/link", LinkCompendiumMonsterToAct)
	r.POST("/oneshot-adventures/:id/scenes/:sid/monsters/link", LinkCompendiumMonsterToScene)
	r.DELETE("/oneshot-monsters/:id/link", UnlinkCompendiumMonster)

	// Monster Library
	r.GET("/monster-library", ListMonsterLibrary)
	r.POST("/monster-library", CreateMonsterLibraryEntry)
	r.PUT("/monster-library/:id", UpdateMonsterLibraryEntry)
	r.DELETE("/monster-library/:id", DeleteMonsterLibraryEntry)

	// Monster Import
	r.POST("/oneshot-adventures/:id/import/compendium", ImportCompendiumMonsterToOneShot)
	r.POST("/encounters/:id/import/compendium", ImportCompendiumMonsterToEncounter)
	r.POST("/oneshot-adventures/:id/import/library", ImportLibraryMonsterToOneShot)
	r.POST("/encounters/:id/import/compendium-entry", ImportCompendiumEntryToEncounter)
	r.POST("/oneshot-adventures/:id/import/compendium-entry", ImportCompendiumEntryToOneShot)

	// Linked Player Characters
	r.GET("/oneshot-adventures/:id/characters", ListLinkedCharacters)
	r.POST("/oneshot-adventures/:id/characters", LinkCharacterToOneShot)
	r.DELETE("/oneshot-adventures/:id/characters/:charId", UnlinkCharacterFromOneShot)

	// Inline editing
	r.PATCH("/oneshot-acts/:id/duration", UpdateActDuration)
	r.PATCH("/oneshot-scenes/:id/duration", UpdateSceneDuration)
	r.PUT("/oneshot-adventures/:id/acts/reorder", ReorderActs)
	r.PUT("/oneshot-acts/:id/scenes/reorder", ReorderScenes)
	r.PUT("/oneshot-scenes/:id/dialogs/reorder", ReorderDialogs)
}
