## ADDED Requirements

### Requirement: Parse and evaluate standard RPG dice notation
The system SHALL accept and evaluate dice expressions using the standard RPG dice notation syntax as defined by `@dice-roller/rpg-dice-roller`.

Supported notation operators:
- **Basic**: `NdS` — roll N dice with S sides (e.g., `2d6`, `1d20`)
- **Keep/Drop**: `NdSkhN` / `NdSklN` — keep highest/lowest N; `NdSdhN` / `NdSdlN` — drop highest/lowest N (e.g., `4d6kh3`)
- **Exploding**: `NdS!` — explode on max; `NdS!>=T` — explode on threshold T; `NdS!!` — compound exploding
- **Reroll**: `NdSrN` — reroll N once; `NdSroN` — reroll N indefinitely (e.g., `1d20r1`)
- **Success counting**: `NdS>T` — count rolls above T; `NdS<C` — count rolls below C (e.g., `8d10>7`)
- **Die grouping**: Implicit via expression combination with `+`/`-` (already supported)
- **Modifiers**: `+N` / `-N` — static modifiers (already supported)
- **Percentile**: `d100` / `d%` — percentile rolls

#### Scenario: Basic expression
- **WHEN** user sends `POST /api/roll` with `{"expression": "2d6+3"}`
- **THEN** the response contains `total` equal to the sum of 2 random 1-6 values plus 3, `breakdown` arrays with two die results and a +3 modifier

#### Scenario: Keep highest (4d6kh3)
- **WHEN** user sends `POST /api/roll` with `{"expression": "4d6kh3"}`
- **THEN** the response contains `total` equal to the sum of the 3 highest values out of 4d6, and `breakdown` lists all 4 rolls with metadata indicating which 3 were kept

#### Scenario: Exploding dice
- **WHEN** user sends `POST /api/roll` with `{"expression": "3d6!"}`
- **THEN** any die that rolls a 6 is rerolled and added to its total, and `breakdown` includes the exploded rolls

#### Scenario: Reroll
- **WHEN** user sends `POST /api/roll` with `{"expression": "1d20r1"}`
- **THEN** a die that rolls 1 is rerolled once, and the breakdown shows the discarded roll and the final result

#### Scenario: Success counting
- **WHEN** user sends `POST /api/roll` with `{"expression": "8d10>7"}`
- **THEN** the `total` is the count of dice showing 8+, and breakdown shows each individual die value

#### Scenario: Invalid expression
- **WHEN** user sends `POST /api/roll` with `{"expression": "invalid!!"}`
- **THEN** the response is 400 with an error message describing the parse failure

### Requirement: Preserve advantage/disadvantage
The system SHALL support advantage and disadvantage as syntactic sugar for `2d20kh1` / `2d20kl1`.

#### Scenario: Advantage via shortcut
- **WHEN** user sends `POST /api/roll` with `{"expression": "1d20", "advantage": "advantage"}`
- **THEN** the result is equivalent to `2d20kh1`, both rolls are shown in the breakdown, and the API response is identical format

### Requirement: Server-authoritative rolls
All dice evaluations SHALL happen server-side. The frontend SHALL NOT compute die results independently.

#### Scenario: Roll result from server
- **WHEN** user triggers a roll from the UI
- **THEN** the expression is sent to `POST /api/roll`, the server evaluates it, saves to history, and returns the result
- **THEN** the frontend displays the returned result and animates based on the server's breakdown

### Requirement: Roll history for notation expressions
The system SHALL save all rolls (including those using advanced notation) to the dice_rolls table with the full expression string and result text.

#### Scenario: History with notation
- **WHEN** user rolls `4d6kh3` and views roll history
- **THEN** the history shows the expression `4d6kh3` and its total
