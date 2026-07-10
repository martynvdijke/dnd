package googlecalendar

import (
	"bufio"
	"bytes"
	"log"
	"strings"
	"time"
)

// ParseICS parses an iCalendar (RFC 5545) feed and returns upcoming events.
// It handles line unfolding, VEVENT blocks, date-only and datetime formats.
// Events with missing DTSTART are skipped. Past events are filtered out.
// Results are capped at maxResults (50).
func ParseICS(data []byte) ([]Event, error) {
	return ParseICSWithLimit(data, 50)
}

// ParseICSWithLimit is like ParseICS but allows specifying the max number of events.
func ParseICSWithLimit(data []byte, maxResults int) ([]Event, error) {
	if maxResults <= 0 {
		maxResults = 50
	}

	lines, err := unfoldLines(data)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var events []Event
	var inEvent bool
	var ev Event

	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r\n")

		if strings.HasPrefix(trimmed, "BEGIN:VEVENT") {
			inEvent = true
			ev = Event{}
			continue
		}

		if strings.HasPrefix(trimmed, "END:VEVENT") {
			if inEvent {
				// Skip events with missing DTSTART
				if ev.StartTime.IsZero() {
					log.Printf("ical_parser: skipping event with missing DTSTART: %q", ev.Title)
				} else if !ev.StartTime.Before(now) {
					// Only include upcoming events
					if ev.ID == "" {
						ev.ID = ev.Title + "-" + ev.StartTime.Format("20060102T150405")
					}
					events = append(events, ev)
				}
			}
			inEvent = false
			continue
		}

		if !inEvent {
			continue
		}

		// Parse property: NAME;PARAMS:VALUE or NAME:VALUE
		colonIdx := strings.Index(trimmed, ":")
		if colonIdx < 0 {
			continue
		}

		nameAndParams := trimmed[:colonIdx]
		value := trimmed[colonIdx+1:]

		// Split name from params on first semicolon
		semiIdx := strings.Index(nameAndParams, ";")
		name := nameAndParams
		params := ""
		if semiIdx >= 0 {
			name = nameAndParams[:semiIdx]
			params = nameAndParams[semiIdx+1:]
		}
		name = strings.ToUpper(name)

		switch name {
		case "SUMMARY":
			ev.Title = unescapeICS(value)
		case "DESCRIPTION":
			ev.Description = unescapeICS(value)
		case "LOCATION":
			ev.Location = unescapeICS(value)
		case "UID":
			ev.ID = value
		case "DTSTART":
			t, allDay := parseICSDate(value, params)
			ev.StartTime = t
			ev.AllDay = allDay
		case "DTEND":
			t, _ := parseICSDate(value, params)
			ev.EndTime = t
		}
	}

	// Sort by start time (earliest first) — simple insertion since events are usually few
	for i := 1; i < len(events); i++ {
		for j := i; j > 0 && events[j].StartTime.Before(events[j-1].StartTime); j-- {
			events[j], events[j-1] = events[j-1], events[j]
		}
	}

	// Cap at maxResults
	if len(events) > maxResults {
		events = events[:maxResults]
	}

	return events, nil
}

// unfoldLines handles RFC 5545 line unfolding: lines starting with space or tab
// are continuations of the previous line.
func unfoldLines(data []byte) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Increase buffer size for large feeds
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	var current string
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			// Continuation line — append to current
			current += line[1:]
		} else {
			if current != "" {
				lines = append(lines, current)
			}
			current = line
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines, scanner.Err()
}

// parseICSDate parses an ICS date or datetime value.
// Returns the parsed time and whether it's an all-day event.
// Handles:
//   - VALUE=DATE:20240101 (all-day)
//   - 20240101T120000Z (UTC datetime)
//   - 20240101T120000 (local datetime, treated as UTC)
//   - TZID=America/New_York:20240101T120000 (timezone-aware, parsed as the offset)
func parseICSDate(value, params string) (time.Time, bool) {
	// Check for VALUE=DATE parameter
	upperParams := strings.ToUpper(params)
	if strings.Contains(upperParams, "VALUE=DATE") || (len(value) == 8 && !strings.Contains(value, "T")) {
		// Date-only: 20240101
		t, err := time.Parse("20060102", value)
		if err != nil {
			return time.Time{}, false
		}
		return t, true
	}

	// Try TZID parameter
	if strings.Contains(upperParams, "TZID=") {
		// Extract timezone name
		tzid := ""
		for _, p := range strings.Split(params, ";") {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(strings.ToUpper(p), "TZID=") {
				tzid = p[5:]
				break
			}
		}
		if tzid != "" {
			loc, err := time.LoadLocation(tzid)
			if err == nil {
				t, err := time.ParseInLocation("20060102T150405", value, loc)
				if err == nil {
					return t, false
				}
			}
		}
	}

	// Datetime with Z suffix (UTC)
	if strings.HasSuffix(value, "Z") {
		t, err := time.Parse("20060102T150405Z", value)
		if err != nil {
			return time.Time{}, false
		}
		return t, false
	}

	// Datetime without Z (treat as UTC)
	if strings.Contains(value, "T") {
		t, err := time.Parse("20060102T150405", value)
		if err != nil {
			return time.Time{}, false
		}
		return t.UTC(), false
	}

	// Fallback: try date-only
	t, err := time.Parse("20060102", value)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// unescapeICS reverses ICS text escaping (RFC 5545 Section 3.3.11).
func unescapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\N", "\n")
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\;", ";")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}
