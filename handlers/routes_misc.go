package handlers

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"villum/db"
	"villum/ent/user"
)

// RegisterPublicRoutes registers non-authenticated routes.
func RegisterPublicRoutes(r *gin.Engine) {
	r.GET("/healthz", HandleHealth)
	r.GET("/metrics", HandleMetrics)
	r.GET("/api/check-setup", CheckSetup)
	r.POST("/api/login", HandleLogin)
	r.GET("/api/share/:token", GetSharedEntity)
}

// RegisterAuthBoilerplate registers auth helper routes (logout, csrf, me).
func RegisterAuthBoilerplate(r *gin.RouterGroup) {
	r.POST("/logout", HandleLogout)
	r.GET("/csrf-token", GetCSRFToken)
	r.GET("/user/me", GetMe)
}

// RegisterBackupRoutes registers backup settings routes.
func RegisterBackupRoutes(r *gin.RouterGroup) {
	r.GET("/backup/settings", GetBackupSettings)
	r.PUT("/backup/settings", SaveBackupSettings)
	r.POST("/backup/trigger", TriggerBackup)
	r.GET("/backup/list", ListBackups)
}

// RegisterMiscAuthRoutes registers miscellaneous auth-protected routes.
func RegisterMiscAuthRoutes(r *gin.RouterGroup) {
	// Dice
	r.POST("/roll", HandleRoll)
	r.POST("/roll/check", HandleCheckRoll)
	r.GET("/dice-rolls", GetDiceRolls)

	// Search
	r.GET("/search", HandleSearch)

	// Import from API
	r.POST("/import/api", ImportFromAPI)

	// Factions & Reputation
	r.GET("/factions", ListFactions)
	r.POST("/factions", CreateFaction)
	r.PUT("/factions/:id", UpdateFaction)
	r.DELETE("/factions/:id", DeleteFaction)
	r.GET("/faction-reputation", GetFactionReputations)
	r.POST("/faction-reputation", SetFactionReputation)
	r.DELETE("/faction-reputation/:id", DeleteFactionReputation)

	// Pregenerated Characters
	r.GET("/pregens", ListPregens)
	r.POST("/pregens", CreatePregen)
	r.GET("/pregens/generate", GeneratePregen)
	r.GET("/pregens/balance", CheckPartyBalance)
	r.GET("/pregens/:id", GetPregen)
	r.PUT("/pregens/:id", UpdatePregen)
	r.DELETE("/pregens/:id", DeletePregen)

	// Generators
	r.GET("/generate/npc", HandleGenerateNPC)
	r.GET("/generate/name", HandleGenerateName)
	r.GET("/generate/encounter", HandleGenerateEncounter)
	r.GET("/generate/loot", HandleGenerateLoot)
	r.GET("/generate/character", HandleGenerateRandomCharacter)
	r.GET("/generate/adventure-hook", HandleGenerateAdventureHook)
	r.GET("/generate/dungeon-dressing", HandleGenerateDungeonDressing)
	r.GET("/generate/tavern", HandleGenerateTavern)
	r.GET("/generate/urban-encounter", HandleGenerateUrbanEncounter)
	r.GET("/generate/road-encounter", HandleGenerateRoadEncounter)
	r.GET("/generate/weather", HandleGenerateWeather)

	// Race Colors
	r.GET("/race-colors", ListRaceColors)
	r.PUT("/race-colors", UpdateRaceColors)

	// Uploads
	r.POST("/upload", HandleUpload)
	r.GET("/uploads", GetUploads)
	r.POST("/uploads/:id/crop", HandleCropUpload)
	r.POST("/upload-links", CreateUploadLink)
	r.DELETE("/upload-links/:id", DeleteUploadLink)

	// Share links
	r.POST("/share", CreateShareLink)
	r.GET("/share", ListShareLinks)
	r.DELETE("/share/:token", DeleteShareLink)

	// Homebrew Content
	r.GET("/homebrew/:type", ListHomebrewContent)
	r.POST("/homebrew/:type", CreateHomebrewContent)
	r.PUT("/homebrew/:type/:id", UpdateHomebrewContent)
	r.DELETE("/homebrew/:type/:id", DeleteHomebrewContent)

	// Quick Reference
	r.GET("/quickref", GetQuickReference)
}

// RegisterStaticRoutes registers static file serving and HTML page routes.
func RegisterStaticRoutes(r *gin.Engine, embedFS embed.FS, mediaPath string, Version string) {
	// Serve uploaded media files
	r.Static("/media", mediaPath)

	// Static file serving for non-API routes
	staticFS, _ := fs.Sub(embedFS, "static")
	r.StaticFS("/static", http.FS(staticFS))

	// Serve HTML pages with version substitution and optional analytics injection
	serveHTML := func(path, fileName string, pageType string) {
		r.GET(path, func(c *gin.Context) {
			data, err := fs.ReadFile(staticFS, fileName)
			if err != nil {
				c.String(http.StatusNotFound, "not found")
				return
			}
			content := strings.ReplaceAll(string(data), "{{VERSION}}", Version)

			// Conditionally inject Umami analytics script
			if pageType == "app" || pageType == "login" || pageType == "admin" {
				script := InjectUmamiScript()
				if script != "" {
					if pageType == "admin" {
						var enableAdminTracking int
						db.DB.QueryRow("SELECT COALESCE(enable_admin_tracking, 0) FROM umami_settings WHERE id = 1").Scan(&enableAdminTracking)
						if enableAdminTracking != 1 {
							script = ""
						}
					}
					if script != "" {
						content = strings.ReplaceAll(content, "</head>", script+"\n</head>")
					}
				}
			}

			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
		})
	}

	serveHTML("/", "app.html", "app")

	// Login page: redirect to /setup if no admin user exists
	r.GET("/login", func(c *gin.Context) {
		count, err := db.Client.User.Query().Where(user.Role("admin")).Count(c.Request.Context())
		if err == nil && count == 0 {
			c.Redirect(http.StatusTemporaryRedirect, "/setup")
			return
		}
		data, err := fs.ReadFile(staticFS, "login.html")
		if err != nil {
			c.String(http.StatusNotFound, "not found")
			return
		}
		content := strings.ReplaceAll(string(data), "{{VERSION}}", Version)
		script := InjectUmamiScript()
		if script != "" {
			content = strings.ReplaceAll(content, "</head>", script+"\n</head>")
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
	})

	serveHTML("/setup", "setup.html", "setup")
	serveHTML("/admin", "admin.html", "admin")

	// Redirect /app to /
	r.GET("/app", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})
}

// RegisterWebSocketRoutes registers WebSocket routes.
func RegisterWebSocketRoutes(r *gin.RouterGroup) {
	r.GET("/ws", HandleWebSocket)
}
