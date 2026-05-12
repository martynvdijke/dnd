package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListCharacterNotes(c *gin.Context) {
	charID := c.Query("character_id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	if charID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character_id required"})
		return
	}

	rows, err := db.DB.Query("SELECT id, character_id, title, content, visibility, category FROM character_notes WHERE character_id=? ORDER BY created_at DESC", charID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var out = make([]models.CharacterNote, 0)
	for rows.Next() {
		var n models.CharacterNote
		rows.Scan(&n.ID, &n.CharacterID, &n.Title, &n.Content, &n.Visibility, &n.Category)
		if n.Visibility == "dm" && role != "admin" {
			var ownerID int64
			db.DB.QueryRow("SELECT user_id FROM characters WHERE id=?", charID).Scan(&ownerID)
			if ownerID != userID {
				var isDM bool
				db.DB.QueryRow(`SELECT COUNT(*) > 0 FROM campaign_members cm JOIN characters c ON c.campaign_id = cm.campaign_id WHERE c.id=? AND cm.user_id=? AND cm.role='dm'`, charID, userID).Scan(&isDM)
				if !isDM {
					continue
				}
			}
		}
		out = append(out, n)
	}
	c.JSON(http.StatusOK, out)
}

func CreateCharacterNote(c *gin.Context) {
	var n models.CharacterNote
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if n.Title == "" {
		n.Title = "Untitled Note"
	}
	result, err := db.DB.Exec("INSERT INTO character_notes(character_id,title,content,visibility,category) VALUES(?,?,?,?,?)",
		n.CharacterID, n.Title, n.Content, n.Visibility, n.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCharacterNote(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var n models.CharacterNote
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE character_notes SET title=?, content=?, visibility=?, category=?, updated_at=datetime('now') WHERE id=?",
		n.Title, n.Content, n.Visibility, n.Category, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCharacterNote(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM character_notes WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
