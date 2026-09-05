package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListPrepChecklist(c *gin.Context) {
	adventureID := c.Param("id")
	rows, err := db.DB.Query("SELECT id, adventure_id, item, category, is_checked, sort_order FROM prep_checklist WHERE adventure_id=? ORDER BY sort_order, id", adventureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []models.PrepChecklistItem
	for rows.Next() {
		var item models.PrepChecklistItem
		var checked int
		if err := rows.Scan(&item.ID, &item.AdventureID, &item.Item, &item.Category, &checked, &item.SortOrder); err == nil {
			item.IsChecked = checked == 1
			items = append(items, item)
		}
	}
	c.JSON(http.StatusOK, items)
}

func CreatePrepChecklistItem(c *gin.Context) {
	adventureID := c.Param("id")
	var req struct {
		Item      string `json:"item"`
		Category  string `json:"category"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Category == "" {
		req.Category = "general"
	}

	result, err := db.DB.Exec("INSERT INTO prep_checklist(adventure_id, item, category, sort_order) VALUES(?,?,?,?)", adventureID, req.Item, req.Category, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdatePrepChecklistItem(c *gin.Context) {
	id := c.Param("cid")
	var req struct {
		Item      string `json:"item,omitempty"`
		Category  string `json:"category,omitempty"`
		IsChecked *bool  `json:"is_checked,omitempty"`
		SortOrder *int   `json:"sort_order,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Item != "" {
		db.DB.Exec("UPDATE prep_checklist SET item=? WHERE id=?", req.Item, id)
	}
	if req.Category != "" {
		db.DB.Exec("UPDATE prep_checklist SET category=? WHERE id=?", req.Category, id)
	}
	if req.IsChecked != nil {
		val := 0
		if *req.IsChecked {
			val = 1
		}
		db.DB.Exec("UPDATE prep_checklist SET is_checked=? WHERE id=?", val, id)
	}
	if req.SortOrder != nil {
		db.DB.Exec("UPDATE prep_checklist SET sort_order=? WHERE id=?", *req.SortOrder, id)
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func DeletePrepChecklistItem(c *gin.Context) {
	id := c.Param("cid")
	_, err := db.DB.Exec("DELETE FROM prep_checklist WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// HTMX handlers for checklist
func ListDmNotes(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	rows, err := db.DB.Query("SELECT id, adventure_id, user_id, title, content, created_at, updated_at FROM dm_notes WHERE adventure_id=? AND user_id=? ORDER BY updated_at DESC", adventureID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var notes []models.DmNote
	for rows.Next() {
		var n models.DmNote
		if err := rows.Scan(&n.ID, &n.AdventureID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err == nil {
			notes = append(notes, n)
		}
	}
	if notes == nil {
		notes = []models.DmNote{}
	}
	c.JSON(http.StatusOK, notes)
}

func CreateDmNote(c *gin.Context) {
	adventureID := c.Param("id")
	userID, _ := c.Get("user_id")

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		ActID   *int64 `json:"act_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.DB.Exec("INSERT INTO dm_notes(adventure_id, user_id, title, content, act_id) VALUES(?,?,?,?,?)", adventureID, userID, input.Title, input.Content, input.ActID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateDmNote(c *gin.Context) {
	noteID := c.Param("nid")
	userID, _ := c.Get("user_id")

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec("UPDATE dm_notes SET title=?, content=?, updated_at=datetime('now') WHERE id=? AND user_id=?", input.Title, input.Content, noteID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func DeleteDmNote(c *gin.Context) {
	noteID := c.Param("nid")
	userID, _ := c.Get("user_id")

	_, err := db.DB.Exec("DELETE FROM dm_notes WHERE id=? AND user_id=?", noteID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ─── One-Shot Items Section (HTMX) ───
