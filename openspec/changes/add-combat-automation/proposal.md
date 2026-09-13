## Why
Villum's combat tracker stores HP, AC, initiative and conditions, but resolving an attack still happens in the players' heads: roll to hit, compare to AC by hand, roll damage, do the subtraction, remember concentration, add the condition. Damage and healing are applied client-side (`ts/characters/combat.ts` `applyHeal`, app-level `applyDamage`) with no server authority. That makes the tracker undependable and leaves no audit trail. This change adds a small, deterministic rules layer that resolves a single attack, save, or damage event server-side and records it.

## What Changes
- Add a derived **attack bonus** for inventory weapons (ability modifier + proficiency + magic, with an explicit override), since `InventoryItem` stores `damage_dice`/`damage_type` but no to-hit.
- Add `POST /api/combat/attack`: rolls d20+bonus vs a target's AC, rolls weapon damage, reports hit/miss/critical, and optionally applies it.
- Add `POST /api/characters/:id/hp`: the single authoritative damage/heal path — temp HP absorption, 0..max clamp, death-save increments at 0, concentration trigger.
- Extend saving throws with a DC-aware `POST /api/roll/save-vs-dc` returning success/failure (and optional half-on-save).
- Auto-apply a failed-save condition (respecting condition immunities) via the existing condition model.
- Roll the concentration save automatically when a concentrating character takes damage; drop concentration on failure.
- Tick round-based conditions from `next-turn`.
- Route every automated resolution through the existing combat log.

## Capabilities
### New Capabilities
- `combat-automation`: deterministic attack/save/damage/condition/concentration resolution with an audit trail.

### Modified Capabilities
<!-- None: existing dice, combat, conditions and concentration APIs keep their current requirements; automation is additive. -->

## Impact
Backend: `handlers/combat.go`, new `handlers/combat_automation.go`, `handlers/hpcalc.go`, `handlers/conditions.go`, `handlers/concentration.go`, `handlers/check.go`, `handlers/dice.go`, `handlers/combat_log.go`, `models/character.go` (attack override), one additive migration. Frontend: `ts/combat-tracker.ts`, `ts/characters/combat.ts`. Reuses the `dice` pool; no new dependencies.
