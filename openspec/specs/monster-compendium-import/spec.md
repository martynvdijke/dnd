# monster-compendium-import Specification

## Purpose
Browse a global monster compendium (SRD-based) and import monsters into adventures or campaigns with one click.

## Requirements

### Requirement: System provides a global monster compendium
The system SHALL maintain a `compendium_monsters` table seeded with SRD monster data that all users can browse.

#### Scenario: Browse compendium monsters
- **WHEN** the DM opens the monster compendium browser
- **THEN** the system SHALL display a searchable, paginated list of all monsters in the compendium with name, CR, type, and source

#### Scenario: Search compendium by name
- **WHEN** the DM types a search query in the compendium search field
- **THEN** the system SHALL filter monsters whose name contains the search string (case-insensitive)

#### Scenario: Filter compendium by CR
- **WHEN** the DM selects a CR range filter
- **THEN** the system SHALL filter monsters to only those within the selected CR range

### Requirement: DM can preview a compendium monster stat block
The system SHALL display the full stat block of any compendium monster when selected.

#### Scenario: View monster details
- **WHEN** the DM clicks on a monster in the compendium list
- **THEN** the system SHALL display the monster's full stat block including AC, HP, ability scores, saves, skills, damage resistances/immunities, senses, languages, special abilities, actions, and legendary actions

### Requirement: DM can import a compendium monster into a one-shot act or scene
The system SHALL allow importing a monster from the global compendium into a one-shot adventure's act or scene as a new `OneShotMonster`.

#### Scenario: Import monster into act
- **WHEN** the DM selects a compendium monster and chooses to add it to a one-shot act
- **THEN** the system creates a new `OneShotMonster` in that act with all stat fields copied from the compendium entry and sets `library_id` to reference the compendium source

#### Scenario: Import monster into scene
- **WHEN** the DM selects a compendium monster and chooses to add it to a one-shot scene
- **THEN** the system creates a new `OneShotMonster` in that scene with all stat fields copied from the compendium entry

### Requirement: DM can import a compendium monster into a campaign encounter
The system SHALL allow importing a monster from the global compendium into a campaign encounter template.

#### Scenario: Import monster into campaign encounter
- **WHEN** the DM selects a compendium monster and chooses to add it to a campaign encounter
- **THEN** the system creates a new `EncounterMonster` in that encounter with stat fields copied from the compendium entry

### Requirement: DM can import from their monster library into acts/scenes
The system SHALL allow selecting a monster from the user's existing `monster_library` and importing it into a one-shot act or scene.

#### Scenario: Import from library into act
- **WHEN** the DM selects a monster from their library and chooses to add it to a one-shot act
- **THEN** the system creates a new `OneShotMonster` with fields pre-filled from the library entry, linked via `library_id`

### Requirement: Library import pre-fills creation form
The system SHALL pre-fill the monster creation form when importing from the user's library, allowing the DM to modify fields before saving.

#### Scenario: Pre-fill and edit
- **WHEN** the DM selects a library monster for import into an act/scene
- **THEN** the system opens the monster creation form with all stat fields pre-filled from the library entry
- **AND** the DM can modify any field before submitting
