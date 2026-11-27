# deep-link-navigation Specification

## Purpose
Client-side hash routing giving every view and entity a stable, shareable, refresh-safe URL.

## Requirements

### Requirement: Stable hash-based URLs for views and entities
The system SHALL assign a stable hash-based URL to every primary view and entity detail (at minimum: characters, campaigns, NPCs, locations, quests, adventures/one-shots, encounters, monsters, shops, factions, journal, timeline, wiki pages, compendium entries). URLs SHALL follow the pattern `#/<type>/<id>` (nested where owned, e.g. `#/campaigns/<id>/oneshots/<id>`).

#### Scenario: Navigating to an entity updates the URL
- **WHEN** a user opens a character's detail view
- **THEN** the browser URL becomes `#/characters/<id>` for that character

#### Scenario: View switch updates the URL
- **WHEN** a user switches to the timeline view
- **THEN** the browser URL reflects the timeline route

### Requirement: URL restores application state
The system SHALL restore the corresponding view and entity detail when a deep-link URL is loaded directly, refreshed, or opened in a new tab, for any authenticated user permitted to view that entity.

#### Scenario: Refresh restores the entity view
- **WHEN** a user viewing `#/npcs/<id>` refreshes the browser
- **THEN** the application renders that NPC's detail view after reload

#### Scenario: Back/forward navigation works
- **WHEN** a user navigates from a quest to an NPC and then presses the browser Back button
- **THEN** the application returns to the quest detail view

#### Scenario: Unauthenticated deep link redirects to login and back
- **WHEN** an unauthenticated user opens a deep-link URL, logs in
- **THEN** the application navigates to the originally requested deep link

### Requirement: In-app navigation entry points use deep links
Search results, mention chips, backlink panels, and linked-entity lists SHALL navigate via deep-link URLs, and unknown or malformed hashes SHALL fall back to the default view without error.

#### Scenario: Unknown hash falls back safely
- **WHEN** a user opens `#/does-not-exist`
- **THEN** the application renders the default view without an error state
