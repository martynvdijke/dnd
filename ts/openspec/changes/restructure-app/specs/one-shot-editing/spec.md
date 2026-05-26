## ADDED Requirements

### Requirement: Inline-editable durations
Act and scene `estimated_minutes` fields SHALL be inline-editable via HTMX. Clicking the duration SHALL replace it with an input field. On blur or Enter, a PUT request SHALL update the value.

#### Scenario: Click duration to edit
- **WHEN** a DM clicks the estimated_minutes display on an act or scene
- **THEN** the display SHALL be replaced with an input field pre-filled with the current value

#### Scenario: Save duration on blur
- **WHEN** a DM edits the duration input and clicks away
- **THEN** a PUT request SHALL be sent to update the duration, and the display SHALL update

#### Scenario: Save duration on Enter
- **WHEN** a DM edits the duration input and presses Enter
- **THEN** a PUT request SHALL be sent to update the duration, and the display SHALL update

### Requirement: Visual act/scene tree
The one-shot detail view SHALL display acts and scenes in a tree structure with visual hierarchy. Acts are top-level, scenes are nested under their parent act with indentation.

#### Scenario: Display act/scene tree
- **WHEN** a DM views a one-shot detail page
- **THEN** acts SHALL be displayed as expandable sections with scenes nested inside

### Requirement: Drag-reorder acts and scenes
The system SHALL support reordering acts and scenes via drag-and-drop using SortableJS.

#### Scenario: Reorder acts
- **WHEN** a DM drags an act to a new position in the tree
- **THEN** a POST request SHALL be sent to `/api/oneshot-acts/reorder` with the new order, and the display SHALL update

#### Scenario: Reorder scenes within act
- **WHEN** a DM drags a scene to a new position within its act
- **THEN** a POST request SHALL be sent to `/api/oneshot-scenes/reorder` with the new order, and the display SHALL update
