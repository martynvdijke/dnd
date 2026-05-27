## Why

The dice roller currently uses a custom parser limited to basic expressions (`2d6+3`), and only the d6 gets a full 3D CSS cube animation—all other dice (d4, d8, d10, d12, d20, d100) are rendered as flat 2D tiles. Adding `@dice-roller/rpg-dice-roller` brings full RPG notation support (keep/drop, exploding, reroll, target, counting successes) while extending the 3D CSS treatment to all polyhedral dice creates a polished, immersive experience worthy of a D&D companion app.

## What Changes

- **Replace custom dice parser** with `@dice-roller/rpg-dice-roller` on the backend (Go port or npm-side evaluation) for full RPG notation: `4d6kh3`, `2d20r1`, `3d6!`, `8d10>7`, `1d20ro<2`, etc.
- **Extend 3D CSS animations** to all polyhedral dice (d4, d8, d10, d12, d20, d100) — currently only d6 has 3D cube rendering; others are flat 2D fallback
- **Add die face markers** (pips for d6, numerals/characters for other dice faces)
- **Update expression input** to accept and validate full RPG notation
- **Preserve** existing advantage/disadvantage, history, and crit detection — integrate them with the new library

## Capabilities

### New Capabilities
- `rpg-dice-notation`: Full RPG dice notation parsing and evaluation engine on the backend, supporting keep/drop, exploding, reroll, target success, and compound operations
- `3d-polyhedral-dice`: 3D CSS-animated rendering for all standard polyhedral dice (d4, d6, d8, d10, d12, d20, d100) with die-face markers

### Modified Capabilities
- *(none — the dice roller was built ad-hoc, not previously spec'd)*

## Impact

- **Backend** (`handlers/dice.go`): Replace `rollDice()` parser with `@dice-roller/rpg-dice-roller` Go equivalent or evaluate expressions via the npm library. The `DiceRequest`/`DiceResult` types may need extension to carry additional notation metadata.
- **Frontend** (`ts/app.ts`): Major update to `renderDiceTab()`, `animateDiceRoll()`, `settleDice()`, `build3DCube()` — rename to `build3DDie()` and support all polyhedral shapes. Update `doRoll()` validation.
- **CSS** (`static/style.css`): New 3D shapes for d4 (tetrahedron), d8 (octahedron), d10 (pentagonal trapezohedron), d12 (dodecahedron), d20 (icosahedron). Each needs unique face layout and rotation transforms.
- **Dependencies**: Add `@dice-roller/rpg-dice-roller` to `package.json`. Evaluate Go-side approach (embedding vs shipping notation to frontend).
- **Tests** (`tests/dice.spec.ts`): Add tests for RPG notation expressions and new 3D dice rendering.
