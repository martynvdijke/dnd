package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListOneShotAdventures(c *gin.Context) {
	userID, _ := c.Get("user_id")
	campaignID := c.Query("campaign_id")

	var rows *sql.Rows
	var err error
	if campaignID != "" {
		rows, err = db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at, COALESCE(is_mini_campaign,0), COALESCE(sort_order,0) FROM oneshot_adventures WHERE campaign_id=? ORDER BY sort_order ASC, updated_at DESC", campaignID)
	} else {
		rows, err = db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at, COALESCE(is_mini_campaign,0), COALESCE(sort_order,0) FROM oneshot_adventures WHERE user_id=? ORDER BY updated_at DESC", userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	out := make([]models.OneShotAdventure, 0)
	for rows.Next() {
		var a models.OneShotAdventure
		var isMiniCampaign, sortOrder int
		rows.Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt, &isMiniCampaign, &sortOrder)
		a.IsMiniCampaign = isMiniCampaign == 1
		a.SortOrder = sortOrder
		out = append(out, a)
	}
	c.JSON(http.StatusOK, out)
}

func CreateOneShotAdventure(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var a models.OneShotAdventure
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO oneshot_adventures(user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes) VALUES(?,?,?,?,?,?,?,?,?)",
		userID, a.CampaignID, a.Title, a.Premise, a.Hook, a.Template, a.EstimatedMinutes, a.Difficulty, a.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()

	// Generate acts/scenes from template if applicable
	if a.Template == "five_room_dungeon" {
		generateFiveRoomDungeonStructure(id, a.Difficulty)
	} else if a.Template == "default" {
		generateDefaultStructure(id)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetOneShotAdventure(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	a, err := loadAdventureDetail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "one-shot not found"})
		return
	}
	if a.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "one-shot not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func UpdateOneShotAdventure(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var a models.OneShotAdventure
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE oneshot_adventures SET title=?, premise=?, hook=?, template=?, estimated_minutes=?, difficulty=?, notes=?, updated_at=datetime('now') WHERE id=?",
		a.Title, a.Premise, a.Hook, a.Template, a.EstimatedMinutes, a.Difficulty, a.Notes, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotAdventure(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	db.DB.Exec("DELETE FROM oneshot_adventures WHERE id=? AND user_id=?", id, userID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Acts ───

type ReorderRequest struct {
	ID        int64 `json:"id"`
	SortOrder int   `json:"sort_order"`
}

func ReorderCampaignOneShots(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req []ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, item := range req {
		db.DB.Exec("UPDATE oneshot_adventures SET sort_order=? WHERE id=? AND campaign_id=?", item.SortOrder, item.ID, campaignID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListCampaignOneShots(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at, COALESCE(is_mini_campaign,0), COALESCE(sort_order,0) FROM oneshot_adventures WHERE campaign_id=? ORDER BY sort_order ASC, updated_at DESC", campaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotAdventure, 0)
	for rows.Next() {
		var a models.OneShotAdventure
		var isMiniCampaign, sortOrder int
		rows.Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt, &isMiniCampaign, &sortOrder)
		a.IsMiniCampaign = isMiniCampaign == 1
		a.SortOrder = sortOrder
		out = append(out, a)
	}
	c.JSON(http.StatusOK, out)
}

// ─── Campaign Overview ───
