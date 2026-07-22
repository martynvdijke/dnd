package handlers

import "github.com/gin-gonic/gin"

// RegisterAdminRoutes registers admin-only routes.
func RegisterAdminRoutes(r *gin.RouterGroup) {
	// User management
	r.GET("/users", AdminListUsers)
	r.POST("/users", AdminCreateUser)
	r.PUT("/users/:id", AdminUpdateUser)
	r.DELETE("/users/:id", AdminDeleteUser)
	r.PUT("/users/:id/password", AdminResetPassword)

	// Email settings
	r.GET("/email-settings", GetEmailSettings)
	r.POST("/email-settings", SaveEmailSettings)
	r.POST("/test-email", TestEmail)
	r.POST("/campaign-highlights", SendCampaignHighlights)

	// Shop management (admin access — global)
	r.POST("/shops", CreateShop)
	r.PUT("/shops/:id", UpdateShop)
	r.DELETE("/shops/:id", DeleteShop)
	r.POST("/shops/:id/items", CreateShopItem)
	r.PUT("/shop-items/:id", UpdateShopItem)
	r.DELETE("/shop-items/:id", DeleteShopItem)

	// Umami analytics
	r.GET("/umami-settings", GetUmamiSettings)
	r.POST("/umami-settings", SaveUmamiSettings)

	// OTel telemetry
	r.GET("/otel-settings", GetOTelSettings)
	r.POST("/otel-settings", SaveOTelSettings)

	// Application Logs
	r.GET("/logs", ListLogs)
	r.GET("/log-sources", ListLogSources)
	r.GET("/log-level", GetLogLevel)
	r.PUT("/log-level", SetLogLevel)

	// Events Settings
	r.GET("/events-settings", GetEventsSettings)
	r.POST("/events-settings", SaveEventsSettings)
	r.POST("/events-settings/clear-cache", ClearEventsCache)

	// Events Share Link & QR
	r.GET("/events/public-link", EventsPublicLink)
	r.GET("/events/qr", EventsQRCode)

	// Campaign Event Settings
	r.GET("/events-campaigns", ListCampaignEventSettings)
	r.GET("/events-campaigns/:id", GetCampaignEventSetting)
	r.POST("/events-campaigns", SaveCampaignEventSetting)
	r.PUT("/events-campaigns/:id", SaveCampaignEventSetting)
	r.DELETE("/events-campaigns/:id", DeleteCampaignEventSetting)

	// Search index management
	r.GET("/search/resync", HandleAdminResyncSearchIndex)
}
