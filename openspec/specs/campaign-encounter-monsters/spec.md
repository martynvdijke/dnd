# campaign-encounter-monsters Specification

## Purpose
Add and manage monsters within campaign-level encounters.

## Requirements

### Requirement: DM can add a monster to a campaign encounter
The system SHALL allow adding a monster (either from the global compendium, manually entered, or from the user's monster library) to a campaign encounter template.

#### Scenario: Add monster to encounter
- **WHEN** the DM opens the campaign encounter detail view and clicks "Add Monster"
- **THEN** the system SHALL present options to: add from compendium, add from library, or enter manually
- **AND** upon selection, create a new `EncounterMonster` record linked to the encounter

#### Scenario: Add monster manually
- **WHEN** the DM chooses to add a monster manually to a campaign encounter
- **THEN** the system SHALL display a form with fields: name, AC, HP, CR, initiative modifier, source, notes, and ability scores
- **AND** upon submit, create a new `EncounterMonster` record

### Requirement: DM can update a monster in a campaign encounter
The system SHALL allow editing the stat fields of an existing `EncounterMonster`.

#### Scenario: Update encounter monster
- **WHEN** the DM edits any field of an encounter monster and submits
- **THEN** the system updates the `EncounterMonster` record with the new values

### Requirement: DM can remove a monster from a campaign encounter
The system SHALL allow deleting a monster from a campaign encounter.

#### Scenario: Remove encounter monster
- **WHEN** the DM clicks delete on an encounter monster
- **THEN** the system removes the `EncounterMonster` record

### Requirement: DM can view all monsters in a campaign encounter
The system SHALL display the full list of monsters in a campaign encounter with their stats.

#### Scenario: View encounter monsters
- **WHEN** the DM opens a campaign encounter detail view
- **THEN** the system SHALL display all monsters in that encounter with name, CR, count, AC, HP, and source

### Requirement: Campaign encounter monster management appears in campaign UI
The system SHALL integrate monster management into the existing campaign detail views.

#### Scenario: Campaign encounters tab
- **WHEN** the DM navigates to the campaign detail page
- **THEN** the system SHALL show an "Encounters" section listing all encounter templates for that campaign
- **AND** each encounter entry SHALL show the number of monsters and allow expanding to manage monsters

#### Scenario: Monster count summary
- **WHEN** the DM views the campaign encounter list
- **THEN** each encounter SHALL display a summary like "5 monsters (3 Goblins, 2 Hobgoblins)"
