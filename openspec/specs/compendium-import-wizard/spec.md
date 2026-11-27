# compendium-import-wizard Specification

## Purpose
Enhanced JSON import with drag-drop, URL fetch, paste, field mapping auto-detection, and duplicate resolution.

## Requirements

### Requirement: Inline import trigger
The compendium schema list view SHALL have an "Import" button per schema row. Clicking it SHALL open the import wizard flow inline within the compendium browser, pre-selecting the target schema.

#### Scenario: Import button opens wizard for selected schema
- **WHEN** an admin clicks "Import" on the "Magic Items" schema
- **THEN** the import wizard opens with "Magic Items" pre-selected as the target schema

### Requirement: Multi-source JSON import
The import wizard SHALL support three data source methods:
1. **File upload** — drag-and-drop or click-to-browse for `.json` files
2. **Paste JSON** — textarea for pasting JSON content
3. **Fetch URL** — URL input that fetches JSON from an external URL

#### Scenario: Import via file upload
- **WHEN** an admin drops a `.json` file onto the import drop zone
- **THEN** the file is read and parsed; a preview table shows the first 5 records with field count

#### Scenario: Import via URL fetch
- **WHEN** an admin enters a URL like `https://raw.githubusercontent.com/.../data.json` and clicks Fetch
- **THEN** the system fetches the URL, parses JSON, and shows a preview

### Requirement: Auto-detect field mapping
After loading JSON data, the system SHALL call `POST /api/admin/compendium-schemas/{id}/import/detect` to auto-detect field mappings between the JSON fields and the schema fields. The detection results SHALL be displayed in a mapping table where each row shows the source JSON field, the auto-suggested target schema field, and a confidence indicator.

#### Scenario: Auto-mapping detects matching fields
- **WHEN** imported JSON has fields "Name", "Description", "Level" and the schema has fields "name", "description", "level"
- **THEN** the mapping table shows each paired with "high" confidence and auto-selected

#### Scenario: Admin can override field mappings
- **WHEN** auto-mapping suggests "Name → name" but the admin wants "Name → title"
- **THEN** the admin can change the target field via a dropdown selector

### Requirement: Duplicate resolution strategy
The import wizard SHALL show detected duplicates (entries with the same "name" field as existing entries) and let the admin choose a resolution strategy per duplicate: skip, overwrite, or create-as-new. A default strategy SHALL be configurable in the import options.

#### Scenario: Duplicate detection during import
- **WHEN** imported JSON has an entry with name "Fireball" and an existing entry with that name already exists in the schema
- **THEN** the wizard shows the duplicate with the incoming data, existing data, and strategy selector (skip/overwrite/create-new)

### Requirement: Import preview and confirmation
Before final import, the wizard SHALL show a summary: total records, new records, duplicates to skip, duplicates to overwrite, and validation errors. The admin MUST confirm before the import executes.

#### Scenario: Import summary shows before execution
- **WHEN** after resolving duplicates, the admin clicks "Review"
- **THEN** a summary card shows: "15 total, 12 new, 2 duplicates (skip), 1 error" with a "Confirm Import" button

### Requirement: Import progress and results
During import, the system SHALL show a progress bar. After completion, the system SHALL display results: count inserted, skipped, overwritten, errors (if any). The entry browser SHALL refresh to show new entries.

#### Scenario: Import completes successfully
- **WHEN** an admin confirms import of 10 entries
- **THEN** a progress bar fills to 100%, and the results show "10 inserted, 0 errors"
