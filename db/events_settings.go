package db

import (
	"log"
	"strings"
)

// EventSettings holds the Google Calendar integration configuration.
type EventSettings struct {
	SourceType        string `json:"source_type"` // "google_api" (default) or "ical"
	ICalURL           string `json:"ical_url"`    // iCal/ICS feed URL when SourceType == "ical"
	CalendarID        string `json:"calendar_id"`
	Tags              string `json:"tags"`
	CacheTTLSeconds   int    `json:"cache_ttl_seconds"`
	CredentialsJSON   string `json:"credentials_json"`
	AuthMethod        string `json:"auth_method"`         // "service_account" or "oauth"
	OAuthClientID     string `json:"oauth_client_id"`     // OAuth 2.0 client ID
	OAuthClientSecret string `json:"oauth_client_secret"` // OAuth 2.0 client secret
	OAuthRefreshToken string `json:"oauth_refresh_token"` // OAuth 2.0 refresh token
	ColorLabels       string `json:"color_labels"`        // Comma-separated GCal color IDs or names
	FilterMode        string `json:"filter_mode"`         // "text", "color", or "both"
}

// defaultEventSettings returns the default settings values.
func defaultEventSettings() EventSettings {
	return EventSettings{
		SourceType:        "google_api",
		ICalURL:           "",
		CalendarID:        "",
		Tags:              "dnd,session,oneshot",
		CacheTTLSeconds:   300,
		CredentialsJSON:   "",
		AuthMethod:        "service_account",
		OAuthClientID:     "",
		OAuthClientSecret: "",
		OAuthRefreshToken: "",
		ColorLabels:       "",
		FilterMode:        "text",
	}
}

// GetEventSettings reads the single settings row. If none exists, returns defaults.
func GetEventSettings() EventSettings {
	var s EventSettings
	err := DB.QueryRow(`
		SELECT COALESCE(source_type,'google_api'), COALESCE(ical_url,''),
			calendar_id, tags, cache_ttl_seconds, COALESCE(credentials_json,''),
			COALESCE(auth_method,'service_account'), COALESCE(oauth_client_id,''), COALESCE(oauth_client_secret,''), COALESCE(oauth_refresh_token,''),
			COALESCE(color_labels,''), COALESCE(filter_mode,'text')
		FROM events_settings WHERE id=1`).
		Scan(&s.SourceType, &s.ICalURL, &s.CalendarID, &s.Tags, &s.CacheTTLSeconds, &s.CredentialsJSON,
			&s.AuthMethod, &s.OAuthClientID, &s.OAuthClientSecret, &s.OAuthRefreshToken,
			&s.ColorLabels, &s.FilterMode)
	if err != nil {
		// Return defaults if no row exists
		return defaultEventSettings()
	}
	if s.AuthMethod == "" {
		s.AuthMethod = "service_account"
	}
	if s.SourceType == "" {
		s.SourceType = "google_api"
	}
	return s
}

// ParseColorLabels splits the comma-separated color labels string and trims whitespace.
func (s EventSettings) ParseColorLabels() []string {
	if s.ColorLabels == "" {
		return nil
	}
	var labels []string
	for _, l := range strings.Split(s.ColorLabels, ",") {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			labels = append(labels, trimmed)
		}
	}
	return labels
}

// SaveEventSettings upserts the settings row.
func SaveEventSettings(s EventSettings) error {
	if s.CacheTTLSeconds <= 0 {
		s.CacheTTLSeconds = 300
	}
	if s.AuthMethod == "" {
		s.AuthMethod = "service_account"
	}
	if s.FilterMode == "" {
		s.FilterMode = "text"
	}
	if s.SourceType == "" {
		s.SourceType = "google_api"
	}
	_, err := DB.Exec(`
		INSERT INTO events_settings(id, source_type, ical_url, calendar_id, tags, cache_ttl_seconds, credentials_json, auth_method, oauth_client_id, oauth_client_secret, oauth_refresh_token, color_labels, filter_mode)
		VALUES(1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			source_type=excluded.source_type,
			ical_url=excluded.ical_url,
			calendar_id=excluded.calendar_id,
			tags=excluded.tags,
			cache_ttl_seconds=excluded.cache_ttl_seconds,
			credentials_json=excluded.credentials_json,
			auth_method=excluded.auth_method,
			oauth_client_id=excluded.oauth_client_id,
			oauth_client_secret=excluded.oauth_client_secret,
			oauth_refresh_token=excluded.oauth_refresh_token,
			color_labels=excluded.color_labels,
			filter_mode=excluded.filter_mode`,
		s.SourceType, s.ICalURL, s.CalendarID, s.Tags, s.CacheTTLSeconds, s.CredentialsJSON, s.AuthMethod,
		s.OAuthClientID, s.OAuthClientSecret, s.OAuthRefreshToken,
		s.ColorLabels, s.FilterMode)
	if err != nil {
		log.Printf("events_settings: save error: %v", err)
	}
	return err
}

// SeedDefaultEventsSettings creates the default settings row if none exists.
func SeedDefaultEventsSettings() {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM events_settings").Scan(&count)
	if err != nil || count > 0 {
		return
	}
	def := defaultEventSettings()
	SaveEventSettings(def)
	log.Println("events_settings: seeded default settings")
}
