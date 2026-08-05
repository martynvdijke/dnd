package db

import (
	"testing"
	"time"

	googlecalendar "villum/google_calendar"
)

func TestCampaignEventSettings(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	t.Run("list is initially empty", func(t *testing.T) {
		list := ListCampaignEventSettings()
		if len(list) != 0 {
			t.Errorf("expected empty list, got %d items", len(list))
		}
	})

	t.Run("create and read", func(t *testing.T) {
		id, err := SaveCampaignEventSettings(CampaignEventSettings{
			CampaignID:  1,
			Slug:        "test-campaign",
			DisplayName: "Test Campaign",
			CalendarID:  "test@example.com",
			Tags:        "dnd,session",
			IsActive:    true,
		})
		if err != nil {
			t.Fatalf("SaveCampaignEventSettings: %v", err)
		}
		if id == 0 {
			t.Fatal("expected non-zero id")
		}

		// Read by slug
		s := GetCampaignEventSettingsBySlug("test-campaign")
		if s == nil {
			t.Fatal("expected to find by slug")
		}
		if s.DisplayName != "Test Campaign" {
			t.Errorf("expected 'Test Campaign', got %q", s.DisplayName)
		}
		if !s.IsActive {
			t.Error("expected is_active=true")
		}

		// Read by ID
		s2 := GetCampaignEventSettingsByID(id)
		if s2 == nil {
			t.Fatal("expected to find by id")
		}
		if s2.Slug != "test-campaign" {
			t.Errorf("expected slug 'test-campaign', got %q", s2.Slug)
		}
	})

	t.Run("slug uniqueness is enforced", func(t *testing.T) {
		_, err := SaveCampaignEventSettings(CampaignEventSettings{
			CampaignID:  2,
			Slug:        "test-campaign",
			DisplayName: "Duplicate Campaign",
			IsActive:    true,
		})
		if err == nil {
			t.Fatal("expected error for duplicate slug, got nil")
		}
	})

	t.Run("list includes created", func(t *testing.T) {
		list := ListCampaignEventSettings()
		if len(list) != 1 {
			t.Errorf("expected 1 item, got %d", len(list))
		}
	})

	t.Run("update", func(t *testing.T) {
		s := GetCampaignEventSettingsBySlug("test-campaign")
		if s == nil {
			t.Fatal("expected to find campaign")
		}
		s.DisplayName = "Updated Campaign"
		s.Tags = "dnd"
		_, err := SaveCampaignEventSettings(*s)
		if err != nil {
			t.Fatalf("SaveCampaignEventSettings update: %v", err)
		}
		updated := GetCampaignEventSettingsBySlug("test-campaign")
		if updated.DisplayName != "Updated Campaign" {
			t.Errorf("expected 'Updated Campaign', got %q", updated.DisplayName)
		}
		if updated.Tags != "dnd" {
			t.Errorf("expected tags 'dnd', got %q", updated.Tags)
		}
	})

	t.Run("deactivate", func(t *testing.T) {
		s := GetCampaignEventSettingsBySlug("test-campaign")
		if s == nil {
			t.Fatal("expected to find campaign")
		}
		s.IsActive = false
		SaveCampaignEventSettings(*s)
		updated := GetCampaignEventSettingsBySlug("test-campaign")
		if updated.IsActive {
			t.Error("expected is_active=false after deactivation")
		}
	})

	t.Run("delete", func(t *testing.T) {
		s := GetCampaignEventSettingsBySlug("test-campaign")
		if s == nil {
			t.Fatal("expected to find campaign before delete")
		}
		if err := DeleteCampaignEventSettings(s.ID); err != nil {
			t.Fatalf("DeleteCampaignEventSettings: %v", err)
		}
		if got := GetCampaignEventSettingsBySlug("test-campaign"); got != nil {
			t.Error("expected nil after delete")
		}
		if got := GetCampaignEventSettingsByID(s.ID); got != nil {
			t.Error("expected nil after delete by ID")
		}
	})
}

func TestPerCampaignCache(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	events := []googlecalendar.Event{
		{ID: "global-1", Title: "Global Event", Description: "Global"},
		{ID: "camp-1", Title: "Campaign Event", Description: "Campaign"},
	}

	t.Run("global cache is isolated from campaign cache", func(t *testing.T) {
		SetCachedEvents(events[:1], "")            // global
		SetCachedEvents(events[1:], "my-campaign") // per-campaign

		global, _ := GetCachedEvents(300, "")
		if len(global) != 1 || global[0].ID != "global-1" {
			t.Errorf("expected 1 global event, got %d", len(global))
		}

		campaign, _ := GetCachedEvents(300, "my-campaign")
		if len(campaign) != 1 || campaign[0].ID != "camp-1" {
			t.Errorf("expected 1 campaign event, got %d", len(campaign))
		}
	})

	t.Run("clearing campaign cache does not affect global", func(t *testing.T) {
		ClearCache("my-campaign")

		global, _ := GetCachedEvents(300, "")
		if len(global) != 1 {
			t.Errorf("expected global cache to survive campaign clear, got %d events", len(global))
		}

		campaign, _ := GetCachedEvents(300, "my-campaign")
		if len(campaign) != 0 {
			t.Errorf("expected empty campaign cache after clear, got %d events", len(campaign))
		}
	})

	t.Run("GetCachedCount is per-campaign", func(t *testing.T) {
		if count := GetCachedCount(""); count != 1 {
			t.Errorf("expected 1 global cached, got %d", count)
		}
		if count := GetCachedCount("my-campaign"); count != 0 {
			t.Errorf("expected 0 campaign cached, got %d", count)
		}
	})

	t.Run("clearing global cache clears all campaigns", func(t *testing.T) {
		SetCachedEvents(events[:1], "")
		SetCachedEvents(events[1:], "my-campaign")

		ClearCache("")

		if global, _ := GetCachedEvents(300, ""); len(global) != 0 {
			t.Errorf("expected global cache empty after global clear, got %d events", len(global))
		}
		if campaign, _ := GetCachedEvents(300, "my-campaign"); len(campaign) != 0 {
			t.Errorf("expected campaign cache empty after global clear, got %d events", len(campaign))
		}
	})
}

func TestCampaignPerCacheFreshFlag(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	SetCachedEvents([]googlecalendar.Event{
		{ID: "evt1", Title: "Campaign Test", StartTime: time.Now()},
	}, "campaign-slug")

	// Fresh via campaign slug
	cached, fresh := GetCachedEvents(99999, "campaign-slug")
	if len(cached) != 1 {
		t.Errorf("expected 1 event for campaign, got %d", len(cached))
	}
	if !fresh {
		t.Error("expected campaign cache to be fresh")
	}

	// Global should be empty
	global, _ := GetCachedEvents(99999, "")
	if len(global) != 0 {
		t.Errorf("expected 0 global events, got %d", len(global))
	}
}
