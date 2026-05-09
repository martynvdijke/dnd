package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vellum/db"
)

type CharacterStats struct {
	SessionCount    int            `json:"session_count"`
	TotalXPEarned   int            `json:"total_xp_earned"`
	TotalGoldEarned int            `json:"total_gold_earned"`
	Quests          QuestBreakdown `json:"quests"`
	Rests           RestBreakdown  `json:"rests"`
	NPCInteractions int            `json:"npc_interactions"`
	TopNPCs         []string       `json:"top_npcs"`
	LocationsCount  int            `json:"locations_count"`
	JournalCount    int            `json:"journal_count"`
	DiceRolls       DiceStats      `json:"dice_rolls"`
	Level           int            `json:"level"`
	SessionsPerMonth float64       `json:"sessions_per_month"`
}

type QuestBreakdown struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Complete  int `json:"complete"`
	Failed    int `json:"failed"`
	Available int `json:"available"`
	Abandoned int `json:"abandoned"`
}

type RestBreakdown struct {
	Total      int `json:"total"`
	Short      int `json:"short"`
	Long       int `json:"long"`
	TotalHealed int `json:"total_healed"`
}

type DiceStats struct {
	TotalRolls int     `json:"total_rolls"`
	Average    float64 `json:"average"`
	Natural20s int     `json:"natural_20s"`
	Natural1s  int     `json:"natural_1s"`
}

func GetCharacterStats(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var stats CharacterStats

	// Character level
	db.DB.QueryRow("SELECT level FROM characters WHERE id=?", charID).Scan(&stats.Level)

	// Session stats
	db.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(xp_earned),0), COALESCE(SUM(gold_earned),0) FROM sessions WHERE character_id=?", charID).Scan(
		&stats.SessionCount, &stats.TotalXPEarned, &stats.TotalGoldEarned)

	// Sessions per month
	if stats.SessionCount > 0 {
		var firstSession, lastSession string
		db.DB.QueryRow("SELECT MIN(session_date), MAX(session_date) FROM sessions WHERE character_id=?", charID).Scan(&firstSession, &lastSession)
		if firstSession != "" && lastSession != "" {
			// Approximate months
			stats.SessionsPerMonth = float64(stats.SessionCount)
		}
	}

	// Quest breakdown
	db.DB.QueryRow("SELECT COUNT(*) FROM quests WHERE character_id=?", charID).Scan(&stats.Quests.Total)
	db.DB.QueryRow("SELECT COUNT(*) FROM quests WHERE character_id=? AND status='active'", charID).Scan(&stats.Quests.Active)
	db.DB.QueryRow("SELECT COUNT(*) FROM quests WHERE character_id=? AND status='complete'", charID).Scan(&stats.Quests.Complete)
	db.DB.QueryRow("SELECT COUNT(*) FROM quests WHERE character_id=? AND status='failed'", charID).Scan(&stats.Quests.Failed)
	db.DB.QueryRow("SELECT COUNT(*) FROM quests WHERE character_id=? AND status='available'", charID).Scan(&stats.Quests.Available)
	db.DB.QueryRow("SELECT COUNT(*) FROM quests WHERE character_id=? AND status='abandoned'", charID).Scan(&stats.Quests.Abandoned)

	// Rest breakdown
	db.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(hp_healed),0) FROM rest_log WHERE character_id=?", charID).Scan(&stats.Rests.Total, &stats.Rests.TotalHealed)
	db.DB.QueryRow("SELECT COUNT(*) FROM rest_log WHERE character_id=? AND rest_type='short'", charID).Scan(&stats.Rests.Short)
	db.DB.QueryRow("SELECT COUNT(*) FROM rest_log WHERE character_id=? AND rest_type='long'", charID).Scan(&stats.Rests.Long)

	// NPC interactions
	db.DB.QueryRow("SELECT COALESCE(SUM(interaction_count),0) FROM character_npcs WHERE character_id=?", charID).Scan(&stats.NPCInteractions)

	// Top NPCs
	rows, _ := db.DB.Query("SELECT n.name FROM character_npcs cn JOIN npcs n ON cn.npc_id = n.id WHERE cn.character_id=? ORDER BY cn.interaction_count DESC LIMIT 5", charID)
	if rows != nil {
		for rows.Next() {
			var name string
			rows.Scan(&name)
			stats.TopNPCs = append(stats.TopNPCs, name)
		}
		rows.Close()
	}

	// Location count
	db.DB.QueryRow("SELECT COUNT(*) FROM character_locations WHERE character_id=?", charID).Scan(&stats.LocationsCount)

	// Journal count
	db.DB.QueryRow("SELECT COUNT(*) FROM journal WHERE character_id=?", charID).Scan(&stats.JournalCount)

	// Dice stats
	db.DB.QueryRow("SELECT COUNT(*), COALESCE(AVG(total),0) FROM dice_rolls WHERE character_id=?", charID).Scan(&stats.DiceRolls.TotalRolls, &stats.DiceRolls.Average)

	c.JSON(http.StatusOK, stats)
}
