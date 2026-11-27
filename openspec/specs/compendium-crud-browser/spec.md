# compendium-crud-browser Specification

## Purpose
Unified admin browser for all compendium schemas and entries with search, filter, sort, and pagination.

## Requirements

### Requirement: Unified schema listing
The system SHALL display all compendium schemas (both built-in and user-created) in a single browsable view. Each schema row SHALL show display name, type name (slug), field count, entry count, and available actions (browse entries, edit schema, delete schema).

#### Scenario: Admin views all schemas
- **WHEN** an admin navigates to the compendium section
- **THEN** the system loads and displays all schemas from `GET /api/admin/compendium-schemas` with entry counts

#### Scenario: Schema list shows correct entry counts
- **WHEN** a schema has entries in `compendium_entries`
- **THEN** the entry count column shows the exact count from the API response

#### Scenario: Schema row actions are functional
- **WHEN** an admin clicks "Browse Entries" on a schema row
- **THEN** the entry browser view loads for that schema showing the first page of entries

### Requirement: Schema filtering and search
The system SHALL provide a search bar at the top of the schema list that filters schemas by display name or type name. The search SHALL be client-side for the schema list (since schemas are few) and server-side for entries.

#### Scenario: Filter schemas by name
- **WHEN** an admin types "magic" in the schema search bar
- **THEN** only schemas with "magic" in their display name or type name are shown

### Requirement: Paginated entry browser
When browsing entries within a schema, the system SHALL display entries in a paginated table with columns derived from the schema's field definitions. The first 3–5 fields SHALL be shown as columns; remaining fields are accessible via an expand/collapse or detail view. Pagination controls SHALL allow navigating between pages.

#### Scenario: Entry table shows schema-appropriate columns
- **WHEN** an admin browses entries for a "Spells" schema with fields [name, level, school, description]
- **THEN** the table shows columns for Name, Level, School, and Description (plus actions)

#### Scenario: Pagination works with many entries
- **WHEN** a schema has 150 entries and page size is 50
- **THEN** the browser shows 3 pages with navigation controls (prev/next, page numbers)

### Requirement: Row-level entry actions
Each entry row SHALL have action buttons for: edit (opens entry editor modal), duplicate (creates copy with modified name), view details (shows all fields), and delete (with confirmation).

#### Scenario: Delete entry with confirmation
- **WHEN** an admin clicks delete on an entry row and confirms the dialog
- **THEN** a DELETE request is sent to `/api/admin/compendium-entries/{id}` and the entry is removed from the table

#### Scenario: Duplicate entry
- **WHEN** an admin clicks duplicate on an entry
- **THEN** a copy is created via POST to `/api/admin/compendium-schemas/{schemaId}/entries` with the same data and name appended with " (copy)"
