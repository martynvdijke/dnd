# flutter-android-app Delta — MODIFIED Requirements

## MODIFIED Requirements

### Requirement: Interactive mobile character sheet

The Flutter Android app SHALL render a tabbed character sheet backed by `/api/characters/:id` supporting 1-tap HP/temp-HP adjustments (`/hp`), spell slot expenditure (`/cast-spell`, `/spellcasting`), inventory and currency management (`/inventory`, `/currency`), and Short/Long Rest recovery. The sheet SHALL render the character's real sub-resources (spellcasting, spells, inventory, currency, features, classes) from the server response, SHALL NOT display fabricated or placeholder values presented as character data, and death saves and exhaustion adjustments SHALL call their server endpoints and reflect the server result.

#### Scenario: Adjust character HP and expend spell slot in combat

- WHEN player taps `-8 HP` or expends a Level 3 slot on phone
- THEN app updates via `/api/characters/:id/hp` or `/cast-spell` and reflects immediately

#### Scenario: Sheet shows only server-provided values

- WHEN a character whose proficiency and bonus values are not provided by the server is opened
- THEN the sheet renders those sections from server data or an explicit empty state, and no hardcoded proficiency or bonus values are displayed as if they were the character's

#### Scenario: Death saves and exhaustion persist

- WHEN the player marks a death save success or increases exhaustion
- THEN the app calls the death-save and exhaustion endpoints and the sheet reflects the persisted state

### Requirement: Pinch-to-zoom battlemap and combat roller

The app SHALL display the active campaign battlemap (`/api/campaigns/:id/battlemap`) with pinch-to-zoom and draggable character tokens (`/api/battlemap-tokens/:id`) alongside an attack/damage roller (`/api/combat/attack`). The roller SHALL use the attacker the user selected and SHALL display only results returned by the server; on API error it SHALL surface the error instead of presenting fabricated numbers.

#### Scenario: Move character token on battlemap

- WHEN player drags token to new grid cell
- THEN posts normalised coords to `/api/battlemap-tokens/:id` and syncs with table

#### Scenario: Attack roll uses the chosen attacker and the server's result

- WHEN the user selects a character and performs an attack roll
- THEN the app sends the attack via `/api/combat/attack` for the selected attacker and displays the server-returned outcome, or the server error when the call fails

### Requirement: Compendium search and AI drafting

The app SHALL provide instant full-text search across `/api/compendium/search`, random generators (`/api/generate/*`), and conversational AI entity drafting (`/api/ai/draft`) for DM/admin users. Search results SHALL parse the identifiers the server returns (numeric ids) so real result sets render without crashing.

#### Scenario: Search spell or monster in compendium

- WHEN user types query in Compendium tab
- THEN app displays matching spells, monsters, items from `/api/compendium/search`

#### Scenario: Real search results render without crashing

- WHEN the server returns results with integer identifiers
- THEN the result list renders each entry without a type error
