## Context

The villum dice roller has a custom Go dice parser (`handlers/dice.go`) that handles only basic expressions like `2d6+3`. It parses by splitting on `+`/`-`, detecting `d` in each term, and rolling via `crypto/rand`. There is no support for RPG standard notation: keep/drop highest/lowest (`kh`/`kl`), exploding (`!`), reroll (`r`/`ro`), success counting (`>t`/`<t`), or compound operations.

On the frontend, only the d6 gets a 3D CSS cube with pip faces and proper rotation to show the correct face-up value. All other dice (d4, d8, d10, d12, d20, d100) use a flat 2D fallback — a styled `<div>` showing just the number. The 3D infrastructure (`perspective: 600px`, `transform-style: preserve-3d`, face transforms) is already in the CSS.

The app is a single-page HTMX + Bootstrap app with Go backend. All dice state (expression input, result, history) is rendered via client-side JS.

## Goals / Non-Goals

**Goals:**
- Replace the custom dice parser with `@dice-roller/rpg-dice-roller` for full RPG notation support
- Evaluate dice expressions on the backend (server-authoritative roll results)
- Build 3D CSS polyhedral shapes for d4, d8, d10, d12, d20, and d100 with proper face-up rotation
- Add face markers (pips, numerals, or symbols) appropriate to each die type
- Preserve existing features: advantage/disadvantage, roll history, crit detection, character-linked rolls
- Keep the same API contract shape so minimal frontend refactoring is needed beyond the 3D rendering

**Non-Goals:**
- Not switching to a WebSocket or streaming roll model — remains request/response
- Not adding real-time shared rolling (multiplayer) in this change
- Not replacing the dice roll database schema or history API
- Not adding 3D physics simulation (CSS animations only, no Three.js/WebGL)

## Decisions

### 1. Backend evaluation via `dice-roller` Go integration vs. frontend evaluation

**Decision:** Evaluate on the backend by wrapping the JavaScript `@dice-roller/rpg-dice-roller` library. Two options:
- **(A)** Embed a JS runtime (goja) to call the npm library server-side
- **(B)** Port the notation parser to Go as a thin adapter

**Chosen:** Option (A) using `goja` — a pure Go JS runtime with no CGo dependency. This keeps the dice logic in a single well-tested library, avoids maintaining a Go port, and allows us to upgrade the library independently. The overhead of embedding a small JS evaluation per roll is negligible.

**Alternatives considered:**
- Frontend-only evaluation was rejected because rolls must be server-authoritative (stored in history, used for character rolls)
- Porting the library to Go was rejected because it adds maintenance burden and risks divergence from the upstream

### 2. 3D CSS for all polyhedral dice vs. SVG/Canvas

**Decision:** Pure CSS 3D transforms. The existing d6 infrastructure shows CSS 3D works well for dice — `perspective`, `preserve-3d`, face transforms, and `backface-visibility`. Complex shapes can be approximated:

| Die | CSS 3D Approach |
|-----|-----------------|
| d4  | Triangular pyramid — 4 triangular faces via `clip-path` |
| d6  | Cube — already implemented (6 square faces) |
| d8  | Octahedron — 8 triangular faces, 2 square pyramids base-to-base |
| d10 | Pentagonal trapezohedron — 10 kite-shaped faces |
| d12 | Dodecahedron — 12 pentagonal faces |
| d20 | Icosahedron — 20 triangular faces |

**Chosen:** CSS 3D with polyhedron face layouts. Each die type gets a dedicated CSS class and face-transform set. The approach:
- Faces are `div` elements positioned in 3D space using `transform: rotateX/Y/Z(...) translateZ(...)`
- Rotation to show the result face uses pre-calculated angle maps (like the existing `rotateToFace` for d6)
- Non-cube faces use `clip-path` for triangular/pentagonal shapes where needed

**Alternatives considered:**
- SVG/Canvas (faster rendering but loses the 3D depth effect)
- Three.js/WebGL (overkill for dice — adds significant bundle size)

### 3. `@dice-roller/rpg-dice-roller` integration for advantage/disadvantage

**Decision:** Map advantage/disadvantage to `2d20kh1` / `2d20kl1` notation directly, removing the separate backend path. The library's output will include both rolls and which was kept, allowing the frontend to highlight the kept die.

### 4. API contract evolution

**Decision:** Extend the existing `/api/roll` endpoint. The `expression` field will accept full RPG notation. The response `DiceResult` will add an optional `notation` field with the parsed/modified notation string and a `rolls_detail` array for richer breakdown info. Backward compatible — old expressions still work.

## Risks / Trade-offs

- **CSS 3D complexity for non-cube dice**: Polyhedral shapes (d4, d8, d10, d12, d20) require careful trigonometry for face positioning. **Mitigation**: Start with the simpler shapes (d8 as double pyramid, d20 as icosahedron) and iterate. The d10 pentagonal trapezohedron is the hardest — may use a simplified approximation.
- **goja performance**: Each roll spins up a small JS evaluation. **Mitigation**: Pre-compile the dice-roller JS source in goja at startup, cache the program, and call into it per-roll. Measured overhead should be <5ms per roll.
- **Bundle size**: `@dice-roller/rpg-dice-roller` is small (~20KB minified). Acceptable.
- **Accessibility**: CSS 3D dice are visual-only. **Mitigation**: The roll result text (`result.text`) and numeric breakdown will always be present as accessible fallback.
- **Reduced motion**: Users with `prefers-reduced-motion` should get static dice showing the result immediately without animation. The existing `@media (prefers-reduced-motion: no-preference)` pattern already exists for some animations.

## Open Questions

- Should dice-roller be bundled for frontend use too (for instant preview/validation) or only server-side?
- How should the d4 be oriented for best visual clarity (triangular base on table vs. point up)?
- Should the d10 show 0-9 or 1-10? (RPG standard is 0-9 for percentile, 1-10 for standalone)
