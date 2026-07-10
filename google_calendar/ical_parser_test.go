package googlecalendar

import (
	"strings"
	"testing"
	"time"
)

func TestParseICS_ValidFeed(t *testing.T) {
	// Build a feed with one upcoming and one past event
	future := time.Now().Add(48 * time.Hour).UTC().Format("20060102T150405Z")
	past := time.Now().Add(-48 * time.Hour).UTC().Format("20060102T150405Z")

	ics := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:future-event@test
SUMMARY:Future Session
DESCRIPTION:D&D night
DTSTART:` + future + `
DTEND:` + future + `
LOCATION:The Tavern
END:VEVENT
BEGIN:VEVENT
UID:past-event@test
SUMMARY:Past Session
DESCRIPTION:Already happened
DTSTART:` + past + `
END:VEVENT
END:VCALENDAR`

	events, err := ParseICS([]byte(ics))
	if err != nil {
		t.Fatalf("ParseICS error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 upcoming event, got %d", len(events))
	}
	if events[0].Title != "Future Session" {
		t.Errorf("expected title 'Future Session', got %q", events[0].Title)
	}
	if events[0].Description != "D&D night" {
		t.Errorf("expected description 'D&D night', got %q", events[0].Description)
	}
	if events[0].Location != "The Tavern" {
		t.Errorf("expected location 'The Tavern', got %q", events[0].Location)
	}
	if events[0].ID != "future-event@test" {
		t.Errorf("expected ID 'future-event@test', got %q", events[0].ID)
	}
}

func TestParseICS_DateOnly(t *testing.T) {
	future := time.Now().Add(72 * time.Hour).UTC().Format("20060102")

	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:All Day Event
DTSTART;VALUE=DATE:` + future + `
END:VEVENT
END:VCALENDAR`

	events, err := ParseICS([]byte(ics))
	if err != nil {
		t.Fatalf("ParseICS error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if !events[0].AllDay {
		t.Error("expected AllDay=true for date-only event")
	}
}

func TestParseICS_LineUnfolding(t *testing.T) {
	future := time.Now().Add(48 * time.Hour).UTC().Format("20060102T150405Z")

	// RFC 5545: continuation lines start with space/tab which is the folding marker.
	// Unfolding removes the CRLF + leading whitespace, reconstructing the original.
	// "This is a very long tit" + " le that is folded" → "This is a very long title that is folded"
	ics := "BEGIN:VCALENDAR\nBEGIN:VEVENT\nSUMMARY:This is a very long tit\n le that is folded\nDESCRIPTION:Long descript\n ion continues here\nDTSTART:" + future + "\nEND:VEVENT\nEND:VCALENDAR"

	events, err := ParseICS([]byte(ics))
	if err != nil {
		t.Fatalf("ParseICS error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Title != "This is a very long title that is folded" {
		t.Errorf("unexpected unfolded title: %q", events[0].Title)
	}
	if events[0].Description != "Long description continues here" {
		t.Errorf("unexpected unfolded description: %q", events[0].Description)
	}
}

func TestParseICS_MissingDTStart(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:No Date Event
DESCRIPTION:This event has no DTSTART
END:VEVENT
END:VCALENDAR`

	events, err := ParseICS([]byte(ics))
	if err != nil {
		t.Fatalf("ParseICS error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events (missing DTSTART), got %d", len(events))
	}
}

func TestParseICS_MalformedFeed(t *testing.T) {
	ics := `BEGIN:VCALENDAR
This is not a valid event
BEGIN:VEVENT
SUMMARY:Broken
DTSTART:NOTADATE
END:VEVENT
BEGIN:VEVENT
SUMMARY:Also Broken
END:VEVENT
END:VCALENDAR`

	events, err := ParseICS([]byte(ics))
	if err != nil {
		t.Fatalf("ParseICS error: %v", err)
	}
	// Malformed events should be skipped, not cause a panic
	if len(events) != 0 {
		t.Fatalf("expected 0 valid events from malformed feed, got %d", len(events))
	}
}

func TestParseICS_EmptyFeed(t *testing.T) {
	ics := `BEGIN:VCALENDAR
VERSION:2.0
END:VCALENDAR`

	events, err := ParseICS([]byte(ics))
	if err != nil {
		t.Fatalf("ParseICS error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events from empty feed, got %d", len(events))
	}
}

func TestParseICS_Escaping(t *testing.T) {
	future := time.Now().Add(48 * time.Hour).UTC().Format("20060102T150405Z")

	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Session #5\, The Big Battle
DESCRIPTION:Dragons\; goblins\, and more\\nNew line here
DTSTART:` + future + `
END:VEVENT
END:VCALENDAR`

	events, err := ParseICS([]byte(ics))
	if err != nil {
		t.Fatalf("ParseICS error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Title != "Session #5, The Big Battle" {
		t.Errorf("unexpected title: %q", events[0].Title)
	}
	if !strings.Contains(events[0].Description, "Dragons; goblins, and more") {
		t.Errorf("unexpected description: %q", events[0].Description)
	}
	if !strings.Contains(events[0].Description, "\n") {
		t.Errorf("expected newline in description, got %q", events[0].Description)
	}
}

func TestParseICS_TZID(t *testing.T) {
	// Use a future date in a specific timezone
	// We'll use UTC to keep the test deterministic
	futureUTC := time.Now().Add(48 * time.Hour).UTC()
	futureStr := futureUTC.Format("20060102T150405")

	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:TZID Event
DTSTART;TZID=UTC:` + futureStr + `
END:VEVENT
END:VCALENDAR`

	events, err := ParseICS([]byte(ics))
	if err != nil {
		t.Fatalf("ParseICS error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Title != "TZID Event" {
		t.Errorf("expected title 'TZID Event', got %q", events[0].Title)
	}
}

func TestParseICSWithLimit(t *testing.T) {
	var ics strings.Builder
	ics.WriteString("BEGIN:VCALENDAR\n")
	for i := 0; i < 10; i++ {
		future := time.Now().Add(time.Duration(48+i*24) * time.Hour).UTC().Format("20060102T150405Z")
		ics.WriteString("BEGIN:VEVENT\n")
		ics.WriteString("SUMMARY:Event " + string(rune('A'+i)) + "\n")
		ics.WriteString("DTSTART:" + future + "\n")
		ics.WriteString("END:VEVENT\n")
	}
	ics.WriteString("END:VCALENDAR")

	events, err := ParseICSWithLimit([]byte(ics.String()), 3)
	if err != nil {
		t.Fatalf("ParseICSWithLimit error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events (capped), got %d", len(events))
	}
}
