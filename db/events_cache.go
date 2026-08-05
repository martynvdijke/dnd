package db

import (
	"fmt"
	"log"
	"time"

	googlecalendar "villum/google_calendar"
)

// GetCachedEvents returns cached events that are within the TTL.
// If no cache exists or the cache is expired, returns nil.
// If the cache is expired but has data, returns the stale data (for fallback).
// campaignSlug empty = global events, non-empty = per-campaign events.
func GetCachedEvents(ttlSeconds int, campaignSlug string) ([]googlecalendar.Event, bool) {
	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}

	rows, err := DB.Query(`
		SELECT event_id, title, COALESCE(description,''), start_time, end_time, COALESCE(location,''), all_day, COALESCE(color_id,''), cached_at
		FROM google_events_cache
		WHERE campaign_slug=?
		ORDER BY start_time ASC`, campaignSlug)
	if err != nil {
		log.Printf("events_cache: query error: %v", err)
		return nil, false
	}
	defer rows.Close()

	var events []googlecalendar.Event
	var newestCache time.Time

	for rows.Next() {
		var e googlecalendar.Event
		var startStr, endStr, cachedAtStr string
		var allDay int
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &startStr, &endStr, &e.Location, &allDay, &e.ColorId, &cachedAtStr); err != nil {
			log.Printf("events_cache: scan error: %v", err)
			continue
		}
		e.AllDay = allDay == 1

		if startStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", startStr); err == nil {
				e.StartTime = t
			} else if t, err := time.Parse(time.RFC3339, startStr); err == nil {
				e.StartTime = t
			}
		}
		if endStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", endStr); err == nil {
				e.EndTime = t
			} else if t, err := time.Parse(time.RFC3339, endStr); err == nil {
				e.EndTime = t
			}
		}
		if t, err := parseCachedAt(cachedAtStr); err == nil {
			if t.After(newestCache) {
				newestCache = t
			}
		}

		events = append(events, e)
	}

	if len(events) == 0 {
		return nil, false
	}

	// Check if cache is still fresh
	elapsed := time.Since(newestCache)
	if elapsed.Seconds() < float64(ttlSeconds) {
		return events, true // fresh
	}

	// Expired — return stale data for fallback
	return events, false
}

// SetCachedEvents upserts events into the cache for a given campaign slug.
// campaignSlug empty = global events.
func SetCachedEvents(events []googlecalendar.Event, campaignSlug string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear old cache for this campaign slug
	if _, err := tx.Exec("DELETE FROM google_events_cache WHERE campaign_slug=?", campaignSlug); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO google_events_cache(event_id, title, description, start_time, end_time, location, all_day, campaign_slug, color_id, cached_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		startStr := ""
		if !e.StartTime.IsZero() {
			startStr = e.StartTime.Format("2006-01-02 15:04:05")
		}
		endStr := ""
		if !e.EndTime.IsZero() {
			endStr = e.EndTime.Format("2006-01-02 15:04:05")
		}
		allDay := 0
		if e.AllDay {
			allDay = 1
		}
		if _, err := stmt.Exec(e.ID, e.Title, e.Description, startStr, endStr, e.Location, allDay, campaignSlug, e.ColorId); err != nil {
			log.Printf("events_cache: insert error: %v", err)
		}
	}

	return tx.Commit()
}

// ClearCache removes all cached events for a given campaign slug.
// campaignSlug empty = global events.
func ClearCache(campaignSlug string) error {
	_, err := DB.Exec("DELETE FROM google_events_cache WHERE campaign_slug=?", campaignSlug)
	if err != nil {
		log.Printf("events_cache: clear error: %v", err)
	}
	return err
}

// GetCachedCount returns the number of cached events for a campaign slug.
func GetCachedCount(campaignSlug string) int {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM google_events_cache WHERE campaign_slug=?", campaignSlug).Scan(&count)
	return count
}

// HasCacheExpired checks whether the cache is older than the given TTL for a campaign slug.
func HasCacheExpired(ttlSeconds int, campaignSlug string) bool {
	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}
	var newest string
	err := DB.QueryRow("SELECT MAX(cached_at) FROM google_events_cache WHERE campaign_slug=?", campaignSlug).Scan(&newest)
	if err != nil || newest == "" {
		return true // no cache = expired
	}
	t, err := parseCachedAt(newest)
	if err != nil {
		return true
	}
	return time.Since(t).Seconds() >= float64(ttlSeconds)
}

// parseCachedAt parses a cached_at timestamp string.
func parseCachedAt(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("cannot parse cached_at: %q", s)
}

// GetStaleCacheAge returns the number of seconds since the cache was last refreshed.
func GetStaleCacheAge(campaignSlug string) int {
	var newest string
	err := DB.QueryRow("SELECT MAX(cached_at) FROM google_events_cache WHERE campaign_slug=?", campaignSlug).Scan(&newest)
	if err != nil || newest == "" {
		return -1
	}
	t, err := parseCachedAt(newest)
	if err != nil {
		return -1
	}
	return int(time.Since(t).Seconds())
}
