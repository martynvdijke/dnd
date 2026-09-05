package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── Scenes ───

func CreateOneShotScene(c *gin.Context) {
	actID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var sc models.OneShotScene
	if err := c.ShouldBindJSON(&sc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if sc.Number == 0 {
		db.DB.QueryRow("SELECT COALESCE(MAX(number),0)+1 FROM oneshot_scenes WHERE act_id=?", actID).Scan(&sc.Number)
	}
	if sc.SortOrder == 0 {
		db.DB.QueryRow("SELECT COALESCE(MAX(sort_order),0)+1 FROM oneshot_scenes WHERE act_id=?", actID).Scan(&sc.SortOrder)
	}
	ctx := c.Request.Context()
	result, err := db.Client.OneShotScene.Create().
		SetActID(actID).
		SetNumber(sc.Number).
		SetSortOrder(sc.SortOrder).
		SetTitle(sc.Title).
		SetDescription(sc.Description).
		SetSceneType(sc.SceneType).
		SetNillableLocationID(sc.LocationID).
		SetNillableEncounterID(sc.EncounterID).
		SetEstimatedMinutes(sc.EstimatedMinutes).
		SetNotes(sc.Notes).
		Save(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateOneShotScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var sc models.OneShotScene
	if err := c.ShouldBindJSON(&sc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	_, err := db.Client.OneShotScene.UpdateOneID(id).
		SetNumber(sc.Number).
		SetSortOrder(sc.SortOrder).
		SetTitle(sc.Title).
		SetDescription(sc.Description).
		SetSceneType(sc.SceneType).
		SetNillableLocationID(sc.LocationID).
		SetNillableEncounterID(sc.EncounterID).
		SetEstimatedMinutes(sc.EstimatedMinutes).
		SetNotes(sc.Notes).
		Save(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteOneShotScene(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()
	db.Client.OneShotScene.DeleteOneID(id).Exec(ctx)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ReorderDialogs(c *gin.Context) {
	sceneID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scene id"})
		return
	}
	var req struct {
		Order []int64 `json:"order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for i, dialogID := range req.Order {
		db.DB.Exec("UPDATE oneshot_scene_dialogs SET sort_order=? WHERE id=? AND scene_id=?", i+1, dialogID, sceneID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
