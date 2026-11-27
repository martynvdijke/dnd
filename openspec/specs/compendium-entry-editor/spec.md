# compendium-entry-editor Specification

## Purpose
Dynamic form-based entry editor that renders schema-appropriate input controls with validation.

## Requirements

### Requirement: Schema-aware dynamic form rendering
The entry editor SHALL render form controls based on `SchemaField.Type` from the schema definition. Each field type SHALL produce the appropriate HTML input:

| SchemaField.Type | HTML Control |
|----------------|-------------|
| `string` | `<input type="text">` |
| `text` | `<textarea>` |
| `integer` | `<input type="number" step="1">` |
| `float` | `<input type="number" step="0.01">` |
| `boolean` | `<input type="checkbox">` |
| `select` | `<select>` with options from `SchemaField.Options` |
| `multi-select` | `<select multiple>` with options |
| `json` | `<textarea class="font-monospace">` with pretty-printed JSON |

#### Scenario: Select field renders dropdown with options
- **WHEN** a schema defines a field with `type: "select"` and `options: ["dnd5e", "pf2e", "homebrew"]`
- **THEN** the editor renders a `<select>` with those three options

#### Scenario: Boolean field renders checkbox
- **WHEN** a schema defines a field with `type: "boolean"`
- **THEN** the editor renders an `<input type="checkbox">`

#### Scenario: JSON field renders monospace textarea
- **WHEN** a schema defines a field with `type: "json"`
- **THEN** the editor renders a `<textarea>` with monospace font and validates JSON on submit

### Requirement: Field validation on create/update
The editor SHALL validate required fields (marked `required: true` in SchemaField) before submission. Empty required fields SHALL show inline validation errors (red border + message). Non-required fields SHALL be optional.

#### Scenario: Required field validation on submit
- **WHEN** a required "name" field is empty and the admin clicks Save
- **THEN** the field gets an `is-invalid` class applied and a "This field is required" message appears

#### Scenario: Optional field can be empty
- **WHEN** a non-required "description" field is empty and the admin clicks Save
- **THEN** the entry is saved successfully with an empty description value

### Requirement: Inline editing in entry table
The entry browser SHALL support an "inline edit" mode where clicking a cell value allows editing it directly in the table without opening the modal. Only string and text fields SHALL support inline editing.

#### Scenario: Inline edit text field
- **WHEN** an admin double-clicks a text cell in the entry table
- **THEN** the cell becomes an inline `<input>` or `<textarea>`; pressing Enter saves via PUT to `/api/admin/compendium-entries/{id}` with the updated field

### Requirement: Entry detail view
The system SHALL provide a detail view showing ALL fields of an entry (not just the 3–5 table columns) with read-only rendering for all field types, including formatted JSON for JSON fields.

#### Scenario: View full entry details
- **WHEN** an admin clicks "View" on an entry row
- **THEN** a modal shows all fields with labels, values, and proper formatting (checkboxes for booleans, code blocks for JSON)

### Requirement: Create compendium entry via modal
The system SHALL allow creating a new compendium entry through a modal form that renders fields dynamically from the schema definition.

#### Scenario: Open create modal for a schema
- **WHEN** user clicks "Add Entry" button above the entry table
- **THEN** the frontend SHALL open a modal with the title "Create Entry"
- **THEN** the modal SHALL render form fields based on the schema's `fields` array
- **THEN** each field SHALL use the appropriate input type (text, number, textarea, checkbox) based on `field.type`
- **THEN** required fields SHALL be marked with a visual indicator

#### Scenario: Successful entry creation
- **WHEN** user fills in all required fields and clicks "Save"
- **THEN** the frontend SHALL send `POST /api/admin/compendium-schemas/:id/entries` with `{ data: { ...fieldValues } }`
- **THEN** the backend SHALL create the entry and return `{ id: <new_id> }`
- **THEN** the frontend SHALL close the modal
- **THEN** the frontend SHALL refresh the entry list to show the new entry
- **THEN** the frontend SHALL show a success toast

#### Scenario: Create entry with missing required fields
- **WHEN** user clicks "Save" with empty required fields
- **THEN** the frontend SHALL highlight the empty required fields
- **THEN** the frontend SHALL display validation errors per field
- **THEN** the backend SHALL NOT be called
- **THEN** the modal SHALL remain open

#### Scenario: Create entry - backend validation failure
- **WHEN** the backend returns a validation error
- **THEN** the frontend SHALL display the error in the modal
- **THEN** the modal SHALL remain open so the user can fix the issue

### Requirement: Edit compendium entry via modal
The system SHALL allow editing an existing compendium entry through a modal pre-filled with the entry's current data.

#### Scenario: Open edit modal for an entry
- **WHEN** user clicks the "Edit" button on an entry row
- **THEN** the frontend SHALL open the same modal with the title "Edit Entry"
- **THEN** the modal SHALL render the same schema-driven form fields
- **THEN** each field SHALL be pre-filled with the entry's current data values
- **THEN** the modal SHALL have a "Save" button (not "Create")

#### Scenario: Successful entry update
- **WHEN** user modifies fields and clicks "Save"
- **THEN** the frontend SHALL send `PUT /api/admin/compendium-entries/:id` with `{ data: { ...updatedValues } }`
- **THEN** the backend SHALL update the entry and return `{ ok: true }`
- **THEN** the frontend SHALL close the modal
- **THEN** the frontend SHALL refresh the entry list
- **THEN** the frontend SHALL show a success toast

#### Scenario: Cancel editing
- **WHEN** user clicks "Cancel" or the modal close button
- **THEN** the modal SHALL close without saving
- **THEN** no API call SHALL be made
- **THEN** the entry list SHALL remain unchanged
