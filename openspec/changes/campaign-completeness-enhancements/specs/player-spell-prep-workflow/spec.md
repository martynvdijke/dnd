## ADDED Requirements

### Requirement: Batch spell preparation modal
The system SHALL provide a "Prepare Spells" button on the spells tab that opens a modal for batch-managing prepared spells.

#### Scenario: Prepare Spells button visible on spells tab
- **WHEN** the spells tab is rendered for a character with spellcasting ability
- **THEN** the system SHALL show a "Prepare Spells" button in the spells tab header

#### Scenario: Modal shows all known spells with checkboxes
- **WHEN** the user clicks "Prepare Spells"
- **THEN** the system SHALL open a modal listing all the character's known spells grouped by level, each with a checkbox for prepared status

#### Scenario: Checkbox reflects current prepared state
- **WHEN** the modal opens
- **THEN** each spell's checkbox SHALL reflect its current `prepared` value

#### Scenario: Batch save updates all selections
- **WHEN** the user toggles checkboxes and clicks "Save Preparation"
- **THEN** the system SHALL send a single PUT request with an array of spell IDs to set as prepared, AND close the modal, AND refresh the spells tab

#### Scenario: Cancel discards changes
- **WHEN** the user clicks "Cancel" or the close button
- **THEN** the modal SHALL close without saving any changes

#### Scenario: Prepared spell count shown
- **WHEN** the modal is open
- **THEN** the system SHALL show "X prepared / Y total" with the maximum number of spells that can be prepared based on class level and spellcasting ability
