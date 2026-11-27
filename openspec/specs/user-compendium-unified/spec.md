# user-compendium-unified Specification

## Purpose
User-facing compendium page that fetches and displays compendium entries from the unified schema-based system (`compendium_entries`), organized by schema type, alongside legacy data. Supports dynamic schema types beyond the hardcoded 5 legacy tabs.

## Requirements

### Requirement: User-facing compendium displays schema-based entries

The system SHALL expose schema-based compendium entries to authenticated users in a read-only view, organized by schema type.

#### Scenario: Auth user views schema entries
- **WHEN** an authenticated user navigates to the compendium page
- **THEN** the system displays legacy tab entries (from legacy tables) AND dynamic tabs for each schema type that has entries in `compendium_entries`
- **AND** dynamic tabs are only shown when at least one schema type has entries

#### Scenario: No schema entries exist
- **WHEN** there are zero rows in `compendium_entries`
- **THEN** the dynamic tabs section is hidden and the compendium page shows only the 5 legacy tabs

#### Scenario: Schema type has entries
- **WHEN** a schema type has one or more entries
- **THEN** a tab is shown with that schema's `display_name`
- **AND** the tab lists up to 20 entries with name and key attributes
- **AND** a "View All N entries" link is provided to browse the full list

### Requirement: Backend endpoint for user-scoped schema entries

The system SHALL provide a `GET /api/compendium/entries-by-schema` endpoint accessible to any authenticated user (auth group) that returns all schemas with their entry counts and first 20 entries.

#### Scenario: Successful response
- **WHEN** an authenticated user calls `GET /api/compendium/entries-by-schema`
- **THEN** the endpoint returns a JSON object with a `schemas` array
- **AND** each schema object contains `id`, `type_name`, `display_name`, `entry_count`, and `entries` array
- **AND** schemas with 0 entries are excluded from the response

#### Scenario: Unauthenticated request
- **WHEN** an unauthenticated user calls `GET /api/compendium/entries-by-schema`
- **THEN** the endpoint returns a 401 Unauthorized response

#### Scenario: Each entry shape
- **WHEN** entries are returned
- **THEN** each entry object contains `id`, `data` (full JSON data map), and `created_at`

### Requirement: Dynamic compendium tabs show schema entries

The frontend SHALL query the schema entries endpoint and render a dynamic nav tab row below the legacy tabs for each schema type with entries.

#### Scenario: Dynamic tabs rendered
- **WHEN** the compendium page loads and the schema entries response contains schemas
- **THEN** a new row of nav tabs/nav-pills is rendered below the legacy tabs
- **AND** clicking a dynamic tab shows a card/list view of entries from that schema

#### Scenario: Dynamic tab switches content
- **WHEN** a user clicks a dynamic tab
- **THEN** the current content area shows entries for that schema type
- **AND** each entry displays its name (from the `data.name` field) and a short preview

#### Scenario: View all entries link
- **WHEN** a dynamic tab has more than 20 entries
- **THEN** the tab shows the first 20 entries
- **AND** a "View All N entries" link is displayed below the list
- **AND** clicking the link navigates to the full entry browser for that schema
