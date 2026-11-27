# compendium-cross-search Specification

## Purpose
Full-text search across all schema entries with schema-aware filters.

## Requirements

### Requirement: Global search across all schemas
The compendium browser SHALL provide a global search bar at the top of the schema list view. When a query is entered, the system SHALL perform an FTS5 search across ALL compendium entries (via `/api/admin/compendium-search`) and display results grouped by schema type.

#### Scenario: Global search finds matching entries
- **WHEN** an admin types "fire" in the global search bar
- **THEN** the system searches via `GET /api/admin/compendium-search?q=fire` and displays results grouped by schema (e.g., "Spells: Fireball, Fire Bolt" under Spells, "Monsters: Fire Elemental" under Monsters)

#### Scenario: Global search with no results
- **WHEN** an admin types "zzznotfound" and no entries match
- **THEN** the system displays "No results found for 'zzznotfound'" with a suggestion to try different terms

### Requirement: Schema-type filter facet
The search results SHALL include a filter facet showing which schema types have matching results. The admin MAY filter results to a specific schema type by clicking the facet.

#### Scenario: Filter search by schema type
- **WHEN** a global search returns results from 3 schemas and the admin clicks "Spells" facet
- **THEN** only spell results are shown

### Requirement: Search result links to entry editor
Each search result item SHALL be a clickable link that opens the entry editor modal for that entry, allowing the admin to view or edit the matched entry directly from search results.

#### Scenario: Navigate from search to entry editor
- **WHEN** an admin clicks a search result for "Fireball"
- **THEN** the entry editor modal opens for the Fireball entry with all fields populated

### Requirement: Search within a specific schema
The entry browser view SHALL have a schema-scoped search bar (separate from the global search) that filters entries within the currently selected schema. This SHALL use the `q` query parameter on `GET /api/admin/compendium-schemas/{id}/entries`.

#### Scenario: Schema-scoped search returns matching entries
- **WHEN** an admin is browsing "Spells" schema and types "heal" in the schema search bar
- **THEN** the system shows only spells with "heal" in their FTS5-indexed content
