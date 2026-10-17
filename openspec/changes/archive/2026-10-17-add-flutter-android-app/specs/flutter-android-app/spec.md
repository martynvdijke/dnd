## ADDED Requirements

### Requirement: Interactive mobile character sheet
The Flutter Android app SHALL render a tabbed character sheet backed by `/api/characters/:id` supporting 1-tap HP/temp-HP adjustments (`/hp`), spell slot expenditure (`/cast-spell`, `/spellcasting`), inventory and currency management (`/inventory`, `/currency`), and Short/Long Rest recovery.

#### Scenario: Adjust character HP and expend spell slot in combat
- **WHEN** a player taps `-8 HP` or expends a Level 3 spell slot on their phone
- **THEN** the app updates the character state via `/api/characters/:id/hp` or `/api/characters/:id/cast-spell` and reflects the change immediately

### Requirement: Pinch-to-zoom battlemap and combat roller
The app SHALL display the active campaign battlemap (`/api/campaigns/:id/battlemap`) with pinch-to-zoom and draggable character tokens (`/api/battlemap-tokens/:id`) alongside an attack/damage roller (`/api/combat/attack`).

#### Scenario: Move character token on battlemap
- **WHEN** a player drags their token to a new grid cell on the mobile battlemap
- **THEN** the app posts the updated normalised token coordinates to `/api/battlemap-tokens/:id` and syncs with the table

### Requirement: Compendium search and AI drafting
The app SHALL provide instant full-text search across `/api/compendium/search`, random generators (`/api/generate/*`), and conversational AI entity drafting (`/api/ai/draft`) for DM/admin users.

#### Scenario: Search spell or monster in compendium
- **WHEN** the user types a query in the Compendium tab
- **THEN** the app displays matching spells, monsters, and items from `/api/compendium/search`
