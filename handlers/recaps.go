package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type CampaignRecap struct {
	ID               int64   `json:"id"`
	CampaignID       int64   `json:"campaign_id"`
	Title            string  `json:"title"`
	Content          string  `json:"content"`
	SessionStartDate *string `json:"session_start_date,omitempty"`
	SessionEndDate   *string `json:"session_end_date,omitempty"`
	WordCount        int     `json:"word_count"`
	IsEdited         bool    `json:"is_edited"`
	IsSent           bool    `json:"is_sent"`
}

type RecapSection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func ListCampaignRecaps(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id,campaign_id,title,content,session_start_date,session_end_date,word_count,is_edited,is_sent FROM campaign_recaps WHERE campaign_id=? ORDER BY created_at DESC", campaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out = make([]CampaignRecap, 0)
	for rows.Next() {
		var r CampaignRecap
		rows.Scan(&r.ID, &r.CampaignID, &r.Title, &r.Content, &r.SessionStartDate, &r.SessionEndDate, &r.WordCount, &r.IsEdited, &r.IsSent)
		out = append(out, r)
	}
	c.JSON(http.StatusOK, out)
}

func GetCampaignRecap(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r CampaignRecap
	err := db.DB.QueryRow("SELECT id,campaign_id,title,content,session_start_date,session_end_date,word_count,is_edited,is_sent FROM campaign_recaps WHERE id=?", id).Scan(&r.ID, &r.CampaignID, &r.Title, &r.Content, &r.SessionStartDate, &r.SessionEndDate, &r.WordCount, &r.IsEdited, &r.IsSent)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recap not found"})
		return
	}
	c.JSON(http.StatusOK, r)
}

func CreateCampaignRecap(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Title            string  `json:"title"`
		Content          string  `json:"content"`
		SessionStartDate *string `json:"session_start_date"`
		SessionEndDate   *string `json:"session_end_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wordCount := len(strings.Fields(req.Content))
	if req.Title == "" {
		req.Title = "Session Recap"
	}
	result, err := db.DB.Exec("INSERT INTO campaign_recaps(campaign_id,title,content,session_start_date,session_end_date,word_count) VALUES(?,?,?,?,?,?)",
		campaignID, req.Title, req.Content, req.SessionStartDate, req.SessionEndDate, wordCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	NotifyCampaignRecapPublished(campaignID, id, req.Title)
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCampaignRecap(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Title            string  `json:"title"`
		Content          string  `json:"content"`
		SessionStartDate *string `json:"session_start_date"`
		SessionEndDate   *string `json:"session_end_date"`
		IsEdited         bool    `json:"is_edited"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wordCount := len(strings.Fields(req.Content))
	db.DB.Exec("UPDATE campaign_recaps SET title=?,content=?,session_start_date=?,session_end_date=?,word_count=?,is_edited=? WHERE id=?",
		req.Title, req.Content, req.SessionStartDate, req.SessionEndDate, wordCount, req.IsEdited, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCampaignRecap(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM campaign_recaps WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GenerateCampaignRecap(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var recaps []RecapSection

	// Get character names
	charRows, _ := db.DB.Query("SELECT id, name, race, class FROM characters WHERE campaign_id=?", campaignID)
	var charNames []string
	if charRows != nil {
		for charRows.Next() {
			var id int64
			var name, race, cls string
			charRows.Scan(&id, &name, &race, &cls)
			charNames = append(charNames, fmt.Sprintf("%s (%s %s, Lvl)", name, race, cls))
		}
		charRows.Close()
	}
	if len(charNames) > 0 {
		recaps = append(recaps, RecapSection{
			Title:   "Party Members",
			Content: "The party consists of: " + strings.Join(charNames, ", ") + ".",
		})
	}

	// Recent timeline events (last 30 days)
	tlRows, _ := db.DB.Query("SELECT title, description, event_type, event_date FROM campaign_timeline_events WHERE campaign_id=? AND event_date >= date('now', '-30 days') ORDER BY event_date DESC LIMIT 10", campaignID)
	var timelineEvents []string
	if tlRows != nil {
		for tlRows.Next() {
			var title, desc, etype, edate string
			tlRows.Scan(&title, &desc, &etype, &edate)
			entry := fmt.Sprintf("[%s] %s: %s", edate, title, desc)
			timelineEvents = append(timelineEvents, entry)
		}
		tlRows.Close()
	}
	if len(timelineEvents) > 0 {
		recaps = append(recaps, RecapSection{
			Title:   "Recent Events",
			Content: strings.Join(timelineEvents, "\n"),
		})
	}

	// Recently completed quests
	questRows, _ := db.DB.Query(`
		SELECT q.name, q.description FROM quests q
		JOIN characters c ON q.character_id=c.id
		WHERE c.campaign_id=? AND q.status='complete' AND q.updated_at >= datetime('now', '-30 days')
		ORDER BY q.updated_at DESC LIMIT 5`, campaignID)
	var completedQuests []string
	if questRows != nil {
		for questRows.Next() {
			var name, desc string
			questRows.Scan(&name, &desc)
			completedQuests = append(completedQuests, fmt.Sprintf("%s: %s", name, desc))
		}
		questRows.Close()
	}
	if len(completedQuests) > 0 {
		recaps = append(recaps, RecapSection{
			Title:   "Completed Quests",
			Content: strings.Join(completedQuests, "\n"),
		})
	}

	// Active quests
	aRows, _ := db.DB.Query(`
		SELECT q.name, q.description FROM quests q
		JOIN characters c ON q.character_id=c.id
		WHERE c.campaign_id=? AND q.status='active' ORDER BY q.name`, campaignID)
	var activeQuests []string
	if aRows != nil {
		for aRows.Next() {
			var name, desc string
			aRows.Scan(&name, &desc)
			activeQuests = append(activeQuests, fmt.Sprintf("%s: %s", name, desc))
		}
		aRows.Close()
	}
	if len(activeQuests) > 0 {
		recaps = append(recaps, RecapSection{
			Title:   "Active Quests",
			Content: strings.Join(activeQuests, "\n"),
		})
	}

	// Recent sessions
	sessRows, _ := db.DB.Query(`
		SELECT s.title, s.notes, s.session_date FROM sessions s
		JOIN characters c ON s.character_id=c.id
		WHERE c.campaign_id=? AND s.created_at >= datetime('now', '-30 days')
		ORDER BY s.session_date DESC LIMIT 5`, campaignID)
	var recentSessions []string
	if sessRows != nil {
		for sessRows.Next() {
			var title, notes, sdate string
			sessRows.Scan(&title, &notes, &sdate)
			recentSessions = append(recentSessions, fmt.Sprintf("%s - %s: %s", sdate, title, notes))
		}
		sessRows.Close()
	}
	if len(recentSessions) > 0 {
		recaps = append(recaps, RecapSection{
			Title:   "Recent Sessions",
			Content: strings.Join(recentSessions, "\n"),
		})
	}

	// Future events
	calRows, _ := db.DB.Query("SELECT title, event_date, event_type FROM campaign_calendar_events WHERE campaign_id=? AND event_date >= date('now') ORDER BY event_date LIMIT 5", campaignID)
	var upcoming []string
	if calRows != nil {
		for calRows.Next() {
			var title, edate, etype string
			calRows.Scan(&title, &edate, &etype)
			upcoming = append(upcoming, fmt.Sprintf("%s - %s [%s]", edate, title, etype))
		}
		calRows.Close()
	}
	if len(upcoming) > 0 {
		recaps = append(recaps, RecapSection{
			Title:   "Upcoming Events",
			Content: strings.Join(upcoming, "\n"),
		})
	}

	// Active conditions
	condRows, _ := db.DB.Query(`
		SELECT cc.name, c.name FROM character_conditions cc
		JOIN characters c ON cc.character_id=c.id
		WHERE c.campaign_id=? AND cc.duration > 0`, campaignID)
	var conditions []string
	if condRows != nil {
		for condRows.Next() {
			var cname, charName string
			condRows.Scan(&cname, &charName)
			conditions = append(conditions, fmt.Sprintf("%s is affected by %s", charName, cname))
		}
		condRows.Close()
	}
	if len(conditions) > 0 {
		recaps = append(recaps, RecapSection{
			Title:   "Active Conditions",
			Content: strings.Join(conditions, "\n"),
		})
	}

	// Build generated content
	var sections []string
	for _, s := range recaps {
		sections = append(sections, fmt.Sprintf("## %s\n%s", s.Title, s.Content))
	}

	var startDate, endDate *string
	if len(timelineEvents) > 0 {
		s := "recent"
		endDate = &s
	}

	generatedContent := strings.Join(sections, "\n\n")
	title := fmt.Sprintf("Campaign Recap - %s", getDateStr())

	var recap CampaignRecap
	recap.CampaignID = campaignID
	recap.Title = title
	recap.Content = generatedContent
	recap.SessionStartDate = startDate
	recap.SessionEndDate = endDate
	recap.WordCount = len(strings.Fields(generatedContent))

	c.JSON(http.StatusOK, recap)
}

func getDateStr() string {
	return strings.Split(getCurrentTimestamp(), " ")[0]
}

func getCurrentTimestamp() string {
	var ts string
	db.DB.QueryRow("SELECT datetime('now')").Scan(&ts)
	return ts
}

func MarkRecapAsSent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var campaignID int64
	var title string
	if err := db.DB.QueryRow("SELECT campaign_id,title FROM campaign_recaps WHERE id=?", id).Scan(&campaignID, &title); err == nil {
		NotifyCampaignRecapPublished(campaignID, id, title)
	}
	db.DB.Exec("UPDATE campaign_recaps SET is_sent=1 WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
