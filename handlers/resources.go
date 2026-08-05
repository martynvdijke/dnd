package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type CharacterResource struct {
	ID                int64  `json:"id"`
	CharacterID       int64  `json:"character_id"`
	Name              string `json:"name"`
	Current           int    `json:"current"`
	Max               int    `json:"max"`
	ShortRestRecovery int    `json:"short_rest_recovery"`
	LongRestRecovery  int    `json:"long_rest_recovery"`
	Icon              string `json:"icon"`
	SortOrder         int    `json:"sort_order"`
}

func ListCharacterResources(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id,character_id,name,current,max,short_rest_recovery,long_rest_recovery,icon,sort_order FROM character_resources WHERE character_id=? ORDER BY sort_order", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]CharacterResource, 0)
	for rows.Next() {
		var r CharacterResource
		rows.Scan(&r.ID, &r.CharacterID, &r.Name, &r.Current, &r.Max, &r.ShortRestRecovery, &r.LongRestRecovery, &r.Icon, &r.SortOrder)
		out = append(out, r)
	}
	c.JSON(http.StatusOK, out)
}

func CreateCharacterResource(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var r CharacterResource
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO character_resources(character_id,name,current,max,short_rest_recovery,long_rest_recovery,icon,sort_order) VALUES(?,?,?,?,?,?,?,?)",
		charID, r.Name, r.Current, r.Max, r.ShortRestRecovery, r.LongRestRecovery, r.Icon, r.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCharacterResource(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditResourceID(c, "character_resources", id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var r CharacterResource
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE character_resources SET name=?,current=?,max=?,short_rest_recovery=?,long_rest_recovery=?,icon=?,sort_order=? WHERE id=?",
		r.Name, r.Current, r.Max, r.ShortRestRecovery, r.LongRestRecovery, r.Icon, r.SortOrder, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCharacterResource(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	// Idempotent no-op when the resource does not exist
	var exists int
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM character_resources WHERE id=?", id).Scan(&exists); err == nil && exists == 0 {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	if !canEditResourceID(c, "character_resources", id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	db.DB.Exec("DELETE FROM character_resources WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func RecoverResourcesOnRest(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	var req struct {
		RestType string `json:"rest_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var col string
	if req.RestType == "short" {
		col = "short_rest_recovery"
	} else {
		col = "long_rest_recovery"
	}

	_, err := db.DB.Exec("UPDATE character_resources SET current = MIN(max, current + "+col+") WHERE character_id=?", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If long rest, also reset spell slots
	if req.RestType == "long" {
		db.DB.Exec(`UPDATE character_spellcasting SET
			slots_1_used=0, slots_2_used=0, slots_3_used=0, slots_4_used=0,
			slots_5_used=0, slots_6_used=0, slots_7_used=0, slots_8_used=0, slots_9_used=0
			WHERE character_id=?`, charID)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
