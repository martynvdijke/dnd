# custom-monster-creator Specification

## Purpose
Form UI for creating custom/homebrew monsters saved to the `monster_library` table with full stat blocks.

## Requirements

### Requirement: Custom monster creation form
The system SHALL provide a form UI for creating custom/homebrew monsters that saves to the `monster_library` table.

#### Scenario: Open creation form
- **WHEN** a user clicks "New Custom Monster" from the compendium browser or monster library section
- **THEN** a full stat-block creation form is displayed
- **AND** the form includes fields for: name, size, type, AC, HP, all six ability scores, CR, saves, skills, damage vulnerabilities/resistances/immunities, condition immunities, senses, languages, special abilities, actions, legendary actions, and description

#### Scenario: Form validation
- **WHEN** a user submits the form with missing required fields
- **THEN** the form is not submitted
- **AND** validation errors are displayed inline for: name (required), HP (numeric, >= 1), AC (numeric, >= 0), ability scores (numeric, 1-30)
- **AND** the CR field accepts values like "0", "1/8", "1/4", "1/2", and integers 1-30

#### Scenario: Create monster from scratch
- **WHEN** a user fills all fields and clicks "Create Monster"
- **THEN** a new entry is inserted into `monster_library` with `source="homebrew"` and `is_full=1`
- **AND** the form is cleared
- **AND** the new monster appears in the library listing
- **AND** a success notification is shown

#### Scenario: Pre-fill from compendium template
- **WHEN** a user clicks "Clone from Compendium" on an existing compendium monster
- **THEN** the creation form opens with all stat fields pre-filled from that compendium entry
- **AND** the name field has " (Variant)" appended
- **AND** the source defaults to "homebrew"
- **AND** the user can modify any field before saving

### Requirement: Custom monster editing
The system SHALL allow users to edit and delete their own custom monsters from the library.

#### Scenario: Edit custom monster
- **WHEN** a user clicks "Edit" on a monster in their library
- **THEN** the creation form opens pre-filled with the monster's current data
- **AND** on submit, the `monster_library` entry is updated
- **AND** the library listing reflects the changes

#### Scenario: Delete custom monster
- **WHEN** a user clicks "Delete" on a monster in their library
- **THEN** a confirmation dialog is shown: "Delete <name>? This cannot be undone."
- **AND** on confirmation, the entry is removed from `monster_library`
- **AND** associated entries in `oneshot_monsters` with `source="homebrew"` retain their data but lose the library link

### Requirement: Monster library listing
The system SHALL display the user's custom monsters in a separate "Monster Library" section accessible from the compendium area.

#### Scenario: Library shows user's monsters
- **WHEN** a user opens the Monster Library section
- **THEN** all monsters in `monster_library` for that user are displayed with name, type, CR
- **AND** each entry shows Edit and Delete buttons
- **AND** a "New Custom Monster" button is prominently displayed

#### Scenario: Empty library
- **WHEN** a user has no custom monsters
- **THEN** the Library section shows "No custom monsters yet. Create one!" with a link to the creation form
