package handlers

import "github.com/gin-gonic/gin"

// RegisterDMARoutes registers DM-only AI generation routes.
func RegisterDMARoutes(r *gin.RouterGroup) {
	r.GET("/ai/endpoints", HandleListEnabledAIEndpoints)
	r.POST("/ai/generate/text", HandleTextGeneration)
	r.POST("/ai/generate/image", HandleImageGeneration)
}

// RegisterAdminAIRoutes registers admin AI endpoint management routes.
func RegisterAdminAIRoutes(r *gin.RouterGroup) {
	r.GET("/ai-endpoints", ListAIEndpoints)
	r.GET("/ai-endpoints/:id", GetAIEndpoint)
	r.POST("/ai-endpoints", CreateAIEndpoint)
	r.PUT("/ai-endpoints/:id", UpdateAIEndpoint)
	r.DELETE("/ai-endpoints/:id", DeleteAIEndpoint)
	r.POST("/ai-endpoints/:id/test", TestAIEndpoint)
}
