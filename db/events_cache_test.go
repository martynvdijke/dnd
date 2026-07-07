package db

import (
	"os"
	"testing"
	"time"

	googlecalendar "villum/google_calendar"
)

func TestEventsCache(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	events := []googlecalendar.Event{
		{ID: "evt1", Title: "DnD Session", Description: "Weekly game", StartTime: time.Date(2026, 7, 15, 19, 0, 0, 0, time.UTC)},
		{ID: "evt2", Title: "One-Shot", Description: "Curse of Strahd", StartTime: time.Date(2026, 7, 20, 18, 0, 0, 0, time.UTC)},
	}

	t.Run("initial cache is empty", func(t *testing.T) {
		cached, fresh := GetCachedEvents(300)
		if len(cached) != 0 {
			t.Errorf("expected empty cache, got %d events", len(cached))
		}
		if fresh {
			t.Error("expected cache to not be fresh")
		}
	})

	t.Run("set and read cached events", func(t *testing.T) {
		if err := SetCachedEvents(events); err != nil {
			t.Fatalf("SetCachedEvents: %v", err)
		}
		cached, fresh := GetCachedEvents(300)
		if len(cached) != 2 {
			t.Fatalf("expected 2 cached events, got %d", len(cached))
		}
		if !fresh {
			t.Error("expected cache to be fresh")
		}
		if cached[0].ID != "evt1" {
			t.Errorf("expected evt1 first, got %s", cached[0].ID)
		}
	})

	t.Run("set replaces old cache", func(t *testing.T) {
		if err := SetCachedEvents(events[:1]); err != nil {
			t.Fatalf("SetCachedEvents: %v", err)
		}
		cached, _ := GetCachedEvents(300)
		if len(cached) != 1 {
			t.Fatalf("expected 1 cached event, got %d", len(cached))
		}
	})

	t.Run("clear cache", func(t *testing.T) {
		if err := ClearCache(); err != nil {
			t.Fatalf("ClearCache: %v", err)
		}
		count := GetCachedCount()
		if count != 0 {
			t.Errorf("expected 0 after clear, got %d", count)
		}
	})

	t.Run("count cached events", func(t *testing.T) {
		SetCachedEvents(events)
		count := GetCachedCount()
		if count != 2 {
			t.Errorf("expected count 2, got %d", count)
		}
	})
}

func TestEventsCacheFreshFlag(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	SetCachedEvents([]googlecalendar.Event{
		{ID: "evt1", Title: "Test Event", StartTime: time.Now()},
	})

	// With large TTL, freshly inserted cache is fresh
	cached, fresh := GetCachedEvents(99999)
	if len(cached) != 1 {
		t.Errorf("expected 1 event, got %d", len(cached))
	}
	if !fresh {
		t.Error("expected newly inserted cache to be fresh with large TTL")
	}

	// With TTL=1, cache should still be fresh (< 1 sec elapsed normally)
	cached, fresh = GetCachedEvents(1)
	if len(cached) != 1 {
		t.Errorf("expected 1 event, got %d", len(cached))
	}
	if !fresh {
		t.Error("expected cache to be fresh with TTL=1 immediately after insert")
	}

	// With TTL=-1 (converted to 300), cache is fresh
	cached, fresh = GetCachedEvents(-1)
	if !fresh {
		t.Error("expected cache to be fresh with negative TTL (defaults to 300)")
	}
}

func TestHasCacheExpired(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// No cache → expired
	if !HasCacheExpired(300) {
		t.Error("expected no cache to be expired")
	}

	SetCachedEvents([]googlecalendar.Event{
		{ID: "evt1", Title: "Test", StartTime: time.Now()},
	})

	// Fresh cache → not expired with large TTL
	if HasCacheExpired(99999) {
		t.Error("expected new cache to not be expired with large TTL")
	}

	// HasCacheExpired internal logic: time.Since(t).Seconds() >= float64(ttlSeconds)
	// Since we can't guarantee test timing, just verify the function runs without error
	// and returns false for a freshly inserted record with large TTL
	ClearCache()
	if !HasCacheExpired(300) {
		t.Error("expected expired after cache cleared")
	}
}

func TestGetStaleCacheAge(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// No cache → -1
	if age := GetStaleCacheAge(); age != -1 {
		t.Errorf("expected -1 for no cache, got %d", age)
	}

	SetCachedEvents([]googlecalendar.Event{
		{ID: "evt1", Title: "Test", StartTime: time.Now()},
	})

	age := GetStaleCacheAge()
	if age < 0 {
		t.Errorf("expected positive age, got %d", age)
	}
}

func setupTestDB(t *testing.T) {
	t.Helper()
	path := "/tmp/villum_db_test_" + t.Name() + ".db"
	os.Remove(path)
	os.Remove(path + "-wal")
	os.Remove(path + "-shm")
	if err := Init(path); err != nil {
		t.Fatalf("db init: %v", err)
	}
}

func teardownTestDB(t *testing.T) {
	t.Helper()
	if DB != nil {
		Close()
	}
	path := "/tmp/villum_db_test_" + t.Name() + ".db"
	os.Remove(path)
	os.Remove(path + "-wal")
	os.Remove(path + "-shm")
}
