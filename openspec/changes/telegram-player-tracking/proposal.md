## Why

Once a player has linked Telegram and claimed a character, the bot renders the sheet but is read-only for everything a player touches at the table: equipment, the spellbook, conditions, and coin. Players still have to open the web UI to equip a shield, mark a spell prepared, note a condition, or pay for a room. Letting them list and manage those from the chat turns the bot into a real companion at the table instead of a lookup tool.

## What Changes

- Add `/inventory` (aliases `/inv`, `/bag`) which lists the claimed character's items with category, quantity, equipped and attunement state, and inline buttons to equip/unequip, change quantity by ±1, and remove an item behind a confirmation step; `/additem <name> [qty]` adds a new item.
- Add `/spells` (alias `/spellbook`) listing spells by level with prepared state and prepare/unprepare buttons (cantrips and always-prepared spells excluded from the toggle); `/prepare <spell>` and `/unprepare <spell>`.
- Add `/conditions` listing active conditions with remove buttons and `/condition <name>` to add one, `/features` for a read-only feature list, and `/money` showing currency with `/money +N|-N <pp|gp|ep|sp|cp>` adjustments clamped at zero.
- Scope every tracking read and write to the claimed character and re-validate edit rights on each operation; all mutations are parameterized and refused for characters the linked user may not edit.
- Cross-reference the tracking commands from the `/sheet` "Other" section so the sheet is the entry point.

## Capabilities

### New Capabilities
- `telegram-player-tracking`: listing and managing the claimed character's inventory, spellbook preparation, conditions, features and currency from the bot, with inline controls and per-character authorization.

### Modified Capabilities
- `telegram-character-commands`: the character sheet's "Other" section now points at the tracking commands.

## Impact

- Backend: new `telegram/tracking.go`; `telegram/commands.go` (command registry), `telegram/keyboard.go` (callback prefixes and routing), `telegram/characters.go` (sheet cross-reference). Reuses `requireActionCharacter`/`editableCharacterSQL` from `telegram/actions.go` and `telegram/claims.go`.
- Schema: none — the tracking tables (`inventory`, `spells`, `character_spellcasting`, `character_conditions`, `character_features`, `character_currency`) already exist.
- Tests: `telegram` unit tests for listing, equip/unequip, quantity changes, remove confirmation, prepare/unprepare, currency clamping, and authorization refusals; coverage floors stay enforced.
