package db

import (
	"testing"
)

func TestEventSettingsDefaults(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	s := GetEventSettings()
	if s.Tags != "dnd,session,oneshot" {
		t.Errorf("expected default tags 'dnd,session,oneshot', got %q", s.Tags)
	}
	if s.CacheTTLSeconds != 300 {
		t.Errorf("expected default TTL 300, got %d", s.CacheTTLSeconds)
	}
	if s.AuthMethod != "service_account" {
		t.Errorf("expected default auth method 'service_account', got %q", s.AuthMethod)
	}
}

func TestEventSettingsSaveAndRead(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	input := EventSettings{
		CalendarID:        "test@example.com",
		Tags:              "dnd,oneshot",
		CacheTTLSeconds:   600,
		CredentialsJSON:   `{"type":"service_account"}`,
		AuthMethod:        "oauth",
		OAuthClientID:     "my-client-id",
		OAuthClientSecret: "my-client-secret",
		OAuthRefreshToken: "my-refresh-token",
		ColorLabels:       "1,5,banana",
		FilterMode:        "both",
	}

	if err := SaveEventSettings(input); err != nil {
		t.Fatalf("SaveEventSettings: %v", err)
	}

	s := GetEventSettings()
	if s.CalendarID != "test@example.com" {
		t.Errorf("calendar_id: expected %q, got %q", "test@example.com", s.CalendarID)
	}
	if s.Tags != "dnd,oneshot" {
		t.Errorf("tags: expected %q, got %q", "dnd,oneshot", s.Tags)
	}
	if s.CacheTTLSeconds != 600 {
		t.Errorf("cache_ttl: expected 600, got %d", s.CacheTTLSeconds)
	}
	if s.AuthMethod != "oauth" {
		t.Errorf("auth_method: expected 'oauth', got %q", s.AuthMethod)
	}
	if s.OAuthClientID != "my-client-id" {
		t.Errorf("oauth_client_id: expected %q, got %q", "my-client-id", s.OAuthClientID)
	}
	if s.OAuthClientSecret != "my-client-secret" {
		t.Errorf("oauth_client_secret: expected %q, got %q", "my-client-secret", s.OAuthClientSecret)
	}
	if s.OAuthRefreshToken != "my-refresh-token" {
		t.Errorf("oauth_refresh_token: expected %q, got %q", "my-refresh-token", s.OAuthRefreshToken)
	}
	if s.ColorLabels != "1,5,banana" {
		t.Errorf("color_labels: expected %q, got %q", "1,5,banana", s.ColorLabels)
	}
	if s.FilterMode != "both" {
		t.Errorf("filter_mode: expected %q, got %q", "both", s.FilterMode)
	}
	if got := s.ParseColorLabels(); len(got) != 3 {
		t.Errorf("ParseColorLabels: expected 3 labels, got %v", got)
	}
}

func TestEventSettingsOverwrite(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Save initial
	SaveEventSettings(EventSettings{
		CalendarID: "first@example.com",
		Tags:       "dnd",
	})

	// Overwrite
	SaveEventSettings(EventSettings{
		CalendarID: "second@example.com",
		Tags:       "oneshot",
	})

	s := GetEventSettings()
	if s.CalendarID != "second@example.com" {
		t.Errorf("expected overwritten calendar_id, got %q", s.CalendarID)
	}
	if s.Tags != "oneshot" {
		t.Errorf("expected overwritten tags, got %q", s.Tags)
	}
}

func TestEventSettingsInvalidTTLUsesDefault(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	SaveEventSettings(EventSettings{
		CacheTTLSeconds: 0,
	})

	s := GetEventSettings()
	if s.CacheTTLSeconds <= 0 {
		t.Errorf("expected TTL to default to positive value, got %d", s.CacheTTLSeconds)
	}
}

func TestSeedDefaultEventSettings(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Seed should create the default row
	SeedDefaultEventsSettings()

	s := GetEventSettings()
	if s.Tags != "dnd,session,oneshot" {
		t.Errorf("expected seeded default tags, got %q", s.Tags)
	}

	// Second seed should not overwrite
	s.Tags = "custom"
	SaveEventSettings(s)
	SeedDefaultEventsSettings()

	s2 := GetEventSettings()
	if s2.Tags != "custom" {
		t.Errorf("expected custom tags preserved after second seed, got %q", s2.Tags)
	}
}
