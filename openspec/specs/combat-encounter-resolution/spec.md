# combat-encounter-resolution Specification

## Purpose
TBD - created by archiving change combat-statblocks-and-encounter-rewards. Update Purpose after archive.
## Requirements
### Requirement: Inline compendium statblock preview in Combat Tracker
The Combat Tracker SHALL allow expanding a monster or NPC combat entry to view its matching compendium statblock inline (including ability scores, traits, actions, and CR/XP), stripping trailing instance numbers (such as `#1` or `2`) when matching by name.

#### Scenario: Expanding a monster's statblock in combat
- **WHEN** the DM clicks the statblock toggle on a `"Goblin #1"` combat entry
- **THEN** the Combat Tracker displays the matching `"Goblin"` compendium statblock inline with its actions

### Requirement: End encounter and reward distribution
The Combat Tracker SHALL provide an End Encounter workflow that calculates total XP from defeated monsters (`hp_current == 0`), allows a manual XP override, distributes XP across selected party characters through `POST /api/characters/:id/xp`, optionally adds generated treasure to Party Inventory, and clears the combat tracker entries through `POST /api/combat/clear`.

#### Scenario: Ending combat awards XP and clears the tracker
- **WHEN** the DM opens the End Encounter modal, confirms the XP reward for party members, and completes the encounter
- **THEN** XP is awarded to the selected party characters and the combat tracker entries for the encounter are cleared
