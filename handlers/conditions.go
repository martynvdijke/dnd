package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListConditions(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character_id required"})
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, name, type, source, duration, duration_type, saving_throw, save_dc, description, started_at FROM character_conditions WHERE character_id=? ORDER BY started_at DESC", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.Condition, 0)
	for rows.Next() {
		var cond models.Condition
		rows.Scan(&cond.ID, &cond.CharacterID, &cond.Name, &cond.Type, &cond.Source, &cond.Duration, &cond.DurationType, &cond.SavingThrow, &cond.SaveDC, &cond.Description, &cond.StartedAt)
		out = append(out, cond)
	}
	c.JSON(http.StatusOK, out)
}

func CreateCondition(c *gin.Context) {
	var cond models.Condition
	if err := c.ShouldBindJSON(&cond); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if cond.Duration < 0 {
		cond.Duration = 0
	}
	if cond.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	result, err := db.DB.Exec("INSERT INTO character_conditions(character_id,name,type,source,duration,duration_type,saving_throw,save_dc,description) VALUES(?,?,?,?,?,?,?,?,?)",
		cond.CharacterID, cond.Name, cond.Type, cond.Source, cond.Duration, cond.DurationType, cond.SavingThrow, cond.SaveDC, cond.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCondition(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cond models.Condition
	if err := c.ShouldBindJSON(&cond); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE character_conditions SET name=?, type=?, source=?, duration=?, duration_type=?, saving_throw=?, save_dc=?, description=? WHERE id=?",
		cond.Name, cond.Type, cond.Source, cond.Duration, cond.DurationType, cond.SavingThrow, cond.SaveDC, cond.Description, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCondition(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM character_conditions WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func TickConditions(c *gin.Context) {
	var req struct {
		CharacterID int64 `json:"character_id"`
		Count       int   `json:"count"`
		DurationType string `json:"duration_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Count < 1 {
		req.Count = 1
	}
	if req.DurationType == "" {
		req.DurationType = "round"
	}

	res, err := db.DB.Exec(`DELETE FROM character_conditions WHERE character_id=? AND duration_type=? AND duration > 0 AND duration <= ?`, req.CharacterID, req.DurationType, req.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	deleted, _ := res.RowsAffected()

	// Decrement remaining durations
	db.DB.Exec(`UPDATE character_conditions SET duration = duration - ? WHERE character_id=? AND duration_type=? AND duration > ?`, req.Count, req.CharacterID, req.DurationType, req.Count)

	expired := int(deleted)
	c.JSON(http.StatusOK, gin.H{"expired": expired, "ticked": req.Count})
}

// Standard 5e condition types
var ConditionTypes = []string{
	"blinded", "charmed", "deafened", "exhaustion", "frightened",
	"grappled", "incapacitated", "invisible", "paralyzed",
	"petrified", "poisoned", "prone", "restrained", "stunned",
	"unconscious", "concentration",
}

func GetConditionTypes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"conditions": ConditionTypes})
}

func GetActiveConditionSummary(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character_id required"})
		return
	}
	rows, err := db.DB.Query("SELECT id, name, type, duration, duration_type FROM character_conditions WHERE character_id=? ORDER BY started_at DESC", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type ConditionBadge struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		Type         string `json:"type"`
		Duration     int    `json:"duration"`
		DurationType string `json:"duration_type"`
		Icon         string `json:"icon"`
		Color        string `json:"color"`
	}

	// Map condition types to icons/colors
	iconMap := map[string]string{
		"blinded":       "fa-eye-slash",
		"charmed":       "fa-heart",
		"deafened":      "fa-ear-deaf",
		"exhaustion":    "fa-battery-quarter",
		"frightened":    "fa-ghost",
		"grappled":      "fa-handcuffs",
		"incapacitated": "fa-bed",
		"invisible":     "fa-ghost",
		"paralyzed":     "fa-snowflake",
		"petrified":     "fa-monument",
		"poisoned":      "fa-skull",
		"prone":         "fa-person-falling",
		"restrained":    "fa-lock",
		"stunned":       "fa-star",
		"unconscious":   "fa-circle",
		"concentration": "fa-brain",
	}
	colorMap := map[string]string{
		"blinded": "#8b0000", "charmed": "#dda0dd", "deafened": "#666",
		"exhaustion": "#ff8c00", "frightened": "#4b0082", "grappled": "#8b4513",
		"incapacitated": "#555", "invisible": "#87ceeb", "paralyzed": "#00bfff",
		"petrified": "#808080", "poisoned": "#32cd32", "prone": "#d2b48c",
		"restrained": "#ffd700", "stunned": "#ff4500", "unconscious": "#2f4f4f",
		"concentration": "#4169e1",
	}

	out := make([]ConditionBadge, 0)
	for rows.Next() {
		var c models.Condition
		rows.Scan(&c.ID, &c.CharacterID, &c.Name, &c.Type, &c.Duration, &c.DurationType)
		durStr := fmt.Sprintf("%d %s", c.Duration, c.DurationType)
		if c.DurationType == "permanent" {
			durStr = "perm"
		}
		icon := "fa-circle"
		if i, ok := iconMap[c.Type]; ok {
			icon = i
		}
		color := "#b8963e"
		if cl, ok := colorMap[c.Type]; ok {
			color = cl
		}
		out = append(out, ConditionBadge{
			ID: c.ID, Name: strings.Title(c.Name), Type: c.Type,
			Duration: c.Duration, DurationType: c.DurationType,
			Icon: icon, Color: color,
		})
		_ = durStr
	}
	c.JSON(http.StatusOK, out)
}
