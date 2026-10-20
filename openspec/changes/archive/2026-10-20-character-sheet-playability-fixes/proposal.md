## Why

During live play, the Character Sheet (`ts/characters/sheet.ts`, `ts/characters/stats.ts`, `ts/characters/combat.ts`) has several mechanical bugs and UX bottlenecks:
1. **Hardcoded Advantage/Disadvantage buttons (`ts/characters/stats.ts:54-55`)**: Clicking "Advantage" or "Disadvantage" calls `rollCheck('check','str','advantage')`, which always rolls a Strength check instead of setting the roll mode (`Normal | Advantage | Disadvantage`) for the player's next skill check, ability check, saving throw, or weapon attack.
2. **Unconditional Saving Throw Proficiency (`ts/characters/combat.ts:158-163`)**: Saving throws add `proficiency_bonus` to all 6 abilities (`const total = ((c['proficiency_bonus'] as number) || 0) + mod`) regardless of which saving throws the character is actually proficient in.
3. **Missing Weapon Attack Cards on the Combat Tab**: Equipped weapons only appear in the `Inventory` loadout panel, forcing players in combat to switch tabs or use the DM's Combat Tracker modal to roll weapon attacks and damage.
4. **Hardcoded Short Rest Hit Dice (`ts/characters/combat.ts:61`)**: Clicking `Short Rest` always sends `hit_dice_count: 1` without letting the player choose how many remaining Hit Dice to spend.
5. **15 Uncategorized Sheet Tabs**: Switching across 15 flat tabs creates cognitive overload during sessions.

## What Changes

- Replace the hardcoded Strength Advantage/Disadvantage buttons in `ts/characters/stats.ts` with a persistent **Roll Mode Toggle** (`Normal`, `Advantage`, `Disadvantage`) stored in character sheet state and respected by `rollCheck()` (ability checks, skill checks, saving throws, initiative) and weapon attacks.
- Fix saving throw calculation in `ts/characters/combat.ts` to apply `proficiency_bonus` only to saving throws in which the character is proficient (derived from standard 5e class proficiencies or explicit `saving_throw_proficiencies` overrides), with visual proficiency indicator dots.
- Add an **Equipped Weapons & Quick Attacks** card to the Combat tab (`ts/characters/combat.ts`) displaying equipped weapons from the character's inventory with 1-click **Attack** (`1d20 + attack bonus`) and **Damage** roll buttons.
- Add a **Hit Dice Count selector** (`1..hit_dice_remaining`) for Short Rests in `ts/characters/combat.ts` so players can spend multiple Hit Dice at once.
- Add **Sheet Mode Filter Pills** (`All`, `Play`, `Build`, `World`, `Insights`) above the character sheet tab bar (`static/app.html`, `ts/characters/sheet.ts`) to declutter the 15 tabs during play while keeping all existing tab buttons in the DOM.

## Capabilities

### New Capabilities

- `character-sheet-playability`: Stateful roll mode toggle (`normal`, `advantage`, `disadvantage`), class-accurate saving throw proficiencies, equipped weapon quick-attack cards on the Combat tab, configurable Short Rest Hit Dice spending, and grouped sheet tab filtering.

### Modified Capabilities

<!-- None -->

## Impact

- **Frontend (`ts/lib/state.ts`, `ts/characters/stats.ts`, `ts/characters/combat.ts`, `ts/characters/sheet.ts`, `static/app.html`)**:
  - Stateful `rollMode` in `ts/lib/state.ts` wired into `rollCheck` and weapon attack rolls.
  - Class saving throw proficiency lookup and indicator badges in `ts/characters/combat.ts`.
  - Equipped weapon attack cards and Short Rest Hit Dice input on the Combat tab.
  - Category filter pills above `#sheetTabs` in `static/app.html` and `ts/characters/sheet.ts`.
- **Tests (`ts/**/*.test.ts`, `tests/`)**:
  - Unit tests for saving throw proficiency derivation, roll mode propagation, and short rest hit dice count.
  - E2E tests covering all new `data-testid` attributes.
