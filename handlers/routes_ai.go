package handlers

import "github.com/gin-gonic/gin"

// RegisterDMARoutes registers DM-only AI generation routes.
func RegisterDMARoutes(r *gin.RouterGroup) {
	r.GET("/ai/endpoints", HandleListEnabledAIEndpoints)
	r.POST("/ai/generate/text", HandleTextGeneration)
	r.POST("/ai/generate/image", HandleImageGeneration)
	r.POST("/ai/save-image", SaveGeneratedImage)

	// Conversational, steerable structured generation.
	r.POST("/ai/draft", StartAIDraft)
	r.GET("/ai/draft", ListAIDrafts)
	r.GET("/ai/draft/:id", GetAIDraft)
	r.POST("/ai/draft/:id/turn", AIDraftTurn)
	r.POST("/ai/draft/:id/commit", CommitAIDraft)
	r.DELETE("/ai/draft/:id", DiscardAIDraft)
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
