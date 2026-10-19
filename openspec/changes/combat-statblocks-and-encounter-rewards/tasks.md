## 1. Backend Monster Attack & Combat Reset Support

- [x] 1.1 Update `HandleCombatAttack` in `handlers/combat_automation.go` to support monster/ad-hoc attackers (`attacker_name`, `attack_bonus`, `damage_dice`, `damage_type`, optional `attacker_entry_id`) when `attacker_id` is omitted.
- [x] 1.2 Add `POST /api/combat/clear` (or campaign-scoped bulk delete) in `handlers/routes_combat.go` to clear combat entries at the end of an encounter.

## 2. Combat Tracker Inline Statblocks, Initiative & End-Encounter Flow

- [x] 2.1 Update `rollAllInitiative()` in `ts/combat-tracker.ts` to use `POST /api/combat/roll-initiative`.
- [x] 2.2 Add an inline Compendium Statblock toggle (`data-testid="combat-statblock-toggle"`) on monster/NPC rows in `ts/combat-tracker.ts` with 1-click action prefill into `showAttackModal()`.
- [x] 2.3 Update `showAttackModal()` in `ts/combat-tracker.ts` to allow selecting monster/NPC combat entries as attackers with configurable attack bonus and damage dice.
- [x] 2.4 Add the **End Encounter & Award Rewards** button (`data-testid="combat-end-encounter-btn"`) and modal in `ts/combat-tracker.ts` to tally defeated monster XP, award XP to party characters, optionally deposit generated treasure into Party Inventory, and clear the encounter.

## 3. Verification & Quality Gates

- [x] 3.1 Add Go handler tests in `handlers/combat_automation_test.go` for monster attacker resolution and encounter clearing.
- [x] 3.2 Add Vitest unit tests and Playwright E2E tests in `tests/` covering `combat-statblock-toggle` and `combat-end-encounter-btn`.
- [x] 3.3 Run `openspec validate combat-statblocks-and-encounter-rewards --strict` and `task ci`.
