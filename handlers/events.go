package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	googlecalendar "villum/google_calendar"
)

// AppVersion holds the application version injected from main.go.
var AppVersion string

// SetAppVersion sets the application version for use in templates.
func SetAppVersion(v string) {
	AppVersion = v
}

// ─── Public Events Page ───

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

	events, errMsg := fetchAndCacheEvents(settings)
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

	events, errMsg := fetchAndCacheEvents(settings)
	empty := len(events) == 0 && errMsg == ""

	renderTemplate(c, "events_list.html", eventsListData{
		Events: events,
		Error:  errMsg,
		Empty:  empty,
	})
}

type eventsPageData struct {
	Events  []googlecalendar.Event
	Error   string
	Empty   bool
	Version string
}

type eventsListData struct {
	Events []googlecalendar.Event
	Error  string
	Empty  bool
}

// fetchAndCacheEvents fetches events from Google Calendar or cache, returns them.
func fetchAndCacheEvents(settings db.EventSettings) ([]googlecalendar.Event, string) {
	// Try cache first
	if cached, fresh := db.GetCachedEvents(settings.CacheTTLSeconds); fresh {
		log.Printf("events: serving %d events from cache (fresh)", len(cached))
		return cached, ""
	}

	// Cache miss or expired — try API
	events, err := fetchFromGoogle(settings)
	if err != nil {
		log.Printf("events: API fetch error: %v", err)
		// Try stale cache fallback
		if stale, _ := db.GetCachedEvents(settings.CacheTTLSeconds); len(stale) > 0 {
			log.Printf("events: serving %d events from stale cache (API failed)", len(stale))
			return stale, ""
		}
		return nil, "Unable to load events at this time. Please check back later."
	}

	// Cache the fresh events
	if len(events) > 0 {
		if err := db.SetCachedEvents(events); err != nil {
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
			// Fall back to env var
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

// ─── ICS / iCal Feed ───

// escapeICS escapes a string for use in ICS property values per RFC 5545.
func escapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// icsDateTime formats a time.Time as ICS UTC datetime (e.g., 20260715T190000Z).
func icsDateTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// icsDate formats a time.Time as ICS date for all-day events (e.g., 20260715).
func icsDate(t time.Time) string {
	return t.UTC().Format("20060102")
}

// EventsICal returns an iCalendar (ICS) file of upcoming events.
// GET /events/ical
func EventsICal(c *gin.Context) {
	settings := db.GetEventSettings()

	events, _ := fetchAndCacheEvents(settings)

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
			// ICS all-day end date is exclusive
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

// ─── Admin Settings ───

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
	// Clear cache since settings changed
	db.ClearCache()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ClearEventsCache(c *gin.Context) {
	if err := db.ClearCache(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
