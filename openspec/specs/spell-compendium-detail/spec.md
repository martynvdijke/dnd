# spell-compendium-detail Specification

## Purpose
Rich spell detail view showing full stats, description, and source information.

## Requirements

### Requirement: Inline spell detail expansion

The system SHALL allow users to click a spell entry in the browse list to expand it inline and view full spell details.

#### Scenario: Click to expand
- **WHEN** a user clicks on a spell entry in the browse list
- **THEN** the system SHALL expand that entry to show full details below the summary line
- **AND** the expanded section SHALL show: description, components, casting time, range, duration, higher levels, and classes

#### Scenario: Click to collapse
- **WHEN** a user clicks on an already-expanded spell entry
- **THEN** the system SHALL collapse the detail section
- **AND** return the entry to summary-only display

#### Scenario: Multiple expansions
- **WHEN** a user expands one spell and then clicks a different spell
- **THEN** the system SHALL collapse the previously expanded spell
- **AND** expand the newly selected spell

### Requirement: Full spell detail modal

The system SHALL provide a "View Full" button or link in the expanded detail that opens a full-screen or modal view of the spell.

#### Scenario: Open full detail modal
- **WHEN** a user clicks "View Full" on an expanded spell
- **THEN** the system SHALL open a modal with the complete spell stat block and description
- **AND** the modal SHALL include: name, level, school, casting time, range, components, duration, description, higher levels, classes, source, and page reference

#### Scenario: Close modal
- **WHEN** a user clicks the close button or clicks outside the modal
- **THEN** the system SHALL close the modal

### Requirement: Spell detail content

The full detail view SHALL display all standard D&D 5e spell fields in a readable format.

#### Scenario: Stat block layout
- **WHEN** a user opens a spell detail
- **THEN** the system SHALL display a formatted stat block with key-value pairs for:
  - **Casting Time**: from `casting_time` field
  - **Range**: from `range` field
  - **Components**: from `components` field
  - **Duration**: from `duration` field
  - **School**: from `school` field
  - **Level**: from `level` field
  - **Classes**: from `classes` field (formatted as comma-separated list)

#### Scenario: Description rendering
- **WHEN** a user opens a spell detail
- **THEN** the system SHALL display the `description` field with whitespace preserved
- **AND** if `higher_levels` is non-empty, display it under an "At Higher Levels" section

#### Scenario: Source reference
- **WHEN** a user opens a spell detail and the spell has a `source_page`
- **THEN** the system SHALL display the source page reference as muted text at the bottom
