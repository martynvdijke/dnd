# ical-url-event-source Specification

## Purpose
Fetch and parse events from a remote iCal/ICS URL as an alternative to the Google Calendar API, with text tag filtering and caching.

## Requirements

### Requirement: iCal URL as event source
The system SHALL support fetching events from a remote iCal/ICS URL as an alternative to the Google Calendar API. The admin SHALL configure the source type ("google_api" or "ical") and, when "ical" is selected, provide a public iCal URL.

#### Scenario: Admin selects iCal URL source type
- **WHEN** the admin selects "iCal URL" as the source type in the events settings
- **THEN** the system hides Google-specific fields (Calendar ID, Auth Method, Credentials, Color Labels, Filter Mode) and shows the iCal URL field and text Tags field

#### Scenario: Admin selects Google Calendar API source type
- **WHEN** the admin selects "Google Calendar API" as the source type
- **THEN** the system shows all existing Google Calendar fields and hides the iCal URL field

#### Scenario: Default source type is Google Calendar API
- **WHEN** a deployment has no source type configured (existing database)
- **THEN** the system defaults to "google_api" and existing behavior is preserved with no changes

### Requirement: iCal URL fetching and parsing
The system SHALL fetch the iCal feed from the configured URL via HTTP GET with a 10-second timeout, parse VEVENT entries, and return upcoming events (DTSTART >= now) as a list of Event structs.

#### Scenario: Successful iCal fetch
- **WHEN** the source type is "ical" and the iCal URL returns a valid ICS feed
- **THEN** the system parses VEVENT entries, filters to upcoming events only, caps at 50 events, and returns them as `[]googlecalendar.Event`

#### Scenario: iCal fetch with text tag filtering
- **WHEN** the source type is "ical", the iCal URL returns events, and tags are configured
- **THEN** the system returns only events whose SUMMARY or DESCRIPTION contains any of the configured tags (case-insensitive)

#### Scenario: iCal fetch failure with stale cache
- **WHEN** the iCal URL is unreachable and a stale cached result exists
- **THEN** the system serves the stale cached events without error

#### Scenario: iCal fetch failure with no cache
- **WHEN** the iCal URL is unreachable and no cached result exists
- **THEN** the system returns an error message: "Unable to load events at this time. Please check back later."

#### Scenario: Malformed ICS feed
- **WHEN** the iCal URL returns content that is not valid ICS or has malformed VEVENT entries
- **THEN** the system skips unparseable entries and events missing DTSTART, parses what it can, and logs a warning

### Requirement: iCal source caching
The system SHALL cache iCal-sourced events using the same cache layer and CacheTTLSeconds as Google Calendar API events.

#### Scenario: Cache hit for iCal source
- **WHEN** the source type is "ical" and fresh cached events exist (within CacheTTLSeconds)
- **THEN** the system serves events from cache without fetching the remote URL

#### Scenario: Cache expiry triggers refetch
- **WHEN** the cache TTL has expired and the source type is "ical"
- **THEN** the system fetches fresh events from the iCal URL and updates the cache

### Requirement: iCal source in per-campaign settings
Per-campaign event settings SHALL support the same source type and iCal URL fields as global settings. Each campaign page can independently use Google Calendar API or iCal URL.

#### Scenario: Campaign page with iCal source
- **WHEN** a campaign event setting has source type "ical" and a valid iCal URL
- **THEN** the campaign events page fetches and displays events from that iCal URL with the campaign's configured tags

#### Scenario: Campaign page inherits global Google credentials
- **WHEN** a campaign has source type "google_api" and no campaign-specific credentials
- **THEN** the campaign falls back to global Google credentials (existing behavior unchanged)

### Requirement: Admin tab naming
The admin tab for events configuration SHALL be labeled "Public Events" to clarify that it configures the public-facing events page.

#### Scenario: Admin tab label
- **WHEN** the admin page loads
- **THEN** the events configuration tab is labeled "Public Events" with a calendar icon
