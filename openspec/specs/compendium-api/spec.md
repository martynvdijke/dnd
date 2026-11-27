# compendium-api Specification

## Purpose
Unified REST API for all compendium entry CRUD with pagination, filtering, sorting, and search.

## Requirements

### Requirement: Unified CRUD API for compendium entries
The system SHALL provide a REST API under /api/compendium/entries/:type that supports create, read, update, delete, list, and search operations for any compendium type.

#### Scenario: List entries with pagination
- **WHEN** a user requests GET /api/compendium/entries/:type with optional page and page_size query params
- **THEN** the system returns a paginated list of entries for that compendium type, including total count, page info, and entry data

#### Scenario: Create entry
- **WHEN** a user POSTs valid entry data to /api/compendium/entries/:type
- **THEN** the system creates the entry, validates against the schema, and returns the created entry with its ID

#### Scenario: Get single entry
- **WHEN** a user requests GET /api/compendium/entries/:type/:id
- **THEN** the system returns the full entry data

#### Scenario: Update entry
- **WHEN** a user PUTs updated data to /api/compendium/entries/:type/:id
- **THEN** the system updates the entry and returns the updated entry

#### Scenario: Delete entry
- **WHEN** a user DELETEs /api/compendium/entries/:type/:id
- **THEN** the system removes the entry

#### Scenario: Delete multiple entries (bulk)
- **WHEN** a user DELETEs /api/compendium/entries/:type with a body containing an array of IDs
- **THEN** the system deletes all specified entries and returns the count deleted

### Requirement: Search and filter entries
The system SHALL support full-text search across all schemas and field-level filtering on any schema field.

#### Scenario: Full-text search
- **WHEN** a user queries ?q=searchterm on any entries endpoint
- **THEN** the system performs a full-text search across all non-JSON fields and returns matching entries

#### Scenario: Field-level filter
- **WHEN** a user queries ?field_name=value on an entries endpoint
- **THEN** the system filters entries where field_name contains the given value

#### Scenario: Combined search and filter
- **WHEN** a user provides both ?q=term and ?field=value
- **THEN** the system applies both filters with AND logic

### Requirement: Cross-type search
The system SHALL support searching across all compendium types at once.

#### Scenario: Global compendium search
- **WHEN** a user requests GET /api/compendium/search?q=fireball
- **THEN** the system returns results from all compendium types, grouped by type, with entry name, type, and a snippet of matching content
