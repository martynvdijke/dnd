package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListFactions(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	query := "SELECT id, campaign_id, name, description, type, headquarters FROM factions"
	args := []any{}
	if campaignID != "" {
		query += " WHERE campaign_id=?"
		args = append(args, campaignID)
	}
	query += " ORDER BY name"
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.Faction, 0)
	for rows.Next() {
		var f models.Faction
		rows.Scan(&f.ID, &f.CampaignID, &f.Name, &f.Description, &f.Type, &f.Headquarters)
		out = append(out, f)
	}
	c.JSON(http.StatusOK, out)
}

func CreateFaction(c *gin.Context) {
	var f models.Faction
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO factions(campaign_id,name,description,type,headquarters) VALUES(?,?,?,?,?)",
		f.CampaignID, f.Name, f.Description, f.Type, f.Headquarters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateFaction(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var f models.Faction
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE factions SET name=?, description=?, type=?, headquarters=? WHERE id=?", f.Name, f.Description, f.Type, f.Headquarters, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteFaction(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM factions WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Faction Reputation ───

func GetFactionReputations(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character_id required"})
		return
	}
	rows, err := db.DB.Query(`
		SELECT fr.id, fr.character_id, fr.faction_id, fr.standing, fr.rank, fr.notes, f.name, f.type
		FROM faction_reputation fr JOIN factions f ON f.id = fr.faction_id
		WHERE fr.character_id=? ORDER BY fr.standing DESC`, charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type RepWithFaction struct {
		models.FactionReputation
		FactionName string `json:"faction_name"`
		FactionType string `json:"faction_type"`
		FactionID   int64  `json:"faction_id"`
	}
	out := make([]RepWithFaction, 0)
	for rows.Next() {
		var r RepWithFaction
		rows.Scan(&r.ID, &r.CharacterID, &r.FactionID, &r.Standing, &r.Rank, &r.Notes, &r.FactionName, &r.FactionType)
		out = append(out, r)
	}
	c.JSON(http.StatusOK, out)
}

func SetFactionReputation(c *gin.Context) {
	var req struct {
		CharacterID int64  `json:"character_id"`
		FactionID   int64  `json:"faction_id"`
		Standing    int    `json:"standing"`
		Rank        string `json:"rank"`
		Notes       string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Standing < -100 {
		req.Standing = -100
	}
	if req.Standing > 100 {
		req.Standing = 100
	}
	var existingID int64
	err := db.DB.QueryRow("SELECT id FROM faction_reputation WHERE character_id=? AND faction_id=?", req.CharacterID, req.FactionID).Scan(&existingID)
	if err == nil {
		_, err = db.DB.Exec("UPDATE faction_reputation SET standing=?, rank=?, notes=? WHERE id=?", req.Standing, req.Rank, req.Notes, existingID)
	} else {
		_, err = db.DB.Exec("INSERT INTO faction_reputation(character_id,faction_id,standing,rank,notes) VALUES(?,?,?,?,?)", req.CharacterID, req.FactionID, req.Standing, req.Rank, req.Notes)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteFactionReputation(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM faction_reputation WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
