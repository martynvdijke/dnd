package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListLinkedCharacters(c *gin.Context) {
	adventureID := c.Param("id")
	rows, err := db.DB.Query(`
		SELECT pc.id, pc.adventure_id, pc.character_id, pc.role, pc.notes, c.name, u.username
		FROM oneshot_player_characters pc
		JOIN characters c ON c.id = pc.character_id
		JOIN users u ON u.id = c.user_id
		WHERE pc.adventure_id=? ORDER BY c.name`, adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]models.OneShotPlayerCharacter, 0)
	for rows.Next() {
		var pc models.OneShotPlayerCharacter
		rows.Scan(&pc.ID, &pc.AdventureID, &pc.CharacterID, &pc.Role, &pc.Notes, &pc.CharName, &pc.Username)
		out = append(out, pc)
	}
	c.JSON(http.StatusOK, out)
}

func LinkCharacterToOneShot(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		CharacterID int64  `json:"character_id"`
		Role        string `json:"role"`
		Notes       string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := db.DB.Exec("INSERT INTO oneshot_player_characters(adventure_id, character_id, role, notes) VALUES(?,?,?,?)",
		adventureID, req.CharacterID, req.Role, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UnlinkCharacterFromOneShot(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	charID, _ := strconv.ParseInt(c.Param("charId"), 10, 64)
	db.DB.Exec("DELETE FROM oneshot_player_characters WHERE adventure_id=? AND character_id=?", adventureID, charID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
