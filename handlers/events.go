package handlers

import (
	"context"
	"fmt"
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
	settings := db.GetEventSettings()

	if settings.CalendarID == "" {
		renderTemplate(c, "events_page.html", eventsPageData{
			Events:  nil,
			Error:   "",
			Empty:   true,
			Version: AppVersion,
		})
		return
	}

	events, errMsg := fetchAndCacheEvents(settings, "")
	empty := len(events) == 0 && errMsg == ""

	renderTemplate(c, "events_page.html", eventsPageData{
		Events:  events,
		Error:   errMsg,
		Empty:   empty,
		Version: AppVersion,
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
}

type eventsListData struct {
	Events       []googlecalendar.Event
	Error        string
	Empty        bool
	CampaignName string
	CampaignSlug string
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

// ─── Per-Campaign Event Pages ───

func CampaignEventsPage(c *gin.Context) {
	slug := c.Param("slug")
	cs := db.GetCampaignEventSettingsBySlug(slug)
	if cs == nil || !cs.IsActive {
		c.HTML(http.StatusNotFound, "events_page.html", eventsPageData{
			Error:  "Campaign not found or inactive.",
			Empty:  true,
			Version: AppVersion,
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
		})
		return
	}

	events, errMsg := fetchAndCacheEvents(settings, slug)
	empty := len(events) == 0 && errMsg == ""

	data := eventsPageData{
		Events:       events,
		Error:        errMsg,
		Empty:        empty,
		Version:      AppVersion,
		CampaignName: cs.DisplayName,
		CampaignSlug: slug,
	}

	// If the campaign's calendar ID matches global, and global has no calendar configured,
	// the empty campaign calendar had no events — but we already showed the page.
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
