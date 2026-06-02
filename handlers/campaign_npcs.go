package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListCampaignNPCs(c *gin.Context) {
	campaignID := c.Param("id")
	rows, err := db.DB.Query(`
		SELECT cn.id, cn.campaign_id, cn.npc_id, cn.role, cn.notes, cn.created_at,
		       COALESCE(n.name,'') as npc_name, COALESCE(n.race,'') as npc_race, COALESCE(n.class,'') as npc_class
		FROM campaign_npcs cn
		JOIN npcs n ON n.id = cn.npc_id
		WHERE cn.campaign_id=? ORDER BY n.name`, campaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.CampaignNPC, 0)
	for rows.Next() {
		var npc models.CampaignNPC
		rows.Scan(&npc.ID, &npc.CampaignID, &npc.NPCID, &npc.Role, &npc.Notes, &npc.CreatedAt,
			&npc.NPCName, &npc.NPCRace, &npc.NPCClass)
		out = append(out, npc)
	}
	c.JSON(http.StatusOK, out)
}

func LinkNPCCampaign(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		NPCID int64  `json:"npc_id"`
		Role  string `json:"role"`
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO campaign_npcs(campaign_id, npc_id, role, notes) VALUES(?,?,?,?)",
		campaignID, req.NPCID, req.Role, req.Notes)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "NPC is already linked to this campaign"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCampaignNPC(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Role  string `json:"role"`
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("UPDATE campaign_npcs SET role=?, notes=? WHERE id=?", req.Role, req.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func UnlinkCampaignNPC(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM campaign_npcs WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func CreateAndLinkCampaignNPC(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Name        string `json:"name"`
		Race        string `json:"race"`
		Class       string `json:"class"`
		Description string `json:"description"`
		Role        string `json:"role"`
		Notes       string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create NPC first
	result, err := db.DB.Exec(`INSERT INTO npcs(user_id, name, race, class, description, notes) VALUES(?,?,?,?,?,?)`,
		0, req.Name, req.Race, req.Class, req.Description, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	npcID, _ := result.LastInsertId()

	// Link to campaign
	linkResult, err := db.DB.Exec("INSERT INTO campaign_npcs(campaign_id, npc_id, role, notes) VALUES(?,?,?,?)",
		campaignID, npcID, req.Role, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	linkID, _ := linkResult.LastInsertId()

	c.JSON(http.StatusCreated, gin.H{"npc_id": npcID, "link_id": linkID})
}
