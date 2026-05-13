package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers"
	"villum/middleware"
)

//go:embed static/*.html static/*.css static/style.css static/js/*.js
var staticFiles embed.FS

const Version = "1.7.0"

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

	// Public routes
	r.GET("/healthz", handlers.HandleHealth)
	r.GET("/metrics", handlers.HandleMetrics)
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

		// Rest & Level Up
		auth.POST("/characters/:id/rest", handlers.DoRest)
		auth.POST("/characters/:id/levelup", handlers.LevelUp)

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

		// Calendar
		auth.GET("/calendar", handlers.ListCalendarEvents)
		auth.POST("/calendar", handlers.CreateCalendarEvent)
		auth.PUT("/calendar/:id", handlers.UpdateCalendarEvent)
		auth.DELETE("/calendar/:id", handlers.DeleteCalendarEvent)

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

		// Media upload
		auth.POST("/upload", handlers.HandleUpload)
		auth.GET("/uploads", handlers.GetUploads)
	}

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
