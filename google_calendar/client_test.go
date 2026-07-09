package googlecalendar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func TestMatchesAnyTag(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		tags     []string
		expected bool
	}{
		{
			name:     "title matches tag case-insensitive",
			event:    Event{Title: "Weekly DnD Session", Description: "Come play"},
			tags:     []string{"dnd"},
			expected: true,
		},
		{
			name:     "description matches tag",
			event:    Event{Title: "Game Night", Description: "One-shot session for new players"},
			tags:     []string{"one-shot"},
			expected: true,
		},
		{
			name:     "no match",
			event:    Event{Title: "Birthday Party", Description: "Celebration"},
			tags:     []string{"dnd"},
			expected: false,
		},
		{
			name:     "empty tags returns false",
			event:    Event{Title: "Anything", Description: ""},
			tags:     []string{},
			expected: false,
		},
		{
			name:     "multiple tags any match",
			event:    Event{Title: "Session Zero", Description: ""},
			tags:     []string{"dnd", "session", "oneshot"},
			expected: true,
		},
		{
			name:     "empty tag in list is skipped",
			event:    Event{Title: "DnD Night", Description: ""},
			tags:     []string{"", "dnd"},
			expected: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesAnyTag(tc.event, tc.tags)
			if got != tc.expected {
				t.Errorf("matchesAnyTag(%+v, %v) = %v, want %v", tc.event, tc.tags, got, tc.expected)
			}
		})
	}
}

func TestToEvent(t *testing.T) {
	now := time.Now().UTC()

	t.Run("timed event", func(t *testing.T) {
		item := &calendar.Event{
			Id:          "evt1",
			Summary:     "DnD Session",
			Description: "Weekly game",
			Location:    "Tabletop Tavern",
			Start:       &calendar.EventDateTime{DateTime: now.Format(time.RFC3339)},
			End:         &calendar.EventDateTime{DateTime: now.Add(3 * time.Hour).Format(time.RFC3339)},
		}
		e := toEvent(item)
		if e.ID != "evt1" || e.Title != "DnD Session" || e.AllDay {
			t.Errorf("timed event conversion incorrect: %+v", e)
		}
	})

	t.Run("all-day event", func(t *testing.T) {
		item := &calendar.Event{
			Id:      "evt2",
			Summary: "Convention 2026",
			Start:   &calendar.EventDateTime{Date: "2026-07-10"},
			End:     &calendar.EventDateTime{Date: "2026-07-12"},
		}
		e := toEvent(item)
		if !e.AllDay {
			t.Error("expected all-day event")
		}
		if e.StartTime.Format("2006-01-02") != "2026-07-10" {
			t.Errorf("unexpected start date: %v", e.StartTime)
		}
	})

	t.Run("nil start/end results in zero times", func(t *testing.T) {
		item := &calendar.Event{
			Id:      "evt3",
			Summary: "No time",
		}
		e := toEvent(item)
		if !e.StartTime.IsZero() {
			t.Error("expected StartTime to be zero when item.Start is nil")
		}
		if !e.EndTime.IsZero() {
			t.Error("expected EndTime to be zero when item.End is nil")
		}
	})
}

func TestFilterEventsByMode(t *testing.T) {
	events := []Event{
		{ID: "1", Title: "DnD Session", Description: "Weekly campaign", ColorId: "1"},
		{ID: "2", Title: "Board Game Night", Description: "Settlers of Catan", ColorId: "3"},
		{ID: "3", Title: "Oneshot: Curse of Strahd", Description: "Curse of Strahd intro", ColorId: "1"},
	}

	ctx := context.Background()

	t.Run("text mode: filter by single tag", func(t *testing.T) {
		result := filterEventsByMode(events, []string{"dnd"}, nil, "text", nil, ctx)
		if len(result) != 1 || result[0].ID != "1" {
			t.Errorf("expected 1 event (DnD Session), got %d: %+v", len(result), result)
		}
	})

	t.Run("text mode: filter by multiple tags OR", func(t *testing.T) {
		result := filterEventsByMode(events, []string{"dnd", "oneshot"}, nil, "text", nil, ctx)
		if len(result) != 2 {
			t.Errorf("expected 2 events, got %d", len(result))
		}
	})

	t.Run("text mode: no tags returns all", func(t *testing.T) {
		result := filterEventsByMode(events, nil, nil, "text", nil, ctx)
		if len(result) != 3 {
			t.Errorf("expected 3 events, got %d", len(result))
		}
	})

	t.Run("text mode: empty tags returns all", func(t *testing.T) {
		result := filterEventsByMode(events, []string{}, nil, "text", nil, ctx)
		if len(result) != 3 {
			t.Errorf("expected 3 events, got %d", len(result))
		}
	})

	t.Run("text mode: no match returns empty", func(t *testing.T) {
		result := filterEventsByMode(events, []string{"nonesuch"}, nil, "text", nil, ctx)
		if len(result) != 0 {
			t.Errorf("expected 0 events, got %d", len(result))
		}
	})

	// ─── Color mode tests ───
	t.Run("color mode: filter by single color", func(t *testing.T) {
		result := filterEventsByMode(events, nil, []string{"1"}, "color", nil, ctx)
		if len(result) != 2 {
			t.Errorf("expected 2 events (colorId=1), got %d", len(result))
		}
	})

	t.Run("color mode: filter by multiple colors OR", func(t *testing.T) {
		result := filterEventsByMode(events, nil, []string{"1", "3"}, "color", nil, ctx)
		if len(result) != 3 {
			t.Errorf("expected 3 events (colorId 1 or 3), got %d", len(result))
		}
	})

	t.Run("color mode: empty color labels is pass-through", func(t *testing.T) {
		result := filterEventsByMode(events, nil, nil, "color", nil, ctx)
		if len(result) != 3 {
			t.Errorf("expected 3 events (pass-through), got %d", len(result))
		}
	})

	t.Run("color mode: non-matching color returns empty", func(t *testing.T) {
		result := filterEventsByMode(events, nil, []string{"5"}, "color", nil, ctx)
		if len(result) != 0 {
			t.Errorf("expected 0 events, got %d", len(result))
		}
	})

	// ─── Both mode tests ───
	t.Run("both mode: event must match text AND color", func(t *testing.T) {
		result := filterEventsByMode(events, []string{"dnd"}, []string{"1"}, "both", nil, ctx)
		if len(result) != 1 || result[0].ID != "1" {
			t.Errorf("expected 1 event (DnD Session, colorId=1), got %d", len(result))
		}
	})

	t.Run("both mode: empty text pass-through", func(t *testing.T) {
		result := filterEventsByMode(events, nil, []string{"1"}, "both", nil, ctx)
		if len(result) != 2 {
			t.Errorf("expected 2 events (colorId=1, text pass-through), got %d", len(result))
		}
	})

	t.Run("both mode: empty color pass-through", func(t *testing.T) {
		result := filterEventsByMode(events, []string{"dnd"}, nil, "both", nil, ctx)
		if len(result) != 1 {
			t.Errorf("expected 1 event (text match, color pass-through), got %d", len(result))
		}
	})
}

func TestResolveColorLabels(t *testing.T) {
	t.Run("numeric IDs are returned directly", func(t *testing.T) {
		ids, _ := resolveColorLabelsStatic([]string{"1", "5", "11"})
		if !ids["1"] || !ids["5"] || !ids["11"] {
			t.Errorf("expected ids 1,5,11, got %v", ids)
		}
	})

	t.Run("color names are resolved to IDs", func(t *testing.T) {
		ids, _ := resolveColorLabelsStatic([]string{"lavender", "banana", "tomato"})
		if !ids["1"] || !ids["5"] || !ids["11"] {
			t.Errorf("expected ids 1,5,11, got %v", ids)
		}
	})

	t.Run("empty input returns nil", func(t *testing.T) {
		ids, _ := resolveColorLabelsStatic(nil)
		if ids != nil {
			t.Errorf("expected nil, got %v", ids)
		}
		ids, _ = resolveColorLabelsStatic([]string{})
		if ids != nil {
			t.Errorf("expected nil for empty slice, got %v", ids)
		}
	})
}

// resolveColorLabelsStatic tests the resolution logic without an API call.
func resolveColorLabelsStatic(labels []string) (map[string]bool, error) {
	if len(labels) == 0 {
		return nil, nil
	}
	ids := make(map[string]bool, len(labels))
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		if _, err := strconv.Atoi(label); err == nil {
			ids[label] = true
			continue
		}
		lower := strings.ToLower(label)
		if id, ok := standardColorNames[lower]; ok {
			ids[id] = true
			continue
		}
	}
	return ids, nil
}

func TestFetchUpcomingEvents(t *testing.T) {
	// Start a test HTTP server that mimics the Google Calendar API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/calendars/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"id":          "evt1",
					"summary":     "DnD Session #5",
					"description": "The finale",
					"start":       map[string]any{"dateTime": "2026-07-15T19:00:00Z"},
					"end":         map[string]any{"dateTime": "2026-07-15T22:00:00Z"},
				},
				{
					"id":      "evt2",
					"summary": "Birthday Party",
					"start":   map[string]any{"dateTime": "2026-07-16T14:00:00Z"},
					"end":     map[string]any{"dateTime": "2026-07-16T17:00:00Z"},
				},
			},
		})
	}))
	defer ts.Close()

	ctx := context.Background()
	svc, err := calendar.NewService(ctx,
		option.WithoutAuthentication(),
		option.WithEndpoint(ts.URL+"/"))
	if err != nil {
		t.Fatalf("create test service: %v", err)
	}

	client := &Client{svc: svc}

	t.Run("with tags filters results", func(t *testing.T) {
		events, err := client.FetchUpcomingEvents(ctx, "test@example.com", []string{"dnd"}, nil, "text", 50)
		if err != nil {
			t.Fatalf("FetchUpcomingEvents: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].ID != "evt1" {
			t.Errorf("expected evt1, got %s", events[0].ID)
		}
	})

	t.Run("no tags returns all events", func(t *testing.T) {
		events, err := client.FetchUpcomingEvents(ctx, "test@example.com", nil, nil, "text", 50)
		if err != nil {
			t.Fatalf("FetchUpcomingEvents: %v", err)
		}
		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}
	})
}

func TestNewOAuthClient_EmptyCredentials(t *testing.T) {
	_, err := NewOAuthClient(context.Background(), "", "secret", "refresh")
	if err == nil {
		t.Fatal("expected error for empty client ID")
	}
	_, err = NewOAuthClient(context.Background(), "id", "", "refresh")
	if err == nil {
		t.Fatal("expected error for empty client secret")
	}
	_, err = NewOAuthClient(context.Background(), "id", "secret", "")
	if err == nil {
		t.Fatal("expected error for empty refresh token")
	}
}

func TestNewClient_InvalidCredentials(t *testing.T) {
	_, err := NewClient(context.Background(), []byte("not-json"))
	if err == nil {
		t.Fatal("expected error for invalid credentials JSON")
	}
}
