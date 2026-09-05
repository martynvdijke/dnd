package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── HTMX Pacing Handlers ───

func ListClues(c *gin.Context) {
	adventureID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	clues := make([]models.Clue, 0)
	for rows.Next() {
		var cl models.Clue
		if err := rows.Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt); err == nil {
			loadClueRelations(&cl)
			clues = append(clues, cl)
		}
	}
	c.JSON(http.StatusOK, clues)
}

func CreateClue(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Verify adventure belongs to user
	var exists bool
	db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM oneshot_adventures WHERE id=? AND user_id=?)", adventureID, userID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "adventure not found"})
		return
	}

	var input struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		ClueType     string `json:"clue_type"`
		IsRedHerring bool   `json:"is_red_herring"`
		SortOrder    int    `json:"sort_order"`
		Notes        string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.ClueType == "" {
		input.ClueType = "direct"
	}

	result, err := db.DB.Exec("INSERT INTO clues(adventure_id, title, description, clue_type, is_red_herring, sort_order, notes) VALUES(?,?,?,?,?,?,?)",
		adventureID, input.Title, input.Description, input.ClueType, boolToInt(input.IsRedHerring), input.SortOrder, input.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetClue(c *gin.Context) {
	id := c.Param("id")
	var cl models.Clue
	err := db.DB.QueryRow("SELECT id, adventure_id, title, description, clue_type, is_red_herring, is_revealed, sort_order, notes, created_at, updated_at FROM clues WHERE id=?", id).
		Scan(&cl.ID, &cl.AdventureID, &cl.Title, &cl.Description, &cl.ClueType, &cl.IsRedHerring, &cl.IsRevealed, &cl.SortOrder, &cl.Notes, &cl.CreatedAt, &cl.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "clue not found"})
		return
	}
	loadClueRelations(&cl)
	c.JSON(http.StatusOK, cl)
}

func UpdateClue(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		ClueType     string `json:"clue_type"`
		IsRedHerring bool   `json:"is_red_herring"`
		SortOrder    int    `json:"sort_order"`
		Notes        string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("UPDATE clues SET title=?, description=?, clue_type=?, is_red_herring=?, sort_order=?, notes=?, updated_at=datetime('now') WHERE id=?",
		input.Title, input.Description, input.ClueType, boolToInt(input.IsRedHerring), input.SortOrder, input.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func DeleteClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("DELETE FROM clues WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func RevealClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("UPDATE clues SET is_revealed=1, updated_at=datetime('now') WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "revealed"})
}

func HideClue(c *gin.Context) {
	id := c.Param("id")
	db.DB.Exec("UPDATE clues SET is_revealed=0, updated_at=datetime('now') WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"status": "hidden"})
}

func AddClueDependency(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		DependsOnID int64 `json:"depends_on_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO clue_dependencies(clue_id, depends_on_id) VALUES(?,?)", id, input.DependsOnID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "linked"})
}

func RemoveClueDependency(c *gin.Context) {
	id := c.Param("id")
	depID := c.Param("did")
	db.DB.Exec("DELETE FROM clue_dependencies WHERE clue_id=? AND depends_on_id=?", id, depID)
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func LinkClueNPC(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		NPCID int64 `json:"npc_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO clue_npcs(clue_id, npc_id) VALUES(?,?)", id, input.NPCID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

func UnlinkClueNPC(c *gin.Context) {
	id := c.Param("id")
	npcID := c.Param("nid")
	db.DB.Exec("DELETE FROM clue_npcs WHERE clue_id=? AND npc_id=?", id, npcID)
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func LinkClueLocation(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		LocationID int64 `json:"location_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT OR IGNORE INTO clue_locations(clue_id, location_id) VALUES(?,?)", id, input.LocationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

func UnlinkClueLocation(c *gin.Context) {
	id := c.Param("id")
	locID := c.Param("lid")
	db.DB.Exec("DELETE FROM clue_locations WHERE clue_id=? AND location_id=?", id, locID)
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}
