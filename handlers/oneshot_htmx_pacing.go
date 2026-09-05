package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

type CampaignOverviewData struct {
	ID               int64
	Name             string
	PartyName        string
	ActiveQuests     int
	UpcomingSessions int
	TotalMembers     int
	ActiveConditions int
	Weather          *WeatherResult
	RecentJournal    int
	Characters       []CharacterDashSummary
	UpcomingEvents   []CalendarEventSummary
	RecentTimeline   []TimelineEventSummary
	OneShots         []models.OneShotAdventure
	RecentRecaps     []RecapSummary
	RecentCombats    []CombatSummary
	RecentDiceRolls  []DiceRollSummary
}

func HtmxCampaignOverview(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var data CampaignOverviewData
	data.ID = campaignID

	db.DB.QueryRow("SELECT name, COALESCE(party_name,'') FROM campaigns WHERE id=?", campaignID).Scan(&data.Name, &data.PartyName)

	db.DB.QueryRow("SELECT COUNT(*) FROM quests q JOIN characters c ON q.character_id=c.id WHERE c.campaign_id=? AND q.status='active'", campaignID).Scan(&data.ActiveQuests)

	rows, _ := db.DB.Query("SELECT COUNT(*) FROM campaign_calendar_events WHERE campaign_id=? AND event_date >= date('now') AND event_type='session'", campaignID)
	if rows != nil {
		rows.Next()
		rows.Scan(&data.UpcomingSessions)
		rows.Close()
	}

	db.DB.QueryRow("SELECT COUNT(*) FROM campaign_members WHERE campaign_id=?", campaignID).Scan(&data.TotalMembers)

	db.DB.QueryRow("SELECT COUNT(*) FROM character_conditions cc JOIN characters c ON cc.character_id=c.id WHERE c.campaign_id=?", campaignID).Scan(&data.ActiveConditions)

	db.DB.QueryRow("SELECT COUNT(*) FROM journal j JOIN characters c ON j.character_id=c.id WHERE c.campaign_id=? AND j.created_at >= datetime('now', '-7 days')", campaignID).Scan(&data.RecentJournal)

	data.Weather = getCampaignWeather(campaignID)

	calRows, _ := db.DB.Query("SELECT id, title, event_date, event_type, color FROM campaign_calendar_events WHERE campaign_id=? AND event_date >= date('now') ORDER BY event_date LIMIT 5", campaignID)
	if calRows != nil {
		for calRows.Next() {
			var ev CalendarEventSummary
			calRows.Scan(&ev.ID, &ev.Title, &ev.EventDate, &ev.EventType, &ev.Color)
			data.UpcomingEvents = append(data.UpcomingEvents, ev)
		}
		calRows.Close()
	}

	tlRows, _ := db.DB.Query("SELECT id, title, event_date, event_type, importance FROM campaign_timeline_events WHERE campaign_id=? ORDER BY event_date DESC LIMIT 5", campaignID)
	if tlRows != nil {
		for tlRows.Next() {
			var tl TimelineEventSummary
			tlRows.Scan(&tl.ID, &tl.Title, &tl.EventDate, &tl.EventType, &tl.Importance)
			data.RecentTimeline = append(data.RecentTimeline, tl)
		}
		tlRows.Close()
	}

	charRows, _ := db.DB.Query("SELECT id, name, race, class, level, hp_current, hp_max, COALESCE(portrait_url,'') FROM characters WHERE campaign_id=? ORDER BY name", campaignID)
	if charRows != nil {
		raceColors := GetRaceColorMap()
		for charRows.Next() {
			var cs CharacterDashSummary
			charRows.Scan(&cs.ID, &cs.Name, &cs.Race, &cs.Class, &cs.Level, &cs.HPCurrent, &cs.HPMax, &cs.PortraitURL)
			cs.RaceColor = raceColors[cs.Race]
			data.Characters = append(data.Characters, cs)
		}
		charRows.Close()
	}

	// One-shots linked to this campaign
	osRows, _ := db.DB.Query("SELECT id, user_id, campaign_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at, COALESCE(is_mini_campaign,0), COALESCE(sort_order,0) FROM oneshot_adventures WHERE campaign_id=? ORDER BY sort_order ASC, updated_at DESC", campaignID)
	if osRows != nil {
		for osRows.Next() {
			var a models.OneShotAdventure
			var isMiniCampaign, sortOrder int
			osRows.Scan(&a.ID, &a.UserID, &a.CampaignID, &a.Title, &a.Premise, &a.Hook, &a.Template, &a.EstimatedMinutes, &a.Difficulty, &a.Notes, &a.CreatedAt, &a.UpdatedAt, &isMiniCampaign, &sortOrder)
			a.IsMiniCampaign = isMiniCampaign == 1
			a.SortOrder = sortOrder
			data.OneShots = append(data.OneShots, a)
		}
		osRows.Close()
	}

	recapRows, _ := db.DB.Query("SELECT id, title, COALESCE(session_start_date,''), created_at FROM campaign_recaps WHERE campaign_id=? ORDER BY created_at DESC LIMIT 3", campaignID)
	if recapRows != nil {
		for recapRows.Next() {
			var r RecapSummary
			recapRows.Scan(&r.ID, &r.Title, &r.SessionStartDate, &r.CreatedAt)
			data.RecentRecaps = append(data.RecentRecaps, r)
		}
		recapRows.Close()
	}

	combatRows, _ := db.DB.Query("SELECT id, name, round, created_at FROM combat_entries WHERE campaign_id=? ORDER BY created_at DESC LIMIT 3", campaignID)
	if combatRows != nil {
		for combatRows.Next() {
			var cs CombatSummary
			combatRows.Scan(&cs.ID, &cs.Name, &cs.Round, &cs.CreatedAt)
			data.RecentCombats = append(data.RecentCombats, cs)
		}
		combatRows.Close()
	}

	rollRows, _ := db.DB.Query(`
		SELECT dr.id, dr.expression, dr.total, dr.created_at
		FROM dice_rolls dr
		JOIN characters c ON dr.character_id = c.id
		WHERE c.campaign_id=?
		ORDER BY dr.created_at DESC LIMIT 5
	`, campaignID)
	if rollRows != nil {
		for rollRows.Next() {
			var dr DiceRollSummary
			rollRows.Scan(&dr.ID, &dr.Expression, &dr.Total, &dr.CreatedAt)
			data.RecentDiceRolls = append(data.RecentDiceRolls, dr)
		}
		rollRows.Close()
	}

	renderTemplate(c, "campaign_overview.html", data)
}

func getCampaignWeather(campaignID int64) *WeatherResult {
	var wr WeatherResult
	err := db.DB.QueryRow("SELECT temperature, description, COALESCE(sky,'') FROM campaign_weather WHERE campaign_id=? ORDER BY created_at DESC LIMIT 1", campaignID).Scan(&wr.Temperature, &wr.Description, &wr.Sky)
	if err != nil {
		return nil
	}
	return &wr
}

// ─── Session Pacing API Handlers ───

func HtmxGetPacingDashboard(c *gin.Context) {
	sessionID := c.Param("id")

	var s models.SessionPacing
	err := db.DB.QueryRow(`
		SELECT sp.id, sp.adventure_id, sp.current_act_id, sp.current_scene_id, sp.status, sp.elapsed_seconds, sp.started_at, COALESCE(sp.completed_at,''),
			COALESCE(oa.title,''), COALESCE(a.title,''), COALESCE(sc.title,''), COALESCE(sc.estimated_minutes,0),
			COALESCE(a.number,0), COALESCE(sc.number,0)
		FROM session_pacing sp
		LEFT JOIN oneshot_adventures oa ON oa.id = sp.adventure_id
		LEFT JOIN oneshot_acts a ON a.id = sp.current_act_id
		LEFT JOIN oneshot_scenes sc ON sc.id = sp.current_scene_id
		WHERE sp.id=?
	`, sessionID).Scan(&s.ID, &s.AdventureID, &s.CurrentActID, &s.CurrentSceneID, &s.Status, &s.ElapsedSeconds, &s.StartedAt, &s.CompletedAt,
		&s.AdventureTitle, &s.ActTitle, &s.SceneTitle, &s.SceneEstimated, &s.ActNumber, &s.SceneNumber)
	if err != nil {
		c.String(http.StatusNotFound, "Session not found")
		return
	}

	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_acts WHERE adventure_id=?", s.AdventureID).Scan(&s.TotalActs)
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_scenes sc JOIN oneshot_acts a ON a.id=sc.act_id WHERE a.adventure_id=?", s.AdventureID).Scan(&s.TotalScenes)

	rows, err := db.DB.Query(`
		SELECT st.id, st.session_id, st.scene_id, st.elapsed_seconds, st.status, COALESCE(st.started_at,''), COALESCE(st.completed_at,''),
			COALESCE(sc.title,''), COALESCE(sc.scene_type,''), COALESCE(sc.estimated_minutes,0)
		FROM scene_timings st
		LEFT JOIN oneshot_scenes sc ON sc.id = st.scene_id
		WHERE st.session_id=?
		ORDER BY st.id
	`, sessionID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st models.SceneTiming
			if err := rows.Scan(&st.ID, &st.SessionID, &st.SceneID, &st.ElapsedSeconds, &st.Status, &st.StartedAt, &st.CompletedAt,
				&st.SceneTitle, &st.SceneType, &st.EstimatedMin); err == nil {
				s.SceneTimings = append(s.SceneTimings, st)
			}
		}
	}

	paceCls, pacePct, paceLbl := computePace(s.SceneEstimated, s.SceneTimings)

	renderTemplate(c, "oneshot_pacing.html", gin.H{
		"Session":     s,
		"Elapsed":     formatDuration(s.ElapsedSeconds),
		"PaceClass":   paceCls,
		"PacePercent": pacePct,
		"PaceLabel":   paceLbl,
	})
}

func computePace(estimatedMin int, timings []models.SceneTiming) (class string, percent int, label string) {
	if estimatedMin == 0 || len(timings) == 0 {
		return "bg-success", 0, "No Estimate"
	}
	for _, t := range timings {
		if t.Status == "active" {
			elapsedMin := t.ElapsedSeconds / 60
			ratio := float64(elapsedMin) / float64(estimatedMin)
			pct := min(int(ratio*100), 100)
			if ratio >= 1.0 {
				return "bg-danger", pct, "Over Time!"
			} else if ratio >= 0.8 {
				return "bg-warning text-dark", pct, "Warning"
			}
			return "bg-success", pct, "On Track"
		}
	}
	return "bg-success", 0, "No Data"
}

func paceClass(estimatedMin int, timings []models.SceneTiming) string {
	if estimatedMin == 0 || len(timings) == 0 {
		return "bg-success"
	}
	for _, t := range timings {
		if t.Status == "active" {
			elapsedMin := t.ElapsedSeconds / 60
			ratio := float64(elapsedMin) / float64(estimatedMin)
			if ratio >= 1.0 {
				return "bg-danger"
			} else if ratio >= 0.8 {
				return "bg-warning text-dark"
			}
			return "bg-success"
		}
	}
	return "bg-success"
}

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// ─── Clue/Mystery Tracker API Handlers ───
