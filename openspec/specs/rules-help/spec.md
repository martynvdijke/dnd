# rules-help Specification

## Purpose
TBD - created by archiving change add-player-experience. Update Purpose after archive.
## Requirements
### Requirement: Inline rules help
The system SHALL provide a static rules glossary that explains the game terms shown on the character sheet, reachable from an info button next to the relevant stat. Opening help SHALL display the rule text in a modal without navigating away from the sheet.

#### Scenario: Help button opens the rule text
- **WHEN** the player activates the help button next to Armor Class
- **THEN** a modal opens containing the Armor Class rule text

#### Scenario: Unknown term falls back gracefully
- **WHEN** help is requested for a term that is not in the glossary
- **THEN** a generic help message is shown instead of an error

#### Scenario: Help does not change the sheet
- **WHEN** the player closes the help modal
- **THEN** the character sheet is still displayed with the same tab and values
