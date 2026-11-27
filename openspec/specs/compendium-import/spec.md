# compendium-import Specification

## Purpose
Multi-JSON file import with format autodetection, validation, duplicate resolution, dry-run preview, and import summary reports.

## Requirements

### Requirement: Upload multiple JSON files for import
The system SHALL accept one or more JSON files uploaded simultaneously. Each file may contain a single object or an array of objects.

#### Scenario: Upload single JSON file with array of entries
- **WHEN** a user uploads a JSON file containing an array of compendium entries
- **THEN** the system parses the file, validates entries, and presents a dry-run preview showing how many entries would be created, updated, or skipped

#### Scenario: Upload single JSON file with single entry
- **WHEN** a user uploads a JSON file containing a single object
- **THEN** the system wraps the object in an array and processes it as a single-entry import

#### Scenario: Upload multiple files simultaneously
- **WHEN** a user selects multiple JSON files for upload
- **THEN** each file is processed independently; results are aggregated into a single import report

### Requirement: Automatic format and schema detection
The system SHALL attempt to auto-detect which compendium schema each uploaded JSON entry belongs to, based on field name matching against known schemas. Users may override the detected schema.

#### Scenario: Auto-detect schema from fields
- **WHEN** uploaded JSON entries contain fields that match a known schema's field names
- **THEN** the system auto-selects that schema and shows the match confidence

#### Scenario: Manual schema assignment
- **WHEN** auto-detection is uncertain or incorrect
- **THEN** the user can manually assign a schema from a dropdown before confirming the import

### Requirement: Field mapping and validation
The system SHALL present a field mapping interface that lets users map JSON fields to schema fields, with validation feedback.

#### Scenario: Auto-map fields
- **WHEN** JSON field names match schema field names (case-insensitive)
- **THEN** the system auto-maps them and highlights successful mappings

#### Scenario: Manual field mapping
- **WHEN** JSON field names do not match schema fields
- **THEN** the user can manually map each JSON field to a schema field or skip it

#### Scenario: Validate entries before import
- **WHEN** the user requests validation
- **THEN** the system validates all entries against the schema (required fields, type constraints) and reports errors per entry

### Requirement: Deduplication detection
The system SHALL detect duplicate entries by matching a "name" or configured unique field, and let users choose how to resolve duplicates.

#### Scenario: Detect duplicate by name
- **WHEN** an imported entry has the same name as an existing entry in the target schema
- **THEN** the system flags it as a duplicate and shows the existing entry for comparison

#### Scenario: Duplicate resolution options
- **WHEN** duplicates are detected
- **THEN** the user can choose: skip, overwrite, create-as-new, or apply-to-all for each duplicate

### Requirement: Import dry-run and confirmation
The system SHALL provide a dry-run preview showing the full import plan before executing.

#### Scenario: Dry-run preview
- **WHEN** files are uploaded and mapped
- **THEN** the system shows a summary table: entries to create, update, skip, and errors per file

#### Scenario: Confirm and execute import
- **WHEN** the user confirms the import
- **THEN** the system executes the import and shows a completion report with counts of created, updated, skipped, and errored entries

### Requirement: Import history and rollback
The system SHALL log all imports and support rolling back the last import.

#### Scenario: View import history
- **WHEN** a user requests import history
- **THEN** the system returns a list of past imports with timestamp, file count, entry counts, and status

#### Scenario: Rollback import
- **WHEN** a user rolls back an import
- **THEN** the system deletes all entries created by that import and restores any entries that were overwritten
