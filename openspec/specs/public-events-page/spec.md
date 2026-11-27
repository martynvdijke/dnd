# public-events-page Specification

## Purpose
Public events pages now render with the site theme (light/dark) and include a theme toggle, matching the rest of the site. The login page links to the public events page so visitors can discover it.

## Requirements

### Requirement: Public events pages use the site theme

The public events pages (`/events`, `/events/c/{slug}`, and `/events/{id}`) SHALL render using the same light/dark theme system as the rest of the site. The pages SHALL read the user's saved theme preference from `localStorage` (key `villum-theme`, default `light`) and set the `data-theme` attribute on the `<html>` element before first paint to avoid a flash of the wrong theme. The pages SHALL NOT hardcode a fixed theme. Each events page SHALL include a theme toggle control that switches between light and dark and persists the choice to `localStorage`, matching the toggle behavior on the login and app pages.

#### Scenario: Events page respects saved light theme
- **GIVEN** the user's `localStorage` has `villum-theme` set to `light`
- **WHEN** the user visits `/events`
- **THEN** the page renders with `data-theme="light"` on the `<html>` element
- **AND** no flash of a dark theme occurs before content paints

#### Scenario: Events page respects saved dark theme
- **GIVEN** the user's `localStorage` has `villum-theme` set to `dark`
- **WHEN** the user visits `/events`
- **THEN** the page renders with `data-theme="dark"` on the `<html>` element

#### Scenario: Events page defaults to light when no preference saved
- **GIVEN** the user's `localStorage` has no `villum-theme` key
- **WHEN** the user visits `/events`
- **THEN** the page renders with `data-theme="light"`

#### Scenario: Theme toggle on events page switches and persists
- **WHEN** the user clicks the theme toggle on `/events`
- **THEN** the page switches to the opposite theme
- **AND** the choice is persisted to `localStorage` under `villum-theme`
- **AND** navigating to `/events/c/{slug}` or `/events/{id}` renders in the same theme

#### Scenario: Campaign event page uses site theme
- **GIVEN** the user's `localStorage` has `villum-theme` set to `dark`
- **WHEN** the user visits `/events/c/lost-mines`
- **THEN** the page renders with `data-theme="dark"`

#### Scenario: Event detail page uses site theme
- **GIVEN** the user's `localStorage` has `villum-theme` set to `light`
- **WHEN** the user visits `/events/123`
- **THEN** the page renders with `data-theme="light"`
- **AND** a theme toggle is present and functional

---

### Requirement: Public events page is linked from the login page

The login page SHALL include a visible link to the public events page (`/events`) so unauthenticated visitors can discover and navigate to the events page without logging in.

#### Scenario: Login page displays events link
- **WHEN** a user visits the login page
- **THEN** the page displays a link labeled to indicate public events (e.g., "View Events")
- **AND** the link's `href` points to `/events`

#### Scenario: Events link works without authentication
- **WHEN** an unauthenticated user clicks the events link on the login page
- **THEN** the browser navigates to `/events`
- **AND** the events page loads without requiring login

### Requirement: Public events page exists at /events

The system SHALL serve a public web page at the `/events` path that displays upcoming DnD events fetched from Google Calendar. The page SHALL NOT require authentication. The `/events` page SHALL serve as the default/global events view. The system SHALL additionally serve campaign-scoped event pages at `/events/c/{slug}` for each enabled campaign.

#### Scenario: Page loads successfully
- **WHEN** a user visits `/events`
- **THEN** the server returns HTTP 200 with an HTML page showing upcoming events
- **AND** the page renders within 5 seconds

#### Scenario: No events found
- **WHEN** a user visits `/events`
- **AND** there are no upcoming events matching the configured tags
- **THEN** the page displays a friendly message: "No upcoming events found"

#### Scenario: Google Calendar API is unreachable
- **WHEN** a user visits `/events`
- **AND** the Google Calendar API returns an error or is unreachable
- **THEN** the page displays a friendly error message: "Unable to load events at this time. Please check back later."

#### Scenario: Campaign page loads successfully
- **WHEN** a user visits `/events/c/{slug}` for an enabled campaign
- **THEN** the server returns HTTP 200 with an HTML page showing that campaign's upcoming events
- **AND** the page title includes the campaign's display name

---

### Requirement: Events are fetched from Google Calendar using configurable auth

The system SHALL connect to Google Calendar API v3 using one of two supported authentication methods (service account or OAuth 2.0), query upcoming events, and filter them against configurable tags/keywords and/or Google Calendar color labels depending on the configured `FilterMode`. The auth method SHALL be configurable via admin settings. When `FilterMode` is `text`, tags SHALL match against event title and description (case-insensitive) and events matching ANY tag SHALL be included. When `FilterMode` is `color`, events SHALL be filtered by their GCal `colorId` against configured color labels. When `FilterMode` is `both`, an event SHALL be included only if it passes both the text-tag filter and the color-label filter, with empty filter lists treated as pass-through.

#### Scenario: Service account auth
- **GIVEN** the system is configured with auth method "service_account"
- **AND** valid service account credentials JSON are provided
- **WHEN** fetching events from Google Calendar
- **THEN** the system authenticates using the service account JWT

#### Scenario: OAuth 2.0 auth with refresh token
- **GIVEN** the system is configured with auth method "oauth"
- **AND** valid OAuth client ID, client secret, and refresh token are provided
- **WHEN** fetching events from Google Calendar
- **THEN** the system authenticates using OAuth 2.0 with automatic token refresh

#### Scenario: OAuth token refresh on expiry
- **GIVEN** the system uses OAuth 2.0 auth
- **AND** the access token has expired
- **WHEN** fetching events
- **THEN** the system automatically refreshes the access token using the stored refresh token

#### Scenario: Invalid credentials are reported
- **GIVEN** the system has invalid or expired Google Calendar credentials
- **WHEN** fetching events
- **THEN** the system returns an error
- **AND** the events page displays "Unable to load events at this time"
- **AND** the error is logged on the server

#### Scenario: Events matching configured tags are displayed (text mode)
- **GIVEN** `FilterMode` is "text"
- **WHEN** the system fetches events from Google Calendar
- **AND** configured tags include "dnd" and "session"
- **AND** an event exists with title "Weekly DnD Session"
- **AND** another event exists with description "Join us for a one-shot session"
- **THEN** both events appear in the results

#### Scenario: Events not matching any tag are excluded (text mode)
- **GIVEN** `FilterMode` is "text"
- **WHEN** the system fetches events from Google Calendar
- **AND** configured tags include "dnd"
- **AND** an event exists with title "Birthday Party"
- **THEN** that event is excluded from results

#### Scenario: Tags are matched case-insensitively (text mode)
- **GIVEN** `FilterMode` is "text"
- **WHEN** configured tags include "DND"
- **AND** an event exists with title "weekly dnd session"
- **THEN** the event is included in results

#### Scenario: Empty tag list returns all events (text mode)
- **GIVEN** `FilterMode` is "text"
- **WHEN** the system fetches events from Google Calendar
- **AND** no tags are configured (empty list)
- **THEN** all upcoming events SHALL be returned

#### Scenario: Events filtered by color label (color mode)
- **GIVEN** `FilterMode` is "color"
- **AND** color labels are configured as "1"
- **WHEN** the system fetches events from Google Calendar
- **AND** an event has `colorId` "1"
- **THEN** the event is included
- **AND** an event with `colorId` "5" is excluded

#### Scenario: Events filtered by both text and color (both mode)
- **GIVEN** `FilterMode` is "both"
- **AND** tags include "dnd"
- **AND** color labels include "1"
- **WHEN** the system fetches events
- **AND** an event has title "DnD Session" and `colorId` "1"
- **THEN** the event is included
- **AND** an event with title "DnD Session" and `colorId` "5" is excluded

---

### Requirement: Events display date, time, title, and description

Each event in the list SHALL display: event title, start date/time, end date/time, and description. The date/time SHALL be formatted in a human-readable format in the viewer's timezone (default to server timezone if unknown). The system SHALL additionally provide a month-grid calendar view in which events are placed on their start-date cells, and event titles in both list and grid views SHALL be clickable links to a per-event detail page at `/events/{id}`.

#### Scenario: Event with all fields is displayed correctly
- **WHEN** an event has title "Session 5", start "2026-07-15T19:00:00", end "2026-07-15T22:00:00", and description "The finale of the Lost Mines campaign"
- **THEN** the list entry shows: title prominently, start and end times, and the description below

#### Scenario: Multi-day event displays date range
- **WHEN** an event spans multiple days (e.g., "Convention 2026" from July 10 to July 12)
- **THEN** the display shows the date range: "Jul 10 – Jul 12, 2026"

#### Scenario: All-day event shows date only (no time)
- **WHEN** an event is marked as an all-day event
- **THEN** only the date is displayed without a specific time

#### Scenario: Missing description is handled gracefully
- **WHEN** an event has no description
- **THEN** the description field is hidden from the list entry

#### Scenario: Event title links to detail page
- **WHEN** an event title is displayed in the list or grid view
- **THEN** the title is a hyperlink to `/events/{id}` for that event

#### Scenario: Grid view places event on start date
- **WHEN** an event starts on July 15, 2026
- **AND** the grid view is displaying July 2026
- **THEN** the event title appears in the July 15 cell

---

### Requirement: Events are cached locally

The system SHALL cache fetched Google Calendar events in a local SQLite table (`google_events_cache`). The cache SHALL have a configurable TTL (default 5 minutes). The system SHALL serve cached events when the API is unreachable (stale cache fallback). The cache SHALL support per-campaign entries keyed by campaign slug, with the global cache using a NULL campaign slug key.

#### Scenario: Cached events are served within TTL
- **WHEN** the cache is populated with events less than 5 minutes ago
- **THEN** the system returns cached events without calling the Google Calendar API

#### Scenario: Cache TTL has expired
- **WHEN** the cache was last refreshed more than 5 minutes ago
- **THEN** the system fetches fresh events from the Google Calendar API
- **AND** updates the cache

#### Scenario: Stale cache fallback on API failure
- **WHEN** the Google Calendar API is unreachable
- **AND** stale cached events exist (older than TTL)
- **THEN** the system serves the stale cached events
- **AND** logs a warning

#### Scenario: Campaign cache is independent
- **WHEN** campaign "lost-mines" has fresh cached events
- **AND** campaign "curse-of-strahd" has an expired cache
- **THEN** a request to the "lost-mines" page serves from cache without fetching
- **AND** a request to the "curse-of-strahd" page fetches fresh events

### Requirement: Search tags and calendar ID are configurable via admin settings

The system SHALL store Google Calendar ID, search tags, color labels, and filter mode in a database table (`events_settings`). An admin user SHALL be able to update these settings via the admin UI without restarting the server. The filter mode SHALL default to "text" for backward compatibility.

#### Scenario: Admin updates search tags
- **WHEN** an admin user updates the search tags field in admin settings
- **THEN** the new tags are persisted to the database
- **AND** the next event fetch uses the updated tags

#### Scenario: Admin updates Google Calendar ID
- **WHEN** an admin user updates the Google Calendar ID in admin settings
- **THEN** the new calendar ID is persisted to the database
- **AND** the cache is invalidated
- **AND** the next event fetch uses the new calendar

#### Scenario: Admin updates color labels
- **WHEN** an admin user updates the color labels field in admin settings
- **THEN** the new color labels are persisted to the database
- **AND** the cache is invalidated
- **AND** the next event fetch uses the updated color labels

#### Scenario: Admin updates filter mode
- **WHEN** an admin user changes the filter mode selector in admin settings
- **THEN** the new filter mode is persisted to the database
- **AND** the cache is invalidated

#### Scenario: Invalid calendar ID is handled gracefully
- **WHEN** an admin sets an invalid Google Calendar ID
- **THEN** the events page displays an appropriate error
- **AND** the admin is notified of the configuration issue

### Requirement: Events page auto-refreshes periodically

The events page SHALL auto-refresh the event list every 5 minutes using an HTMX trigger to keep the display current without requiring a full page reload. The grid view SHALL support HTMX-based month navigation via the `GET /htmx/events/grid?month=YYYY-MM` endpoint.

#### Scenario: Page auto-refreshes
- **WHEN** the events page is open in a browser
- **AND** 5 minutes pass
- **THEN** the event list is refreshed via HTMX without a full page reload

#### Scenario: Manual refresh available
- **WHEN** a user clicks a "Refresh" button on the events page
- **THEN** the event list is immediately refreshed

#### Scenario: Grid month navigation via HTMX
- **WHEN** a user clicks "Next" or "Previous" on the grid view
- **THEN** the grid updates to the adjacent month via HTMX without a full page reload

### Requirement: Public events page supports multiple source types
The public events page (`/events`) SHALL display events sourced from either the Google Calendar API or an iCal URL feed, depending on the admin-configured source type. The source type is transparent to the frontend — the same page template, list/grid views, and iCal export work regardless of source.

#### Scenario: Events page with Google Calendar API source
- **WHEN** the admin has configured source type "google_api" with a valid Calendar ID and credentials
- **THEN** the public events page displays events fetched from the Google Calendar API

#### Scenario: Events page with iCal URL source
- **WHEN** the admin has configured source type "ical" with a valid iCal URL
- **THEN** the public events page displays events fetched and parsed from the iCal URL

#### Scenario: Events page with no source configured
- **WHEN** no Calendar ID and no iCal URL are configured
- **THEN** the public events page shows an empty state with no error

### Requirement: Login page links to public events page
The login page SHALL include a visible link to the public events page (`/events`) so unauthenticated visitors can always access it.

#### Scenario: Visitor on login page
- **WHEN** an unauthenticated visitor views the login page
- **THEN** a "View Events" link is visible in the footer, linking to `/events`

### Requirement: Public events page uses site theme
The public events page SHALL use the same CSS theme (light/dark) as the rest of the site, with a theme toggle control, rather than hardcoding a single theme.

#### Scenario: Theme consistency
- **WHEN** a visitor has selected "dark" theme on the app or login page
- **THEN** the public events page renders in dark theme on load, with a toggle to switch to light

#### Scenario: Theme toggle on events page
- **WHEN** a visitor clicks the theme toggle on the events page
- **THEN** the page switches between light and dark theme and persists the choice in localStorage
