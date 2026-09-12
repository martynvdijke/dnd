## Why

`tsconfig.json` enables `strict: true` but 13 files carry blanket `// @ts-nocheck`, nullifying strict mode for ~60% of frontend code. With 797 `any` occurrences and 201 `!` non-null assertions, type safety is illusory and regressions slip through unchecked.

## What Changes

- Make `ts/lib/api.ts:api()` generic: `api<T>(method, path, body?): Promise<T>` and define shared API response types for campaign/character/compendium endpoints
- Add a typed `Window` augmentation (e.g. `ts/lib/window.d.ts`) replacing the ~80 `(window as any)` casts used via `expose()` bridge
- Add a central DOM lookup helper `$id()`/`getEl()` with proper narrowing to replace `getElementById(...)!` assertions
- Remove `// @ts-nocheck` file-by-file in phased order (leaf files first: `stats`, `resources`, `combat`, `spells`, `inventory`, then `compendium`, `encounter`, `combat-tracker`, `factions`, `timeline`, `party`), fixing real type errors — no new `@ts-ignore`/`@ts-nocheck` added
- Add a CI guard (grep/AST check via `package.json` lint script or vitest test) that fails if a new `// @ts-nocheck` is introduced
- Explicitly out of scope: removing `@ts-nocheck` from `ts/app.ts` and `ts/admin.ts` (depends on monolith split)

## Capabilities

### New Capabilities

- `frontend-type-safety`: Phased restoration of TypeScript strict-mode coverage — generic API client, typed window bridge, safe DOM helpers, incremental `@ts-nocheck` removal, and a CI guard preventing regression

### Modified Capabilities

- None

## Impact

- **Frontend**: `ts/lib/api.ts`, `ts/lib/expose.ts`, `ts/lib/dom.ts` (new `$id` helper), `ts/lib/window.d.ts` (new), 11 files losing `@ts-nocheck` + their callers; no runtime behavior change
- **Tooling**: `package.json` lint script or vitest guard for `@ts-nocheck`; `tsconfig.json` unchanged (`strict` stays `true`)
- **Risk if not done**: Continued silent type errors, `any` propagation from `api(): Promise<any>`, and no compiler protection on the majority of frontend code
