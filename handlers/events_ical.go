package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"villum/db"
)

func CampaignEventsICal(c *gin.Context) {
	slug := c.Param("slug")
	cs := db.GetCampaignEventSettingsBySlug(slug)
	if cs == nil || !cs.IsActive {
		c.String(http.StatusNotFound, "Campaign not found")
		return
	}

	settings := campaignToGlobalSettings(cs)
	events, _ := fetchAndCacheEvents(settings, slug)

	c.Header("Content-Type", "text/calendar")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.ics"`, slug))

	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//villum//Events//EN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")
	fmt.Fprintf(&b, "X-WR-CALNAME:%s\r\n", escapeICS(cs.DisplayName))

	now := time.Now()
	for _, e := range events {
		b.WriteString("BEGIN:VEVENT\r\n")
		fmt.Fprintf(&b, "UID:%s@%s.villum.events\r\n", e.ID, slug)
		fmt.Fprintf(&b, "DTSTAMP:%s\r\n", icsDateTime(now))
		if e.AllDay {
			fmt.Fprintf(&b, "DTSTART;VALUE=DATE:%s\r\n", icsDate(e.StartTime))
			endTime := e.EndTime
			if endTime.IsZero() || endTime.Equal(e.StartTime) {
				endTime = e.StartTime.Add(24 * time.Hour)
			}
			fmt.Fprintf(&b, "DTEND;VALUE=DATE:%s\r\n", icsDate(endTime))
		} else {
			fmt.Fprintf(&b, "DTSTART:%s\r\n", icsDateTime(e.StartTime))
			if !e.EndTime.IsZero() {
				fmt.Fprintf(&b, "DTEND:%s\r\n", icsDateTime(e.EndTime))
			}
		}
		fmt.Fprintf(&b, "SUMMARY:%s\r\n", escapeICS(e.Title))
		if e.Description != "" {
			fmt.Fprintf(&b, "DESCRIPTION:%s\r\n", escapeICS(e.Description))
		}
		if e.Location != "" {
			fmt.Fprintf(&b, "LOCATION:%s\r\n", escapeICS(e.Location))
		}
		b.WriteString("END:VEVENT\r\n")
	}
	b.WriteString("END:VCALENDAR\r\n")
	c.String(http.StatusOK, b.String())
}

// campaignToGlobalSettings merges campaign settings with global defaults for auth.
func campaignToGlobalSettings(cs *db.CampaignEventSettings) db.EventSettings {
	gs := db.GetEventSettings()
	s := db.EventSettings{
		SourceType:      cs.SourceType,
		ICalURL:         cs.ICalURL,
		CalendarID:      cs.CalendarID,
		Tags:            cs.Tags,
		ColorLabels:     cs.ColorLabels,
		FilterMode:      cs.FilterMode,
		CacheTTLSeconds: cs.CacheTTLSeconds,
	}
	// Use campaign-specific credentials if set, else fall back to global
	if cs.CredentialsJSON != "" {
		s.CredentialsJSON = cs.CredentialsJSON
		s.AuthMethod = cs.AuthMethod
		s.OAuthClientID = cs.OAuthClientID
		s.OAuthClientSecret = cs.OAuthClientSecret
		s.OAuthRefreshToken = cs.OAuthRefreshToken
	} else {
		s.CredentialsJSON = gs.CredentialsJSON
		s.AuthMethod = gs.AuthMethod
		s.OAuthClientID = gs.OAuthClientID
		s.OAuthClientSecret = gs.OAuthClientSecret
		s.OAuthRefreshToken = gs.OAuthRefreshToken
	}
	// If campaign has no source type set, inherit from global
	if s.SourceType == "" {
		s.SourceType = gs.SourceType
	}
	return s
}

// ─── ICS / iCal Feed (Global) ───

func escapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func icsDateTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

func icsDate(t time.Time) string {
	return t.UTC().Format("20060102")
}

func EventsICal(c *gin.Context) {
	settings := db.GetEventSettings()
	events, _ := fetchAndCacheEvents(settings, "")

	c.Header("Content-Type", "text/calendar")
	c.Header("Content-Disposition", `attachment; filename="events.ics"`)

	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//villum//Events//EN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")
	b.WriteString("X-WR-CALNAME:Upcoming Events\r\n")

	now := time.Now()
	for _, e := range events {
		b.WriteString("BEGIN:VEVENT\r\n")
		fmt.Fprintf(&b, "UID:%s@villum.events\r\n", e.ID)
		fmt.Fprintf(&b, "DTSTAMP:%s\r\n", icsDateTime(now))
		if e.AllDay {
			fmt.Fprintf(&b, "DTSTART;VALUE=DATE:%s\r\n", icsDate(e.StartTime))
			endTime := e.EndTime
			if endTime.IsZero() || endTime.Equal(e.StartTime) {
				endTime = e.StartTime.Add(24 * time.Hour)
			}
			fmt.Fprintf(&b, "DTEND;VALUE=DATE:%s\r\n", icsDate(endTime))
		} else {
			fmt.Fprintf(&b, "DTSTART:%s\r\n", icsDateTime(e.StartTime))
			if !e.EndTime.IsZero() {
				fmt.Fprintf(&b, "DTEND:%s\r\n", icsDateTime(e.EndTime))
			}
		}
		fmt.Fprintf(&b, "SUMMARY:%s\r\n", escapeICS(e.Title))
		if e.Description != "" {
			fmt.Fprintf(&b, "DESCRIPTION:%s\r\n", escapeICS(e.Description))
		}
		if e.Location != "" {
			fmt.Fprintf(&b, "LOCATION:%s\r\n", escapeICS(e.Location))
		}
		b.WriteString("END:VEVENT\r\n")
	}
	b.WriteString("END:VCALENDAR\r\n")
	c.String(http.StatusOK, b.String())
}

// ─── Admin Settings (Global) ───

func GetEventsSettings(c *gin.Context) {
	s := db.GetEventSettings()
	c.JSON(http.StatusOK, s)
}

func SaveEventsSettings(c *gin.Context) {
	var s db.EventSettings
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.SaveEventSettings(s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Clear global cache since settings changed
	db.ClearCache("")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ClearEventsCache(c *gin.Context) {
	if err := db.ClearCache(""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Admin Campaign Event Settings CRUD ───

func ListCampaignEventSettings(c *gin.Context) {
	settings := db.ListCampaignEventSettings()
	c.JSON(http.StatusOK, settings)
}

func GetCampaignEventSetting(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	s := db.GetCampaignEventSettingsByID(id)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func SaveCampaignEventSetting(c *gin.Context) {
	var s db.CampaignEventSettings
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Validate slug
	if s.Slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}
	if !slugRe.MatchString(s.Slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug must contain only lowercase letters, numbers, and hyphens"})
		return
	}
	id, err := db.SaveCampaignEventSettings(s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Clear per-campaign cache
	db.ClearCache(s.Slug)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// ClearCampaignCache clears the cached events for a single campaign.
func ClearCampaignCache(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	s := db.GetCampaignEventSettingsByID(id)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := db.ClearCache(s.Slug); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCampaignEventSetting(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	s := db.GetCampaignEventSettingsByID(id)
	if s != nil {
		db.ClearCache(s.Slug)
	}
	if err := db.DeleteCampaignEventSettings(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Calendar Grid View Handlers ───

// EventsGridPartial returns HTML for the grid partial (HTMX month navigation).
func EventsGridPartial(c *gin.Context) {
	settings := db.GetEventSettings()

	if settings.CalendarID == "" {
		renderTemplate(c, "events_grid.html", eventsGridData{
			Error: "",
			Empty: true,
		})
		return
	}

	now := time.Now()
	year := now.Year()
	month := now.Month()
	if mt, ok := parseMonthParam(c.Query("month")); ok {
		year = mt.Year()
		month = mt.Month()
	}
	monthTime := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	events, errMsg := fetchAndCacheEvents(settings, "")
	filtered := filterEventsByMonth(events, year, monthTime)
	weeks := buildGrid(filtered, year, monthTime)
	empty := len(filtered) == 0 && errMsg == ""
	prevMonth := monthTime.AddDate(0, -1, 0)
	nextMonth := monthTime.AddDate(0, 1, 0)
	showToday := monthTime.Year() != now.Year() || monthTime.Month() != now.Month()

	renderTemplate(c, "events_grid.html", eventsGridData{
		Weeks:     weeks,
		Error:     errMsg,
		Empty:     empty,
		Month:     monthTime,
		PrevMonth: prevMonth.Format("2006-01"),
		NextMonth: nextMonth.Format("2006-01"),
		ShowToday: showToday,
		Version:   AppVersion,
	})
}

// EventsGridCampaignPartial is the grid partial for a campaign page.
func EventsGridCampaignPartial(c *gin.Context) {
	slug := c.Param("slug")
	cs := db.GetCampaignEventSettingsBySlug(slug)
	if cs == nil || !cs.IsActive {
		renderTemplate(c, "events_grid.html", eventsGridData{
			Error: "Campaign not found or inactive.",
			Empty: true,
		})
		return
	}

	settings := campaignToGlobalSettings(cs)
	now := time.Now()
	year := now.Year()
	month := now.Month()
	if mt, ok := parseMonthParam(c.Query("month")); ok {
		year = mt.Year()
		month = mt.Month()
	}
	monthTime := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	events, errMsg := fetchAndCacheEvents(settings, slug)
	filtered := filterEventsByMonth(events, year, monthTime)
	weeks := buildGrid(filtered, year, monthTime)
	empty := len(filtered) == 0 && errMsg == ""
	prevMonth := monthTime.AddDate(0, -1, 0)
	nextMonth := monthTime.AddDate(0, 1, 0)
	showToday := monthTime.Year() != now.Year() || monthTime.Month() != now.Month()

	renderTemplate(c, "events_grid.html", eventsGridData{
		Weeks:        weeks,
		Error:        errMsg,
		Empty:        empty,
		Month:        monthTime,
		PrevMonth:    prevMonth.Format("2006-01"),
		NextMonth:    nextMonth.Format("2006-01"),
		ShowToday:    showToday,
		CampaignName: cs.DisplayName,
		CampaignSlug: slug,
		Version:      AppVersion,
	})
}

// googleMapsURL constructs a Google Maps search URL from a location string.
func googleMapsURL(location string) string {
	if location == "" {
		return ""
	}
	return "https://www.google.com/maps/search/" + url.QueryEscape(location)
}

// EventDetail renders the event detail page.
func EventDetail(c *gin.Context) {
	eventID := c.Param("id")

	// Try global cache first
	settings := db.GetEventSettings()
	if events, _ := fetchAndCacheEvents(settings, ""); len(events) > 0 {
		for _, e := range events {
			if e.ID == eventID {
				renderTemplate(c, "event_detail.html", eventDetailData{
					Event:    e,
					Version:  AppVersion,
					Location: googleMapsURL(e.Location),
				})
				return
			}
		}
	}

	// Try campaign caches
	campaigns := db.ListCampaignEventSettings()
	for _, cs := range campaigns {
		if !cs.IsActive {
			continue
		}
		campSettings := campaignToGlobalSettings(&cs)
		if campEvents, _ := fetchAndCacheEvents(campSettings, cs.Slug); len(campEvents) > 0 {
			for _, e := range campEvents {
				if e.ID == eventID {
					renderTemplate(c, "event_detail.html", eventDetailData{
						Event:        e,
						Version:      AppVersion,
						CampaignName: cs.DisplayName,
						CampaignSlug: cs.Slug,
						Location:     googleMapsURL(e.Location),
					})
					return
				}
			}
		}
	}

	c.Status(http.StatusNotFound)
	renderTemplate(c, "event_detail.html", eventDetailData{
		Error:   "Event not found.",
		Version: AppVersion,
	})
}
