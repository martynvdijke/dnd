# per-campaign-events-pages Specification

## Purpose
Public, shareable per-campaign event pages with unique slugs, each backed by its own Google Calendar (or filtered view of a shared calendar), with independent caching and ICS export.

## Requirements

### Requirement: Campaign event pages exist at /events/c/{slug}

The system SHALL serve public campaign-scoped event pages at `/events/c/{slug}` where `{slug}` matches `[a-z0-9-]+`. Each campaign page SHALL display upcoming events filtered by that campaign's configured calendar ID, tags, color labels, and filter mode. The page SHALL NOT require authentication. Unknown or disabled campaign slugs SHALL return HTTP 404.

#### Scenario: Campaign page loads successfully
- **WHEN** a user visits `/events/c/lost-mines`
- **AND** a campaign with slug "lost-mines" exists and is enabled
- **THEN** the server returns HTTP 200 with an HTML page showing that campaign's upcoming events
- **AND** the page title includes the campaign's display name

#### Scenario: Unknown campaign slug returns 404
- **WHEN** a user visits `/events/c/nonexistent`
- **AND** no campaign with slug "nonexistent" exists
- **THEN** the server returns HTTP 404

#### Scenario: Disabled campaign returns 404
- **WHEN** a user visits `/events/c/old-campaign`
- **AND** a campaign with slug "old-campaign" exists but is disabled
- **THEN** the server returns HTTP 404

#### Scenario: Campaign page with no events
- **WHEN** a user visits an enabled campaign page
- **AND** there are no upcoming events for that campaign
- **THEN** the page displays "No upcoming events found"

---

### Requirement: Campaign settings are managed via admin UI

The system SHALL provide admin CRUD operations for campaign event settings. An admin SHALL be able to create, edit, delete, enable, and disable campaigns. Each campaign SHALL have a unique slug, a display name, a calendar ID, tags, color labels, a filter mode, a cache TTL, an active flag, AND Google Calendar auth configuration (auth method, and either service-account credentials JSON or OAuth client ID/secret/refresh token). The campaign add/edit modal SHALL expose ALL of these fields with parity to the global events settings card. Auth and credential fields SHALL be optional per-campaign overrides; when left empty, the campaign SHALL inherit the corresponding global events settings. Slug uniqueness SHALL be enforced. The admin events settings section SHALL be organized into clear, grouped sub-sections (public sharing, calendar source, filtering, Google auth) with consistent labeling and helpful hints for usability.

#### Scenario: Admin creates a campaign
- **WHEN** an admin submits a new campaign with display name "Lost Mines" and slug "lost-mines"
- **THEN** the campaign is persisted to the database
- **AND** the campaign appears in the admin campaign list

#### Scenario: Slug must be unique
- **WHEN** an admin creates a campaign with slug "lost-mines"
- **AND** a campaign with that slug already exists
- **THEN** the system rejects the creation with an error message

#### Scenario: Admin edits a campaign
- **WHEN** an admin updates the calendar ID for an existing campaign
- **THEN** the change is persisted
- **AND** that campaign's cache is invalidated

#### Scenario: Admin deletes a campaign
- **WHEN** an admin deletes a campaign and confirms
- **THEN** the campaign is removed from the database
- **AND** its cached events are removed
- **AND** its public page returns 404

#### Scenario: Admin disables a campaign
- **WHEN** an admin sets a campaign's enabled flag to false
- **THEN** the campaign's public page returns 404
- **AND** the campaign remains in the admin list for re-enabling

#### Scenario: Admin sets campaign color labels and filter mode
- **WHEN** an admin opens the campaign edit modal
- **AND** sets the filter mode to "color"
- **AND** enters color labels "1,5"
- **AND** saves
- **THEN** the color labels and filter mode are persisted for that campaign
- **AND** the campaign's public page filters events by those color labels

#### Scenario: Admin configures campaign-specific Google auth
- **WHEN** an admin opens the campaign edit modal
- **AND** selects auth method "service_account"
- **AND** pastes service-account credentials JSON
- **AND** saves
- **THEN** the credentials are persisted for that campaign
- **AND** the campaign's event fetches authenticate using those credentials instead of the global ones

#### Scenario: Campaign inherits global auth when left empty
- **WHEN** an admin creates a campaign without filling in auth method or credentials
- **AND** saves
- **THEN** the campaign uses the global events auth settings for fetching
- **AND** no per-campaign credentials are stored

#### Scenario: Campaign modal shows credential fields based on auth method
- **WHEN** an admin selects auth method "oauth" in the campaign modal
- **THEN** the OAuth client ID, client secret, and refresh token fields are shown
- **AND** the service-account credentials JSON field is hidden
- **WHEN** an admin selects auth method "service_account"
- **THEN** the service-account credentials JSON field is shown
- **AND** the OAuth fields are hidden

#### Scenario: Admin events settings are neatly grouped
- **WHEN** an admin views the events settings section
- **THEN** the settings are organized into clearly labeled groups (public sharing, calendar source, filtering, Google auth)
- **AND** helper hints are displayed under fields that benefit from explanation
- **AND** the layout is consistent between the global settings card and the campaign modal
### Requirement: Campaign pages have independent caching

The system SHALL cache events per-campaign, keyed by campaign slug. Each campaign's cache SHALL be independent of the global cache and other campaigns' caches. The cache TTL SHALL be the global configured value. Clearing the global cache SHALL also clear all campaign caches.

#### Scenario: Campaign cache is independent
- **WHEN** campaign "lost-mines" has fresh cached events
- **AND** campaign "curse-of-strahd" has an expired cache
- **THEN** a request to "lost-mines" serves from cache without fetching
- **AND** a request to "curse-of-strahd" fetches fresh events from Google Calendar

#### Scenario: Clearing global cache clears all campaigns
- **WHEN** an admin clears the events cache
- **THEN** the global cache and all campaign caches are cleared

#### Scenario: Per-campaign cache clear
- **WHEN** an admin clears the cache for a specific campaign
- **THEN** only that campaign's cache is cleared
- **AND** the global cache and other campaigns' caches are unaffected

---

### Requirement: Campaign-scoped ICS feed

The system SHALL provide a public ICS feed at `/events/c/{slug}/ical` returning events for that campaign, respecting the campaign's filter settings and cache. The ICS format SHALL match the global ICS feed format.

#### Scenario: Campaign ICS returns valid ICS
- **WHEN** a client requests `GET /events/c/lost-mines/ical`
- **AND** the campaign exists and is enabled
- **THEN** the response has `Content-Type: text/calendar`
- **AND** the response body is valid ICS with that campaign's events

#### Scenario: Campaign ICS for unknown slug returns 404
- **WHEN** a client requests `GET /events/c/nonexistent/ical`
- **THEN** the response is HTTP 404

#### Scenario: Campaign ICS respects campaign filters
- **WHEN** campaign "lost-mines" is configured with tags "dnd"
- **THEN** only events matching those tags appear in the campaign's ICS feed
