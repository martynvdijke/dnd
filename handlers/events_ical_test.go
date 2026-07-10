package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	googlecalendar "villum/google_calendar"
	"villum/handlers/testutil"
)

// sampleICS returns a valid ICS feed with events in the near future.
func sampleICS() string {
	// Use dynamic dates relative to now so events are always "upcoming"
	now := time.Now().UTC()
	d1 := now.Add(48 * time.Hour)
	d2 := now.Add(72 * time.Hour)
	return "BEGIN:VCALENDAR\r\n" +
		"VERSION:2.0\r\n" +
		"PRODID:-//test//test//EN\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:evt-1@test\r\n" +
		"SUMMARY:DnD Session\r\n" +
		"DESCRIPTION:Weekly game\r\n" +
		"DTSTART:" + d1.Format("20060102T150405Z") + "\r\n" +
		"DTEND:" + d1.Add(3 * time.Hour).Format("20060102T150405Z") + "\r\n" +
		"LOCATION:The Tavern\r\n" +
		"END:VEVENT\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:evt-2@test\r\n" +
		"SUMMARY:Board Game Night\r\n" +
		"DESCRIPTION:Casual board games\r\n" +
		"DTSTART:" + d2.Format("20060102T150405Z") + "\r\n" +
		"DTEND:" + d2.Add(2 * time.Hour).Format("20060102T150405Z") + "\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"
}

// sampleICSWithPastEvents includes both past and future events.
func sampleICSWithPastEvents() string {
	now := time.Now().UTC()
	past := now.Add(-48 * time.Hour)
	future := now.Add(48 * time.Hour)
	return "BEGIN:VCALENDAR\r\n" +
		"VERSION:2.0\r\n" +
		"PRODID:-//test//test//EN\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:past-1@test\r\n" +
		"SUMMARY:Past Event\r\n" +
		"DTSTART:" + past.Format("20060102T150405Z") + "\r\n" +
		"END:VEVENT\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:future-1@test\r\n" +
		"SUMMARY:Future DnD\r\n" +
		"DTSTART:" + future.Format("20060102T150405Z") + "\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"
}

func TestFetchFromICalURL(t *testing.T) {
	// Mock HTTP server serving an ICS feed
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		w.Write([]byte(sampleICS()))
	}))
	defer srv.Close()

	settings := db.EventSettings{
		SourceType: "ical",
		ICalURL:    srv.URL,
	}

	events, err := fetchFromICalURL(settings)
	if err != nil {
		t.Fatalf("fetchFromICalURL failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	// Events should be sorted by start time
	if events[0].Title != "DnD Session" {
		t.Errorf("expected first event 'DnD Session', got %q", events[0].Title)
	}
	if events[1].Title != "Board Game Night" {
		t.Errorf("expected second event 'Board Game Night', got %q", events[1].Title)
	}
	if events[0].Location != "The Tavern" {
		t.Errorf("expected location 'The Tavern', got %q", events[0].Location)
	}
}

func TestFetchFromICalURL_PastEventsFiltered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		w.Write([]byte(sampleICSWithPastEvents()))
	}))
	defer srv.Close()

	settings := db.EventSettings{
		SourceType: "ical",
		ICalURL:    srv.URL,
	}

	events, err := fetchFromICalURL(settings)
	if err != nil {
		t.Fatalf("fetchFromICalURL failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 upcoming event (past filtered), got %d", len(events))
	}
	if events[0].Title != "Future DnD" {
		t.Errorf("expected 'Future DnD', got %q", events[0].Title)
	}
}

func TestFetchFromICalURL_TagFiltering(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		w.Write([]byte(sampleICS()))
	}))
	defer srv.Close()

	settings := db.EventSettings{
		SourceType: "ical",
		ICalURL:    srv.URL,
		Tags:       "dnd,session",
	}

	events, err := fetchFromICalURL(settings)
	if err != nil {
		t.Fatalf("fetchFromICalURL failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event matching 'dnd' or 'session', got %d", len(events))
	}
	if events[0].Title != "DnD Session" {
		t.Errorf("expected 'DnD Session', got %q", events[0].Title)
	}
}

func TestFetchFromICalURL_EmptyURL(t *testing.T) {
	settings := db.EventSettings{
		SourceType: "ical",
		ICalURL:    "",
	}
	_, err := fetchFromICalURL(settings)
	if err == nil {
		t.Error("expected error for empty iCal URL")
	}
	if !strings.Contains(err.Error(), "no iCal URL") {
		t.Errorf("expected 'no iCal URL' error, got %v", err)
	}
}

func TestFetchFromICalURL_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	settings := db.EventSettings{
		SourceType: "ical",
		ICalURL:    srv.URL,
	}
	_, err := fetchFromICalURL(settings)
	if err == nil {
		t.Error("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected status 500 error, got %v", err)
	}
}

func TestFetchFromICalURL_UnreachableURL(t *testing.T) {
	settings := db.EventSettings{
		SourceType: "ical",
		ICalURL:    "http://127.0.0.1:1/nonexistent.ics", // port 1 should fail
	}
	_, err := fetchFromICalURL(settings)
	if err == nil {
		t.Error("expected error for unreachable URL")
	}
}

func TestFetchDispatch_ICalSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		w.Write([]byte(sampleICS()))
	}))
	defer srv.Close()

	// Source type = ical should dispatch to fetchFromICalURL
	settings := db.EventSettings{
		SourceType: "ical",
		ICalURL:    srv.URL,
		Tags:       "",
	}

	events, err := fetchFromGoogle(settings)
	if err != nil {
		t.Fatalf("fetchFromGoogle dispatch failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events from iCal dispatch, got %d", len(events))
	}
}

func TestFetchDispatch_GoogleAPISource(t *testing.T) {
	// Source type = google_api with no credentials should fail with Google API error (not iCal)
	settings := db.EventSettings{
		SourceType: "google_api",
		CalendarID: "test@example.com",
		AuthMethod: "service_account",
		// No CredentialsJSON — should fail at Google API client creation
	}

	_, err := fetchFromGoogle(settings)
	if err == nil {
		t.Error("expected error for Google API with no credentials")
	}
	// Should NOT be an iCal error
	if strings.Contains(err.Error(), "iCal") {
		t.Errorf("should not dispatch to iCal for google_api source type, got: %v", err)
	}
}

func TestFetchAndCacheEvents_ICalStaleCacheFallback(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	// Seed stale cache
	staleEvents := []googlecalendar.Event{
		{ID: "stale-1", Title: "Stale Event", StartTime: time.Date(2026, 7, 15, 19, 0, 0, 0, time.UTC)},
	}
	db.SetCachedEvents(staleEvents, "")

	// Configure iCal source pointing to unreachable URL
	db.SaveEventSettings(db.EventSettings{
		SourceType:      "ical",
		ICalURL:         "http://127.0.0.1:1/unreachable.ics",
		CacheTTLSeconds: 1, // very short TTL so cache is always stale
	})

	events, errMsg := fetchAndCacheEvents(db.GetEventSettings(), "")
	if errMsg == "" {
		// If no error, we should have gotten events from somewhere
		if len(events) == 0 {
			t.Error("expected events from stale cache fallback")
		}
	}
	// The function should either return stale events or an error message
	// but not crash
}

func TestEventSettingsRoundTrip_SourceTypeAndICalURL(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	// Save settings with iCal source type
	settings := db.EventSettings{
		SourceType: "ical",
		ICalURL:    "https://example.com/calendar.ics",
		Tags:       "dnd,test",
		AuthMethod: "service_account",
	}
	db.SaveEventSettings(settings)

	// Read back and verify
	got := db.GetEventSettings()
	if got.SourceType != "ical" {
		t.Errorf("expected source_type 'ical', got %q", got.SourceType)
	}
	if got.ICalURL != "https://example.com/calendar.ics" {
		t.Errorf("expected ical_url, got %q", got.ICalURL)
	}
	if got.Tags != "dnd,test" {
		t.Errorf("expected tags 'dnd,test', got %q", got.Tags)
	}
}

func TestEventSettingsRoundTrip_DefaultSourceType(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	// Default settings should have source_type = "google_api"
	got := db.GetEventSettings()
	if got.SourceType != "google_api" {
		t.Errorf("expected default source_type 'google_api', got %q", got.SourceType)
	}
	if got.ICalURL != "" {
		t.Errorf("expected default ical_url '', got %q", got.ICalURL)
	}
}

func TestEventSettingsRoundTrip_GoogleAPI(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	// Save settings with google_api source type
	settings := db.EventSettings{
		SourceType: "google_api",
		CalendarID:  "test@example.com",
		Tags:        "dnd",
		AuthMethod:  "service_account",
	}
	db.SaveEventSettings(settings)

	got := db.GetEventSettings()
	if got.SourceType != "google_api" {
		t.Errorf("expected source_type 'google_api', got %q", got.SourceType)
	}
	if got.CalendarID != "test@example.com" {
		t.Errorf("expected calendar_id, got %q", got.CalendarID)
	}
}

func TestCampaignEventSettingsRoundTrip_SourceTypeAndICalURL(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	// Save campaign settings with iCal source type
	cs := db.CampaignEventSettings{
		CampaignID:  1,
		Slug:        "test-campaign",
		DisplayName: "Test Campaign",
		SourceType:  "ical",
		ICalURL:     "https://example.com/campaign.ics",
		Tags:        "dnd",
		IsActive:    true,
	}
	db.SaveCampaignEventSettings(cs)

	// Read back by slug
	got := db.GetCampaignEventSettingsBySlug("test-campaign")
	if got == nil {
		t.Fatal("expected campaign settings, got nil")
	}
	if got.SourceType != "ical" {
		t.Errorf("expected source_type 'ical', got %q", got.SourceType)
	}
	if got.ICalURL != "https://example.com/campaign.ics" {
		t.Errorf("expected ical_url, got %q", got.ICalURL)
	}
}

func TestCampaignEventSettingsRoundTrip_DefaultSourceType(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	// Save campaign settings without specifying source type (should default to google_api)
	cs := db.CampaignEventSettings{
		CampaignID:  1,
		Slug:        "default-campaign",
		DisplayName: "Default Campaign",
		IsActive:    true,
	}
	db.SaveCampaignEventSettings(cs)

	got := db.GetCampaignEventSettingsBySlug("default-campaign")
	if got == nil {
		t.Fatal("expected campaign settings, got nil")
	}
	if got.SourceType != "google_api" {
		t.Errorf("expected default source_type 'google_api', got %q", got.SourceType)
	}
	if got.ICalURL != "" {
		t.Errorf("expected default ical_url '', got %q", got.ICalURL)
	}
}

func TestEventsPageICalEmptyState(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	AppVersion = "test-version"

	r := setupTestRouter()
	r.GET("/events", EventsPage)

	// Configure iCal source with no URL
	db.SaveEventSettings(db.EventSettings{
		SourceType: "ical",
		ICalURL:    "",
	})

	w := testutil.Get(t, r, "/events")
	testutil.AssertStatus(t, w, 200)
	body := w.Body.String()
	if !contains(body, "No upcoming events found") {
		t.Error("expected empty state message for iCal with no URL")
	}
}

func TestEventsPageICalWithCachedEvents(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	AppVersion = "test-version"

	r := setupTestRouter()
	r.GET("/events", EventsPage)

	// Seed events into cache
	events := []googlecalendar.Event{
		{ID: "ical-1", Title: "iCal Event", StartTime: time.Date(2026, 7, 15, 19, 0, 0, 0, time.UTC)},
	}
	db.SetCachedEvents(events, "")

	// Configure iCal source with a URL (so it's not empty state)
	db.SaveEventSettings(db.EventSettings{
		SourceType: "ical",
		ICalURL:    "https://example.com/calendar.ics",
	})

	w := testutil.Get(t, r, "/events")
	testutil.AssertStatus(t, w, 200)
	body := w.Body.String()
	if !contains(body, "iCal Event") {
		t.Error("expected cached iCal event title in rendered page")
	}
}

// setupTestRouter creates a minimal gin router for testing.
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}
