package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"villum/db"
	"villum/ent/user"
	"villum/handlers"
	"villum/middleware"
)

//go:embed static/*.html static/*.css static/style.css static/js/*.js static/sw.js static/manifest.json
var staticFiles embed.FS

const Version = "2.29.5"

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
	handlers.SeedCompendiumSchemas()
	handlers.SeedCraftingRecipes()
	db.SeedDefaultEventsSettings()
	handlers.SetAppVersion(Version)
	handlers.SetBaseURL(os.Getenv("BASE_URL"))

	// AUTO_SETUP=true creates the admin user automatically (used for per-worker Playwright isolation).
	if os.Getenv("AUTO_SETUP") == "true" {
		ctx := context.Background()
		count, err := db.Client.User.Query().Where(user.Role("admin")).Count(ctx)
		if err == nil && count == 0 {
			hash, err := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)
			if err == nil {
				_, err = db.Client.User.Create().
					SetUsername("admin").
					SetPassword(string(hash)).
					SetDisplayName("Admin").
					SetRole("admin").
					Save(ctx)
				if err != nil {
					log.Printf("Warning: AUTO_SETUP failed to create admin: %v", err)
				} else {
					log.Println("AUTO_SETUP: admin user created")
				}
			}
		}
	}

	handlers.SetDBPath(dbPath)

	// Replace in-memory session store with SQLite-backed persistence
	middleware.Store = middleware.NewDBSessionStore(db.DB)
	middleware.StartCleanupTask()
	handlers.StartBackupScheduler()
	handlers.StartDBCleanupTask()

	port := os.Getenv("PORT")
	if port == "" {
		port = "6270"
	}

	r := gin.New()
	r.Use(middleware.RequestLogger(), gin.Recovery())
	r.Use(middleware.SecurityHeaders())

	// Initialize structured application logger
	middleware.InitAppLoggerFromEnv()
	handlers.InitLogSettings()
	if middleware.AppLog != nil {
		middleware.AppLog.Info("system", "AppLogger initialized")
	}

	tp, promExporter, lp, err := initTelemetry()
	if err != nil {
		log.Printf("Warning: telemetry init failed (%v), running without OTel", err)
	}
	if tp != nil {
		// otelgin creates spans for all incoming requests
		r.Use(otelgin.Middleware("villum"))
		defer func() {
			shutdownTelemetry(tp, lp)
		}()

		// Wire OTel log bridge — sends structured logs via the OTel Logs SDK
		if lp != nil && middleware.AppLog != nil {
			otelHandler := otelslog.NewHandler("villum", otelslog.WithLoggerProvider(lp))
			middleware.AppLog.Handler().SetExportFn(func(ctx context.Context, r slog.Record) {
				_ = otelHandler.Handle(ctx, r)
			})
			middleware.AppLog.Info("system", "OTel log bridge enabled")
		}
	}
	// Initialize OTel metrics middleware (records both Prometheus and OTel metrics)
	om := initOTelMetrics()
	r.Use(newOTelMetricsMiddleware(om))
	_ = promExporter // OTel Prometheus exporter is registered with the metrics pipeline

	// Public routes
	handlers.RegisterPublicRoutes(r)

	// WebSocket (auth required, no CSRF)
	ws := r.Group("/api")
	ws.Use(middleware.AuthRequired())
	handlers.RegisterWebSocketRoutes(ws)

	// HTMX endpoints (auth + CSRF)
	htmxGroup := r.Group("/")
	htmxGroup.Use(middleware.AuthRequired(), middleware.CSRFRequired())
	handlers.HtmxRegisterRoutes(htmxGroup)

	// Authenticated routes
	auth := r.Group("/api")
	auth.Use(middleware.AuthRequired(), middleware.CSRFRequired())
	{
		handlers.RegisterAuthBoilerplate(auth)
		handlers.RegisterCharacterRoutes(auth)
		handlers.RegisterCombatRoutes(auth)
		handlers.RegisterEncounterRoutes(auth)
		handlers.RegisterCompendiumRoutes(auth)
		handlers.RegisterCampaignRoutes(auth)
		handlers.RegisterShopRoutes(auth)
		handlers.RegisterBackupRoutes(auth)
		handlers.RegisterMiscAuthRoutes(auth)
		handlers.RegisterTransferRoutes(auth)
	}

	// DM / One-Shot routes (DM or admin only)
	dm := r.Group("/api")
	dm.Use(middleware.AuthRequired(), middleware.DMRequired(), middleware.CSRFRequired())
	{
		handlers.RegisterDMCharacterRoutes(dm)
		handlers.RegisterDMCompendiumRoutes(dm)
		handlers.RegisterDMOneShotRoutes(dm)
		handlers.RegisterDMCampaignRoutes(dm)
		handlers.RegisterDMARoutes(dm)
	}

	// Admin routes
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(), middleware.AdminRequired(), middleware.CSRFRequired())
	{
		handlers.RegisterAdminRoutes(admin)
		handlers.RegisterAdminAIRoutes(admin)
		handlers.RegisterAdminCompendiumRoutes(admin)
	}

	// Static & HTML page serving
	handlers.RegisterStaticRoutes(r, staticFiles, mediaPath, Version)

	log.Printf("villum v%s starting on :%s", Version, port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
