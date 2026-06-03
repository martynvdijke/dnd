package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"villum/db"
	"villum/handlers"
	"villum/middleware"
)

//go:embed static/*.html static/*.css static/style.css static/js/*.js static/sw.js static/manifest.json
var staticFiles embed.FS

const Version = "2.5.2"

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		if os.Getenv("DOCKER") == "true" {
			dbPath = "/db/villum.db"
		} else {
			dbPath = "villum.db"
		}
	}

	if err := db.Init(dbPath); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	// Media path setup
	mediaPath := os.Getenv("MEDIA_PATH")
	if mediaPath == "" {
		basePath := filepath.Dir(dbPath)
		if basePath == "." {
			basePath = "."
		}
		mediaPath = filepath.Join(basePath, "media")
	}
	if err := os.MkdirAll(mediaPath, 0755); err != nil {
		log.Printf("Warning: could not create media directory: %v", err)
	}
	handlers.SetMediaPath(mediaPath)

	db.Seed()
	handlers.SeedCraftingRecipes()
	handlers.SetDBPath(dbPath)
	middleware.StartCleanupTask()
	handlers.StartBackupScheduler()
	handlers.StartDBCleanupTask()

	port := os.Getenv("PORT")
	if port == "" {
		port = "6270"
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.SecurityHeaders())

	tp, err := initTelemetry()
	if err != nil {
		log.Printf("Failed to initialize telemetry: %v", err)
	} else {
		r.Use(metricsMiddleware())
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}()
	}

	// Public routes
	r.GET("/healthz", handlers.HandleHealth)
	r.GET("/metrics", handlers.HandleMetrics)
	r.GET("/metrics/prometheus", gin.WrapH(promhttp.Handler()))
	r.GET("/api/check-setup", handlers.CheckSetup)
	r.POST("/api/login", handlers.HandleLogin)

	// Authenticated routes
	auth := r.Group("/api")
	auth.Use(middleware.AuthRequired(), middleware.CSRFRequired())
	{
		auth.POST("/logout", handlers.HandleLogout)
		auth.GET("/csrf-token", handlers.GetCSRFToken)
		auth.GET("/user/me", handlers.GetMe)

		// Characters
		auth.GET("/characters", handlers.ListCharacters)
		auth.POST("/characters", handlers.CreateCharacter)
		auth.GET("/characters/:id", handlers.GetCharacter)
		auth.PUT("/characters/:id", handlers.UpdateCharacter)
		auth.DELETE("/characters/:id", handlers.DeleteCharacter)
		auth.GET("/characters/:id/export", handlers.ExportCharacter)
		auth.GET("/characters/:id/print", handlers.PrintCharacter)
		auth.POST("/characters/import", handlers.ImportJSON)

		// Currency
		auth.PUT("/characters/:id/currency", handlers.UpdateCurrency)

		// Spellcasting
		auth.PUT("/characters/:id/spellcasting", handlers.UpdateSpellcasting)

		// Inventory sub-resource (GET handled by GetCharacter full load)
		auth.POST("/characters/:id/inventory", handlers.CreateInventory)
		auth.PUT("/inventory/:iid", handlers.UpdateInventory)
		auth.DELETE("/inventory/:iid", handlers.DeleteInventory)

		// Spells sub-resource
		auth.POST("/characters/:id/spells", handlers.CreateSpell)
		auth.PUT("/spells/:sid", handlers.UpdateSpell)
		auth.DELETE("/spells/:sid", handlers.DeleteSpell)

		// Features sub-resource
		auth.POST("/characters/:id/features", handlers.CreateFeature)
		auth.PUT("/features/:fid", handlers.UpdateFeature)
		auth.DELETE("/features/:fid", handlers.DeleteFeature)

		// Proficiencies
		auth.POST("/proficiencies", handlers.CreateProficiency)
		auth.DELETE("/proficiencies/:pid", handlers.DeleteProficiency)

		// Dice
		auth.POST("/roll", handlers.HandleRoll)
		auth.POST("/roll/check", handlers.HandleCheckRoll)
		auth.GET("/dice-rolls", handlers.GetDiceRolls)

		// Compendium (read-only)
		auth.GET("/compendium/races", handlers.ListCompendiumRaces)
		auth.GET("/compendium/classes", handlers.ListCompendiumClasses)
		auth.GET("/compendium/spells", handlers.ListCompendiumSpells)
		auth.GET("/compendium/feats", handlers.ListCompendiumFeats)
		auth.GET("/compendium/backgrounds", handlers.ListCompendiumBackgrounds)
		auth.GET("/compendium/equipment", handlers.ListCompendiumEquipment)
		auth.GET("/compendium/search", handlers.SearchCompendium)
		auth.GET("/search", handlers.HandleSearch)

		// D&D 5e API fallback lookup
		auth.GET("/compendium/api/:category", handlers.FetchFromDnDApi)

		// Import from API
		auth.POST("/import/api", handlers.ImportFromAPI)

		// Backup
		auth.GET("/backup/settings", handlers.GetBackupSettings)
		auth.PUT("/backup/settings", handlers.SaveBackupSettings)
		auth.POST("/backup/trigger", handlers.TriggerBackup)
		auth.GET("/backup/list", handlers.ListBackups)

		// Locations
		auth.GET("/locations", handlers.ListLocations)
		auth.POST("/locations", handlers.CreateLocation)
		auth.PUT("/locations/:id", handlers.UpdateLocation)
		auth.DELETE("/locations/:id", handlers.DeleteLocation)
		// Character-Location links
		auth.GET("/characters/:id/locations", handlers.GetCharacterLocations)
		auth.POST("/characters/:id/locations", handlers.LinkLocation)
		auth.DELETE("/locations/link/:lid", handlers.UnlinkLocation)

		// NPCs
		auth.GET("/npcs", handlers.ListNPCs)
		auth.POST("/npcs", handlers.CreateNPC)
		auth.PUT("/npcs/:id", handlers.UpdateNPC)
		auth.DELETE("/npcs/:id", handlers.DeleteNPC)
		// Character-NPC links
		auth.GET("/characters/:id/npcs", handlers.GetCharacterNPCs)
		auth.POST("/characters/:id/npcs", handlers.LinkNPC)
		auth.DELETE("/npcs/link/:nid", handlers.UnlinkNPC)
		auth.POST("/npcs/link/:nid/interact", handlers.LogNPCInteraction)

		// Sessions
		auth.GET("/characters/:id/sessions", handlers.ListSessions)
		auth.POST("/characters/:id/sessions", handlers.CreateSession)
		auth.PUT("/sessions/:sid", handlers.UpdateSession)
		auth.DELETE("/sessions/:sid", handlers.DeleteSession)

		// Quests
		auth.GET("/characters/:id/quests", handlers.ListQuests)
		auth.POST("/characters/:id/quests", handlers.CreateQuest)
		auth.PUT("/quests/:qid", handlers.UpdateQuest)
		auth.DELETE("/quests/:qid", handlers.DeleteQuest)

		// Journal
		auth.GET("/characters/:id/journal", handlers.ListJournal)
		auth.POST("/characters/:id/journal", handlers.CreateJournalEntry)
		auth.PUT("/journal/:jid", handlers.UpdateJournalEntry)
		auth.DELETE("/journal/:jid", handlers.DeleteJournalEntry)

		// Graph
		auth.GET("/characters/:id/graph", handlers.GetGraphData)

		// Stats
		auth.GET("/characters/:id/stats", handlers.GetCharacterStats)

		// Campaigns
		auth.GET("/campaigns", handlers.ListCampaigns)
		auth.POST("/campaigns", handlers.CreateCampaign)
		auth.PUT("/campaigns/:id", handlers.UpdateCampaign)
		auth.DELETE("/campaigns/:id", handlers.DeleteCampaign)
		auth.GET("/campaigns/:id/members", handlers.ListCampaignMembers)
		auth.POST("/campaigns/:id/members", handlers.AddCampaignMember)
		auth.PUT("/campaigns/:id/members/:userId", handlers.SetCampaignMemberRole)
		auth.DELETE("/campaigns/:id/members/:userId", handlers.RemoveCampaignMember)

		// Party Inventory (campaign shared items)
		auth.GET("/campaigns/:id/party-items", handlers.ListCampaignPartyItems)
		auth.POST("/campaigns/:id/party-items", handlers.CreateCampaignPartyItem)
		auth.DELETE("/party-items/:id", handlers.DeleteCampaignPartyItem)

		// Session Plans
		auth.GET("/campaigns/:id/session-plans", handlers.ListSessionPlans)
		auth.POST("/campaigns/:id/session-plans", handlers.CreateSessionPlan)
		auth.PUT("/session-plans/:id", handlers.UpdateSessionPlan)
		auth.DELETE("/session-plans/:id", handlers.DeleteSessionPlan)

		// Rest & Level Up
		auth.POST("/characters/:id/rest", handlers.DoRest)
		auth.POST("/characters/:id/levelup", handlers.LevelUp)

		// Exhaustion
		auth.PATCH("/characters/:id/exhaustion", handlers.UpdateExhaustion)

		// Batch Spell Prep
		auth.PUT("/characters/:id/spells/prepare", handlers.BatchPrepareSpells)

		// Party view
		auth.GET("/party", handlers.GetPartyView)
		auth.GET("/users/search", handlers.SearchUsers)

		// Combat
		auth.GET("/combat", handlers.ListCombatEntries)
		auth.POST("/combat", handlers.CreateCombatEntry)
		auth.PUT("/combat/:id", handlers.UpdateCombatEntry)
		auth.DELETE("/combat/:id", handlers.DeleteCombatEntry)
		auth.POST("/combat/initiative", handlers.RollInitiative)
		auth.POST("/combat/next-turn", handlers.NextTurn)
		auth.GET("/combat/current-turn", handlers.GetCurrentTurn)

		// Shops
		auth.GET("/shops", handlers.ListShops)
		auth.GET("/shops/:id/items", handlers.ListShopItems)
		auth.POST("/shops/:id/buy", handlers.BuyItem)
		auth.POST("/shops/:id/sell", handlers.SellItem)
		auth.GET("/shop-transactions", handlers.ListShopTransactions)

		// Wiki
		auth.GET("/campaigns/:id/wiki", handlers.ListWikiPages)
		auth.POST("/campaigns/:id/wiki", handlers.CreateWikiPage)
		auth.GET("/wiki/:id", handlers.GetWikiPage)
		auth.PUT("/wiki/:id", handlers.UpdateWikiPage)
		auth.DELETE("/wiki/:id", handlers.DeleteWikiPage)

		// Campaign graph
		auth.GET("/campaigns/:id/graph", handlers.GetCampaignGraphData)

		// Multi-class
		auth.POST("/characters/:id/classes", handlers.CreateCharacterClass)
		auth.PUT("/classes/:ccid", handlers.UpdateCharacterClass)
		auth.DELETE("/classes/:ccid", handlers.DeleteCharacterClass)

		// Encounter Builder
		auth.GET("/encounters", handlers.ListEncounters)
		auth.POST("/encounters", handlers.CreateEncounter)
		auth.GET("/encounters/:id", handlers.GetEncounter)
		auth.PUT("/encounters/:id", handlers.UpdateEncounter)
		auth.DELETE("/encounters/:id", handlers.DeleteEncounter)
		auth.POST("/encounters/:id/monsters", handlers.AddEncounterMonster)
		auth.PUT("/encounter-monsters/:mid", handlers.UpdateEncounterMonster)
		auth.DELETE("/encounter-monsters/:mid", handlers.DeleteEncounterMonster)
		auth.POST("/encounters/calculate-xp", handlers.CalculateEncounterXP)
		auth.GET("/monster-xp", handlers.GetMonsterXP)

		// Pregenerated Characters
		auth.GET("/pregens", handlers.ListPregens)
		auth.POST("/pregens", handlers.CreatePregen)
		auth.GET("/pregens/generate", handlers.GeneratePregen)
		auth.GET("/pregens/balance", handlers.CheckPartyBalance)
		auth.GET("/pregens/:id", handlers.GetPregen)
		auth.PUT("/pregens/:id", handlers.UpdatePregen)
		auth.DELETE("/pregens/:id", handlers.DeletePregen)

		// Timeline
		auth.GET("/timeline", handlers.ListTimelineEvents)
		auth.POST("/timeline", handlers.CreateTimelineEvent)
		auth.PUT("/timeline/:id", handlers.UpdateTimelineEvent)
		auth.DELETE("/timeline/:id", handlers.DeleteTimelineEvent)

		// Conditions / Ailments
		auth.GET("/conditions", handlers.ListConditions)
		auth.POST("/conditions", handlers.CreateCondition)
		auth.PUT("/conditions/:id", handlers.UpdateCondition)
		auth.DELETE("/conditions/:id", handlers.DeleteCondition)
		auth.POST("/conditions/tick", handlers.TickConditions)
		auth.GET("/conditions/types", handlers.GetConditionTypes)
		auth.GET("/conditions/summary", handlers.GetActiveConditionSummary)

		// Concentration
		auth.POST("/characters/:id/check-concentration", handlers.CheckConcentration)

		// Feats
		auth.GET("/feats", handlers.ListFeats)
		auth.POST("/feats", handlers.CreateFeat)
		auth.PUT("/feats/:id", handlers.UpdateFeat)
		auth.DELETE("/feats/:id", handlers.DeleteFeat)

		// Companions
		auth.GET("/companions", handlers.ListCompanions)
		auth.POST("/companions", handlers.CreateCompanion)
		auth.PUT("/companions/:id", handlers.UpdateCompanion)
		auth.DELETE("/companions/:id", handlers.DeleteCompanion)

		// Factions & Reputation
		auth.GET("/factions", handlers.ListFactions)
		auth.POST("/factions", handlers.CreateFaction)
		auth.PUT("/factions/:id", handlers.UpdateFaction)
		auth.DELETE("/factions/:id", handlers.DeleteFaction)
		auth.GET("/faction-reputation", handlers.GetFactionReputations)
		auth.POST("/faction-reputation", handlers.SetFactionReputation)
		auth.DELETE("/faction-reputation/:id", handlers.DeleteFactionReputation)

		// Weather
		auth.GET("/generate/weather", handlers.HandleGenerateWeather)

		// Notes
		auth.GET("/notes", handlers.ListCharacterNotes)
		auth.POST("/notes", handlers.CreateCharacterNote)
		auth.PUT("/notes/:id", handlers.UpdateCharacterNote)
		auth.DELETE("/notes/:id", handlers.DeleteCharacterNote)

		// HP Auto-Calc
		auth.POST("/characters/:id/calc-hp", handlers.CalculateHP)

		// Crafting
		auth.GET("/crafting/recipes", handlers.ListCraftingRecipes)
		auth.POST("/crafting/recipes", handlers.CreateCraftingRecipe)
		auth.DELETE("/crafting/recipes/:id", handlers.DeleteCraftingRecipe)
		auth.GET("/characters/:id/crafting", handlers.ListCharacterCrafting)
		auth.POST("/characters/:id/crafting", handlers.CreateCharacterCrafting)
		auth.PUT("/crafting/:id", handlers.UpdateCharacterCrafting)
		auth.DELETE("/crafting/:id", handlers.DeleteCharacterCrafting)

		// Character Comparison
		auth.GET("/characters/compare", handlers.CompareCharacters)

		// Generators
		auth.GET("/generate/npc", handlers.HandleGenerateNPC)
		auth.GET("/generate/name", handlers.HandleGenerateName)
		auth.GET("/generate/encounter", handlers.HandleGenerateEncounter)
		auth.GET("/generate/loot", handlers.HandleGenerateLoot)
		auth.GET("/generate/character", handlers.HandleGenerateRandomCharacter)
		auth.GET("/generate/adventure-hook", handlers.HandleGenerateAdventureHook)
		auth.GET("/generate/dungeon-dressing", handlers.HandleGenerateDungeonDressing)
		auth.GET("/generate/tavern", handlers.HandleGenerateTavern)
		auth.GET("/generate/urban-encounter", handlers.HandleGenerateUrbanEncounter)
		auth.GET("/generate/road-encounter", handlers.HandleGenerateRoadEncounter)

		// Media upload
		auth.POST("/upload", handlers.HandleUpload)
		auth.GET("/uploads", handlers.GetUploads)
		auth.POST("/uploads/:id/crop", handlers.HandleCropUpload)
		auth.POST("/upload-links", handlers.CreateUploadLink)
		auth.DELETE("/upload-links/:id", handlers.DeleteUploadLink)

		// Share links
		auth.POST("/share", handlers.CreateShareLink)
		auth.GET("/share", handlers.ListShareLinks)
		auth.DELETE("/share/:token", handlers.DeleteShareLink)

		// Campaign Dashboard
		auth.GET("/campaigns/:id/dashboard", handlers.GetCampaignDashboard)
		auth.GET("/campaigns/:id/one-shots", handlers.ListCampaignOneShots)
		auth.PUT("/campaigns/:id/one-shots/reorder", handlers.ReorderCampaignOneShots)

		// Character Resources
		auth.GET("/characters/:id/resources", handlers.ListCharacterResources)
		auth.POST("/characters/:id/resources", handlers.CreateCharacterResource)
		auth.PUT("/resources/:id", handlers.UpdateCharacterResource)
		auth.DELETE("/resources/:id", handlers.DeleteCharacterResource)
		auth.POST("/characters/:id/recover-resources", handlers.RecoverResourcesOnRest)

		// Combat Log
		auth.GET("/combat-log", handlers.ListCombatLogEntries)
		auth.POST("/combat-log", handlers.CreateCombatLogEntry)
		auth.GET("/combat-log/stats", handlers.GetCombatLogStats)

		// Homebrew Content Manager
		auth.GET("/homebrew/:type", handlers.ListHomebrewContent)
		auth.POST("/homebrew/:type", handlers.CreateHomebrewContent)
		auth.PUT("/homebrew/:type/:id", handlers.UpdateHomebrewContent)
		auth.DELETE("/homebrew/:type/:id", handlers.DeleteHomebrewContent)

		// Campaign Maps
		auth.GET("/campaigns/:id/maps", handlers.ListCampaignMaps)
		auth.POST("/campaigns/:id/maps", handlers.CreateCampaignMap)
		auth.PUT("/maps/:id", handlers.UpdateCampaignMap)
		auth.DELETE("/maps/:id", handlers.DeleteCampaignMap)
		auth.POST("/campaigns/:id/maps/:mapId/activate", handlers.SetActiveCampaignMap)
		auth.PUT("/maps/:id/fog", handlers.UpdateFogOfWar)
		auth.GET("/campaigns/:id/maps/active", handlers.GetActiveCampaignMap)
		auth.GET("/maps/:mapId/pins", handlers.ListMapPins)
		auth.POST("/maps/:mapId/pins", handlers.CreateMapPin)
		auth.PUT("/map-pins/:id", handlers.UpdateMapPin)
		auth.DELETE("/map-pins/:id", handlers.DeleteMapPin)

		// Quick Reference
		auth.GET("/quickref", handlers.GetQuickReference)

		// Downtime Activities
		auth.GET("/characters/:id/downtime", handlers.ListDowntimeActivities)
		auth.POST("/characters/:id/downtime", handlers.CreateDowntimeActivity)
		auth.PUT("/downtime/:id", handlers.UpdateDowntimeActivity)
		auth.DELETE("/downtime/:id", handlers.DeleteDowntimeActivity)
		auth.POST("/downtime/:id/advance", handlers.AdvanceDowntimeDay)
		auth.GET("/downtime/types", handlers.GetDowntimeTypes)

		// Level Up Planner
		auth.GET("/characters/:id/level-plan", handlers.GetLevelUpPlan)
		auth.POST("/characters/:id/level-plan", handlers.SaveLevelUpPlan)
		auth.DELETE("/characters/:id/level-plan", handlers.DeleteLevelUpPlan)
		auth.GET("/characters/:id/level-suggestions", handlers.GetLevelUpSuggestions)

		// Campaign Recaps
		auth.GET("/campaigns/:id/recaps", handlers.ListCampaignRecaps)
		auth.POST("/campaigns/:id/recaps", handlers.CreateCampaignRecap)
		auth.GET("/recaps/:id", handlers.GetCampaignRecap)
		auth.PUT("/recaps/:id", handlers.UpdateCampaignRecap)
		auth.DELETE("/recaps/:id", handlers.DeleteCampaignRecap)
		auth.POST("/campaigns/:id/recaps/generate", handlers.GenerateCampaignRecap)
		auth.POST("/recaps/:id/mark-sent", handlers.MarkRecapAsSent)

	}

	// DM / One-Shot routes (DM or admin only)
	dm := r.Group("/api")
	dm.Use(middleware.AuthRequired(), middleware.DMRequired(), middleware.CSRFRequired())
	{
		// One-Shot Adventures
		dm.GET("/oneshot-adventures", handlers.ListOneShotAdventures)
		dm.POST("/oneshot-adventures", handlers.CreateOneShotAdventure)
		dm.GET("/oneshot-adventures/:id", handlers.GetOneShotAdventure)
		dm.PUT("/oneshot-adventures/:id", handlers.UpdateOneShotAdventure)
		dm.DELETE("/oneshot-adventures/:id", handlers.DeleteOneShotAdventure)
		dm.POST("/oneshot-adventures/generate", handlers.GenerateOneShotFromTemplate)
		dm.POST("/oneshot-adventures/:id/acts", handlers.CreateOneShotAct)
		dm.GET("/oneshot-adventures/:id/npcs", handlers.GetOneShotNPCs)
		dm.POST("/oneshot-adventures/:id/npcs", handlers.LinkOneShotNPC)
		dm.DELETE("/oneshot-adventures/:id/npcs/:nid", handlers.UnlinkOneShotNPC)
		dm.GET("/oneshot-adventures/:id/locations", handlers.GetOneShotLocations)
		dm.POST("/oneshot-adventures/:id/locations", handlers.LinkOneShotLocation)
		dm.DELETE("/oneshot-adventures/:id/locations/:lid", handlers.UnlinkOneShotLocation)
		dm.GET("/oneshot-adventures/:id/encounters", handlers.GetOneShotEncounters)
		dm.POST("/oneshot-adventures/:id/encounters", handlers.LinkOneShotEncounter)
		dm.DELETE("/oneshot-adventures/:id/encounters/:eid", handlers.UnlinkOneShotEncounter)

		// Acts & Scenes
		dm.PUT("/oneshot-acts/:id", handlers.UpdateOneShotAct)
		dm.DELETE("/oneshot-acts/:id", handlers.DeleteOneShotAct)
		dm.POST("/oneshot-acts/:id/scenes", handlers.CreateOneShotScene)
		dm.PUT("/oneshot-scenes/:id", handlers.UpdateOneShotScene)
		dm.DELETE("/oneshot-scenes/:id", handlers.DeleteOneShotScene)

		// Act-level NPCs
		dm.GET("/oneshot-acts/:id/npcs", handlers.ListActNPCs)
		dm.POST("/oneshot-acts/:id/npcs", handlers.CreateActNPC)
		dm.DELETE("/oneshot-acts/:id/npcs/:nid", handlers.DeleteActNPC)

		// Act-level Notes
		dm.GET("/oneshot-acts/:id/notes", handlers.ListActNotes)
		dm.POST("/oneshot-acts/:id/notes", handlers.CreateActNote)

		// Act-level Details (HTMX)
		dm.GET("/htmx/oneshot-acts/:id/details", handlers.HtmxActDetails)

		// Session Pacing
		dm.POST("/oneshot-adventures/:id/pacing/start", handlers.StartPacingSession)
		dm.GET("/oneshot-adventures/:id/pacing", handlers.GetPacingSession)
		dm.GET("/session-pacing/:id", handlers.GetPacingSession)
		dm.POST("/session-pacing/:id/pause", handlers.PausePacingSession)
		dm.POST("/session-pacing/:id/resume", handlers.ResumePacingSession)
		dm.POST("/session-pacing/:id/next-scene", handlers.AdvanceToNextScene)
		dm.POST("/session-pacing/:id/complete", handlers.CompletePacingSession)
		dm.POST("/session-pacing/:id/tick", handlers.UpdatePacingTimers)

		// Clue/Mystery Tracker
		dm.GET("/oneshot-adventures/:id/clues", handlers.ListClues)
		dm.POST("/oneshot-adventures/:id/clues", handlers.CreateClue)
		dm.GET("/clues/:id", handlers.GetClue)
		dm.PUT("/clues/:id", handlers.UpdateClue)
		dm.DELETE("/clues/:id", handlers.DeleteClue)
		dm.POST("/clues/:id/reveal", handlers.RevealClue)
		dm.POST("/clues/:id/hide", handlers.HideClue)
		dm.POST("/clues/:id/dependencies", handlers.AddClueDependency)
		dm.DELETE("/clues/:id/dependencies/:did", handlers.RemoveClueDependency)
		dm.POST("/clues/:id/npcs", handlers.LinkClueNPC)
		dm.DELETE("/clues/:id/npcs/:nid", handlers.UnlinkClueNPC)
		dm.POST("/clues/:id/locations", handlers.LinkClueLocation)
		dm.DELETE("/clues/:id/locations/:lid", handlers.UnlinkClueLocation)

		// Prep Dashboard & Checklist
		dm.GET("/oneshot-adventures/:id/checklist", handlers.ListPrepChecklist)
		dm.POST("/oneshot-adventures/:id/checklist", handlers.CreatePrepChecklistItem)
		dm.PUT("/prep-checklist/:cid", handlers.UpdatePrepChecklistItem)
		dm.DELETE("/prep-checklist/:cid", handlers.DeletePrepChecklistItem)

		// DM Screen / Quick Reference
		dm.GET("/oneshot-adventures/:id/dm-screen", handlers.HtmxDmScreen)
		dm.GET("/oneshot-adventures/:id/notes", handlers.ListDmNotes)
		dm.POST("/oneshot-adventures/:id/notes", handlers.CreateDmNote)
		dm.PUT("/dm-notes/:nid", handlers.UpdateDmNote)
		dm.DELETE("/oneshot-adventures/:id/notes/:nid", handlers.DeleteDmNote)

		// Characters all (DM can see all user characters)
		dm.GET("/characters/all", handlers.ListAllCharacters)

		// Parties
		dm.GET("/parties", handlers.ListParties)
		dm.POST("/parties", handlers.CreateParty)
		dm.GET("/parties/:id", handlers.GetParty)
		dm.PUT("/parties/:id", handlers.UpdateParty)
		dm.DELETE("/parties/:id", handlers.DeleteParty)
		dm.GET("/parties/:id/factions", handlers.ListPartyFactions)
		dm.POST("/parties/:id/factions", handlers.CreatePartyFaction)
		dm.GET("/parties/:id/uploads", handlers.ListPartyUploads)

		// One-Shot Items
		dm.GET("/oneshot-adventures/:id/items", handlers.ListOneShotItems)
		dm.POST("/oneshot-adventures/:id/items", handlers.CreateOneShotItem)
		dm.PUT("/oneshot-items/:id", handlers.UpdateOneShotItem)
		dm.DELETE("/oneshot-items/:id", handlers.DeleteOneShotItem)
		dm.GET("/oneshot-items/:id/uploads", handlers.ListItemUploads)
		dm.GET("/oneshot-items/:id/npcs", handlers.ListNPCsForItem)

		// NPC-Item Links
		dm.POST("/oneshot-adventures/:id/npc-item-links", handlers.CreateNPCItemLink)
		dm.GET("/oneshot-adventures/:id/npcs/:nid/items", handlers.ListItemsForNPC)
		dm.DELETE("/npc-item-links/:id", handlers.DeleteNPCItemLink)

		// One-Shot Shops
		dm.GET("/oneshot-adventures/:id/shops", handlers.ListOneShotShops)
		dm.POST("/oneshot-adventures/:id/shops", handlers.CreateOneShotShop)
		dm.POST("/oneshot-adventures/:id/shops/:shopId/items", handlers.CreateOneShotShopItem)
		dm.DELETE("/oneshot-adventures/:id/shops/:shopId", handlers.DeleteOneShotShop)

		// One-Shot Monsters
		dm.GET("/oneshot-acts/:id/monsters", handlers.ListActMonsters)
		dm.POST("/oneshot-acts/:id/monsters", handlers.CreateActMonster)
		dm.PUT("/oneshot-monsters/:id", handlers.UpdateOneShotMonster)
		dm.DELETE("/oneshot-monsters/:id", handlers.DeleteOneShotMonster)
		dm.GET("/oneshot-scenes/:id/monsters", handlers.ListSceneMonsters)
		dm.POST("/oneshot-scenes/:id/monsters", handlers.CreateSceneMonster)

		// AI Generation
		dm.POST("/ai/generate/text", handlers.HandleTextGeneration)
		dm.POST("/ai/generate/image", handlers.HandleImageGeneration)

		// Monster Library
		dm.GET("/monster-library", handlers.ListMonsterLibrary)
		dm.POST("/monster-library", handlers.CreateMonsterLibraryEntry)
		dm.PUT("/monster-library/:id", handlers.UpdateMonsterLibraryEntry)
		dm.DELETE("/monster-library/:id", handlers.DeleteMonsterLibraryEntry)

		// Monster Compendium
		dm.GET("/compendium-monsters", handlers.ListCompendiumMonsters)
		dm.GET("/compendium-monsters/:id", handlers.GetCompendiumMonster)

		// Monster Import
		dm.POST("/oneshot-adventures/:id/import/compendium", handlers.ImportCompendiumMonsterToOneShot)
		dm.POST("/encounters/:id/import/compendium", handlers.ImportCompendiumMonsterToEncounter)
		dm.POST("/oneshot-adventures/:id/import/library", handlers.ImportLibraryMonsterToOneShot)

		// Campaign NPCs
		dm.GET("/campaigns/:id/npcs", handlers.ListCampaignNPCs)
		dm.POST("/campaigns/:id/npcs", handlers.LinkNPCCampaign)
		dm.PUT("/campaign-npcs/:id", handlers.UpdateCampaignNPC)
		dm.DELETE("/campaign-npcs/:id", handlers.UnlinkCampaignNPC)
		dm.POST("/campaigns/:id/npcs/create-and-link", handlers.CreateAndLinkCampaignNPC)

		// Campaign Encounter Monsters
		dm.GET("/encounters/:id/monsters", handlers.ListCampaignEncounterMonsters)

		// Linked Player Characters
		dm.GET("/oneshot-adventures/:id/characters", handlers.ListLinkedCharacters)
		dm.POST("/oneshot-adventures/:id/characters", handlers.LinkCharacterToOneShot)
		dm.DELETE("/oneshot-adventures/:id/characters/:charId", handlers.UnlinkCharacterFromOneShot)

		// Inline editing
		dm.PATCH("/oneshot-acts/:id/duration", handlers.UpdateActDuration)
		dm.PATCH("/oneshot-scenes/:id/duration", handlers.UpdateSceneDuration)
		dm.PUT("/oneshot-adventures/:id/acts/reorder", handlers.ReorderActs)
		dm.PUT("/oneshot-acts/:id/scenes/reorder", handlers.ReorderScenes)

		// Scene Dialogs
		dm.PUT("/oneshot-scenes/:id/dialogs/reorder", handlers.ReorderDialogs)
	}

	// htmx endpoints (separate group, outside /api, with auth + CSRF)
	htmxGroup := r.Group("/")
	htmxGroup.Use(middleware.AuthRequired(), middleware.CSRFRequired())
	handlers.HtmxRegisterRoutes(htmxGroup)

	// Public share routes (no auth required)
	r.GET("/api/share/:token", handlers.GetSharedEntity)

	// WebSocket (auth required, but no CSRF)
	ws := r.Group("/api")
	ws.Use(middleware.AuthRequired())
	{
		ws.GET("/ws", handlers.HandleWebSocket)
	}

	// Admin routes
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(), middleware.AdminRequired(), middleware.CSRFRequired())
	{
		admin.GET("/users", handlers.AdminListUsers)
		admin.POST("/users", handlers.AdminCreateUser)
		admin.PUT("/users/:id", handlers.AdminUpdateUser)
		admin.DELETE("/users/:id", handlers.AdminDeleteUser)
		admin.PUT("/users/:id/password", handlers.AdminResetPassword)

		// Email settings
		admin.GET("/email-settings", handlers.GetEmailSettings)
		admin.POST("/email-settings", handlers.SaveEmailSettings)
		admin.POST("/test-email", handlers.TestEmail)
		admin.POST("/campaign-highlights", handlers.SendCampaignHighlights)

		// Shop management (admin only)
		admin.POST("/shops", handlers.CreateShop)
		admin.PUT("/shops/:id", handlers.UpdateShop)
		admin.DELETE("/shops/:id", handlers.DeleteShop)
		admin.POST("/shops/:id/items", handlers.CreateShopItem)
		admin.DELETE("/shop-items/:id", handlers.DeleteShopItem)

		// AI Endpoint management
		admin.GET("/ai-endpoints", handlers.ListAIEndpoints)
		admin.GET("/ai-endpoints/:id", handlers.GetAIEndpoint)
		admin.POST("/ai-endpoints", handlers.CreateAIEndpoint)
		admin.PUT("/ai-endpoints/:id", handlers.UpdateAIEndpoint)
		admin.DELETE("/ai-endpoints/:id", handlers.DeleteAIEndpoint)
		admin.POST("/ai-endpoints/:id/test", handlers.TestAIEndpoint)

		// Compendium CRUD
		admin.POST("/compendium/:type", handlers.AdminCreateCompendiumEntry)
		admin.PUT("/compendium/:type/:id", handlers.AdminUpdateCompendiumEntry)
		admin.DELETE("/compendium/:type/:id", handlers.AdminDeleteCompendiumEntry)
	}

	// Serve uploaded media files
	r.Static("/media", mediaPath)

	// Static file serving for non-API routes
	staticFS, _ := fs.Sub(staticFiles, "static")
	r.StaticFS("/static", http.FS(staticFS))

	// Serve HTML pages with version substitution
	serveHTML := func(path, fileName string) {
		r.GET(path, func(c *gin.Context) {
			data, err := fs.ReadFile(staticFS, fileName)
			if err != nil {
				c.String(http.StatusNotFound, "not found")
				return
			}
			content := strings.ReplaceAll(string(data), "{{VERSION}}", Version)
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
		})
	}

	serveHTML("/", "app.html")
	serveHTML("/login", "login.html")
	serveHTML("/setup", "setup.html")
	serveHTML("/admin", "admin.html")

	// Redirect /app to / for backward compatibility
	r.GET("/app", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	log.Printf("villum v%s starting on :%s", Version, port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
