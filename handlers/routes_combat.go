package handlers

import "github.com/gin-gonic/gin"

// RegisterCombatRoutes registers combat tracker and encounter routes.
func RegisterCombatRoutes(r *gin.RouterGroup) {
	// Combat entries
	r.GET("/combat", ListCombatEntries)
	r.POST("/combat", CreateCombatEntry)
	r.PUT("/combat/:id", UpdateCombatEntry)
	r.DELETE("/combat/:id", DeleteCombatEntry)
	r.POST("/combat/initiative", RollInitiative)
	r.POST("/combat/next-turn", NextTurn)
	r.GET("/combat/current-turn", GetCurrentTurn)

	// Combat Log
	r.GET("/combat-log", ListCombatLogEntries)
	r.POST("/combat-log", CreateCombatLogEntry)
	r.GET("/combat-log/stats", GetCombatLogStats)
}

// RegisterEncounterRoutes registers encounter builder routes.
func RegisterEncounterRoutes(r *gin.RouterGroup) {
	// Encounter Builder
	r.GET("/encounters", ListEncounters)
	r.POST("/encounters", CreateEncounter)
	r.GET("/encounters/:id", GetEncounter)
	r.PUT("/encounters/:id", UpdateEncounter)
	r.DELETE("/encounters/:id", DeleteEncounter)
	r.POST("/encounters/:id/monsters", AddEncounterMonster)
	r.PUT("/encounter-monsters/:mid", UpdateEncounterMonster)
	r.DELETE("/encounter-monsters/:mid", DeleteEncounterMonster)
	r.POST("/encounters/calculate-xp", CalculateEncounterXP)
	r.GET("/monster-xp", GetMonsterXP)
}
