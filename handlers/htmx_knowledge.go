package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

func HtmxKnowledgeList(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	if !knowledgeVisible(campaignID, uid, r) {
		c.String(http.StatusNotFound, "not found")
		return
	}
	owner := isOwner(campaignID, uid, r)
	statusFilter := c.Query("status")
	query := `SELECT id,title,content,source,status,shared FROM campaign_knowledge WHERE campaign_id=?`
	args := []any{campaignID}
	if statusFilter != "" {
		query += ` AND status=?`
		args = append(args, statusFilter)
	}
	if !owner {
		query += ` AND shared=1`
	}
	query += ` ORDER BY created_at DESC`
	rows, _ := db.DB.Query(query, args...)
	defer func() { if rows != nil { rows.Close() } }()
	type Card struct {
		ID     int64
		Title  string
		Status string
		Source string
		Shared bool
	}
	var cards []Card
	if rows != nil {
		for rows.Next() {
			var ca Card
			var shared int
			var content string
			rows.Scan(&ca.ID, &ca.Title, &content, &ca.Source, &ca.Status, &shared)
			ca.Shared = shared != 0
			cards = append(cards, ca)
		}
	}
	c.Header("Content-Type", "text/html")
	// simple html with data-testids
	html := `<div data-testid="knowledge-list"><div data-testid="knowledge-status-filter"><a href="?status=">All</a> <a href="?status=rumor">Rumor</a> <a href="?status=confirmed">Confirmed</a> <a href="?status=revealed">Revealed</a> <a href="?status=false">False</a></div>`
	for _, ca := range cards {
		html += `<div data-testid="knowledge-card" data-id="` + strconv.FormatInt(ca.ID, 10) + `" data-status="` + ca.Status + `">` + ca.Title + ` - ` + ca.Status + `</div>`
	}
	html += `</div>`
	c.String(http.StatusOK, html)
}

func HtmxKnowledgeDetail(c *gin.Context) {
	kid, _ := strconv.ParseInt(c.Param("kid"), 10, 64)
	userID, _ := c.Get("user_id")
	uid := userID.(int64)
	role, _ := c.Get("role")
	r, _ := role.(string)
	k, err := knowledgeRowToStruct(kid)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	owner := isOwner(k.CampaignID, uid, r)
	if !k.Shared && !owner {
		c.String(http.StatusNotFound, "not found")
		return
	}
	// known_by ids
	rows, _ := db.DB.Query(`SELECT character_id FROM campaign_knowledge_known_by WHERE knowledge_id=?`, kid)
	known := []int64{}
	if rows != nil {
		for rows.Next() {
			var id int64
			rows.Scan(&id)
			known = append(known, id)
		}
		rows.Close()
	}
	c.Header("Content-Type", "text/html")
	html := `<div data-testid="knowledge-detail" data-id="` + strconv.FormatInt(k.ID, 10) + `">`
	html += `<h3>` + k.Title + `</h3>`
	html += `<div>` + k.Content + `</div>`
	html += `<div>Source: ` + k.Source + `</div>`
	html += `<div>Status: ` + k.Status + `</div>`
	if owner {
		html += `<button data-testid="share-toggle">Share: ` + strconv.FormatBool(k.Shared) + `</button>`
		html += `<div data-testid="known-by-picker">Known by: `
		for _, id := range known {
			html += strconv.FormatInt(id, 10) + ` `
		}
		html += `</div>`
		html += `<button data-testid="bulk-reveal-btn">Reveal to party</button>`
		html += `<div data-testid="entity-link-picker"></div>`
	}
	html += `</div>`
	c.String(http.StatusOK, html)
}
