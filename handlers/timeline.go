package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListTimelineEvents(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campaign_id required"})
		return
	}

	rows, err := db.DB.Query("SELECT id, campaign_id, title, description, event_date, event_type, importance, icon, linked_entity_type, linked_entity_id, created_at FROM campaign_timeline_events WHERE campaign_id=? ORDER BY event_date DESC, created_at DESC", campaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var out = make([]models.TimelineEvent, 0)
	for rows.Next() {
		var e models.TimelineEvent
		rows.Scan(&e.ID, &e.CampaignID, &e.Title, &e.Description, &e.EventDate, &e.EventType, &e.Importance, &e.Icon, &e.LinkedEntityType, &e.LinkedEntityID, &e.CreatedAt)
		out = append(out, e)
	}
	c.JSON(http.StatusOK, out)
}

func CreateTimelineEvent(c *gin.Context) {
	var e models.TimelineEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if e.Icon == "" {
		e.Icon = "fa-star"
	}
	if e.Importance < 1 {
		e.Importance = 1
	}
	if e.Importance > 5 {
		e.Importance = 5
	}
	result, err := db.DB.Exec("INSERT INTO campaign_timeline_events(campaign_id,title,description,event_date,event_type,importance,icon,linked_entity_type,linked_entity_id) VALUES(?,?,?,?,?,?,?,?,?)",
		e.CampaignID, e.Title, e.Description, e.EventDate, e.EventType, e.Importance, e.Icon, e.LinkedEntityType, e.LinkedEntityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateTimelineEvent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var e models.TimelineEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE campaign_timeline_events SET title=?, description=?, event_date=?, event_type=?, importance=?, icon=?, linked_entity_type=?, linked_entity_id=? WHERE id=?",
		e.Title, e.Description, e.EventDate, e.EventType, e.Importance, e.Icon, e.LinkedEntityType, e.LinkedEntityID, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteTimelineEvent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM campaign_timeline_events WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
