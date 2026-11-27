# compendium-export Specification

## Purpose
Export compendium entries (by type, source, system, search query) to JSON.

## Requirements

### Requirement: Export compendium entries to JSON
The system SHALL allow exporting compendium entries as downloadable JSON files, with filtering options.

#### Scenario: Export all entries of a type
- **WHEN** a user requests GET /api/compendium/export/:type
- **THEN** the system returns a JSON file containing all entries of that compendium type

#### Scenario: Export entries with filters
- **WHEN** a user requests GET /api/compendium/export/:type?source=homebrew&system=pf2e
- **THEN** the system returns a JSON file containing only entries matching the filter criteria

#### Scenario: Export search results
- **WHEN** a user requests GET /api/compendium/export/:type?q=fire
- **THEN** the system returns a JSON file containing entries matching the search query

### Requirement: Export selected entries
The system SHALL support exporting a specific set of entries by their IDs.

#### Scenario: Export by ID list
- **WHEN** a user POSTs /api/compendium/export with a body containing an array of entry IDs with their types
- **THEN** the system returns a JSON file containing only the specified entries

### Requirement: Export file format
Exported JSON SHALL include metadata: export timestamp, source system version, schema version, and entry count.

#### Scenario: Export includes metadata
- **WHEN** a user downloads an export file
- **THEN** the JSON includes a top-level "metadata" object with timestamp, export_type, count, and schema info, plus an "entries" array
