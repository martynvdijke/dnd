package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

// ─── Inline Duration Editing ───

func UpdateActDuration(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		EstimatedMinutes int `json:"estimated_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("UPDATE oneshot_acts SET estimated_minutes=? WHERE id=?", req.EstimatedMinutes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "estimated_minutes": req.EstimatedMinutes})
}

func UpdateSceneDuration(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		EstimatedMinutes int `json:"estimated_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("UPDATE oneshot_scenes SET estimated_minutes=? WHERE id=?", req.EstimatedMinutes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "estimated_minutes": req.EstimatedMinutes})
}

// ─── Act/Scene Reordering ───

type reorderRequest struct {
	Order []int64 `json:"order"`
}

func ReorderActs(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i, actID := range req.Order {
		_, err := tx.Exec("UPDATE oneshot_acts SET sort_order=? WHERE id=? AND adventure_id=?", i+1, actID, adventureID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ReorderScenes(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i, sceneID := range req.Order {
		_, err := tx.Exec("UPDATE oneshot_scenes SET sort_order=? WHERE id=? AND act_id=?", i+1, sceneID, actID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
