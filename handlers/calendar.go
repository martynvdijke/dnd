package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func ListCalendarEvents(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campaign_id required"})
		return
	}

	rows, err := db.DB.Query("SELECT id, campaign_id, title, description, event_date, event_type, color, created_at FROM campaign_calendar_events WHERE campaign_id=? ORDER BY event_date", campaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var out = make([]models.CalendarEvent, 0)
	for rows.Next() {
		var e models.CalendarEvent
		rows.Scan(&e.ID, &e.CampaignID, &e.Title, &e.Description, &e.EventDate, &e.EventType, &e.Color, &e.CreatedAt)
		out = append(out, e)
	}
	c.JSON(http.StatusOK, out)
}

func CreateCalendarEvent(c *gin.Context) {
	var e models.CalendarEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if e.Color == "" {
		e.Color = "#b8963e"
	}
	result, err := db.DB.Exec("INSERT INTO campaign_calendar_events(campaign_id,title,description,event_date,event_type,color) VALUES(?,?,?,?,?,?)",
		e.CampaignID, e.Title, e.Description, e.EventDate, e.EventType, e.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCalendarEvent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var e models.CalendarEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE campaign_calendar_events SET title=?, description=?, event_date=?, event_type=?, color=? WHERE id=?",
		e.Title, e.Description, e.EventDate, e.EventType, e.Color, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCalendarEvent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("DELETE FROM campaign_calendar_events WHERE id=?", id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
