## MODIFIED Requirements

### Requirement: Character sheet command

The bot SHALL render a character sheet summary for the claimed character, or for a character given by id or name, including identity (race, class, subclass, level), ability scores with modifiers, hit points, armor class, initiative, speed, proficiency bonus, passive perception, and counters for features, spells, and inventory. The sheet's other-content section SHALL point at the tracking commands for inventory, spells, conditions, features and currency. Rendered content SHALL escape user-supplied text.

#### Scenario: Sheet defaults to the claim

- **WHEN** a linked user with a claimed character sends `/sheet`
- **THEN** the reply renders that character's sheet summary

#### Scenario: Sheet by name

- **WHEN** a linked user sends `/sheet <name>` for a character they may edit
- **THEN** the reply renders that character's sheet summary

#### Scenario: Foreign character is denied

- **WHEN** a linked user requests a sheet for a character they may not edit
- **THEN** the bot replies with the same "not found" response as for a nonexistent character and reveals no sheet data

#### Scenario: Unclaimed user is guided

- **WHEN** a linked user without a claim sends `/sheet` and owns several characters
- **THEN** the bot replies with a selection list instead of guessing a character

#### Scenario: Hostile text is escaped

- **WHEN** a character name or trait contains HTML-significant characters
- **THEN** the message renders the text literally without enabling markup

#### Scenario: Other section points at tracking

- **WHEN** the sheet renders its other-content counters
- **THEN** it names the tracking commands for inventory, spells, conditions, features and currency
