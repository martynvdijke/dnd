# character-builder Specification

## Purpose
TBD - created by archiving change add-player-experience. Update Purpose after archive.
## Requirements
### Requirement: Guided character builder
The system SHALL provide a multi-step character builder that collects identity, ability scores, and proficiencies, then creates the character through the existing character and proficiency endpoints and opens the created character's sheet. The builder SHALL be reachable from the New Character modal without removing the existing quick-create fields.

#### Scenario: Builder opens from the New Character modal
- **WHEN** the player clicks "Guided builder" in the New Character modal
- **THEN** the builder opens on the identity step with name, race, class, background, and level inputs

#### Scenario: Standard array assignment
- **WHEN** the player selects standard array mode
- **THEN** the six values 15, 14, 13, 12, 10, and 8 are each assigned to exactly one ability through selects, and the resulting modifiers are shown

#### Scenario: Manual ability entry
- **WHEN** the player selects manual mode and enters ability scores
- **THEN** the entered values are used for the created character

#### Scenario: Proficiencies are persisted
- **WHEN** the player selects skill and saving-throw proficiencies and creates the character
- **THEN** one proficiency record is created per selection with type `skill` or `save`

#### Scenario: Review creates the character
- **WHEN** the player confirms the review step with a non-empty name
- **THEN** the character is created with the chosen identity and abilities, and the sheet for the new character opens

#### Scenario: Name is required
- **WHEN** the player attempts to advance from the identity step with an empty name
- **THEN** the builder stays on the identity step and reports that a name is required
