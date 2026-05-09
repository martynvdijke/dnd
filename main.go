package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers"
	"villum/middleware"
)

//go:embed static/*.html static/*.css static/style.css static/js/*.js
var staticFiles embed.FS

const Version = "1.0.0"

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "villum.db"
	}

	if err := db.Init(dbPath); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	db.Seed()
	handlers.SetDBPath(dbPath)
	middleware.StartCleanupTask()
	handlers.StartBackupScheduler()

	port := os.Getenv("PORT")
	if port == "" {
		port = "6280"
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

		// Rest & Level Up
		auth.POST("/characters/:id/rest", handlers.DoRest)
		auth.POST("/characters/:id/levelup", handlers.LevelUp)

		// Party view
		auth.GET("/party", handlers.GetPartyView)

		// Combat
		auth.GET("/combat", handlers.ListCombatEntries)
		auth.POST("/combat", handlers.CreateCombatEntry)
		auth.PUT("/combat/:id", handlers.UpdateCombatEntry)
		auth.DELETE("/combat/:id", handlers.DeleteCombatEntry)
		auth.POST("/combat/initiative", handlers.RollInitiative)
		auth.POST("/combat/next-turn", handlers.NextTurn)
		auth.GET("/combat/current-turn", handlers.GetCurrentTurn)

		// Generators
		auth.GET("/generate/npc", handlers.HandleGenerateNPC)
		auth.GET("/generate/name", handlers.HandleGenerateName)
		auth.GET("/generate/encounter", handlers.HandleGenerateEncounter)
		auth.GET("/generate/loot", handlers.HandleGenerateLoot)
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

		// Compendium CRUD
		admin.POST("/compendium/:type", handlers.AdminCreateCompendiumEntry)
		admin.PUT("/compendium/:type/:id", handlers.AdminUpdateCompendiumEntry)
		admin.DELETE("/compendium/:type/:id", handlers.AdminDeleteCompendiumEntry)
	}

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
	serveHTML("/app", "app.html")
	serveHTML("/admin", "admin.html")

	log.Printf("villum v%s starting on :%s", Version, port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
