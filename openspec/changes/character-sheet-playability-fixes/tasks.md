## 1. Roll Mode State & Saving Throw Proficiency Fixes

- [x] 1.1 Add `rollMode` (`'normal' | 'advantage' | 'disadvantage'`) to `ts/lib/state.ts` and update `ts/characters/stats.ts` and `ts/characters/combat.ts` so `Advantage` / `Disadvantage` buttons set the active roll mode for all checks, saves, skills, and attacks.
- [x] 1.2 Add 5e class saving throw proficiency lookup (`isSaveProficient`) in `ts/characters/combat.ts` so only proficient saving throws add `proficiency_bonus` and render a proficiency indicator badge.

## 2. Combat Tab Quick Attacks, Short Rest Hit Dice & Tab Grouping

- [x] 2.1 Render an `Equipped Weapons` quick-attack card in `renderCombatTab()` (`ts/characters/combat.ts`) with 1-click To-Hit and Damage roll buttons for equipped weapons.
- [x] 2.2 Add a Hit Dice count input (`data-testid="short-rest-hit-dice-input"`) to the Rest panel in `ts/characters/combat.ts` and wire `doRest('short')` to submit the selected count.
- [x] 2.3 Add sheet tab category filter pills (`All`, `Play`, `Build`, `World`, `Insights`) in `static/app.html` and `ts/characters/sheet.ts`.

## 3. Verification & Quality Gates

- [x] 3.1 Add Vitest unit tests covering `isSaveProficient()`, `setRollMode()`, equipped weapon quick-attack rendering, and Short Rest Hit Dice count submission.
- [x] 3.2 Add Playwright E2E tests in `tests/` covering all new `data-testid` selectors (`roll-mode-advantage`, `roll-mode-disadvantage`, `short-rest-hit-dice-input`, `combat-quick-attacks`, `data-filter` tab filter pills).
- [ ] 3.3 Run `openspec validate character-sheet-playability-fixes --strict` and `task ci`.
