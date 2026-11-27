# fab-cleanup Specification

## Purpose
TBD - created by archiving change character-selection-ownership. Update Purpose after archive.
## Requirements
### Requirement: Remove "Generate with AI" FAB entry

The system SHALL remove the "Generate with AI" item from the floating action button (FAB) menu in both the static FAB markup and the dynamically generated per-view actions.

#### Scenario: FAB menu without AI entry

- **WHEN** the user opens the FAB menu
- **THEN** the menu SHALL not contain a "Generate with AI" item

#### Scenario: Dynamically rendered views

- **WHEN** the FAB menu is rebuilt for any view
- **THEN** the rendered actions SHALL not include the AI generation action

### Requirement: Tidy FAB and quick-action UI

The system SHALL render the FAB menu with consistent iconography, spacing, and styling, and remove any stale or duplicate entries so the menu shows only the actions relevant to the current view.

#### Scenario: FAB menu reflects current view

- **WHEN** the user switches views
- **THEN** the FAB menu SHALL show only the actions configured for that view, with consistent icon and label styling

#### Scenario: No duplicate entries

- **WHEN** the FAB menu is open
- **THEN** each action SHALL appear exactly once

### Requirement: Selection UI cleanliness

The campaign and character selection UI introduced by the selection flow SHALL follow the same visual polish as the FAB cleanup: consistent cards, spacing, and empty states.

#### Scenario: Selection screens render cleanly

- **WHEN** the user completes the campaign and character selection
- **THEN** the screens SHALL render with consistent styling, spacing, and empty states matching the rest of the app
