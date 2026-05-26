## ADDED Requirements

### Requirement: Act/Scene tree view
The system SHALL provide a tree view of acts and scenes for a one-shot adventure. Each act is a collapsible node containing its scenes as child nodes.

#### Scenario: Render act/scene tree
- **WHEN** a DM navigates to the one-shot editor
- **THEN** the system renders a tree view with acts as top-level nodes and scenes as children

#### Scenario: Expand/collapse act
- **WHEN** a DM clicks on an act header in the tree
- **THEN** the system toggles visibility of that act's scenes

### Requirement: Inline-editable durations
Acts and scenes SHALL support inline editing of their estimated_minutes field via HTMX hx-trigger="blur".

#### Scenario: Edit act duration inline
- **WHEN** a DM clicks and edits the duration field of an act, then blurs away
- **THEN** the system sends a PATCH request and updates the duration

#### Scenario: Edit scene duration inline
- **WHEN** a DM clicks and edits the duration field of a scene, then blurs away
- **THEN** the system sends a PATCH request and updates the duration

### Requirement: Drag-reorder acts and scenes
The system SHALL support drag-and-drop reordering of acts (within a one-shot) and scenes (within an act) using SortableJS.

#### Scenario: Reorder acts
- **WHEN** a DM drags an act to a new position in the tree
- **THEN** the system sends a PUT request to update act numbers and re-renders the tree

#### Scenario: Reorder scenes within an act
- **WHEN** a DM drags a scene to a new position within its act
- **THEN** the system sends a PUT request to update scene numbers and re-renders the tree
