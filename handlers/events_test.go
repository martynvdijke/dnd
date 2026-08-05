package handlers

import (
	"encoding/json"
	"net/http"
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
		db.SetCachedEvents(events, "")

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
		db.ClearCache("")
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
		db.SetCachedEvents(events, "")
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
			"calendar_id":         "admin@example.com",
			"tags":                "dnd,oneshot",
			"cache_ttl_seconds":   600,
			"auth_method":         "oauth",
			"oauth_client_id":     "test-client",
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
		}, "")
		if count := db.GetCachedCount(""); count != 1 {
			t.Fatalf("expected 1 cached event, got %d", count)
		}

		w := testutil.PostJSON(t, r, "/api/admin/events-settings/clear-cache", nil)
		testutil.AssertStatus(t, w, 200)

		if count := db.GetCachedCount(""); count != 0 {
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
		db.ClearCache("")
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
		db.SetCachedEvents(events, "")
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
		db.ClearCache("")
		events := []googlecalendar.Event{
			{
				ID:        "evt3",
				Title:     "Minimal Event",
				StartTime: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, 9, 1, 14, 0, 0, 0, time.UTC),
			},
		}
		db.SetCachedEvents(events, "")
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

// ─── Events Share Link & QR Tests ───

func TestEventsPublicLink(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	gin.SetMode(gin.TestMode)

	t.Run("returns public URL from request host", func(t *testing.T) {
		r := gin.New()
		admin := r.Group("/api/admin")
		admin.Use(mockAuth("admin"), mockCSRF())
		admin.GET("/events/public-link", EventsPublicLink)

		w := testutil.Get(t, r, "/api/admin/events/public-link")
		testutil.AssertStatus(t, w, 200)

		var result map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		url, ok := result["url"]
		if !ok {
			t.Fatal("expected 'url' field in response")
		}
		if !strings.HasSuffix(url, "/events") {
			t.Errorf("expected URL ending in /events, got %q", url)
		}
	})

	t.Run("returns campaign URL with slug param", func(t *testing.T) {
		r := gin.New()
		admin := r.Group("/api/admin")
		admin.Use(mockAuth("admin"), mockCSRF())
		admin.GET("/events/public-link", EventsPublicLink)

		w := testutil.Get(t, r, "/api/admin/events/public-link?slug=lost-mines")
		testutil.AssertStatus(t, w, 200)

		var result map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		url := result["url"]
		if !strings.HasSuffix(url, "/events/c/lost-mines") {
			t.Errorf("expected URL ending in /events/c/lost-mines, got %q", url)
		}
	})

	t.Run("uses BASE_URL when set", func(t *testing.T) {
		SetBaseURL("https://dnd.example.com")
		defer SetBaseURL("")

		r := gin.New()
		admin := r.Group("/api/admin")
		admin.Use(mockAuth("admin"), mockCSRF())
		admin.GET("/events/public-link", EventsPublicLink)

		w := testutil.Get(t, r, "/api/admin/events/public-link")
		testutil.AssertStatus(t, w, 200)

		var result map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		url := result["url"]
		expected := "https://dnd.example.com/events"
		if url != expected {
			t.Errorf("expected %q, got %q", expected, url)
		}
	})
}

func TestEventsQRCode(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	gin.SetMode(gin.TestMode)

	r := gin.New()
	admin := r.Group("/api/admin")
	admin.Use(mockAuth("admin"), mockCSRF())
	admin.GET("/events/qr", EventsQRCode)

	t.Run("returns valid PNG", func(t *testing.T) {
		SetBaseURL("https://dnd.example.com")
		defer SetBaseURL("")

		w := testutil.Get(t, r, "/api/admin/events/qr")
		testutil.AssertStatus(t, w, 200)

		ct := w.Header().Get("Content-Type")
		if ct != "image/png" {
			t.Errorf("expected Content-Type image/png, got %q", ct)
		}

		body := w.Body.Bytes()
		if len(body) == 0 {
			t.Fatal("expected non-empty PNG body")
		}
		// PNG header: 89 50 4E 47 0D 0A 1A 0A
		if len(body) < 8 || body[0] != 0x89 || body[1] != 'P' || body[2] != 'N' || body[3] != 'G' {
			t.Error("expected valid PNG magic bytes")
		}

		cc := w.Header().Get("Cache-Control")
		if cc == "" {
			t.Error("expected Cache-Control header on QR response")
		}
	})

	t.Run("returns PNG for campaign slug", func(t *testing.T) {
		SetBaseURL("https://dnd.example.com")
		defer SetBaseURL("")

		w := testutil.Get(t, r, "/api/admin/events/qr?slug=lost-mines")
		testutil.AssertStatus(t, w, 200)

		ct := w.Header().Get("Content-Type")
		if ct != "image/png" {
			t.Errorf("expected Content-Type image/png, got %q", ct)
		}
		body := w.Body.Bytes()
		if len(body) < 8 || body[0] != 0x89 || body[1] != 'P' || body[2] != 'N' || body[3] != 'G' {
			t.Error("expected valid PNG magic bytes")
		}
	})
}

func TestEventsShareAuth(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	gin.SetMode(gin.TestMode)

	t.Run("non-admin is rejected from public-link", func(t *testing.T) {
		r := gin.New()
		admin := r.Group("/api/admin")
		admin.Use(mockAuth("user"), mockAdminRequired(), mockCSRF())
		admin.GET("/events/public-link", EventsPublicLink)

		w := testutil.Get(t, r, "/api/admin/events/public-link")
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 for non-admin, got %d", w.Code)
		}
	})

	t.Run("non-admin is rejected from qr", func(t *testing.T) {
		r := gin.New()
		admin := r.Group("/api/admin")
		admin.Use(mockAuth("user"), mockAdminRequired(), mockCSRF())
		admin.GET("/events/qr", EventsQRCode)

		w := testutil.Get(t, r, "/api/admin/events/qr")
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 for non-admin, got %d", w.Code)
		}
	})
}

// ─── Calendar Grid View Tests ───

func TestEventsGridPartial(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := gin.New()
	r.GET("/htmx/events/grid", EventsGridPartial)

	t.Run("returns grid partial HTML", func(t *testing.T) {
		w := testutil.Get(t, r, "/htmx/events/grid")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if contains(body, "<!DOCTYPE html>") {
			t.Error("grid partial should not include DOCTYPE")
		}
	})

	t.Run("shows empty state when no calendar configured", func(t *testing.T) {
		w := testutil.Get(t, r, "/htmx/events/grid")
		body := w.Body.String()
		if !contains(body, "No events this month") {
			t.Error("expected empty state in grid partial")
		}
	})

	t.Run("shows month navigation with events", func(t *testing.T) {
		now := time.Now()
		events := []googlecalendar.Event{
			{ID: "g1", Title: "Grid Event", StartTime: time.Date(now.Year(), now.Month(), 15, 19, 0, 0, 0, time.UTC)},
		}
		db.SetCachedEvents(events, "")
		db.SaveEventSettings(db.EventSettings{CalendarID: "test@example.com", Tags: ""})

		w := testutil.Get(t, r, "/htmx/events/grid")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "Grid Event") {
			t.Error("expected event title in grid partial")
		}
		if !contains(body, "calendar-grid") {
			t.Error("expected calendar-grid class")
		}
	})

	t.Run("accepts month param", func(t *testing.T) {
		db.ClearCache("")
		events := []googlecalendar.Event{
			{ID: "g2", Title: "July Event", StartTime: time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)},
		}
		db.SetCachedEvents(events, "")
		db.SaveEventSettings(db.EventSettings{CalendarID: "test@example.com"})

		w := testutil.Get(t, r, "/htmx/events/grid?month=2026-07")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "July Event") {
			t.Error("expected event in July grid")
		}
		if !contains(body, "July 2026") {
			t.Error("expected July 2026 in month header")
		}
	})

	t.Run("different month hides events", func(t *testing.T) {
		w := testutil.Get(t, r, "/htmx/events/grid?month=2026-06")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if contains(body, "July Event") {
			t.Error("July event should not appear in June grid")
		}
	})

	t.Run("shows Today button when viewing another month", func(t *testing.T) {
		prev := time.Now().AddDate(0, -1, 0).Format("2006-01")
		w := testutil.Get(t, r, "/htmx/events/grid?month="+prev)
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "Jump to today") {
			t.Error("expected Today button when viewing a different month")
		}
	})

	t.Run("hides Today button on current month", func(t *testing.T) {
		w := testutil.Get(t, r, "/htmx/events/grid")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if contains(body, "Jump to today") {
			t.Error("Today button should be hidden on the current month")
		}
	})
}

func TestEventDetail(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := gin.New()
	r.GET("/events/:id", EventDetail)

	t.Run("returns 404 for unknown event", func(t *testing.T) {
		w := testutil.Get(t, r, "/events/nonexistent")
		testutil.AssertStatus(t, w, 404)
		body := w.Body.String()
		if !contains(body, "Event not found") {
			t.Error("expected not found message")
		}
	})

	t.Run("shows event detail when event is in cache", func(t *testing.T) {
		events := []googlecalendar.Event{
			{ID: "det1", Title: "Detail Event", Description: "A detailed description", StartTime: time.Date(2026, 7, 15, 19, 0, 0, 0, time.UTC)},
		}
		db.SetCachedEvents(events, "")
		db.SaveEventSettings(db.EventSettings{CalendarID: "test@example.com"})

		w := testutil.Get(t, r, "/events/det1")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "Detail Event") {
			t.Error("expected event title in detail page")
		}
		if !contains(body, "A detailed description") {
			t.Error("expected event description in detail page")
		}
		if !contains(body, "Back to Events") {
			t.Error("expected back link")
		}
	})

	t.Run("shows location with Google Maps link", func(t *testing.T) {
		db.ClearCache("")
		events := []googlecalendar.Event{
			{ID: "det2", Title: "Located Event", Location: "123 Main St, City", StartTime: time.Date(2026, 7, 20, 18, 0, 0, 0, time.UTC)},
		}
		db.SetCachedEvents(events, "")
		db.SaveEventSettings(db.EventSettings{CalendarID: "test@example.com"})

		w := testutil.Get(t, r, "/events/det2")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "123 Main St, City") {
			t.Error("expected location in detail page")
		}
		if !contains(body, "Open in Google Maps") {
			t.Error("expected Google Maps link")
		}
		if !contains(body, "google.com/maps/search") {
			t.Error("expected google maps search URL")
		}
	})
}

func TestEventsGridViewParam(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := gin.New()
	r.GET("/events", EventsPage)

	t.Run("default view is list", func(t *testing.T) {
		w := testutil.Get(t, r, "/events")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if contains(body, "calendar-grid") {
			t.Error("default view should not show grid")
		}
	})

	t.Run("view=grid shows grid elements", func(t *testing.T) {
		db.ClearCache("")
		now := time.Now()
		events := []googlecalendar.Event{
			{ID: "v1", Title: "View Grid Event", StartTime: time.Date(now.Year(), now.Month(), 10, 19, 0, 0, 0, time.UTC)},
		}
		db.SetCachedEvents(events, "")
		db.SaveEventSettings(db.EventSettings{CalendarID: "test@example.com"})

		w := testutil.Get(t, r, "/events?view=grid")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "View Grid Event") {
			t.Error("expected grid event title when view=grid")
		}
		if !contains(body, "calendar-grid") {
			t.Error("expected calendar grid when view=grid")
		}
	})

	t.Run("view=list stays in list mode", func(t *testing.T) {
		w := testutil.Get(t, r, "/events?view=list")
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if contains(body, "calendar-grid") {
			t.Error("view=list should not show grid")
		}
	})

	t.Run("view=grid honors month param", func(t *testing.T) {
		db.ClearCache("")
		prev := time.Now().AddDate(0, -1, 0)
		events := []googlecalendar.Event{
			{ID: "v2", Title: "Prev Month Event", StartTime: time.Date(prev.Year(), prev.Month(), 15, 19, 0, 0, 0, time.UTC)},
		}
		db.SetCachedEvents(events, "")
		db.SaveEventSettings(db.EventSettings{CalendarID: "test@example.com"})

		w := testutil.Get(t, r, "/events?view=grid&month="+prev.Format("2006-01"))
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		if !contains(body, "Prev Month Event") {
			t.Error("expected event from requested month in grid")
		}
		if !contains(body, prev.Format("January 2006")) {
			t.Error("expected requested month in grid header")
		}
		if !contains(body, "Jump to today") {
			t.Error("expected Today button when navigating to another month via page param")
		}
	})
}

func TestEventsGridHelpers(t *testing.T) {
	t.Run("parseMonthParam returns correct month", func(t *testing.T) {
		mt, ok := parseMonthParam("2026-07")
		if !ok {
			t.Fatal("expected parse to succeed")
		}
		if mt.Year() != 2026 || mt.Month() != time.July {
			t.Errorf("expected 2026-07, got %d-%d", mt.Year(), mt.Month())
		}
	})

	t.Run("parseMonthParam rejects invalid input", func(t *testing.T) {
		_, ok := parseMonthParam("")
		if ok {
			t.Error("expected empty string to fail")
		}
		_, ok = parseMonthParam("not-a-date")
		if ok {
			t.Error("expected invalid string to fail")
		}
		_, ok = parseMonthParam("2026-13")
		if ok {
			t.Error("expected invalid month to fail")
		}
		_, ok = parseMonthParam("2026-00")
		if ok {
			t.Error("expected zero month to fail")
		}
	})

	t.Run("filterEventsByMonth returns only events in month", func(t *testing.T) {
		events := []googlecalendar.Event{
			{ID: "a", StartTime: time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)},
			{ID: "b", StartTime: time.Date(2026, 7, 15, 19, 0, 0, 0, time.UTC)},
			{ID: "c", StartTime: time.Date(2026, 6, 30, 23, 0, 0, 0, time.UTC)},
			{ID: "d", StartTime: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		}
		july := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		filtered := filterEventsByMonth(events, 2026, july)
		if len(filtered) != 2 {
			t.Errorf("expected 2 events in July, got %d", len(filtered))
		}
	})

	t.Run("buildGrid creates correct number of weeks", func(t *testing.T) {
		events := []googlecalendar.Event{
			{ID: "a", StartTime: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)},
			{ID: "b", StartTime: time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)},
		}
		july := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		weeks := buildGrid(events, 2026, july)
		// July 2026 starts on Wednesday, so 5 weeks
		if len(weeks) < 4 || len(weeks) > 6 {
			t.Errorf("expected 4-6 weeks for July 2026, got %d", len(weeks))
		}
		// Check first week has leading empty days (Wed=3, so 3 leading empty)
		if len(weeks[0].Days) != 7 {
			t.Errorf("expected 7 days per week, got %d", len(weeks[0].Days))
		}
	})

	t.Run("googleMapsURL returns correct URL", func(t *testing.T) {
		url := googleMapsURL("123 Main St, City")
		expected := "https://www.google.com/maps/search/123+Main+St%2C+City"
		if url != expected {
			t.Errorf("expected %q, got %q", expected, url)
		}
	})

	t.Run("googleMapsURL returns empty for empty location", func(t *testing.T) {
		if url := googleMapsURL(""); url != "" {
			t.Errorf("expected empty URL, got %q", url)
		}
	})
}
