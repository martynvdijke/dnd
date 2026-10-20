## Context

In `ts/characters/stats.ts`, lines 54–55 render:
```html
<button class="btn btn-sm btn-outline-success" onclick="rollCheck('check','str','advantage')">Advantage</button>
<button class="btn btn-sm btn-outline-danger" onclick="rollCheck('check','str','disadvantage')">Disadvantage</button>
```
Clicking either button immediately triggers a Strength check rather than toggling advantage/disadvantage for the skill or save the player wants to roll.

In `ts/characters/combat.ts`, lines 158–163 compute saving throws as:
```ts
const mod = getAbilityMod((c[s] as number) || 10);
const total = ((c['proficiency_bonus'] as number) || 0) + mod;
```
which erroneously adds `proficiency_bonus` to all 6 saving throws. Additionally, `doRest('short')` hardcodes `hit_dice_count: 1`, and equipped weapons are only rendered in `ts/characters/inventory.ts` (`#loadoutPanel`), leaving the Combat tab without weapon attack actions.

## Goals / Non-Goals

**Goals:**
- Provide a stateful Roll Mode (`normal` | `advantage` | `disadvantage`) that applies to ability checks, skill checks, saving throws, initiative rolls, and weapon attack rolls from the Character Sheet.
- Compute saving throw modifiers accurately using the character's 5e class proficiencies (and any explicit `saving_throw_proficiencies` list on the character record).
- Surface equipped weapon attacks (To-Hit and Damage rolls) directly on the Combat tab.
- Allow choosing how many Hit Dice (`1..hit_dice_remaining`) to spend when taking a Short Rest.
- Group the 15 sheet tabs under quick filter pills (`All`, `Play`, `Build`, `World`, `Insights`) without breaking existing tab selectors.

**Non-Goals:**
- Altering backend database columns for `characters` (saving throw proficiencies can be derived from `class` + optional JSON field).
- Removing any of the 15 existing sheet tab buttons.

## Decisions

1. **Stateful Roll Mode (`rollMode` in `ts/lib/state.ts`)**:
   - *Decision*: Add `rollMode: 'normal' | 'advantage' | 'disadvantage'` (default `'normal'`) to `ts/lib/state.ts` and expose `setRollMode(mode)` on `window`. When `rollCheck(type, ability, mode)` is called without an explicit `mode` override, it reads `state.rollMode`. Clicking an already-active `Advantage` or `Disadvantage` button toggles back to `'normal'`.
   - *Rationale*: Fixes the hardcoded `'str'` bug while keeping the exact same `Advantage` / `Disadvantage` button placement and making every check, save, and skill button respect the active roll mode.

2. **5e Class Saving Throw Proficiency Map**:
   - *Decision*: Define `CLASS_SAVE_PROFICIENCIES: Record<string, string[]>` in `ts/characters/combat.ts` covering all standard 5e classes (Barbarian: `str, con`; Bard: `dex, cha`; Cleric: `wis, cha`; Druid: `int, wis`; Fighter: `str, con`; Monk: `str, dex`; Paladin: `wis, cha`; Ranger: `str, dex`; Rogue: `dex, int`; Sorcerer: `con, cha`; Warlock: `wis, cha`; Wizard: `int, wis`; Artificer: `con, int`), plus any entries in `c.saving_throw_proficiencies`. Only proficient saves add `proficiency_bonus`.
   - *Rationale*: Immediately fixes the 5e rules bug for all characters with zero schema migration required.

3. **Equipped Weapons Panel on the Combat Tab**:
   - *Decision*: Reuse the character's loaded inventory items (`equipped === true` and `category === 'weapon'` or `damage_dice` present) to render an `Equipped Weapons` section in `renderCombatTab()` with 1-click To-Hit (`rollCheck('attack', ...)`) and Damage (`POST /api/roll` with weapon `damage_dice` + ability mod) buttons.
   - *Rationale*: Eliminates switching between the Combat and Inventory tabs on every turn.

4. **Configurable Short Rest Hit Dice**:
   - *Decision*: Add a numeric input (`#shortRestHitDiceInput`, `data-testid="short-rest-hit-dice-input"`, clamped to `1..Math.max(1, hitDiceRemaining)`) next to the `Short Rest` button in `renderCombatTab()`, and pass its value in `doRest('short')`.
   - *Rationale*: Zero modal friction while letting players spend multiple Hit Dice on a Short Rest.

## Risks / Trade-offs

- **[Risk] Existing tests expecting `Advantage` button click behavior** → *Mitigation*: Verify all existing tests in `tests/` and `ts/`; update/add unit and E2E tests for `setRollMode` and saving throw proficiency calculation.
- **[Risk] Strict `data-testid` lint gate** → *Mitigation*: Reference every new `data-testid` in Playwright E2E tests (`tests/`).
