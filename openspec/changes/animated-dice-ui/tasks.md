## 1. Setup & Dependencies

- [x] 1.1 Add `@dice-roller/rpg-dice-roller` to `package.json`
- [x] 1.2 Add `goja` (Go JS runtime) dependency to `go.mod`
- [x] 1.3 Create dice engine wrapper that embeds `@dice-roller/rpg-dice-roller` via goja and exposes a `Evaluate(expression string) (*DiceResult, error)` function
- [x] 1.4 Write Go unit tests for the dice engine wrapper (basic expressions, keep/drop, exploding, reroll, edge cases)

## 2. Backend: Replace Dice Parser

- [x] 2.1 Replace `rollDice()` in `handlers/dice.go` to call the new dice engine instead of the custom parser
- [x] 2.2 Extend `DieRollDetail` / `BreakdownGroup` types to include notation metadata (kept/dropped flags, exploded rolls, reroll history)
- [x] 2.3 Refactor advantage/disadvantage handler to map to `2d20kh1` / `2d20kl1` notation using the new engine
- [x] 2.4 Update API response format and ensure backward compatibility
- [x] 2.5 Run existing dice tests to verify nothing broke

## 3. Backend: Validate and Save

- [x] 3.1 Add input validation for RPG notation expressions (reject syntactically invalid expressions with clear error messages)
- [x] 3.2 Verify roll history correctly stores full notation expressions (no schema changes needed)
- [x] 3.3 Update `GetDiceRolls` handler if breakdown format changed (ensure frontend can still parse)

## 4. Frontend: 3D Polyhedral Shapes (CSS)

- [x] 4.1 Build d20 icosahedron — 20 triangular faces with CSS `clip-path` and 3D transforms; implement `rotateToFace(face, 20)` rotation mapping
- [x] 4.2 Build d12 dodecahedron — 12 pentagonal faces with CSS `clip-path` and 3D transforms
- [x] 4.3 Build d8 octahedron — 8 triangular faces, double square pyramid layout
- [x] 4.4 Build d4 tetrahedron — 4 triangular faces, triangular pyramid layout
- [x] 4.5 Build d10 pentagonal trapezohedron (or reasonable CSS approximation) — 10 faces
- [x] 4.6 Ensure d10 and d100 share the same shape but use distinct color themes
- [x] 4.7 Update `build3DCube` → `build3DDie` to dispatch to the correct shape builder per die type

## 5. Frontend: Animation System

- [x] 5.1 Extend rolling animation CSS (`diceTumble` keyframes) to vary per die shape (different tumble patterns for d4 vs d20)
- [x] 5.2 Add `prefers-reduced-motion` support: skip rolling animation, show final state immediately
- [x] 5.3 Update `animateDiceRoll` to pass correct die type info to the animation
- [x] 5.4 Update `settleDice` to rotate each die type to its correct face-up orientation
- [x] 5.5 Verify crit effects (gold/red glow) work on d20 across all die-shape code paths

## 6. Frontend: Notation UI

- [x] 6.1 Update expression input placeholder and help text to indicate RPG notation support (e.g., "2d6+3, 4d6kh3, 1d20!")
- [x] 6.2 Add input validation/feedback for RPG notation syntax on the frontend (basic regex check before sending)
- [x] 6.3 Add quick-preset buttons for common notation patterns (advantage, disadvantage, 4d6kh3 for stats, etc.)
- [x] 6.4 Ensure result breakdown display works with new notation metadata (kept/dropped indicators, exploded markers)

## 7. Tests & Verification

- [x] 7.1 Update `tests/dice.spec.ts` — add tests for RPG notation expressions, 3D dice rendering visibility, result breakdown display
- [x] 7.2 Run full test suite: `go test -tags fts5 ./dice/... ./handlers/...` — all pass. Frontend `vite build` passes.
- [x] 7.3 Manual QA: test all die types (d4-d100), test edge cases (exploding, reroll, keep/drop), test advantage/disadvantage
- [x] 7.4 Verify reduced-motion behavior via browser DevTools
