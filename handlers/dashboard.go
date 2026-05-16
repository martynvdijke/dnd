package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
)

type CampaignDashboard struct {
	ID               int64                   `json:"id"`
	Name             string                  `json:"name"`
	PartyName        string                  `json:"party_name"`
	ActiveQuests     int                     `json:"active_quests"`
	UpcomingSessions int                     `json:"upcoming_sessions"`
	TotalMembers     int                     `json:"total_members"`
	ActiveConditions int                     `json:"active_conditions"`
	Weather          *WeatherResult          `json:"weather,omitempty"`
	RecentJournal    int                     `json:"recent_journal"`
	UpcomingEvents   []CalendarEventSummary  `json:"upcoming_events"`
	RecentTimeline   []TimelineEventSummary  `json:"recent_timeline"`
	CharacterSummary []CharacterDashSummary  `json:"characters"`
	DowntimeCount    int                     `json:"downtime_count"`
}

type CalendarEventSummary struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	EventDate string `json:"event_date"`
	EventType string `json:"event_type"`
	Color     string `json:"color"`
}

type TimelineEventSummary struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	EventDate   string `json:"event_date"`
	EventType   string `json:"event_type"`
	Importance  int    `json:"importance"`
}

type CharacterDashSummary struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Race      string `json:"race"`
	Class     string `json:"class"`
	Level     int    `json:"level"`
	HPCurrent int    `json:"hp_current"`
	HPMax     int    `json:"hp_max"`
}

func GetCampaignDashboard(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var dash CampaignDashboard
	dash.ID = campaignID

	db.DB.QueryRow("SELECT name, COALESCE(party_name,'') FROM campaigns WHERE id=?", campaignID).Scan(&dash.Name, &dash.PartyName)

	db.DB.QueryRow("SELECT COUNT(*) FROM quests q JOIN characters c ON q.character_id=c.id WHERE c.campaign_id=? AND q.status='active'", campaignID).Scan(&dash.ActiveQuests)

	rows, _ := db.DB.Query("SELECT COUNT(*) FROM campaign_calendar_events WHERE campaign_id=? AND event_date >= date('now') AND event_type='session'", campaignID)
	if rows != nil {
		rows.Next()
		rows.Scan(&dash.UpcomingSessions)
		rows.Close()
	}

	db.DB.QueryRow("SELECT COUNT(*) FROM campaign_members WHERE campaign_id=?", campaignID).Scan(&dash.TotalMembers)
	_ = dash.TotalMembers

	db.DB.QueryRow("SELECT COUNT(*) FROM character_conditions cc JOIN characters c ON cc.character_id=c.id WHERE c.campaign_id=?", campaignID).Scan(&dash.ActiveConditions)

	_ = db.DB.QueryRow("SELECT COUNT(*) FROM journal j JOIN characters c ON j.character_id=c.id WHERE c.campaign_id=? AND j.created_at >= datetime('now', '-7 days')", campaignID).Scan(&dash.RecentJournal)

	// Upcoming calendar events
	calRows, _ := db.DB.Query("SELECT id, title, event_date, event_type, color FROM campaign_calendar_events WHERE campaign_id=? AND event_date >= date('now') ORDER BY event_date LIMIT 5", campaignID)
	if calRows != nil {
		for calRows.Next() {
			var ev CalendarEventSummary
			calRows.Scan(&ev.ID, &ev.Title, &ev.EventDate, &ev.EventType, &ev.Color)
			dash.UpcomingEvents = append(dash.UpcomingEvents, ev)
		}
		calRows.Close()
	}

	// Recent timeline events
	tlRows, _ := db.DB.Query("SELECT id, title, event_date, event_type, importance FROM campaign_timeline_events WHERE campaign_id=? ORDER BY event_date DESC LIMIT 5", campaignID)
	if tlRows != nil {
		for tlRows.Next() {
			var tl TimelineEventSummary
			tlRows.Scan(&tl.ID, &tl.Title, &tl.EventDate, &tl.EventType, &tl.Importance)
			dash.RecentTimeline = append(dash.RecentTimeline, tl)
		}
		tlRows.Close()
	}

	// Character summaries
	charRows, _ := db.DB.Query("SELECT id, name, race, class, level, hp_current, hp_max FROM characters WHERE campaign_id=? ORDER BY name", campaignID)
	if charRows != nil {
		for charRows.Next() {
			var cs CharacterDashSummary
			charRows.Scan(&cs.ID, &cs.Name, &cs.Race, &cs.Class, &cs.Level, &cs.HPCurrent, &cs.HPMax)
			dash.CharacterSummary = append(dash.CharacterSummary, cs)
		}
		charRows.Close()
	}

	// Active downtime activities
	db.DB.QueryRow("SELECT COUNT(*) FROM downtime_activities da JOIN characters c ON da.character_id=c.id WHERE c.campaign_id=? AND da.status='in-progress'", campaignID).Scan(&dash.DowntimeCount)

	c.JSON(http.StatusOK, dash)
}
