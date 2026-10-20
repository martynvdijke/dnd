## Why

In the Combat Tracker (`ts/combat-tracker.ts`, `handlers/routes_combat.go`):
1. **No inline Monster Statblock or Monster Attacker support**: Combat entries only show `name, type, ac, hp, initiative`. When a monster's turn arrives, the DM cannot expand its compendium statblock/actions inline, and the `Attack` modal (`showAttackModal` in `ts/combat-tracker.ts:242`) only fetches `/api/characters`, preventing monsters/NPCs in combat from resolving attacks through the modal.
2. **Client-side sequential initiative loop**: `rollAllInitiative()` in `ts/combat-tracker.ts` loops `POST /api/roll` + `PUT /api/combat/:id` sequentially for each entry instead of calling the existing server batch endpoint `POST /api/combat/roll-initiative`.
3. **No End-Encounter XP & Loot flow**: When combat concludes, the DM must manually delete combat entries one by one, manually calculate monster XP, manually add XP to each character, and separately generate/add treasure to Party Inventory.

## What Changes

- Add an inline **Monster Statblock Drawer/Popover** in `ts/combat-tracker.ts` for non-player combat entries, matching the combatant's name against `/api/compendium/monsters` and rendering its ability scores, traits, and clickable actions.
- Update the Combat Attack flow (`ts/combat-tracker.ts`, `handlers/combat_automation.go`) so monsters/combat entries (or custom attack bonus + damage dice) can act as the attacker against any combat target.
- Switch `rollAllInitiative()` in `ts/combat-tracker.ts` to use `POST /api/combat/roll-initiative` (`campaign_id`) for fast, atomic server-side initiative rolling and ordering.
- Add an **End Encounter & Award Rewards** flow (`data-testid="combat-end-encounter-btn"`) that tallies XP from defeated monsters, distributes XP evenly across party characters via `POST /api/characters/:id/xp`, optionally generates/deposits treasure into Party Inventory (`/api/campaigns/:id/party-items`), and clears completed combat entries.

## Capabilities

### New Capabilities

- `combat-encounter-resolution`: Inline compendium statblock lookup in the Combat Tracker, monster attacker support in the Combat Attack modal, atomic batch initiative rolling, and end-of-encounter XP and treasure distribution.

### Modified Capabilities

- `combat-automation`: Allow attack resolution to accept an explicit ad-hoc/monster attacker (`attacker_name`, `attack_bonus`, `damage_dice`, `damage_type`) in addition to a `character_id` attacker.

## Impact

- **Backend (`handlers/combat_automation.go`, `handlers/routes_combat.go`)**:
  - Support optional `attacker_name` / `attack_bonus` / `damage_dice` when `attacker_id` is omitted or refers to a monster combat entry in `POST /api/combat/attack`.
  - Add `POST /api/combat/clear` (batch clear by `campaign_id`) to cleanly reset finished encounters.
- **Frontend (`ts/combat-tracker.ts`, `static/app.html`)**:
  - Inline statblock toggle button on monster combat entries.
  - Include monster combat entries and compendium actions in `showAttackModal()`.
  - End Encounter & Rewards modal in `ts/combat-tracker.ts`.
- **Tests (`handlers/*_test.go`, `ts/**/*.test.ts`, `tests/`)**:
  - Backend unit tests for monster attack resolution and encounter clear.
  - E2E tests for inline statblock preview and End Encounter rewards modal.
