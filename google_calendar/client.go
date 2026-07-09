package googlecalendar

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Event represents a single Google Calendar event with only the fields we need.
type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	AllDay      bool      `json:"all_day"`
	Location    string    `json:"location"`
	ColorId     string    `json:"color_id"`
}

// Standard Google Calendar event color name to colorId mapping.
// These are the 11 standard GCal event colors (independent of calendar palette).
var standardColorNames = map[string]string{
	"lavender":  "1",
	"sage":      "2",
	"grape":     "3",
	"flamingo":  "4",
	"banana":    "5",
	"tangerine": "6",
	"peacock":   "7",
	"graphite":  "8",
	"blueberry": "9",
	"basil":     "10",
	"tomato":    "11",
}

// colorNameCache caches the GCal color name→ID mapping from the API.
type colorCache struct {
	names     map[string]string // color name → colorId
	fetchedAt time.Time
}

var colorsCache colorCache

// resolveColorLabels parses a comma-separated list of color labels (names or IDs)
// and returns a set of resolved color IDs. Names are resolved case-insensitively.
func (c *Client) resolveColorLabels(ctx context.Context, labels []string) (map[string]bool, error) {
	if len(labels) == 0 {
		return nil, nil
	}

	// Try to refresh color names from the API for authoritative name→ID resolution.
	_, _ = c.refreshColorsFromAPI(ctx)

	ids := make(map[string]bool, len(labels))
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		// Check if it's already a numeric ID
		if id, err := strconv.Atoi(label); err == nil && id >= 1 && id <= 11 {
			ids[strconv.Itoa(id)] = true
			continue
		}
		// Try name lookup (case-insensitive)
		lower := strings.ToLower(label)
		if id, ok := standardColorNames[lower]; ok {
			ids[id] = true
			continue
		}
		// Check cached API names
		if colorsCache.names != nil {
			if id, ok := colorsCache.names[lower]; ok {
				ids[id] = true
				continue
			}
		}
		// Unknown color label — log and skip
		log.Printf("google_calendar: unknown color label %q (supported: lavender, sage, grape, flamingo, banana, tangerine, peacock, graphite, blueberry, basil, tomato, or IDs 1-11)", label)
	}
	return ids, nil
}

// refreshColorsFromAPI fetches the GCal Colors API and caches color definitions.
// The API does not return human-readable names, so this primarily validates IDs.
// The name→ID mapping is maintained as a hardcoded standard.
func (c *Client) refreshColorsFromAPI(ctx context.Context) (map[string]string, error) {
	// Return cached if still fresh (24h TTL)
	if colorsCache.names != nil && time.Since(colorsCache.fetchedAt) < 24*time.Hour {
		return colorsCache.names, nil
	}

	// If client is nil (tests, etc.), use standard names without API call
	if c == nil || c.svc == nil {
		colorsCache.names = standardColorNames
		colorsCache.fetchedAt = time.Now()
		return colorsCache.names, nil
	}

	colors, err := c.svc.Colors.Get().Context(ctx).Do()
	if err != nil {
		// Return stale cache on error
		if colorsCache.names != nil {
			return colorsCache.names, nil
		}
		return nil, fmt.Errorf("google_calendar: fetch colors: %w", err)
	}

	// Build name→ID mapping from event color definitions
	names := make(map[string]string, len(colors.Event))
	for id := range colors.Event {
		names[id] = id // Store ID as fallback
	}

	// Merge with hardcoded name mapping (the API provides hex colors, not names)
	// We use the well-known standard names
	colorsCache.names = standardColorNames
	colorsCache.fetchedAt = time.Now()

	return colorsCache.names, nil
}

// filterEventsByColor filters events to only those whose ColorId is in the allowed set.
// If allowed is nil or empty, all events pass through (no color filter).
func filterEventsByColor(events []Event, allowed map[string]bool) []Event {
	if len(allowed) == 0 {
		return events
	}
	var out []Event
	for _, e := range events {
		if allowed[e.ColorId] {
			out = append(out, e)
		}
	}
	return out
}

// Client wraps the Google Calendar API v3 service.
type Client struct {
	svc *calendar.Service
}

// NewClient creates a new Google Calendar client from a service account credentials JSON.
func NewClient(ctx context.Context, credentialsJSON []byte) (*Client, error) {
	svc, err := calendar.NewService(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("google_calendar: create service: %w", err)
	}
	return &Client{svc: svc}, nil
}

// NewOAuthClient creates a new Google Calendar client from OAuth 2.0 credentials.
// clientID and clientSecret are the OAuth 2.0 client credentials.
// refreshToken is a stored refresh token that will be used to obtain new access tokens automatically.
func NewOAuthClient(ctx context.Context, clientID, clientSecret, refreshToken string) (*Client, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, fmt.Errorf("google_calendar: OAuth client requires client_id, client_secret, and refresh_token")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{calendar.CalendarReadonlyScope},
	}

	token := &oauth2.Token{RefreshToken: refreshToken}
	tokenSource := config.TokenSource(ctx, token)

	svc, err := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("google_calendar: create service with OAuth: %w", err)
	}
	return &Client{svc: svc}, nil
}

// FetchUpcomingEvents queries events from the given calendar starting from now.
// Tags are used to filter events (case-insensitive OR match on title or description).
// If tags is nil or empty, all events are returned (when filterMode is "text" or "both").
// colorLabels are comma-separated GCal color IDs or names. If empty, color filter is a pass-through.
// filterMode controls how filters combine: "text" (tags only), "color" (colors only), "both" (AND).
// maxResults caps the number of events returned by the API (before filtering).
func (c *Client) FetchUpcomingEvents(ctx context.Context, calendarID string, tags []string, colorLabels []string, filterMode string, maxResults int64) ([]Event, error) {
	if maxResults <= 0 {
		maxResults = 50
	}

	now := time.Now().UTC().Format(time.RFC3339)
	items, err := c.svc.Events.List(calendarID).
		TimeMin(now).
		OrderBy("startTime").
		SingleEvents(true).
		MaxResults(maxResults).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("google_calendar: list events: %w", err)
	}

	// Convert all items to our Event struct first
	allEvents := make([]Event, len(items.Items))
	for i, item := range items.Items {
		allEvents[i] = toEvent(item)
	}

	// Apply filters based on filterMode
	return filterEventsByMode(allEvents, tags, colorLabels, filterMode, c, ctx), nil
}

// filterEventsByMode applies the configured filter mode and returns matching events.
func filterEventsByMode(events []Event, tags []string, colorLabels []string, filterMode string, c *Client, ctx context.Context) []Event {
	hasTags := len(tags) > 0
	resolvedColorIDs, _ := c.resolveColorLabels(ctx, colorLabels)
	hasColors := len(resolvedColorIDs) > 0

	switch filterMode {
	case "color":
		// Color-only mode: ignore text tags
		if !hasColors {
			return events // pass-through
		}
		return filterEventsByColor(events, resolvedColorIDs)

	case "both":
		// Both mode: AND combination, empty set = pass-through
		var out []Event
		for _, e := range events {
			passText := !hasTags || matchesAnyTag(e, tags)
			passColor := !hasColors || resolvedColorIDs[e.ColorId]
			if passText && passColor {
				out = append(out, e)
			}
		}
		return out

	default:
		// "text" mode (default): existing behavior
		if !hasTags {
			return events
		}
		var out []Event
		for _, e := range events {
			if matchesAnyTag(e, tags) {
				out = append(out, e)
			}
		}
		return out
	}
}

// matchesAnyTag checks if the event title or description contains any of the given tags (case-insensitive).
func matchesAnyTag(e Event, tags []string) bool {
	title := strings.ToLower(e.Title)
	desc := strings.ToLower(e.Description)
	for _, tag := range tags {
		t := strings.ToLower(strings.TrimSpace(tag))
		if t == "" {
			continue
		}
		if strings.Contains(title, t) || strings.Contains(desc, t) {
			return true
		}
	}
	return false
}

// toEvent converts a Google Calendar API event to our simplified Event struct.
func toEvent(item *calendar.Event) Event {
	e := Event{
		ID:          item.Id,
		Title:       item.Summary,
		Description: item.Description,
		Location:    item.Location,
		ColorId:     item.ColorId,
	}

	// Determine start time
	if item.Start != nil {
		if item.Start.DateTime != "" {
			t, err := time.Parse(time.RFC3339, item.Start.DateTime)
			if err == nil {
				e.StartTime = t
			}
		} else if item.Start.Date != "" {
			t, err := time.Parse("2006-01-02", item.Start.Date)
			if err == nil {
				e.StartTime = t
				e.AllDay = true
			}
		}
	}

	// Determine end time
	if item.End != nil {
		if item.End.DateTime != "" {
			t, err := time.Parse(time.RFC3339, item.End.DateTime)
			if err == nil {
				e.EndTime = t
			}
		} else if item.End.Date != "" {
			t, err := time.Parse("2006-01-02", item.End.Date)
			if err == nil {
				e.EndTime = t
			}
		}
	}

	return e
}

// ValidateCredentials tests whether the current client configuration can authenticate
// and access the specified calendar. Returns nil on success or a descriptive error.
func (c *Client) ValidateCredentials(ctx context.Context, calendarID string) error {
	// Try a minimal API call to verify auth works
	_, err := c.svc.Events.List(calendarID).
		MaxResults(1).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("google_calendar: credential validation failed for calendar %q: %w", calendarID, err)
	}
	return nil
}

// LogEvents prints event summaries for debugging.
func LogEvents(events []Event) {
	for _, e := range events {
		prefix := ""
		if e.AllDay {
			prefix = "[ALL-DAY] "
		}
		log.Printf("  %s%s (%s → %s)", prefix, e.Title, e.StartTime.Format("2006-01-02 15:04"), e.EndTime.Format("2006-01-02 15:04"))
	}
}
