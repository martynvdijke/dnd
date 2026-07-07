package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	googlecalendar "villum/google_calendar"
	"villum/handlers/testutil"
)

// contains and searchString are defined in oneshot_planning_test.go

func TestEventsPage(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	AppVersion = "test-version"

	r := gin.New()

	// Register the public events page route
	r.GET("/events", EventsPage)

	t.Run("renders with no calendar configured", func(t *testing.T) {
		w := testutil.Get(t, r, "/events")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "Upcoming Events") {
			t.Error("expected page title 'Upcoming Events'")
		}
		if !contains(body, "No upcoming events found") {
			t.Error("expected empty state message")
		}
	})

	t.Run("renders with events from cache", func(t *testing.T) {
		// Seed events directly into cache
		events := []googlecalendar.Event{
			{ID: "evt1", Title: "DnD Session", StartTime: time.Date(2026, 7, 15, 19, 0, 0, 0, time.UTC)},
			{ID: "evt2", Title: "One-Shot", StartTime: time.Date(2026, 7, 20, 18, 0, 0, 0, time.UTC)},
		}
		db.SetCachedEvents(events)

		// Configure a calendar ID so the handler tries to fetch
		db.SaveEventSettings(db.EventSettings{
			CalendarID: "test@example.com",
			Tags:       "dnd",
			AuthMethod: "service_account",
		})

		w := testutil.Get(t, r, "/events")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "DnD Session") {
			t.Error("expected cached event title in rendered page")
		}
		if !contains(body, "One-Shot") {
			t.Error("expected second cached event title in rendered page")
		}
		if !contains(body, "test-version") {
			t.Error("expected version string in page")
		}
	})

	t.Run("shows error state when API is unreachable and no cache", func(t *testing.T) {
		// Set calendar_id but clear cache
		db.ClearCache()
		db.SaveEventSettings(db.EventSettings{
			CalendarID:      "nonexistent@example.com",
			Tags:            "dnd",
			AuthMethod:      "service_account",
			CredentialsJSON: `{"type":"service_account","project_id":"test"}`,
		})

		w := testutil.Get(t, r, "/events")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "Unable to load events") {
			t.Error("expected error message when API unreachable and no cache")
		}
	})
}

func TestEventsListPartial(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := gin.New()
	r.GET("/htmx/events/list", EventsListPartial)

	t.Run("returns partial HTML", func(t *testing.T) {
		w := testutil.Get(t, r, "/htmx/events/list")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		// Should NOT contain full page HTML elements
		if contains(body, "<!DOCTYPE html>") {
			t.Error("partial should not include DOCTYPE")
		}
		if contains(body, "<head>") {
			t.Error("partial should not include <head>")
		}
	})

	t.Run("shows empty state when no calendar configured", func(t *testing.T) {
		w := testutil.Get(t, r, "/htmx/events/list")
		body := w.Body.String()
		if !contains(body, "No upcoming events found") {
			t.Error("expected empty state in partial")
		}
	})

	t.Run("shows cached events in partial", func(t *testing.T) {
		events := []googlecalendar.Event{
			{ID: "evt3", Title: "Test Session", StartTime: time.Date(2026, 7, 15, 19, 0, 0, 0, time.UTC)},
		}
		db.SetCachedEvents(events)
		db.SaveEventSettings(db.EventSettings{CalendarID: "test@example.com"})

		w := testutil.Get(t, r, "/htmx/events/list")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "Test Session") {
			t.Error("expected event title in partial")
		}
	})
}

func TestAdminEventsSettings(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/events-settings", GetEventsSettings)
		auth.POST("/admin/events-settings", SaveEventsSettings)
		auth.POST("/admin/events-settings/clear-cache", ClearEventsCache)
	})

	t.Run("GET returns default settings", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/events-settings")
		testutil.AssertStatus(t, w, 200)
		var s db.EventSettings
		testutil.ParseJSON(t, w, &s)
		if s.Tags != "dnd,session,oneshot" {
			t.Errorf("expected default tags, got %q", s.Tags)
		}
		if s.AuthMethod != "service_account" {
			t.Errorf("expected default auth method, got %q", s.AuthMethod)
		}
	})

	t.Run("POST saves settings", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/events-settings", map[string]any{
			"calendar_id":        "admin@example.com",
			"tags":               "dnd,oneshot",
			"cache_ttl_seconds":  600,
			"auth_method":        "oauth",
			"oauth_client_id":    "test-client",
			"oauth_client_secret": "test-secret",
			"oauth_refresh_token": "test-refresh",
		})
		testutil.AssertStatus(t, w, 200)

		// Verify by reading back
		s := db.GetEventSettings()
		if s.CalendarID != "admin@example.com" {
			t.Errorf("expected saved calendar_id, got %q", s.CalendarID)
		}
		if s.AuthMethod != "oauth" {
			t.Errorf("expected saved auth method 'oauth', got %q", s.AuthMethod)
		}
		if s.OAuthClientID != "test-client" {
			t.Errorf("expected saved oauth_client_id, got %q", s.OAuthClientID)
		}
	})

	t.Run("POST clear cache", func(t *testing.T) {
		// Seed some cache data
		db.SetCachedEvents([]googlecalendar.Event{
			{ID: "test", Title: "Test", StartTime: time.Now()},
		})
		if count := db.GetCachedCount(); count != 1 {
			t.Fatalf("expected 1 cached event, got %d", count)
		}

		w := testutil.PostJSON(t, r, "/api/admin/events-settings/clear-cache", nil)
		testutil.AssertStatus(t, w, 200)

		if count := db.GetCachedCount(); count != 0 {
			t.Errorf("expected 0 cached events after clear, got %d", count)
		}
	})

	t.Run("POST invalid settings returns 400", func(t *testing.T) {
		// Invalid JSON
		w := testutil.PostJSON(t, r, "/api/admin/events-settings", "not-json")
		if w.Code == 200 {
			t.Error("expected non-200 status for invalid JSON")
		}
	})
}

func TestEventsPageWithSettingsLabels(t *testing.T) {
	// Test that the admin events settings endpoints are registered under the admin routes
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/events-settings", GetEventsSettings)
	})

	w := testutil.Get(t, r, "/api/admin/events-settings")
	testutil.AssertStatus(t, w, 200)

	var result map[string]any
	testutil.ParseJSON(t, w, &result)

	// Verify all expected settings fields are present
	expectedFields := []string{"calendar_id", "tags", "cache_ttl_seconds", "auth_method", "oauth_client_id", "oauth_client_secret", "oauth_refresh_token"}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("expected field %q in response", field)
		}
	}
}

// ─── ICS / iCal Tests ───

func TestEscapeICS(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"plain text", "plain text"},
		{"comma, here", `comma\, here`},
		{"semi;colon", `semi\;colon`},
		{"back\\slash", `back\\slash`},
		{"line\nbreak", `line\nbreak`},
		{"all, special; chars\nand\\more", `all\, special\; chars\nand\\more`},
		{"carriage\r\nreturn", `carriage\nreturn`},
	}
	for _, tc := range tests {
		got := escapeICS(tc.input)
		if got != tc.expected {
			t.Errorf("escapeICS(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestIcsDateTime(t *testing.T) {
	// 2026-07-15T19:00:00 UTC
	tm := time.Date(2026, 7, 15, 19, 0, 0, 0, time.UTC)
	result := icsDateTime(tm)
	expected := "20260715T190000Z"
	if result != expected {
		t.Errorf("icsDateTime(%v) = %q, want %q", tm, result, expected)
	}

	// Verify it always outputs UTC (converts from non-UTC)
	tmLocal := time.Date(2026, 7, 15, 19, 0, 0, 0, time.FixedZone("EST", -5*60*60))
	result = icsDateTime(tmLocal)
	// EST 19:00 = UTC 00:00 next day
	expected = "20260716T000000Z"
	if result != expected {
		t.Errorf("icsDateTime(%v) = %q, want %q", tmLocal, result, expected)
	}
}

func TestIcsDate(t *testing.T) {
	tm := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	result := icsDate(tm)
	expected := "20260715"
	if result != expected {
		t.Errorf("icsDate(%v) = %q, want %q", tm, result, expected)
	}
}

func TestEventsICalHandler(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := gin.New()
	r.GET("/events/ical", EventsICal)

	t.Run("returns valid ICS with no events configured", func(t *testing.T) {
		// No calendar configured = no events
		db.ClearCache()
		db.SaveEventSettings(db.EventSettings{CalendarID: "", Tags: "dnd"})

		w := testutil.Get(t, r, "/events/ical")
		testutil.AssertStatus(t, w, 200)

		body := w.Body.String()
		if !strings.Contains(body, "BEGIN:VCALENDAR") {
			t.Error("expected BEGIN:VCALENDAR")
		}
		if !strings.Contains(body, "END:VCALENDAR") {
			t.Error("expected END:VCALENDAR")
		}
		if !strings.Contains(body, "VERSION:2.0") {
			t.Error("expected VERSION:2.0")
		}
		if !strings.Contains(body, "PRODID:-//villum//Events//EN") {
			t.Error("expected PRODID")
		}
		if !strings.Contains(body, "METHOD:PUBLISH") {
			t.Error("expected METHOD:PUBLISH")
		}
		if strings.Contains(body, "BEGIN:VEVENT") {
			t.Error("expected no VEVENT when no events")
		}

		ct := w.Header().Get("Content-Type")
		if ct != "text/calendar" {
			t.Errorf("expected Content-Type text/calendar, got %q", ct)
		}
		disp := w.Header().Get("Content-Disposition")
		if !strings.Contains(disp, "events.ics") {
			t.Errorf("expected Content-Disposition with events.ics, got %q", disp)
		}
	})

	t.Run("returns ICS with cached events", func(t *testing.T) {
		events := []googlecalendar.Event{
			{
				ID:          "evt1",
				Title:       "DnD Session",
				Description: "Weekly game, bring dice!",
				StartTime:   time.Date(2026, 7, 15, 19, 0, 0, 0, time.UTC),
				EndTime:     time.Date(2026, 7, 15, 22, 0, 0, 0, time.UTC),
				Location:    "Tabletop Tavern",
			},
			{
				ID:        "evt2",
				Title:     "All-Day Con",
				StartTime: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
				AllDay:    true,
			},
		}
		db.SetCachedEvents(events)
		db.SaveEventSettings(db.EventSettings{CalendarID: "test@example.com", Tags: "dnd"})

		w := testutil.Get(t, r, "/events/ical")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()

		// Check timed event
		if !strings.Contains(body, "UID:evt1@villum.events") {
			t.Error("expected UID for evt1")
		}
		if !strings.Contains(body, "DTSTART:20260715T190000Z") {
			t.Error("expected DTSTART for timed event")
		}
		if !strings.Contains(body, "DTEND:20260715T220000Z") {
			t.Error("expected DTEND for timed event")
		}
		if !strings.Contains(body, "SUMMARY:DnD Session") {
			t.Error("expected summary")
		}
		if !strings.Contains(body, "DESCRIPTION:Weekly game\\, bring dice!") {
			t.Error("expected escaped description")
		}
		if !strings.Contains(body, "LOCATION:Tabletop Tavern") {
			t.Error("expected location")
		}

		// Check all-day event
		if !strings.Contains(body, "UID:evt2@villum.events") {
			t.Error("expected UID for evt2")
		}
		if !strings.Contains(body, "DTSTART;VALUE=DATE:20260801") {
			t.Error("expected all-day DTSTART with VALUE=DATE")
		}
		if !strings.Contains(body, "DTEND;VALUE=DATE:20260803") {
			t.Error("expected all-day DTEND with VALUE=DATE")
		}
		if !strings.Contains(body, "SUMMARY:All-Day Con") {
			t.Error("expected all-day summary")
		}
	})

	t.Run("empty event with no description omits DESCRIPTION", func(t *testing.T) {
		db.ClearCache()
		events := []googlecalendar.Event{
			{
				ID:        "evt3",
				Title:     "Minimal Event",
				StartTime: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, 9, 1, 14, 0, 0, 0, time.UTC),
			},
		}
		db.SetCachedEvents(events)
		db.SaveEventSettings(db.EventSettings{CalendarID: "test@example.com"})

		w := testutil.Get(t, r, "/events/ical")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()

		if !strings.Contains(body, "SUMMARY:Minimal Event") {
			t.Error("expected summary")
		}
		if strings.Contains(body, "DESCRIPTION:") {
			t.Error("expected no DESCRIPTION for event without description")
		}
		if strings.Contains(body, "LOCATION:") {
			t.Error("expected no LOCATION for event without location")
		}
	})
}

// contains and searchString are available from oneshot_planning_test.go
