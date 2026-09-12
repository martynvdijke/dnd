## Context

`tsconfig.json` has `"strict": true` but 13 files opt out entirely with `// @ts-nocheck`:

`ts/app.ts:1`, `ts/party.ts:1`, `ts/combat-tracker.ts:1`, `ts/compendium.ts:1`, `ts/encounter.ts:1`, `ts/factions.ts:1`, `ts/timeline.ts:1`, `ts/characters/combat.ts:1`, `ts/characters/inventory.ts:1`, `ts/characters/resources.ts:1`, `ts/characters/sheet.ts:1`, `ts/characters/spells.ts:1`, `ts/characters/stats.ts:1` (plus `ts/lib/dom.ts:1` has a narrow `@ts-ignore` for bootstrap types, which is acceptable).

There are 797 `any` occurrences (worst: `app.ts` ~199, `admin.ts` ~141, `party.ts` 57, `party-subtabs.ts` 34). Root causes are `ts/lib/api.ts:26` (`api(method, path, body?: any): Promise<any>` propagates `any` everywhere), `ts/lib/expose.ts:11` (`expose(name, value: any)` + `(window as any)` bridge, ~80 casts), and untyped `(e:any)`/`(r:any)` callbacks. 201× `getElementById(...)!` non-null assertions lack a central helper.

## Goals / Non-Goals

**Goals:**
- Restore compiler checking incrementally without a big-bang rewrite
- Fix systemic `any` sources (`api()`, `expose()`, window bridge, DOM lookups) so leaf-file cleanup is tractable
- Prevent reintroduction of `@ts-nocheck` via CI guard
- Keep behavior byte-identical; types only

**Non-Goals:**
- Removing `@ts-nocheck` from `ts/app.ts` and `ts/admin.ts` (explicitly out of scope; depends on `split-ts-monolith`)
- Achieving zero `any` globally — reduce hotspots, add types where value is high
- Changing runtime APIs or adding runtime validation (e.g. zod) — types are compile-time only this change
- Fixing `admin.ts` any hotspots beyond what `api<T>` naturally improves

## Decisions

### 1. Generic `api<T>()` with shared response types

Change `ts/lib/api.ts:26` from `api(method, path, body?: any): Promise<any>` to `async function api<T>(method: string, path: string, body?: unknown): Promise<T>`. Define shared types (e.g. `ts/lib/api-types.ts` or `ts/types.ts`) for common responses: `Character`, `Campaign`, `CompendiumEntry`, `Paginated<T>`, etc., reusing existing ent/Go shapes where practical.

**Why:** Single fix propagates type safety to every caller. `unknown` for `body` forces callers to type payloads. **Alternative:** per-endpoint typed wrappers (`getCharacter()`, `listCampaigns()`) — rejected as too much boilerplate this phase; generic `api<T>` is minimal and can be wrapped later. **Alternative:** `zod` runtime validation — deferred; compile-time types first.

### 2. Typed Window augmentation for the expose bridge

Add `ts/lib/window.d.ts` (or `ts/types/window.d.ts`) with `declare global { interface Window { /* exposed names */ } }` and tighten `expose(name, value: unknown)` to use typed keys where possible. Replace `(window as any)` casts with `window.<name>`.

**Why:** ~80 casts disappear; new `expose()` calls get type-checked. **Alternative:** keep `expose(name: string, value: any)` — rejected; preserves the `any` sink. Tighten to `unknown` at minimum; ideally a union of known bridge names.

### 3. Central `$id()`/`getEl()` DOM helper

Add to `ts/lib/dom.ts` (or `ts/lib/query.ts`):

```ts
export function $id<T extends HTMLElement>(id: string): T | null
export function $idStrict<T extends HTMLElement>(id: string): T // throws if missing
```

Replace `document.getElementById(...)!` (201 occurrences) incrementally as files are detyped. Prefer nullable return with caller narrowing; `!` only where existence is truly invariant (and then via helper that throws with a clear message).

**Why:** One helper removes a class of assertions and gives a place to add diagnostics. **Alternative:** `as HTMLElement` casts — rejected; same unsafety.

### 4. Phased `@ts-nocheck` removal, leaf-first

Order:
1. Infra: `api<T>`, window types, `$id()` helper, CI guard (no file detyped yet)
2. Leaf character modules: `stats` → `resources` → `combat` → `spells` → `inventory` → `sheet`
3. Mid-size domains: `compendium` → `encounter` → `combat-tracker` → `factions` → `timeline`
4. `party.ts` (largest of the in-scope files)

Each file: remove pragma, run `npm run typecheck`, fix real errors (add types, narrow `any`, fix `null` handling), run `npm run test:unit` + relevant e2e. No new `@ts-ignore`/`@ts-expect-error` to silence errors — fix the types.

**Why:** Smallest files have fewest dependencies and fastest feedback; `party.ts` last because it touches many domains. **Alternative:** alphabetical or random — rejected; dependency order minimizes rework.

### 5. CI guard against new `// @ts-nocheck`

Add a check that fails CI if any `// @ts-nocheck` appears outside the current allowlist (`app.ts`, `admin.ts`, plus any not-yet-migrated file during the transition). Implement as either:
- `package.json` script: `grep -r "@ts-nocheck" ts/ --include="*.ts" | check-allowlist`
- Vitest test: reads `ts/` files and asserts no disallowed pragma

Guard is tightened as files are migrated (allowlist shrinks to just `app.ts`/`admin.ts` at end).

**Why:** Prevents regression; allowlist makes phased migration CI-green. **Alternative:** eslint `no-ts-nocheck` rule — viable but adds eslint config; grep/vitest is simpler.

## Risks / Trade-offs

- **Fixing types surfaces real bugs** → Some `any`-masked errors are actual runtime bugs. Mitigation: fix bugs as found; add tests where behavior was wrong.
- **Large diff per file detyped** → Reviewer load. Mitigation: one file (or small cluster) per commit/PR; keep commits reviewable.
- **Generic `api<T>` still requires caller to specify `T`** → Callers that omit `<T>` get `unknown`. Mitigation: codemod call sites to add type args where response shape is known; `unknown` is safer than `any`.
- **`app.ts`/`admin.ts` remain unchecked** → Majority of lines still unchecked until monolith split. Mitigation: documented as follow-up; leaf-file work still covers ~40% of frontend and establishes patterns.
- **Window augmentation drift** → New `expose()` names must be added to `Window` interface. Mitigation: CI guard or type error when exposing an undeclared name (if `expose` is typed with `keyof Window`).

## Migration Plan

1. Phase 0 — Infra: land `api<T>`, window types, `$id()` helper, CI guard with full allowlist. `npm run typecheck` still passes (nocheck files unchanged).
2. Phase 1 — Character leaves: `stats` → `resources` → `combat` → `spells` → `inventory` → `sheet`, one PR per file or pair.
3. Phase 2 — Domain modules: `compendium` → `encounter` → `combat-tracker` → `factions` → `timeline`.
4. Phase 3 — `party.ts` (largest), then tighten allowlist to `app.ts`/`admin.ts` only.
5. Each phase verifies `npm run typecheck`, `npm run test:unit`, and relevant e2e before next phase.

## Open Questions

- Where should shared API response types live — `ts/lib/api-types.ts`, `ts/types/api.ts`, or co-located with handlers? (Proposal: `ts/lib/api-types.ts` initially, move if it grows.)
- Should `api<T>` also type the `body` param as `unknown` or `Partial<T>`? (Proposal: `unknown` — request and response shapes differ.)
- Exact shape of `Window` augmentation — enumerate all `expose()` names now or incrementally? (Proposal: enumerate known names now, add as needed.)
