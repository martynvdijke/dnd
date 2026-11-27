# compendium-ui Specification

## Purpose
Admin HTMX UI for browsing, searching, creating, editing, deleting, and bulk-managing compendium entries. Includes a schema editor for custom types.

## Requirements

### Requirement: Compendium manager admin page
The system SHALL provide an HTMX-powered admin page at /admin/compendium for managing all compendium content.

#### Scenario: Navigate to compendium manager
- **WHEN** an admin clicks "Compendium Manager" in the admin nav
- **THEN** the page loads showing a schema type selector, search bar, and entry listing

### Requirement: Schema management UI
The system SHALL provide a schema editor UI for viewing, creating, editing, and deleting compendium schemas.

#### Scenario: View schemas list
- **WHEN** an admin opens the schema management tab
- **THEN** a list of all schemas is displayed with type name, field count, entry count, and actions column

#### Scenario: Create new schema
- **WHEN** an admin clicks "New Schema" and fills in the type name, display name, and field definitions
- **THEN** the schema is saved and becomes available for creating entries

#### Scenario: Add fields to schema
- **WHEN** an admin adds fields with name, label, type (string, text, integer, float, boolean, json, select, multi-select), required flag, and options (for select types)
- **THEN** fields are added to the schema definition

#### Scenario: Edit existing schema
- **WHEN** an admin edits a schema's field definitions
- **THEN** the schema is updated in place

### Requirement: Entry browser UI
The system SHALL provide a browse interface for viewing, searching, filtering, and paginating compendium entries.

#### Scenario: Browse entries by type
- **WHEN** an admin selects a compendium type from the type selector
- **THEN** entries are loaded via HTMX with pagination, showing a table of entries with key fields as columns

#### Scenario: Search entries
- **WHEN** an admin types a search query
- **THEN** results are fetched via HTMX and the entry table updates in real-time

#### Scenario: Filter by field value
- **WHEN** an admin applies a field filter
- **THEN** entries are filtered accordingly, with filter controls generated from the schema field definitions

#### Scenario: Paginate through results
- **WHEN** an admin clicks a pagination control
- **THEN** the next/previous page of entries loads via HTMX

### Requirement: Entry CRUD forms
The system SHALL provide forms for creating and editing compendium entries, with form fields generated dynamically from the schema.

#### Scenario: Create entry form
- **WHEN** an admin clicks "New Entry" for a selected schema
- **THEN** a modal or inline form is shown with input fields matching the schema definition, including type-appropriate inputs (text, number, textarea, checkbox, select, JSON editor)

#### Scenario: Edit entry form
- **WHEN** an admin clicks "Edit" on an entry
- **THEN** an edit form is shown pre-filled with the entry's data

#### Scenario: Submit form
- **WHEN** an admin submits the create/edit form
- **THEN** the entry is saved via HTMX and the entry list refreshes

### Requirement: Bulk operations
The system SHALL support bulk select, delete, and export operations.

#### Scenario: Select multiple entries
- **WHEN** an admin checks multiple entries in the browser
- **THEN** a bulk action toolbar appears with available actions

#### Scenario: Bulk delete
- **WHEN** an admin selects entries and clicks "Delete Selected"
- **THEN** all selected entries are deleted after confirmation

#### Scenario: Bulk export
- **WHEN** an admin selects entries and clicks "Export Selected"
- **THEN** a JSON file containing the selected entries is downloaded

### Requirement: Entry detail view
The system SHALL provide a detail view for any compendium entry showing all field values with proper formatting.

#### Scenario: View entry detail
- **WHEN** an admin clicks an entry's name
- **THEN** a detail view shows all fields with rendered values (JSON fields formatted, images shown, etc.)
