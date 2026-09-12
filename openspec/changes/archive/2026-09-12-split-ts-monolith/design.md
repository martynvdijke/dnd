## Context

`ts/app.ts` is 4067 lines and `ts/admin.ts` is 2730 lines; together 39% of all `ts/` lines (17,402). Both already have `// ─── Section ───` seam comments marking domain boundaries.

`app.ts` seams (line anchors):
Sort L51, Campaign/Character switching L112, Character Sheet L129, Roll/Combat Actions L131, Features L373, Proficiencies L430, Details L464, D3 Force Graph L553, Graph L730, Analytics Chart.js L752, New Character L843, Import/Export L877, Delete/Logout/Portrait L954-980, Race Colors L1017, Multi-Class L1074, Calendar L1141, Conditions/Feats/Companions L1146-1464, Notes TipTap L1534, Shops L1646-2312 (~666 lines), Random Gen/Comparison L2479-2554, Wiki/Campaign Graph L2602-2816, One-Shot Tree/Items/Shops/Monsters/NPCs L2816-3517 (~700 lines), Polymorphic File Uploads L3517, Campaign Dashboard/Party Inventory/Session Planner L3553-3700, Encounter Difficulty/Treasure L3796-3920.

`admin.ts` seams:
E-ink L119, AI Features L144, API Tokens L194, Unified Compendium L325, Global Search L365, Entry Browser L446, Bulk Ops L571, Entry CRUD L692, Schema-Aware Editor L786, Import Integration L922, Logs L945-1146, Users L1147, Schemas L1263, Backup/Email/Push/Umami/OTel/AI settings L1382-1593, Import Wizard L1716-2090, Utils/PDF/Events/Camera L2091-2508.

Extraction precedent already exists: prior change extracted `ts/party.ts`, `ts/combat-tracker.ts`, `ts/compendium.ts`, `ts/encounter.ts`, `ts/characters/*.ts`, `ts/fab.ts` etc. using window-level self-registration via `expose()` from `ts/lib/expose.ts` — module self-registers on import; `init.ts` imports modules. Duplication risk: some functions exist both in `app.ts` and `ts/characters/*` and must be reconciled. Cross-cutting patterns: 23× `setCurrentChar(await api('GET', /api/characters/${id}))` and ~40× `catch(e){toast(e.message,true)}`.

`restore-ts-type-safety` handles `@ts-nocheck` removal separately — this change keeps existing pragmas on moved code where needed.

## Goals / Non-Goals

**Goals:**
- Reduce `app.ts` and `admin.ts` to < 800 lines each (thin orchestrators)
- Preserve behavior byte-identically through each extraction phase
- Follow the established `expose()` self-registration pattern — no new module system
- Reconcile `app.ts` ↔ `ts/characters/*` duplication during extraction
- Extract `refreshChar()` and `renderError()` helpers to de-duplicate cross-cutting patterns

**Non-Goals:**
- Removing `// @ts-nocheck` from moved code (belongs to `restore-ts-type-safety`)
- Changing runtime behavior, APIs, or bundle count (still 5 vite IIFE bundles)
- Introducing a new state-management framework or build-system change
- Splitting `ts/party.ts`, `ts/compendium.ts`, etc. further (already modular)

## Decisions

### 1. Follow the existing expose() self-registration pattern

Each extracted module imports `expose` from `ts/lib/expose.ts`, registers its public functions on `window` (e.g. `expose('renderWiki', renderWiki)`), and is imported for side effects in `ts/init.ts` (or `ts/app.ts` barrel). No ES re-exports or new dependency injection.

**Why:** Consistency with prior extractions (`party.ts`, `combat-tracker.ts`, `characters/*`, `fab.ts`). Zero new abstraction to learn. **Alternative:** ES module re-exports and explicit imports — cleaner but inconsistent with current codebase and would require touching every call site at once. Deferred to a later import-hygiene pass.

### 2. Domain-clustered phases, not one big move

Extract one domain cluster per phase, verify, then next. Proposed `app.ts` phase order (smallest/leafiest first to de-risk):

1. Helpers — `refreshChar()` + `renderError()` (touches many files, land first so later phases use them)
2. Sort (L51) + Campaign/Character switching (L112) + New Character (L843) + Import/Export (L877) + Delete/Logout/Portrait (L954-980)
3. Features/Proficiencies/Details (L373-464) + Conditions/Feats/Companions (L1146-1464) + Character Sheet (L129) + Roll/Combat Actions (L131)
4. Race Colors (L1017) + Multi-Class (L1074) + Calendar (L1141) + Notes TipTap (L1534)
5. Shops (L1646-2312) + Random Gen/Comparison (L2479-2554)
6. Wiki/Campaign Graph (L2602-2816) + Polymorphic File Uploads (L3517)
7. One-Shot Tree/Items/Shops/Monsters/NPCs (L2816-3517) + Campaign Dashboard/Party Inventory/Session Planner (L3553-3700) + Encounter Difficulty/Treasure (L3796-3920) + D3/Graph/Analytics (L553-752)

`admin.ts` phase order:

1. E-ink + AI Features + API Tokens (L119-325)
2. Unified Compendium + Global Search + Entry Browser + Bulk Ops (L325-692)
3. Entry CRUD + Schema-Aware Editor + Import Integration (L692-945)
4. Logs + Users + Schemas (L945-1382)
5. Backup/Email/Push/Umami/OTel/AI settings (L1382-1593) + Import Wizard (L1716-2090) + Utils/PDF/Events/Camera (L2091-2508)

Each phase: move code → wire `expose()` + `init.ts` imports → `npm run typecheck` → `npm run test:unit` → `task test:e2e`.

**Why:** Small, verifiable increments; bisectable if e2e breaks. **Alternative:** one giant PR — rejected; unreviewable and high risk.

### 3. Reconcile app.ts ↔ ts/characters/* duplication on contact

When an extracted `app.ts` seam duplicates a function already in `ts/characters/*`, keep the `ts/characters/*` version as canonical, delete the `app.ts` copy, and ensure `app.ts` (or its replacement module) imports/uses the canonical one via the `window` bridge or direct import if already modular.

**Why:** `ts/characters/*` was the intentional extraction target; `app.ts` copies are stale forks.

### 4. Extract refreshChar() and renderError() as shared helpers

- `refreshChar(id?: string)` — `setCurrentChar(await api('GET', `/api/characters/${id ?? currentId}`))`
- `renderError(e: unknown)` — `toast(e instanceof Error ? e.message : String(e), true)` (or the hardened `toast` signature after sanitize change)

Place in `ts/lib/refresh.ts` or `ts/lib/dom.ts` as appropriate.

**Why:** 23 and ~40 call sites respectively; extracting early reduces churn in later phases and makes error handling consistent.

### 5. Keep // @ts-nocheck on moved code

Moved modules retain `// @ts-nocheck` if the source had it. Do not fix types during the move — that belongs to `restore-ts-type-safety`.

**Why:** Keeps this change behavior-only and reviewable as pure moves. Mixing type fixes with moves obscures diffs.

### 6. Bundle count and vite config unchanged

The 5 vite IIFE bundles (app, admin, pwa, setup, login) stay as is. Extracted modules are bundled via the existing entry points (`ts/app.ts` barrel re-imports, `ts/admin.ts` barrel). No new entry points unless a module is truly standalone (unlikely).

**Why:** Minimal infra change; avoids `vite.config.ts` churn.

## Risks / Trade-offs

- **Window bridge ordering / race conditions** → Module must be imported before use. Mitigation: `init.ts` import order mirrors old in-file order; verify with e2e after each phase.
- **Hidden coupling via shared mutable state** → `app.ts` globals (`currentChar`, `currentCampaign`) accessed across seams. Mitigation: keep globals in orchestrator `app.ts` or a small `ts/lib/state.ts`; modules access via `window` bridge as before.
- **Duplication reconciliation breaks callers** → Deleting an `app.ts` duplicate may break an untyped caller. Mitigation: grep for function name before deleting; e2e catches runtime breaks.
- **Large git diff obscures review** → Mitigation: `git mv` + minimal edits per phase; reviewer can verify moved code is unchanged (aside from `expose()` wiring).
- **Dependency on restore-ts-type-safety ordering** → If type-safety work lands mid-split, moved files may have diverged. Mitigation: sequence — split first or coordinate; document that this change keeps nocheck pragmas so either order is safe but simultaneous edits to same file should be avoided.

## Migration Plan

1. Land helpers phase (`refreshChar`/`renderError`) first.
2. Extract `app.ts` domain clusters in order above, one phase per PR/commit, verifying `npm run typecheck` + `npm run test:unit` + `task test:e2e` after each.
3. Extract `admin.ts` clusters similarly.
4. Final verification: `app.ts` < 800 lines, `admin.ts` < 800 lines, `npm run build:vite` produces 5 bundles, full e2e green.
5. Rollback per phase: revert the phase commit; each phase is self-contained.

## Open Questions

- Exact file paths for extracted modules — `ts/app/shops.ts` vs `ts/shops.ts` vs `ts/domains/shops.ts`? (Proposal: `ts/app/<domain>.ts` and `ts/admin/<domain>.ts` to keep entry-point adjacency clear.)
- Should `refreshChar`/`renderError` live in `ts/lib/` or `ts/app/helpers.ts`? (Proposal: `ts/lib/` so admin can reuse `renderError`.)
- How to handle D3/Chart.js imports that are only used in analytics/graph seams — keep in orchestrator or move with seam? (Proposal: move with seam; import where used.)
