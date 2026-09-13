## Context

The roller is `dice.Pool.Roll(expr)` (`dice/engine.go:75`), wrapped by `getDicePool()` (`handlers/dice.go:38`). `handlers/check.go:61` (`HandleCheckRoll`) already resolves ability→modifier, proficiency, advantage/disadvantage and writes to `dice_rolls` — the template for any new roll. Characters carry `ac`, `hp_max`, `hp_current`, `temp_hp`, `spell_save_dc`, `spell_attack_bonus`, `concentrating_on`, `exhaustion`, `death_saves` (`models/character.go`). Weapons are `InventoryItem{damage_dice, damage_type, weapon_properties, ac_bonus, is_magical}` with **no to-hit field**. Combat state lives in `combat_entries` (`handlers/combat.go:14`) and `handlers/combat_log.go` already records combat events. Conditions are `character_conditions{name,type,source,duration,duration_type,saving_throw,save_dc}` with `TickConditions` (`conditions.go:88`) and 16 standard types (`conditions.go:124`). `CheckConcentration` (`concentration.go:19`) computes the save DC but never rolls or breaks concentration. Damage/heal currently live in the browser, so the server has no idea it happened.

## Goals / Non-Goals

**Goals**
- One call resolves an attack: to-hit vs AC, crit/fumble, damage on hit, applied atomically.
- One authoritative damage/heal path that handles temp HP, clamping, death saves, and concentration.
- Saving throws that know their DC and outcome; failed saves can apply conditions.
- Every automated resolution is logged and returns its full breakdown for the DM to trust.

**Non-Goals**
- Parsing freeform monster `actions` text into structured attacks — monster attacks are resolved from manual input in this change.
- Full spell effect automation (areas, upcasting, multi-target) beyond a single save/damage.
- Grid positioning, cover, or line-of-sight.
- Auto-rolling the *attacker's* exact weapon selection UI beyond passing an inventory item id.

## Decisions

### 1. Attack bonus is derived, with an explicit override
Add `attack_ability` (and optional `attack_bonus`) to `InventoryItem`. Default derivation: melee → STR, ranged/finesse (from `weapon_properties`) → DEX, plus proficiency bonus by level and `is_magical` bonus; an explicit stored value wins.

**Why:** Existing data has damage dice but no to-hit; deriving with a sensible default keeps imports working, and the override prevents silent wrong math for magic/unusual weapons. **Alternative:** require every weapon to store a bonus — rejected (breaks existing inventory).

### 2. One `POST /api/combat/attack` endpoint, hit resolution separate from application
The handler accepts `{attacker_type, attacker_id, target_type, target_id, item_id|attack_bonus, damage_dice, advantage, apply}`. It rolls the d20 (reusing the `HandleCheckRoll` advantage pattern), compares to target AC (`combat_entries.ac` or `characters.ac`), rolls damage on hit, and when `apply` is true calls the damage path. Natural 20 is a crit (double damage dice); natural 1 is a miss.

**Why:** Keeps "show me the roll" and "apply it" in one auditable request while letting the UI preview before committing. **Alternative:** separate roll + apply calls — rejected (racy, two round-trips).

### 3. `POST /api/characters/:id/hp` is the single mutation point for HP
Input `{delta, type, source}`; behaviour: absorb into `temp_hp` first, clamp `hp_current` to `0..hp_max`, increment death saves when reaching 0, reset on healing, and invoke the concentration trigger on damage. The browser's `applyDamage`/`applyHeal` are rewritten to call it.

**Why:** Server authority removes client drift and makes combat-log entries truthful. **Alternative:** keep applying client-side — rejected (no authority, no audit).

### 4. Save-vs-DC is a thin extension of the existing check roller
`POST /api/roll/save-vs-dc {character_id, ability, dc, advantage, half_on_save}` reuses `HandleCheckRoll`, adds the DC comparison, and returns `{success, total, dc}`; when `half_on_save` and success, the caller applies half damage via the HP path.

### 5. Conditions and concentration reuse their existing models
A failed save / hit effect calls the existing `CreateCondition` (`conditions.go:35`) after checking `condition_immunities`; concentration failure reuses the existing DC calculation, rolls a CON save through the check roller, and on failure deletes the concentration condition and clears `concentrating_on`. `NextTurn` additionally calls `TickConditions` with `duration_type=round`.

**Why:** No new condition machinery; the immunity check is the only new guard.

### 6. Everything is written to the combat log
Each automated resolution appends to `combat_log` with the breakdown (rolls, modifiers, hit/miss/crit, damage, condition, concentration result).

## Risks / Trade-offs

- **Wrong derivation of attack bonus silently corrupts play** → explicit override field; unit tests for finesse/ranged/magic/multiclass proficiency.
- **Damage path becomes a chokepoint** → covered by focused tests: temp HP, overheal, 0-crossing death saves, concentration trigger, damage at 0.
- **Monster attacks are unparsed freeform text** → out of scope; DM enters the attack manually, automation still resolves and applies it.
- **Double application (UI calls it twice)** → HP endpoint is idempotency-free but every call is logged; UI disables the control while in flight. Revisit with an idempotency key only if it proves a real problem.

## Migration Plan
Additive migration for `inventory.attack_ability` / `attack_bonus` (nullable). New endpoints ship alongside the current client math; the client is switched to the HP endpoint in the same release. Rollback: revert the client switch (endpoints unused) and drop the columns.

## Open Questions
- Should crit damage double the dice or the total? (Proposal: double the dice, 5e default.)
- Should the attack endpoint accept multiple targets for multiattack? (Deferred; one target per call.)
