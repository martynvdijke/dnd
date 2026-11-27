# universal-entity-search Specification

## Purpose
Full-text search across all entity types via FTS5 with type filtering, keyboard-first palette UI, and consistent navigable results.

## Requirements

### Requirement: Full-text index covers all searchable entity types
The system SHALL maintain a unified SQLite FTS5 search index containing, at minimum: characters, NPCs, notes, quests, journal entries, sessions, campaigns, locations, encounters, custom monsters, shops, factions, one-shot adventures, wiki pages, homebrew entries, timeline events, items, and compendium entries. The index SHALL be updated on create, update, and delete of each indexed entity, and SHALL store entity type, entity ID, title, subtitle, and searchable body text.

#### Scenario: Newly created entity is findable
- **WHEN** a user creates a location named "Emberhold Tavern"
- **THEN** a global search for "Emberhold" returns the location in the results with type "location"

#### Scenario: Deleted entity disappears from results
- **WHEN** a user deletes an NPC and then searches for that NPC's name
- **THEN** the NPC does not appear in the search results

#### Scenario: Index backfill for existing data
- **WHEN** the migration introducing the unified index runs on a database with existing content
- **THEN** all pre-existing searchable entities are present in the index and findable via search

### Requirement: Global search endpoint with type filtering
The system SHALL provide an authenticated search endpoint accepting a query string and optional entity-type filters, returning ranked results grouped by entity type, each result including entity type, entity ID, title, subtitle, and a matched-text snippet. Results SHALL only include entities the requesting user is permitted to view.

#### Scenario: Unfiltered search returns grouped results
- **WHEN** a user searches for "dragon" without type filters
- **THEN** the response contains matching entities grouped by type with ranked ordering within each group

#### Scenario: Type-filtered search
- **WHEN** a user searches for "dragon" with the filter limited to "monsters"
- **THEN** only monster results are returned

#### Scenario: Permission scoping
- **WHEN** a player-role user searches for a term that only matches entities in a campaign they are not a member of
- **THEN** those entities are not present in the results

### Requirement: Keyboard-first command palette UI
The system SHALL provide a command palette opened via Cmd/Ctrl+K (and `/` outside text inputs) that queries the global search endpoint with debouncing, supports entity-type filter chips, full keyboard navigation (arrow keys, Enter to open, Esc to close), and shows recent entities when the query is empty.

#### Scenario: Open palette and navigate by keyboard
- **WHEN** the user presses Cmd+K, types "ember", presses ArrowDown then Enter
- **THEN** the highlighted result's entity page is opened via its stable URL

#### Scenario: Filter results by type chip
- **WHEN** the user opens the palette, types a query, and activates the "NPCs" filter chip
- **THEN** only NPC results are displayed

#### Scenario: Empty query shows recents
- **WHEN** the user opens the palette without typing
- **THEN** a list of recently viewed entities is displayed

### Requirement: Search results navigate to deep links
Each search result SHALL navigate to the target entity's stable deep-link URL when activated.

#### Scenario: Result opens entity detail
- **WHEN** a user activates a search result for a quest
- **THEN** the application navigates to that quest's deep-link URL and renders the quest detail view
