package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"villum/handlers"
	"villum/middleware"
)

// setupRouter creates the Gin engine, attaches middleware, and registers
// all route groups. mediaPath is used for static file serving.
// It returns the engine and a shutdown func that flushes telemetry
// (callers running a server must defer shutdown()).
func setupRouter(mediaPath string) (*gin.Engine, func()) {
	shutdown := func() {}
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
		r.Use(otelgin.Middleware("villum"))
		shutdown = func() {
			shutdownTelemetry(tp, lp)
		}
		if lp != nil && middleware.AppLog != nil {
			otelHandler := otelslog.NewHandler("villum", otelslog.WithLoggerProvider(lp))
			middleware.AppLog.Handler().SetExportFn(func(ctx context.Context, rec slog.Record) {
				_ = otelHandler.Handle(ctx, rec)
			})
			middleware.AppLog.Info("system", "OTel log bridge enabled")
		}
	}
	om := initOTelMetrics()
	r.Use(newOTelMetricsMiddleware(om))
	_ = promExporter

	// Public routes
	handlers.RegisterPublicRoutes(r)

	// WebSocket (auth required, no CSRF)
	ws := r.Group("/api")
	ws.Use(middleware.AuthRequired())
	handlers.RegisterWebSocketRoutes(ws)

	// HTMX endpoints (auth + CSRF + API token on mutations)
	htmxGroup := r.Group("/")
	htmxGroup.Use(middleware.AuthRequired(), middleware.APITokenRequired(), middleware.CSRFRequired())
	handlers.HtmxRegisterRoutes(htmxGroup)

	// Authenticated routes
	auth := r.Group("/api")
	auth.Use(middleware.AuthRequired(), middleware.APITokenRequired(), middleware.CSRFRequired())
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

	// API token lifecycle (session + CSRF protected, but NOT API-token protected)
	tokenGroup := r.Group("/api")
	tokenGroup.Use(middleware.AuthRequired(), middleware.CSRFRequired())
	handlers.RegisterTokenRoutes(tokenGroup)

	// DM / One-Shot routes (DM or admin only)
	dm := r.Group("/api")
	dm.Use(middleware.AuthRequired(), middleware.DMRequired(), middleware.APITokenRequired(), middleware.CSRFRequired())
	{
		handlers.RegisterDMCharacterRoutes(dm)
		handlers.RegisterDMCompendiumRoutes(dm)
		handlers.RegisterDMOneShotRoutes(dm)
		handlers.RegisterDMCampaignRoutes(dm)
		handlers.RegisterDMARoutes(dm)
	}

	// Admin routes
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(), middleware.AdminRequired(), middleware.APITokenRequired(), middleware.CSRFRequired())
	{
		handlers.RegisterAdminRoutes(admin)
		handlers.RegisterAdminAIRoutes(admin)
		handlers.RegisterAdminCompendiumRoutes(admin)
	}

	// Static & HTML page serving
	handlers.RegisterStaticRoutes(r, staticFiles, mediaPath, Version)

	return r, shutdown
}
