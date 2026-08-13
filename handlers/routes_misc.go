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
	r.GET("/share/:token", GetSharedPage)
	r.GET("/events", EventsPage)
	r.GET("/events/ical", EventsICal)
	r.GET("/htmx/events/list", EventsListPartial)
	r.GET("/htmx/events/grid", EventsGridPartial)
	r.GET("/events/c/:slug", CampaignEventsPage)
	r.GET("/events/c/:slug/ical", CampaignEventsICal)
	r.GET("/htmx/events/c/:slug/list", CampaignEventsListPartial)
	r.GET("/htmx/events/c/:slug/grid", EventsGridCampaignPartial)
	// Event detail route (after campaign/c routes since literal 'ical' matches first)
	r.GET("/events/:id", EventDetail)

	// TRMNL e-ink display polling endpoints (token-guarded, read-only)
	r.GET("/api/trmnl/character-stats", GetTRMNLCharacterStats)
	r.GET("/api/trmnl/campaign-stats", GetTRMNLCampaignStats)
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
	r.GET("/uploads/entity/:type/:id", GetUploadsForEntity)

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

// einkActive reports whether e-ink mode should be enabled for this request:
// explicit ?eink=1 param (persisted as a cookie), existing eink cookie, or the
// site-wide admin setting.
func einkActive(c *gin.Context) bool {
	if c.Query("eink") == "1" {
		c.SetCookie("eink", "1", 31536000, "/", "", false, true)
		return true
	}
	if c.Query("eink") == "0" {
		c.SetCookie("eink", "", -1, "/", "", false, true)
		// Fall through: still honor an existing site-wide setting.
	}
	if v, err := c.Cookie("eink"); err == nil && v == "1" {
		return true
	}
	var value string
	if err := db.DB.QueryRow("SELECT value FROM app_settings WHERE key = 'eink'").Scan(&value); err == nil && value == "1" {
		return true
	}
	return false
}

// RegisterStaticRoutes registers static file serving and HTML page routes.
func RegisterStaticRoutes(r *gin.Engine, embedFS embed.FS, mediaPath string, Version string) {
	// Serve uploaded media files
	r.Static("/media", mediaPath)

	// Static file serving for non-API routes
	staticFS, _ := fs.Sub(embedFS, "static")
	r.StaticFS("/static", http.FS(staticFS))

	// Service worker at root scope so it can control the whole app (offline support)
	r.GET("/sw.js", func(c *gin.Context) {
		data, err := fs.ReadFile(staticFS, "sw.js")
		if err != nil {
			c.String(http.StatusNotFound, "not found")
			return
		}
		c.Data(http.StatusOK, "application/javascript", data)
	})

	// Serve HTML pages with version substitution and optional analytics injection
	serveHTML := func(path, fileName string, pageType string) {
		r.GET(path, func(c *gin.Context) {
			data, err := fs.ReadFile(staticFS, fileName)
			if err != nil {
				c.String(http.StatusNotFound, "not found")
				return
			}
			content := strings.ReplaceAll(string(data), "{{VERSION}}", Version)

			// E-ink mode: add class to the <html> element
			if einkActive(c) {
				content = strings.Replace(content, `<html lang="en" data-theme="light">`, `<html lang="en" data-theme="light" class="eink">`, 1)
			}

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

		// E-ink mode: add class to the <html> element
		if einkActive(c) {
			content = strings.Replace(content, `<html lang="en" data-theme="light">`, `<html lang="en" data-theme="light" class="eink">`, 1)
		}

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
