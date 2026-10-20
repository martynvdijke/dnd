## Context

`ts/combat-tracker.ts` manages the active initiative order (`/api/combat`). Currently:
- `showAttackModal()` fetches `/api/characters` for the Attacker dropdown, meaning only player characters can initiate automated attacks, even when it is a monster's turn.
- Monster combatants have no link to their `/api/compendium/monsters` statblock (traits, actions, attack bonuses, damage dice, CR/XP).
- `rollAllInitiative()` makes $2N$ sequential HTTP calls (`POST /api/roll` + `PUT /api/combat/:id`) despite `POST /api/combat/roll-initiative` already existing in `handlers/routes_combat.go`.
- Ending an encounter requires deleting every combatant row manually with no XP or loot handoff.

## Goals / Non-Goals

**Goals:**
- Let DMs expand a monster combatant's compendium statblock inline in the Combat Tracker and click any monster action to pre-fill the Attack modal.
- Allow `POST /api/combat/attack` and `showAttackModal()` to resolve attacks from monster/NPC combat entries (using explicit `attack_bonus` and `damage_dice`) as well as player characters.
- Use `POST /api/combat/roll-initiative` in `rollAllInitiative()`.
- Provide an **End Encounter** workflow that summarizes defeated monster XP, awards split XP to party characters, optionally adds generated treasure to Party Inventory, and clears the combat tracker.

**Non-Goals:**
- Changing the underlying `combat_entries` schema.
- Replacing the Encounter Builder view (`ts/encounter.ts`).

## Decisions

1. **Compendium Monster Matching by Name**:
   - *Decision*: When a combat entry has `type === 'monster'` or `type === 'npc'`, strip trailing numbering (e.g. `"Goblin #2"` or `"Goblin 2"` → `"Goblin"`) and query `/api/compendium/monsters?q=<baseName>` (cached in memory per session) to display an inline collapsible statblock (`AC`, `HP`, `Speed`, `STR..CHA`, `Actions`, `CR / XP`).
   - *Rationale*: Works with all existing combat entries created from Encounter templates or added manually without requiring a schema migration.

2. **Monster Attacker Support in `POST /api/combat/attack`**:
   - *Decision*: Extend `AttackRequest` in `handlers/combat_automation.go` so that if `AttackerID == 0`, the handler uses `AttackerName`, `AttackBonus`, `DamageDice`, `DamageBonus`, and `DamageType` supplied in the request (and optionally `AttackerCombatEntryID` to inspect attacker conditions).
   - *Rationale*: Backward-compatible with character-based attacks while enabling full automated attack + damage + concentration + combat log resolution for monsters.

3. **End Encounter & Award Rewards Modal**:
   - *Decision*: Add an `End Encounter` button (`data-testid="combat-end-encounter-btn"`) to the Combat Tracker header. Clicking it opens a modal showing defeated monsters (`hp_current === 0`), their matched Compendium XP total, the per-character XP share across active party characters, an optional "Roll Treasure Hoard to Party Loot" checkbox, and a "Complete & Clear Encounter" button.
   - *Rationale*: Connects Combat Tracker → Character XP (`POST /api/characters/:id/xp`) → Party Inventory (`/api/campaigns/:id/party-items`) in one cohesive step.

## Risks / Trade-offs

- **[Risk] Custom homebrew monsters not found in `/api/compendium/monsters`** → *Mitigation*: When no compendium match is found, the inline statblock panel gracefully shows the combat entry's known stats (`AC`, `HP`, `Init Mod`) and allows entering custom `Attack Bonus` + `Damage Dice` in the Attack modal, plus a manual XP override in the End Encounter modal.
