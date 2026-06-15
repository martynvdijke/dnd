package handlers

import "github.com/gin-gonic/gin"

// RegisterCompendiumRoutes registers compendium read routes.
func RegisterCompendiumRoutes(r *gin.RouterGroup) {
	r.GET("/compendium/races", ListCompendiumRaces)
	r.GET("/compendium/classes", ListCompendiumClasses)
	r.GET("/compendium/spells", ListCompendiumSpells)
	r.GET("/compendium/feats", ListCompendiumFeats)
	r.GET("/compendium/backgrounds", ListCompendiumBackgrounds)
	r.GET("/compendium/equipment", ListCompendiumEquipment)
	r.GET("/compendium/monsters", ListCompendiumMonsters)
	r.GET("/compendium/search", SearchCompendium)
	r.GET("/compendium/api/:category", FetchFromDnDApi)
	r.GET("/compendium/entries-by-schema", ListUserCompendiumEntriesBySchema)
	r.GET("/compendium/schemas/:id/entries", ListCompendiumEntries)
}

// RegisterDMCompendiumRoutes registers DM-scoped compendium routes.
func RegisterDMCompendiumRoutes(r *gin.RouterGroup) {
	r.GET("/compendium-monsters", ListCompendiumMonsters)
	r.GET("/compendium-monsters/:id", GetCompendiumMonster)
	r.GET("/compendium-search", SearchCompendiumEntries)
}

// RegisterAdminCompendiumRoutes registers admin compendium management routes.
func RegisterAdminCompendiumRoutes(r *gin.RouterGroup) {
	// Compendium Schema System
	r.GET("/compendium-schemas", ListCompendiumSchemas)
	r.POST("/compendium-schemas", CreateCompendiumSchema)
	r.GET("/compendium-schemas/:id", GetCompendiumSchema)
	r.PUT("/compendium-schemas/:id", UpdateCompendiumSchema)
	r.DELETE("/compendium-schemas/:id", DeleteCompendiumSchema)

	// Compendium Entries
	r.GET("/compendium-schemas/:id/entries", ListCompendiumEntries)
	r.POST("/compendium-schemas/:id/entries", CreateCompendiumEntry)
	r.GET("/compendium-entries/:id", GetCompendiumEntry)
	r.PUT("/compendium-entries/:id", UpdateCompendiumEntry)
	r.DELETE("/compendium-entries/:id", DeleteCompendiumEntry)

	// Bulk Operations
	r.POST("/compendium-entries/batch-delete", BatchDeleteCompendiumEntries)
	r.POST("/compendium-entries/batch-update", BatchUpdateCompendiumEntries)

	// Search
	r.GET("/compendium-search", SearchCompendiumEntries)

	// Import/Export
	r.POST("/compendium-schemas/:id/import", ImportCompendiumEntries)
	r.POST("/compendium-schemas/:id/import/with-mapping", ImportCompendiumEntriesWithMapping)
	r.POST("/compendium-schemas/:id/import/detect", DetectImportFields)
	r.GET("/compendium-schemas/:id/export", ExportCompendiumEntries)
	r.POST("/compendium-import", ImportCompendiumBatchJSON)

	// Import Logs
	r.GET("/compendium-import-logs", ListCompendiumImportLogs)
	r.POST("/compendium-import-logs/:id/rollback", RollbackCompendiumImport)
}
