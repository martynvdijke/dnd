package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type WikiPage struct {
	ID         int64  `json:"id"`
	CampaignID int64  `json:"campaign_id"`
	UserID     int64  `json:"user_id"`
	ParentID   *int64 `json:"parent_id,omitempty"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
	Tags       string `json:"tags"`
	SortOrder  int    `json:"sort_order"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func ListWikiPages(c *gin.Context) {
	campaignID := c.Param("id")
	rows, err := db.DB.Query("SELECT id,campaign_id,user_id,parent_id,title,content,visibility,tags,sort_order,created_at,updated_at FROM campaign_wiki_pages WHERE campaign_id=? ORDER BY sort_order, title", campaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var pages []WikiPage
	for rows.Next() {
		var p WikiPage
		rows.Scan(&p.ID, &p.CampaignID, &p.UserID, &p.ParentID, &p.Title, &p.Content, &p.Visibility, &p.Tags, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
		pages = append(pages, p)
	}
	c.JSON(http.StatusOK, pages)
}

func CreateWikiPage(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var p WikiPage
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := db.DB.Exec("INSERT INTO campaign_wiki_pages(campaign_id,user_id,parent_id,title,content,visibility,tags,sort_order) VALUES(?,?,?,?,?,?,?,?)",
		p.CampaignID, userID, p.ParentID, p.Title, p.Content, p.Visibility, p.Tags, p.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id, "ok": true})
}

func GetWikiPage(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p WikiPage
	err := db.DB.QueryRow("SELECT id,campaign_id,user_id,parent_id,title,content,visibility,tags,sort_order,created_at,updated_at FROM campaign_wiki_pages WHERE id=?", id).
		Scan(&p.ID, &p.CampaignID, &p.UserID, &p.ParentID, &p.Title, &p.Content, &p.Visibility, &p.Tags, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func UpdateWikiPage(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p WikiPage
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE campaign_wiki_pages SET title=?,content=?,visibility=?,tags=?,sort_order=?,updated_at=datetime('now') WHERE id=?", p.Title, p.Content, p.Visibility, p.Tags, p.SortOrder, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteWikiPage(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM campaign_wiki_pages WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
