package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func CreateNPCItemLink(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var l models.NPCItemLink
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO npc_item_links(npc_id, adventure_id, item_id, relationship_type, notes) VALUES(?,?,?,?,?)",
		l.NPCID, adventureID, l.ItemID, l.RelationshipType, l.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func ListItemsForNPC(c *gin.Context) {
	npcID := c.Param("nid")
	adventureID := c.Param("id")
	rows, err := db.DB.Query(`
		SELECT l.id, l.npc_id, l.adventure_id, l.item_id, l.relationship_type, l.notes, i.name
		FROM npc_item_links l JOIN oneshot_items i ON i.id = l.item_id
		WHERE l.npc_id=? AND l.adventure_id=? ORDER BY i.name`, npcID, adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.NPCItemLink, 0)
	for rows.Next() {
		var l models.NPCItemLink
		rows.Scan(&l.ID, &l.NPCID, &l.AdventureID, &l.ItemID, &l.RelationshipType, &l.Notes, &l.ItemName)
		out = append(out, l)
	}
	c.JSON(http.StatusOK, out)
}

func DeleteNPCItemLink(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM npc_item_links WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
