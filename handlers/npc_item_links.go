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
	// Exactly one source: a one-shot item OR a compendium equipment entry.
	hasItem := l.ItemID > 0
	hasCompendium := l.CompendiumEquipmentID != nil && *l.CompendiumEquipmentID > 0
	if hasItem == hasCompendium {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exactly one of item_id or compendium_equipment_id is required"})
		return
	}
	var result interface{ LastInsertId() (int64, error) }
	if hasCompendium {
		result, err := db.DB.Exec("INSERT INTO npc_item_links(npc_id, adventure_id, compendium_equipment_id, relationship_type, notes) VALUES(?,?,?,?,?)",
			l.NPCID, adventureID, *l.CompendiumEquipmentID, l.RelationshipType, l.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": id})
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
		SELECT l.id, l.npc_id, l.adventure_id, COALESCE(l.item_id,0), COALESCE(l.compendium_equipment_id,0),
			l.relationship_type, l.notes,
			COALESCE(i.name, ce.name, '') AS item_name
		FROM npc_item_links l
		LEFT JOIN oneshot_items i ON i.id = l.item_id
		LEFT JOIN compendium_equipment ce ON ce.id = l.compendium_equipment_id
		WHERE l.npc_id=? AND l.adventure_id=? ORDER BY item_name`, npcID, adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.NPCItemLink, 0)
	for rows.Next() {
		var l models.NPCItemLink
		var compID int64
		rows.Scan(&l.ID, &l.NPCID, &l.AdventureID, &l.ItemID, &compID, &l.RelationshipType, &l.Notes, &l.ItemName)
		if compID > 0 {
			id := compID
			l.CompendiumEquipmentID = &id
		}
		out = append(out, l)
	}
	c.JSON(http.StatusOK, out)
}

func DeleteNPCItemLink(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM npc_item_links WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
