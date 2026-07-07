package googlecalendar

import (
	"context"
	"fmt"
	"log"
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
// If tags is nil or empty, all events are returned.
// maxResults caps the number of events returned by the API (before tag filtering).
func (c *Client) FetchUpcomingEvents(ctx context.Context, calendarID string, tags []string, maxResults int64) ([]Event, error) {
	if maxResults <= 0 {
		maxResults = 50
	}

	now := time.Now().UTC().Format(time.RFC3339)
	events, err := c.svc.Events.List(calendarID).
		TimeMin(now).
		OrderBy("startTime").
		SingleEvents(true).
		MaxResults(maxResults).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("google_calendar: list events: %w", err)
	}

	return filterEvents(events.Items, tags), nil
}

// filterEvents applies tag filtering and converts API items to our Event struct.
func filterEvents(items []*calendar.Event, tags []string) []Event {
	var out []Event
	hasTags := len(tags) > 0

	for _, item := range items {
		e := toEvent(item)

		if hasTags && !matchesAnyTag(e, tags) {
			continue
		}

		out = append(out, e)
	}
	return out
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
