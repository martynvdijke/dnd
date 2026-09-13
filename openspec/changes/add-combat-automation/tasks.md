## 1. Data model

- [x] 1.1 Add nullable `attack_ability` and `attack_bonus` to the inventory item schema + additive migration.
- [x] 1.2 Add attack-bonus derivation (STR melee / DEX ranged+finesse, proficiency by level, magic bonus, explicit override) as a testable helper.

## 2. HP authority

- [x] 2.1 Add `POST /api/characters/:id/hp {delta,type,source}` with temp-HP absorption, clamping, death-save transitions, and concentration trigger.
- [x] 2.2 Switch `ts/characters/combat.ts` `applyHeal`/`applyDamage` to call the new endpoint.

## 3. Resolution endpoints

- [x] 3.1 Add `POST /api/combat/attack` (d20+bonus vs AC, crit/fumble, damage roll, optional apply) reusing the check-roll advantage pattern.
- [x] 3.2 Add `POST /api/roll/save-vs-dc` extending `HandleCheckRoll` with DC comparison and optional half-on-save.
- [x] 3.3 Add condition application after failed saves/hit effects with immunity checks.
- [x] 3.4 Roll the concentration save automatically on damage and drop concentration on failure.
- [x] 3.5 Tick round-based conditions from `next-turn`.

## 4. Audit

- [x] 4.1 Write every automated resolution to `combat_log` with its breakdown.

## 5. Frontend

- [x] 5.1 Combat tracker: attack flow (pick attacker/target/weapon, preview, apply).
- [x] 5.2 Show save/condition/concentration outcomes in the combat log view.

## 6. Tests & verification

- [x] 6.1 Unit tests: attack hit/miss/crit/nat-1; bonus derivation (finesse, ranged, magic, override).
- [x] 6.2 Unit tests: temp HP, overheal, 0-crossing death saves, concentration trigger/break/hold.
- [x] 6.3 e2e: resolve an attack from the tracker and see HP and the log update.
- [x] 6.4 `go vet ./...`, `task test`, `npm run typecheck`, `npm run build:vite`, `task lint:e2e`.
