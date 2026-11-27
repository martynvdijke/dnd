# ical-export-feed Specification

## Purpose
Public iCalendar (.ics) endpoint that exports upcoming filtered events in the standard ICS format, enabling calendar app subscriptions.

## Requirements

### Requirement: ICS feed endpoint
The system SHALL provide a public HTTP endpoint `GET /events/ical` that returns an iCalendar (ICS) file containing upcoming events from the configured Google Calendar, respecting the same tag filter settings and cache layer as the events web page. The system SHALL additionally provide campaign-scoped ICS feeds at `GET /events/c/{slug}/ical` returning events for that campaign, respecting the campaign's filter settings and cache.

#### Scenario: GET /events/ical returns valid ICS
- **WHEN** a client requests `GET /events/ical`
- **THEN** the response SHALL have `Content-Type: text/calendar` header
- **THEN** the response SHALL have `Content-Disposition: attachment; filename="events.ics"` header
- **THEN** the response body SHALL be valid ICS format beginning with `BEGIN:VCALENDAR` and ending with `END:VCALENDAR`
- **THEN** the response SHALL include a `VERSION:2.0` and `PRODID:-//villum//Events//EN` in the VCALENDAR block
- **THEN** the response SHALL include a `METHOD:PUBLISH` property

#### Scenario: ICS feed returns 200 when no calendar configured
- **WHEN** a client requests `GET /events/ical` and no calendar ID is configured
- **THEN** the response SHALL still return a valid empty ICS file (no VEVENT entries)

#### Scenario: Cache miss triggers Google Calendar fetch
- **WHEN** a client requests `GET /events/ical` and the cache is empty or expired
- **THEN** the system SHALL attempt to fetch fresh events from Google Calendar before generating the ICS

#### Scenario: ICS feed respects tag filters
- **WHEN** tags are configured in event settings (e.g., "dnd,session")
- **THEN** only events matching those tags SHALL appear in the ICS feed
- **WHEN** no tags are configured
- **THEN** all upcoming events SHALL appear in the ICS feed

#### Scenario: Campaign-scoped ICS feed
- **WHEN** a client requests `GET /events/c/{slug}/ical` for an enabled campaign
- **THEN** the response SHALL have `Content-Type: text/calendar` header
- **AND** the response body SHALL be valid ICS containing only that campaign's filtered events
- **WHEN** the campaign slug is unknown or disabled
- **THEN** the response SHALL be HTTP 404

### Requirement: ICS VEVENT format
Each event in the ICS feed SHALL be formatted as a VEVENT component with the following properties.

#### Scenario: Timed event VEVENT
- **WHEN** an event has a specific start and end time (not all-day)
- **THEN** the VEVENT SHALL include `DTSTART:{start-time-in-UTC}` with `Z` suffix
- **THEN** the VEVENT SHALL include `DTEND:{end-time-in-UTC}` with `Z` suffix
- **THEN** the VEVENT SHALL include `SUMMARY:{event-title}` with proper escaping
- **THEN** the VEVENT SHALL include a unique `UID:{event-id}@villum.events`
- **THEN** the VEVENT SHALL include `DTSTAMP:{current-time-in-UTC}` with `Z` suffix

#### Scenario: All-day event VEVENT
- **WHEN** an event is an all-day event
- **THEN** the VEVENT SHALL use `VALUE=DATE` format for `DTSTART` and `DTEND` (e.g., `DTSTART;VALUE=DATE:20260715`)
- **THEN** the VEVENT SHALL NOT include a time component in DTSTART/DTEND

#### Scenario: Event with description
- **WHEN** an event has a description
- **THEN** the VEVENT SHALL include `DESCRIPTION:{event-description}` with proper ICS escaping (commas, semicolons, newlines)
- **WHEN** an event has no description
- **THEN** the VEVENT SHALL omit the DESCRIPTION property

#### Scenario: Event with location
- **WHEN** an event has a location
- **THEN** the VEVENT SHALL include `LOCATION:{event-location}` with proper ICS escaping
- **WHEN** an event has no location
- **THEN** the VEVENT SHALL omit the LOCATION property

### Requirement: ICS escaping
The system SHALL properly escape ICS property values according to RFC 5545.

#### Scenario: Special character escaping
- **WHEN** a title, description, or location contains commas, semicolons, or backslashes
- **THEN** commas SHALL be escaped as `\,`
- **THEN** semicolons SHALL be escaped as `\;`
- **THEN** backslashes SHALL be escaped as `\\`
- **WHEN** a description contains newlines
- **THEN** newlines SHALL be escaped as `\n`

### Requirement: Subscribe button on events page
The events page SHALL include a visible link to the ICS feed so users can subscribe.

#### Scenario: Subscribe button visible
- **WHEN** a user visits `/events`
- **THEN** the page SHALL display a button or link labeled "Subscribe (iCal)" or similar
- **THEN** clicking the button SHALL navigate to `/events/ical` (triggering a download/subscription prompt)
- **THEN** the button SHALL be styled consistently with the existing "Refresh" button

### Requirement: Reuse existing auth and cache
The ICS feed SHALL reuse the same authentication, settings, and caching infrastructure as the events web page.

#### Scenario: Auth method reuse
- **WHEN** the ICS handler is called
- **THEN** it SHALL call the same `fetchAndCacheEvents` function used by `EventsPage`
- **THEN** it SHALL respect the configured auth method (service_account or oauth) from event settings
- **THEN** it SHALL use the same cache layer to avoid duplicate Google Calendar API calls
