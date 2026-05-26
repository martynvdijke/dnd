package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListParties(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, user_id, name, description, created_at FROM parties ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.Party, 0)
	for rows.Next() {
		var p models.Party
		rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.CreatedAt)
		out = append(out, p)
	}
	c.JSON(http.StatusOK, out)
}

func CreateParty(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var p models.Party
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO parties(user_id, name, description) VALUES(?,?,?)",
		userID, p.Name, p.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetParty(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p models.Party
	err := db.DB.QueryRow("SELECT id, user_id, name, description, created_at FROM parties WHERE id=?", id).
		Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "party not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func UpdateParty(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p models.Party
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("UPDATE parties SET name=?, description=? WHERE id=?", p.Name, p.Description, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteParty(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM parties WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListPartyFactions(c *gin.Context) {
	partyID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, campaign_id, party_id, name, description, type, headquarters FROM factions WHERE party_id=? ORDER BY name", partyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.Faction, 0)
	for rows.Next() {
		var f models.Faction
		var campaignID, partyID sql.NullInt64
		rows.Scan(&f.ID, &campaignID, &partyID, &f.Name, &f.Description, &f.Type, &f.Headquarters)
		if campaignID.Valid { f.CampaignID = &campaignID.Int64 }
		if partyID.Valid { f.PartyID = &partyID.Int64 }
		out = append(out, f)
	}
	c.JSON(http.StatusOK, out)
}

func CreatePartyFaction(c *gin.Context) {
	partyID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var f models.Faction
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO factions(party_id, name, description, type, headquarters) VALUES(?,?,?,?,?)",
		partyID, f.Name, f.Description, f.Type, f.Headquarters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func ListPartyUploads(c *gin.Context) {
	partyID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, hash, ext, url, COALESCE(resized_url,''), COALESCE(thumbnail_url,''), owner_type, owner_id, COALESCE(created_at,'') FROM uploads WHERE owner_type='party' AND owner_id=? ORDER BY created_at DESC", partyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	uploads := make([]models.Upload, 0)
	for rows.Next() {
		var u models.Upload
		rows.Scan(&u.ID, &u.Hash, &u.Ext, &u.URL, &u.ResizedURL, &u.ThumbnailURL, &u.OwnerType, &u.OwnerID, &u.CreatedAt)
		uploads = append(uploads, u)
	}
	c.JSON(http.StatusOK, uploads)
}
