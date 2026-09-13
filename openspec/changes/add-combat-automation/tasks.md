## 1. Data model

- [ ] 1.1 Add nullable `attack_ability` and `attack_bonus` to the inventory item schema + additive migration.
- [ ] 1.2 Add attack-bonus derivation (STR melee / DEX ranged+finesse, proficiency by level, magic bonus, explicit override) as a testable helper.

## 2. HP authority

- [ ] 2.1 Add `POST /api/characters/:id/hp {delta,type,source}` with temp-HP absorption, clamping, death-save transitions, and concentration trigger.
- [ ] 2.2 Switch `ts/characters/combat.ts` `applyHeal`/`applyDamage` to call the new endpoint.

## 3. Resolution endpoints

- [ ] 3.1 Add `POST /api/combat/attack` (d20+bonus vs AC, crit/fumble, damage roll, optional apply) reusing the check-roll advantage pattern.
- [ ] 3.2 Add `POST /api/roll/save-vs-dc` extending `HandleCheckRoll` with DC comparison and optional half-on-save.
- [ ] 3.3 Add condition application after failed saves/hit effects with immunity checks.
- [ ] 3.4 Roll the concentration save automatically on damage and drop concentration on failure.
- [ ] 3.5 Tick round-based conditions from `next-turn`.

## 4. Audit

- [ ] 4.1 Write every automated resolution to `combat_log` with its breakdown.

## 5. Frontend

- [ ] 5.1 Combat tracker: attack flow (pick attacker/target/weapon, preview, apply).
- [ ] 5.2 Show save/condition/concentration outcomes in the combat log view.

## 6. Tests & verification

- [ ] 6.1 Unit tests: attack hit/miss/crit/nat-1; bonus derivation (finesse, ranged, magic, override).
- [ ] 6.2 Unit tests: temp HP, overheal, 0-crossing death saves, concentration trigger/break/hold.
- [ ] 6.3 e2e: resolve an attack from the tracker and see HP and the log update.
- [ ] 6.4 `go vet ./...`, `task test`, `npm run typecheck`, `npm run build:vite`, `task lint:e2e`.
