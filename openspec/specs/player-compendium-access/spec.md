# player-compendium-access Specification

## Purpose
Currently, compendium items (monsters, races, spells, etc.) and dynamic compendium entries are only accessible to DMs and admins.

## Requirements

### Requirement: Public read-only compendium API
The system SHALL expose GET endpoints under `/api/compendium/` for all authenticated users to list schemas and browse entries.

#### Scenario: List schemas
- **WHEN** an authenticated user calls GET `/api/compendium/schemas`
- **THEN** the response SHALL include all compendium schema definitions (type_name, display_name, fields)
- **AND** the response SHALL exclude entry count and admin-only metadata

#### Scenario: List entries by schema
- **WHEN** an authenticated user calls GET `/api/compendium/schemas/:id/entries`
- **THEN** the response SHALL include paginated entries for that schema
- **AND** the response SHALL include only the data field (not internal IDs or timestamps)

#### Scenario: Get single entry
- **WHEN** an authenticated user calls GET `/api/compendium/schemas/:schemaId/entries/:entryId`
- **THEN** the response SHALL include the full entry data

#### Scenario: Search entries
- **WHEN** an authenticated user calls GET `/api/compendium/search?q=<query>`
- **THEN** the response SHALL include matching entries across all schemas

### Requirement: Frontend compendium browser
The system SHALL provide a compendium browser page at `/compendium/` accessible to all authenticated users.

#### Scenario: Browse compendium categories
- **WHEN** an authenticated user visits `/compendium/`
- **THEN** they SHALL see a list of compendium categories (schemas)
- **AND** clicking a category SHALL show entries in that category

#### Scenario: View entry details
- **WHEN** a player clicks a compendium entry
- **THEN** they SHALL see a detail view with all fields rendered using schema-driven display

### Requirement: Link compendium entries in acts and scenes
The system SHALL support linking to compendium entries from act and scene content using `[[compendium:<schema>:<name>]]` syntax.

#### Scenario: Render compendium link in scene
- **WHEN** an act or scene description contains `[[compendium:monster:goblin]]`
- **THEN** the rendered content SHALL display a clickable link
- **AND** clicking the link SHALL navigate to the compendium entry detail view

#### Scenario: Link resolves by name
- **WHEN** a compendium link uses an entry name
- **THEN** the system SHALL resolve the entry by name within the specified schema
