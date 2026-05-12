package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type CombatEntry struct {
	ID             int64  `json:"id"`
	CharacterID    *int64 `json:"character_id,omitempty"`
	CampaignID     *int64 `json:"campaign_id,omitempty"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	InitiativeRoll int    `json:"initiative_roll"`
	InitiativeMod  int    `json:"initiative_mod"`
	HPMax          int    `json:"hp_max"`
	HPCurrent      int    `json:"hp_current"`
	AC             int    `json:"ac"`
	IsActive       bool   `json:"is_active"`
	TurnOrder      int    `json:"turn_order"`
	ConditionIDs   string `json:"condition_ids"`
	Notes          string `json:"notes"`
}

func CreateCombatEntry(c *gin.Context) {
	var e CombatEntry
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec(`INSERT INTO combat_entries(character_id,campaign_id,name,type,initiative_roll,initiative_mod,hp_max,hp_current,ac,turn_order,condition_ids,notes,is_active) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,1)`,
		e.CharacterID, e.CampaignID, e.Name, e.Type, e.InitiativeRoll, e.InitiativeMod, e.HPMax, e.HPCurrent, e.AC, e.TurnOrder, e.ConditionIDs, e.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func ListCombatEntries(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	var rows *sql.Rows
	var err error

	if campaignID != "" {
		rows, err = db.DB.Query("SELECT id,character_id,campaign_id,name,type,initiative_roll,initiative_mod,hp_max,hp_current,ac,is_active,turn_order,condition_ids,notes FROM combat_entries WHERE campaign_id=? ORDER BY initiative_roll DESC, turn_order", campaignID)
	} else {
		rows, err = db.DB.Query("SELECT id,character_id,campaign_id,name,type,initiative_roll,initiative_mod,hp_max,hp_current,ac,is_active,turn_order,condition_ids,notes FROM combat_entries ORDER BY initiative_roll DESC, turn_order")
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var entries []CombatEntry
	for rows.Next() {
		var e CombatEntry
		var isActive int
		rows.Scan(&e.ID, &e.CharacterID, &e.CampaignID, &e.Name, &e.Type,
			&e.InitiativeRoll, &e.InitiativeMod, &e.HPMax, &e.HPCurrent, &e.AC,
			&isActive, &e.TurnOrder, &e.ConditionIDs, &e.Notes)
		e.IsActive = isActive == 1
		entries = append(entries, e)
	}
	c.JSON(http.StatusOK, entries)
}

func UpdateCombatEntry(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var e CombatEntry
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isActive := 0
	if e.IsActive {
		isActive = 1
	}
	db.DB.Exec(`UPDATE combat_entries SET name=?,type=?,initiative_roll=?,initiative_mod=?,hp_max=?,hp_current=?,ac=?,is_active=?,turn_order=?,condition_ids=?,notes=? WHERE id=?`,
		e.Name, e.Type, e.InitiativeRoll, e.InitiativeMod, e.HPMax, e.HPCurrent, e.AC, isActive, e.TurnOrder, e.ConditionIDs, e.Notes, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCombatEntry(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM combat_entries WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func RollInitiative(c *gin.Context) {
	var req struct {
		CharacterID int64 `json:"character_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var name, class string
	var dex, initiative int
	err := db.DB.QueryRow("SELECT name, class, dex, initiative FROM characters WHERE id=?", req.CharacterID).
		Scan(&name, &class, &dex, &initiative)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	initMod := abilityMod(dex) + initiative
	d20, _ := randInt(1, 20)
	roll := d20 + initMod

	c.JSON(http.StatusOK, gin.H{
		"character_id": req.CharacterID,
		"name":         name,
		"d20":          d20,
		"modifier":     initMod,
		"total":        roll,
	})
}

func NextTurn(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campaign_id required"})
		return
	}
	// Find the highest turn_order, advance it (wrap around)
	var maxOrder int
	db.DB.QueryRow("SELECT COALESCE(MAX(turn_order),0) FROM combat_entries WHERE campaign_id=? AND is_active=1", campaignID).Scan(&maxOrder)
	db.DB.Exec("UPDATE combat_entries SET turn_order = CASE WHEN turn_order >= ? THEN 0 ELSE turn_order + 1 END WHERE campaign_id=? AND is_active=1", maxOrder, campaignID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetCurrentTurn(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campaign_id required"})
		return
	}
	var entry CombatEntry
	var isActive int
	err := db.DB.QueryRow("SELECT id,character_id,campaign_id,name,type,initiative_roll,initiative_mod,hp_max,hp_current,ac,is_active,turn_order,condition_ids,notes FROM combat_entries WHERE campaign_id=? AND is_active=1 ORDER BY turn_order DESC LIMIT 1", campaignID).
		Scan(&entry.ID, &entry.CharacterID, &entry.CampaignID, &entry.Name, &entry.Type,
			&entry.InitiativeRoll, &entry.InitiativeMod, &entry.HPMax, &entry.HPCurrent, &entry.AC,
			&isActive, &entry.TurnOrder, &entry.ConditionIDs, &entry.Notes)
	entry.IsActive = isActive == 1
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"current": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"current": entry})
}
