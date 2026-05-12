package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListFeats(c *gin.Context) {
	charID := c.Query("character_id")
	if charID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character_id required"})
		return
	}
	rows, err := db.DB.Query("SELECT id, character_id, name, description, prerequisites, source, level_gained FROM character_feats WHERE character_id=? ORDER BY level_gained, name", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]models.CharacterFeat, 0)
	for rows.Next() {
		var f models.CharacterFeat
		rows.Scan(&f.ID, &f.CharacterID, &f.Name, &f.Description, &f.Prerequisites, &f.Source, &f.LevelGained)
		out = append(out, f)
	}
	c.JSON(http.StatusOK, out)
}

func CreateFeat(c *gin.Context) {
	var f models.CharacterFeat
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO character_feats(character_id,name,description,prerequisites,source,level_gained) VALUES(?,?,?,?,?,?)",
		f.CharacterID, f.Name, f.Description, f.Prerequisites, f.Source, f.LevelGained)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateFeat(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var f models.CharacterFeat
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE character_feats SET name=?, description=?, prerequisites=?, source=?, level_gained=? WHERE id=?",
		f.Name, f.Description, f.Prerequisites, f.Source, f.LevelGained, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteFeat(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM character_feats WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
