package handlers

import "github.com/gin-gonic/gin"

// RegisterCampaignRoutes registers campaign, party, wiki, and related routes.
func RegisterCampaignRoutes(r *gin.RouterGroup) {
	// Campaigns
	r.GET("/campaigns", ListCampaigns)
	r.GET("/campaigns/mine", ListCampaigns)
	r.GET("/campaigns/:id/characters", ListCampaignCharacters)
	r.GET("/campaigns/:id/character-candidates", ListCampaignCharacterCandidates)
	r.POST("/campaigns/:id/characters", AddCampaignCharacter)
	r.DELETE("/campaigns/:id/characters/:characterId", RemoveCampaignCharacter)
	r.POST("/campaigns", CreateCampaign)
	r.PUT("/campaigns/:id", UpdateCampaign)
	r.DELETE("/campaigns/:id", DeleteCampaign)
	r.GET("/campaigns/:id/members", ListCampaignMembers)
	r.POST("/campaigns/:id/members", AddCampaignMember)
	r.PUT("/campaigns/:id/members/:userId", SetCampaignMemberRole)
	r.DELETE("/campaigns/:id/members/:userId", RemoveCampaignMember)

	// Party view
	r.GET("/party", GetPartyView)
	r.GET("/users/search", SearchUsers)

	// Party Inventory
	r.GET("/campaigns/:id/party-items", ListCampaignPartyItems)
	r.POST("/campaigns/:id/party-items", CreateCampaignPartyItem)
	r.DELETE("/party-items/:id", DeleteCampaignPartyItem)

	// Session Plans
	r.GET("/campaigns/:id/session-plans", ListSessionPlans)
	r.POST("/campaigns/:id/session-plans", CreateSessionPlan)
	r.PUT("/session-plans/:id", UpdateSessionPlan)
	r.DELETE("/session-plans/:id", DeleteSessionPlan)

	// Wiki
	r.GET("/campaigns/:id/wiki", ListWikiPages)
	r.POST("/campaigns/:id/wiki", CreateWikiPage)
	r.GET("/wiki/:id", GetWikiPage)
	r.PUT("/wiki/:id", UpdateWikiPage)
	r.DELETE("/wiki/:id", DeleteWikiPage)

	// Campaign graph
	r.GET("/campaigns/:id/graph", GetCampaignGraphData)

	// Campaign Dashboard
	r.GET("/campaigns/:id/dashboard", GetCampaignDashboard)
	r.GET("/campaigns/:id/one-shots", ListCampaignOneShots)
	r.PUT("/campaigns/:id/one-shots/reorder", ReorderCampaignOneShots)

	// Campaign Maps
	r.GET("/campaigns/:id/maps", ListCampaignMaps)
	r.POST("/campaigns/:id/maps", CreateCampaignMap)
	r.PUT("/maps/:id", UpdateCampaignMap)
	r.DELETE("/maps/:id", DeleteCampaignMap)
	r.POST("/campaigns/:id/maps/:mapId/activate", SetActiveCampaignMap)
	r.PUT("/maps/:id/fog", UpdateFogOfWar)
	r.GET("/campaigns/:id/maps/active", GetActiveCampaignMap)
	r.GET("/maps/:mapId/pins", ListMapPins)
	r.POST("/maps/:mapId/pins", CreateMapPin)
	r.PUT("/map-pins/:id", UpdateMapPin)
	r.DELETE("/map-pins/:id", DeleteMapPin)

	// Calendar
	r.GET("/calendar", ListCalendarEvents)
	r.POST("/calendar", CreateCalendarEvent)
	r.PUT("/calendar/:id", UpdateCalendarEvent)
	r.DELETE("/calendar/:id", DeleteCalendarEvent)

	// Timeline
	r.GET("/timeline", ListTimelineEvents)
	r.POST("/timeline", CreateTimelineEvent)
	r.PUT("/timeline/:id", UpdateTimelineEvent)
	r.DELETE("/timeline/:id", DeleteTimelineEvent)

	// Campaign Recaps
	r.GET("/campaigns/:id/recaps", ListCampaignRecaps)
	r.POST("/campaigns/:id/recaps", CreateCampaignRecap)
	r.GET("/recaps/:id", GetCampaignRecap)
	r.PUT("/recaps/:id", UpdateCampaignRecap)
	r.DELETE("/recaps/:id", DeleteCampaignRecap)
	r.POST("/campaigns/:id/recaps/generate", GenerateCampaignRecap)
	r.POST("/recaps/:id/mark-sent", MarkRecapAsSent)

	// Knowledge
	r.GET("/campaigns/:id/knowledge", ListKnowledge)
	r.POST("/campaigns/:id/knowledge", CreateKnowledge)
	r.GET("/knowledge/:kid", GetKnowledge)
	r.PUT("/knowledge/:kid", UpdateKnowledge)
	r.DELETE("/knowledge/:kid", DeleteKnowledge)
	r.GET("/knowledge/:kid/known-by", ListKnowledgeKnownBy)
	r.POST("/knowledge/:kid/known-by", AddKnowledgeKnownBy)
	r.DELETE("/knowledge/:kid/known-by/:cid", RemoveKnowledgeKnownBy)
	r.POST("/knowledge/:kid/reveal", BulkRevealKnowledge)

	// Per-user push mute for a campaign
	r.GET("/campaigns/:id/push-mute", GetCampaignPushMute)
	r.PUT("/campaigns/:id/push-mute", SetCampaignPushMute)
}

// RegisterDM Campaign routes
func RegisterDMCampaignRoutes(r *gin.RouterGroup) {
	// Campaign NPCs
	r.GET("/campaigns/:id/npcs", ListCampaignNPCs)
	r.POST("/campaigns/:id/npcs", LinkNPCCampaign)
	r.PUT("/campaign-npcs/:id", UpdateCampaignNPC)
	r.DELETE("/campaign-npcs/:id", UnlinkCampaignNPC)
	r.POST("/campaigns/:id/npcs/create-and-link", CreateAndLinkCampaignNPC)

	// Campaign Encounter Monsters
	r.GET("/encounters/:id/monsters", ListCampaignEncounterMonsters)

	// Campaign-scoped DM Shops
	r.POST("/campaigns/:id/shops", CreateShop)
	r.PUT("/shops/:id", UpdateShop)
	r.DELETE("/shops/:id", DeleteShop)
	r.POST("/shops/:id/items", CreateShopItem)
	r.PUT("/shop-items/:id", UpdateShopItem)
	r.DELETE("/shop-items/:id", DeleteShopItem)
	r.DELETE("/shop-items/:id/link", UnlinkShopItem)
}

// RegisterShopRoutes registers auth shops routes.
func RegisterShopRoutes(r *gin.RouterGroup) {
	r.GET("/shops", ListShops)
	r.GET("/shops/:id/items", ListShopItems)
	r.POST("/shops/:id/buy", BuyItem)
	r.POST("/shops/:id/sell", SellItem)
	r.GET("/shop-transactions", ListShopTransactions)
}
