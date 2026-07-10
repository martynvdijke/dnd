package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"

	"villum/db"
	googlecalendar "villum/google_calendar"
)

// AppVersion holds the application version injected from main.go.
var AppVersion string

// BaseURL holds the configured public base URL (from BASE_URL env var or request host).
var BaseURL string

// SetAppVersion sets the application version for use in templates.
// ViewMode holds the current events view ("list" or "grid").
type ViewMode string

const (
	ViewList ViewMode = "list"
	ViewGrid ViewMode = "grid"
)

func SetAppVersion(v string) {
	AppVersion = v
}

// SetBaseURL sets the public base URL for constructing shareable links.
func SetBaseURL(u string) {
	BaseURL = u
}

// resolvePublicURL constructs the public events URL, using the optional slug for campaign pages.
func resolvePublicURL(c *gin.Context, slug string) string {
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := c.Request.Host
	base := BaseURL
	if base == "" {
		base = scheme + "://" + host
	} else {
		// Ensure base URL does not have a trailing slash
		base = strings.TrimRight(base, "/")
	}
	if slug != "" {
		return base + "/events/c/" + url.PathEscape(slug)
	}
	return base + "/events"
}

// EventsPublicLink returns JSON with the public events URL (optional ?slug param for campaign pages).
func EventsPublicLink(c *gin.Context) {
	slug := c.Query("slug")
	link := resolvePublicURL(c, slug)
	c.JSON(http.StatusOK, gin.H{"url": link})
}

// EventsQRCode generates a QR code PNG for the public events URL.
func EventsQRCode(c *gin.Context) {
	slug := c.Query("slug")
	link := resolvePublicURL(c, slug)

	png, err := qrcode.Encode(link, qrcode.Medium, 256)
	if err != nil {
		log.Printf("qr: encode error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate QR code"})
		return
	}

	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="events-qr%s.png"`, slugParamForFilename(slug)))
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/png", png)
}

// slugParamForFilename returns a filename-safe slug suffix.
func slugParamForFilename(slug string) string {
	if slug == "" {
		return ""
	}
	return "-" + slug
}



// ─── Public Events Page (Global) ───

func EventsPage(c *gin.Context) {
	view := c.DefaultQuery("view", "list")
	settings := db.GetEventSettings()

	// Show empty state if no source is configured
	if settings.SourceType == "google_api" && settings.CalendarID == "" {
		renderTemplate(c, "events_page.html", eventsPageData{
			Events:  nil,
			Error:   "",
			Empty:   true,
			Version: AppVersion,
			View:    view,
		})
		return
	}
	if settings.SourceType == "ical" && settings.ICalURL == "" {
		renderTemplate(c, "events_page.html", eventsPageData{
			Events:  nil,
			Error:   "",
			Empty:   true,
			Version: AppVersion,
			View:    view,
		})
		return
	}

	events, errMsg := fetchAndCacheEvents(settings, "")

	if view == "grid" {
		now := time.Now()
		year := now.Year()
		month := now.Month()
		monthTime := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		filtered := filterEventsByMonth(events, year, monthTime)
		weeks := buildGrid(filtered, year, monthTime)
		prevMonth := monthTime.AddDate(0, -1, 0)
		nextMonth := monthTime.AddDate(0, 1, 0)
		empty := len(filtered) == 0 && errMsg == ""

		renderTemplate(c, "events_page.html", eventsPageData{
			Events:    filtered,
			Error:     errMsg,
			Empty:     empty,
			Version:   AppVersion,
			View:      view,
			Weeks:     weeks,
			Month:     monthTime,
			PrevMonth: prevMonth.Format("2006-01"),
			NextMonth: nextMonth.Format("2006-01"),
		})
		return
	}

	empty := len(events) == 0 && errMsg == ""

	renderTemplate(c, "events_page.html", eventsPageData{
		Events:  events,
		Error:   errMsg,
		Empty:   empty,
		Version: AppVersion,
		View:    view,
	})
}

func EventsListPartial(c *gin.Context) {
	settings := db.GetEventSettings()

	if settings.CalendarID == "" {
		renderTemplate(c, "events_list.html", eventsListData{
			Events: nil,
			Error:  "",
			Empty:  true,
		})
		return
	}

	events, errMsg := fetchAndCacheEvents(settings, "")
	empty := len(events) == 0 && errMsg == ""

	renderTemplate(c, "events_list.html", eventsListData{
		Events: events,
		Error:  errMsg,
		Empty:  empty,
	})
}

type eventsPageData struct {
	Events       []googlecalendar.Event
	Error        string
	Empty        bool
	Version      string
	CampaignName string
	CampaignSlug string
	View         string // "list" or "grid"

	// Grid-specific fields
	Weeks     []gridWeek
	Month     time.Time
	PrevMonth string
	NextMonth string
}

type eventsListData struct {
	Events       []googlecalendar.Event
	Error        string
	Empty        bool
	CampaignName string
	CampaignSlug string
}

// ─── Calendar Grid View Data Types ───

type gridDay struct {
	Day            int
	Events         []googlecalendar.Event
	Overflow       int
	IsToday        bool
	IsCurrentMonth bool
}

type gridWeek struct {
	Days []gridDay // always 7 (Sun-Sat)
}

type eventsGridData struct {
	Weeks        []gridWeek
	Error        string
	Empty        bool
	Month        time.Time
	PrevMonth    string
	NextMonth    string
	CampaignName string
	CampaignSlug string
	Version      string
}

type eventDetailData struct {
	Event        googlecalendar.Event
	Error        string
	CampaignName string
	CampaignSlug string
	Version      string
	Location     string // Google Maps URL
}

// ─── Calendar Grid Helpers ───

func filterEventsByMonth(events []googlecalendar.Event, year int, month time.Time) []googlecalendar.Event {
	startOfMonth := time.Date(year, month.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	var filtered []googlecalendar.Event
	for _, e := range events {
		if !e.StartTime.IsZero() && (e.StartTime.Equal(startOfMonth) || e.StartTime.After(startOfMonth)) && e.StartTime.Before(endOfMonth) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

type gridBuilder struct{}

func (gridBuilder) build(events []googlecalendar.Event, year int, month time.Time) []gridWeek {
	firstDay := time.Date(year, month.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)
	daysInMonth := lastDay.Day()

	dayEvents := make(map[int][]googlecalendar.Event)
	for _, e := range events {
		day := e.StartTime.Day()
		dayEvents[day] = append(dayEvents[day], e)
	}

	today := time.Now().In(time.UTC)
	const maxShown = 3

	// Leading empty cells
	weekday := int(firstDay.Weekday()) // 0=Sun
	allDays := make([]gridDay, 0, daysInMonth+6)
	for i := 0; i < weekday; i++ {
		allDays = append(allDays, gridDay{Day: 0, IsCurrentMonth: false})
	}

	// Actual days
	for d := 1; d <= daysInMonth; d++ {
		evts := dayEvents[d]
		dayDate := time.Date(year, month.Month(), d, 0, 0, 0, 0, time.UTC)
		isToday := dayDate.Year() == today.Year() && dayDate.YearDay() == today.YearDay()

		gd := gridDay{Day: d, IsToday: isToday, IsCurrentMonth: true}
		if len(evts) > maxShown {
			gd.Events = evts[:maxShown]
			gd.Overflow = len(evts) - maxShown
		} else {
			gd.Events = evts
			gd.Overflow = 0
		}
		allDays = append(allDays, gd)
	}

	// Trailing empty cells to complete last week
	for len(allDays)%7 != 0 {
		allDays = append(allDays, gridDay{Day: 0, IsCurrentMonth: false})
	}

	weeks := make([]gridWeek, 0, len(allDays)/7)
	for i := 0; i < len(allDays); i += 7 {
		weeks = append(weeks, gridWeek{Days: allDays[i : i+7]})
	}
	return weeks
}

func buildGrid(events []googlecalendar.Event, year int, month time.Time) []gridWeek {
	return gridBuilder{}.build(events, year, month)
}

// parseMonthParam parses a "YYYY-MM" string into a time.Time (first of month).
func parseMonthParam(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return time.Time{}, false
	}
	y, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, false
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 1 || m > 12 {
		return time.Time{}, false
	}
	return time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC), true
}

// fetchAndCacheEvents fetches events from Google Calendar or cache.
// campaignSlug empty = global events.
func fetchAndCacheEvents(settings db.EventSettings, campaignSlug string) ([]googlecalendar.Event, string) {
	// Try cache first
	if cached, fresh := db.GetCachedEvents(settings.CacheTTLSeconds, campaignSlug); fresh {
		log.Printf("events: serving %d events from cache (fresh, slug=%q)", len(cached), campaignSlug)
		return cached, ""
	}

	// Cache miss or expired — try API
	events, err := fetchFromGoogle(settings)
	if err != nil {
		log.Printf("events: API fetch error: %v", err)
		// Try stale cache fallback
		if stale, _ := db.GetCachedEvents(settings.CacheTTLSeconds, campaignSlug); len(stale) > 0 {
			log.Printf("events: serving %d events from stale cache (API failed, slug=%q)", len(stale), campaignSlug)
			return stale, ""
		}
		return nil, "Unable to load events at this time. Please check back later."
	}

	// Cache the fresh events
	if len(events) > 0 {
		if err := db.SetCachedEvents(events, campaignSlug); err != nil {
			log.Printf("events: cache write error: %v", err)
		}
	}

	return events, ""
}

func fetchFromGoogle(settings db.EventSettings) ([]googlecalendar.Event, error) {
	// Dispatch based on source type
	if settings.SourceType == "ical" {
		return fetchFromICalURL(settings)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var client *googlecalendar.Client
	var err error

	switch settings.AuthMethod {
	case "oauth":
		clientID := settings.OAuthClientID
		clientSecret := settings.OAuthClientSecret
		refreshToken := settings.OAuthRefreshToken
		if clientID == "" || clientSecret == "" || refreshToken == "" {
			return nil, fmt.Errorf("OAuth credentials not fully configured (client_id, client_secret, refresh_token required)")
		}
		client, err = googlecalendar.NewOAuthClient(ctx, clientID, clientSecret, refreshToken)
		if err != nil {
			return nil, fmt.Errorf("create OAuth client: %w", err)
		}
	case "service_account":
		fallthrough
	default:
		creds := settings.CredentialsJSON
		if creds == "" {
			creds = os.Getenv("GOOGLE_CALENDAR_CREDENTIALS")
		}
		if creds == "" {
			return nil, fmt.Errorf("no Google Calendar credentials configured (set credentials_json or GOOGLE_CALENDAR_CREDENTIALS env var)")
		}
		client, err = googlecalendar.NewClient(ctx, []byte(creds))
		if err != nil {
			return nil, fmt.Errorf("create service account client: %w", err)
		}
	}

	var tags []string
	if settings.Tags != "" {
		for _, t := range strings.Split(settings.Tags, ",") {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	colorLabels := settings.ParseColorLabels()

	events, err := client.FetchUpcomingEvents(ctx, settings.CalendarID, tags, colorLabels, settings.FilterMode, 50)
	if err != nil {
		return nil, fmt.Errorf("fetch events: %w", err)
	}

	log.Printf("events: fetched %d events from Google Calendar (method=%s, tags=%v, colors=%v, filter=%s)", len(events), settings.AuthMethod, tags, colorLabels, settings.FilterMode)
	googlecalendar.LogEvents(events)
	return events, nil
}

// fetchFromICalURL fetches and parses an iCal/ICS feed from the given URL.
// Applies text tag filtering (color labels and filter mode are ignored for iCal sources).
func fetchFromICalURL(settings db.EventSettings) ([]googlecalendar.Event, error) {
	if settings.ICalURL == "" {
		return nil, fmt.Errorf("no iCal URL configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", settings.ICalURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create iCal request: %w", err)
	}
	req.Header.Set("Accept", "text/calendar, text/plain, */*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch iCal URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("iCal URL returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read iCal response: %w", err)
	}

	events, err := googlecalendar.ParseICS(body)
	if err != nil {
		return nil, fmt.Errorf("parse iCal feed: %w", err)
	}

	// Apply text tag filtering (iCal sources only support text tags)
	var tags []string
	if settings.Tags != "" {
		for _, t := range strings.Split(settings.Tags, ",") {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	if len(tags) > 0 {
		var filtered []googlecalendar.Event
		for _, e := range events {
			if googlecalendar.MatchesAnyTag(e, tags) {
				filtered = append(filtered, e)
			}
		}
		events = filtered
	}

	log.Printf("events: fetched %d events from iCal URL (tags=%v, url=%s)", len(events), tags, settings.ICalURL)
	googlecalendar.LogEvents(events)
	return events, nil
}

// ─── Per-Campaign Event Pages ───

func CampaignEventsPage(c *gin.Context) {
	view := c.DefaultQuery("view", "list")
	slug := c.Param("slug")
	cs := db.GetCampaignEventSettingsBySlug(slug)
	if cs == nil || !cs.IsActive {
		c.Status(http.StatusNotFound)
		renderTemplate(c, "events_page.html", eventsPageData{
			Error:   "Campaign not found or inactive.",
			Empty:   true,
			Version: AppVersion,
			View:    view,
		})
		return
	}

	settings := campaignToGlobalSettings(cs)

	if settings.CalendarID == "" {
		renderTemplate(c, "events_page.html", eventsPageData{
			Events:       nil,
			Error:        "",
			Empty:        true,
			Version:      AppVersion,
			CampaignName: cs.DisplayName,
			View:         view,
		})
		return
	}

	events, errMsg := fetchAndCacheEvents(settings, slug)

	if view == "grid" {
		now := time.Now()
		year := now.Year()
		month := now.Month()
		monthTime := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		filtered := filterEventsByMonth(events, year, monthTime)
		weeks := buildGrid(filtered, year, monthTime)
		prevMonth := monthTime.AddDate(0, -1, 0)
		nextMonth := monthTime.AddDate(0, 1, 0)
		empty := len(filtered) == 0 && errMsg == ""

		renderTemplate(c, "events_page.html", eventsPageData{
			Events:       filtered,
			Error:        errMsg,
			Empty:        empty,
			Version:      AppVersion,
			CampaignName: cs.DisplayName,
			CampaignSlug: slug,
			View:         view,
			Weeks:        weeks,
			Month:        monthTime,
			PrevMonth:    prevMonth.Format("2006-01"),
			NextMonth:    nextMonth.Format("2006-01"),
		})
		return
	}

	empty := len(events) == 0 && errMsg == ""

	data := eventsPageData{
		Events:       events,
		Error:        errMsg,
		Empty:        empty,
		Version:      AppVersion,
		CampaignName: cs.DisplayName,
		CampaignSlug: slug,
		View:         view,
	}
	renderTemplate(c, "events_page.html", data)
}

func CampaignEventsListPartial(c *gin.Context) {
	slug := c.Param("slug")
	cs := db.GetCampaignEventSettingsBySlug(slug)
	if cs == nil || !cs.IsActive {
		renderTemplate(c, "events_list.html", eventsListData{
			Events: nil,
			Error:  "Campaign not found or inactive.",
			Empty:  true,
		})
		return
	}

	settings := campaignToGlobalSettings(cs)

	if settings.CalendarID == "" {
		renderTemplate(c, "events_list.html", eventsListData{
			Events:       nil,
			Error:        "",
			Empty:        true,
			CampaignName: cs.DisplayName,
		})
		return
	}

	events, errMsg := fetchAndCacheEvents(settings, slug)
	empty := len(events) == 0 && errMsg == ""

	renderTemplate(c, "events_list.html", eventsListData{
		Events:       events,
		Error:        errMsg,
		Empty:        empty,
		CampaignName: cs.DisplayName,
		CampaignSlug: slug,
	})
}

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
	id, err := db.SaveCampaignEventSettings(s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Clear per-campaign cache
	db.ClearCache(s.Slug)
	c.JSON(http.StatusOK, gin.H{"id": id})
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

	renderTemplate(c, "events_grid.html", eventsGridData{
		Weeks:     weeks,
		Error:     errMsg,
		Empty:     empty,
		Month:     monthTime,
		PrevMonth: prevMonth.Format("2006-01"),
		NextMonth: nextMonth.Format("2006-01"),
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

	renderTemplate(c, "events_grid.html", eventsGridData{
		Weeks:        weeks,
		Error:        errMsg,
		Empty:        empty,
		Month:        monthTime,
		PrevMonth:    prevMonth.Format("2006-01"),
		NextMonth:    nextMonth.Format("2006-01"),
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
