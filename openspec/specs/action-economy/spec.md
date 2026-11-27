# action-economy Specification

## Purpose
TBD - created by archiving change add-combat-depth. Update Purpose after archive.
## Requirements
### Requirement: Per-turn action economy
The system SHALL track, for each combat entry, whether its action, bonus action, and reaction have been used, and how much movement it has spent this turn, up to a movement maximum. Consuming a resource SHALL record it and reject a spend that exceeds the available amount. The action, bonus action, reaction, and movement SHALL reset at the start of the combat entry's turn.

#### Scenario: Spending the action is recorded
- **WHEN** a combat entry with an unused action spends its action
- **THEN** the entry reports its action as used

#### Scenario: Overspending is rejected
- **WHEN** a combat entry attempts to spend an action it has already used
- **THEN** the spend is rejected and the economy is unchanged

#### Scenario: Movement is bounded
- **WHEN** a combat entry with 30 feet of movement spends 40 feet
- **THEN** the spend is rejected

#### Scenario: Turn advance resets the economy
- **WHEN** the turn advances to a combat entry
- **THEN** that entry's action, bonus action, reaction, and movement used are reset
