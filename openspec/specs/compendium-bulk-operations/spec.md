# compendium-bulk-operations Specification

## Purpose
Multi-select batch edit, batch delete, and batch export across entries.

## Requirements

### Requirement: Multi-select entry checkboxes
Each entry row in the entry browser SHALL have a checkbox. A "select all" checkbox in the table header SHALL toggle all visible entries. The selection state SHALL be tracked client-side. Page changes SHALL NOT clear existing selections.

#### Scenario: Select all visible entries
- **WHEN** an admin checks the "Select All" checkbox
- **THEN** all visible entry rows are checked and the selection count shows the total across all pages?

#### Scenario: Selection persists across page navigation
- **WHEN** an admin selects entries on page 1, navigates to page 2, then navigates back to page 1
- **THEN** the previously selected entries on page 1 remain checked

### Requirement: Floating bulk action toolbar
When one or more entries are selected, a floating toolbar SHALL appear showing the count of selected items and action buttons: "Delete Selected", "Edit Selected", "Export Selected".

#### Scenario: Bulk toolbar appears with selection
- **WHEN** an admin selects 3 entries
- **THEN** a toolbar appears showing "3 selected" with Delete, Edit, and Export buttons

#### Scenario: Bulk toolbar hides when no selection
- **WHEN** an admin deselects all entries
- **THEN** the bulk toolbar disappears

### Requirement: Batch delete
The "Delete Selected" button SHALL prompt a confirmation dialog showing the count of entries to delete. On confirmation, the system SHALL send a POST to `/api/admin/compendium-entries/batch-delete` with the selected entry IDs. The table SHALL refresh after deletion.

#### Scenario: Batch delete with confirmation
- **WHEN** an admin selects 5 entries and clicks "Delete Selected" and confirms
- **THEN** a POST to `/api/admin/compendium-entries/batch-delete` with `{ids: [1,2,3,4,5]}` is sent, the entries are removed, and the table refreshes

### Requirement: Batch edit
The "Edit Selected" button SHALL open a modal with a field selector (dropdown of schema fields) and a value input. On submission, the system SHALL send a POST to `/api/admin/compendium-entries/batch-update` with the selected entry IDs and the field-value pair.

#### Scenario: Batch edit sets a field value
- **WHEN** an admin selects 3 entries, chooses the "source" field, enters "homebrew", and clicks Update
- **THEN** a POST to `/api/admin/compendium-entries/batch-update` is sent with `{ids: [1,2,3], data: {source: "homebrew"}}` and 3 entries are updated

### Requirement: Batch export
The "Export Selected" button SHALL download a JSON file containing the selected entries with their schema metadata. The export SHALL use the existing `/api/admin/compendium-schemas/{id}/export?format=json` endpoint but filtered to only the selected entry IDs.

#### Scenario: Export selected entries as JSON
- **WHEN** an admin selects 10 entries and clicks "Export Selected"
- **THEN** a JSON file is downloaded containing the 10 entries with schema metadata

### Requirement: Bulk delete compendium entries
The system SHALL allow deleting multiple compendium entries in a single API call by providing an array of entry IDs.

#### Scenario: Bulk delete selected entries
- **WHEN** user selects 3 entries in the entry browser and clicks "Delete Selected"
- **THEN** the frontend SHALL show a confirmation dialog with the count (3)
- **WHEN** user confirms the deletion
- **THEN** the frontend SHALL send `POST /api/admin/compendium-entries/batch-delete` with `{ ids: [1, 2, 3] }`
- **THEN** the backend SHALL delete those entries from `compendium_entries` table
- **THEN** the backend SHALL return `{ deleted: 3 }`
- **THEN** the frontend SHALL refresh the entry list

#### Scenario: Bulk delete with invalid IDs
- **WHEN** user sends batch-delete with IDs that include non-existent entries
- **THEN** the backend SHALL delete only the entries that exist
- **THEN** the backend SHALL return `{ deleted: <count>, errors: [...] }` for invalid IDs
- **THEN** the frontend SHALL show a toast with the deletion summary

### Requirement: Bulk update compendium entries
The system SHALL allow updating a specific field on multiple compendium entries in a single API call.

#### Scenario: Bulk update field on selected entries
- **WHEN** user selects 5 entries and clicks "Edit Selected"
- **THEN** the frontend SHALL open a modal showing a field selector (dropdown of schema fields) and a value input
- **WHEN** user picks a field, enters a value, and confirms
- **THEN** the frontend SHALL send `POST /api/admin/compendium-entries/batch-update` with `{ ids: [1,2,3,4,5], data: { "field_name": "new_value" } }`
- **THEN** the backend SHALL update the specified field in each entry's data JSON
- **THEN** the backend SHALL return `{ updated: 5 }`
- **THEN** the frontend SHALL refresh the entry list

#### Scenario: Bulk update with empty field value
- **WHEN** user sends batch-update with a field set to empty string
- **THEN** the backend SHALL still update the field (allow clearing a field)
- **THEN** the backend SHALL return `{ updated: <count> }`

#### Scenario: Bulk update with non-existent entries
- **WHEN** user sends batch-update with IDs that include non-existent entries
- **THEN** the backend SHALL update only the entries that exist
- **THEN** the backend SHALL return `{ updated: <count>, errors: [...] }` for invalid IDs
