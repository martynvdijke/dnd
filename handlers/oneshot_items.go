package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListOneShotItems(c *gin.Context) {
	adventureID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, adventure_id, name, description, category, quantity, weight, price_gp, is_magical, attunement, notes, created_at FROM oneshot_items WHERE adventure_id=? ORDER BY name", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotItem, 0)
	for rows.Next() {
		var it models.OneShotItem
		var isMag, att int
		rows.Scan(&it.ID, &it.AdventureID, &it.Name, &it.Description, &it.Category, &it.Quantity, &it.Weight, &it.PriceGP, &isMag, &att, &it.Notes, &it.CreatedAt)
		it.IsMagical = isMag == 1
		it.Attunement = att == 1
		out = append(out, it)
	}
	c.JSON(http.StatusOK, out)
}

func CreateOneShotItem(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var it models.OneShotItem
	if err := c.ShouldBindJSON(&it); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isMag := 0
	if it.IsMagical {
		isMag = 1
	}
	att := 0
	if it.Attunement {
		att = 1
	}
	result, err := db.DB.Exec("INSERT INTO oneshot_items(adventure_id, name, description, category, quantity, weight, price_gp, is_magical, attunement, notes) VALUES(?,?,?,?,?,?,?,?,?,?)",
		adventureID, it.Name, it.Description, it.Category, it.Quantity, it.Weight, it.PriceGP, isMag, att, it.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateOneShotItem(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var it models.OneShotItem
	if err := c.ShouldBindJSON(&it); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isMag := 0
	if it.IsMagical {
		isMag = 1
	}
	att := 0
	if it.Attunement {
		att = 1
	}
	_, err := db.DB.Exec("UPDATE oneshot_items SET name=?, description=?, category=?, quantity=?, weight=?, price_gp=?, is_magical=?, attunement=?, notes=? WHERE id=?",
		it.Name, it.Description, it.Category, it.Quantity, it.Weight, it.PriceGP, isMag, att, it.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotItem(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_items WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListItemUploads(c *gin.Context) {
	itemID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, hash, ext, url, COALESCE(resized_url,''), COALESCE(thumbnail_url,''), owner_type, owner_id, COALESCE(created_at,'') FROM uploads WHERE owner_type='item' AND owner_id=? ORDER BY created_at DESC", itemID)
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

func ListNPCsForItem(c *gin.Context) {
	itemID := c.Param("id")
	rows, err := db.DB.Query(`
		SELECT l.id, l.npc_id, l.adventure_id, l.item_id, l.relationship_type, l.notes, n.name
		FROM npc_item_links l JOIN npcs n ON n.id = l.npc_id
		WHERE l.item_id=? ORDER BY n.name`, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.NPCItemLink, 0)
	for rows.Next() {
		var l models.NPCItemLink
		rows.Scan(&l.ID, &l.NPCID, &l.AdventureID, &l.ItemID, &l.RelationshipType, &l.Notes, &l.NPCName)
		out = append(out, l)
	}
	c.JSON(http.StatusOK, out)
}
