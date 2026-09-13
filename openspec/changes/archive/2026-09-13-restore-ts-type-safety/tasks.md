## 1. Infra — fix systemic any sources

- [x] 1.1 Make `ts/lib/api.ts:api()` generic: `api<T>(method, path, body?: unknown): Promise<T>` and define shared response types (`ts/lib/api-types.ts` or `ts/types.ts`) for `Character`, `Campaign`, `CompendiumEntry`, etc.
- [x] 1.2 Update representative `api()` call sites to pass explicit `<T>` type args; verify `npm run typecheck` still passes with nocheck files untouched
- [x] 1.3 Add typed `Window` augmentation (`ts/lib/window.d.ts`) for all `expose()` bridge names; tighten `expose(name, value)` from `any` to `unknown`/typed keys and replace `(window as any)` casts
- [x] 1.4 Add central DOM helper `$id<T extends HTMLElement>(id: string): T | null` (and strict throwing variant) to `ts/lib/dom.ts` or `ts/lib/query.ts`
- [x] 1.5 Add CI guard that fails on disallowed `// @ts-nocheck` (grep script in `package.json` lint step or vitest test with allowlist); initial allowlist = current 13 files; wire into `npm run typecheck` or `npm run test:unit`

## 2. Leaf character modules (smallest first)

- [x] 2.1 Remove `// @ts-nocheck` from `ts/characters/stats.ts` — fix type errors, no new suppressions, `npm run typecheck` clean
- [x] 2.2 Remove `// @ts-nocheck` from `ts/characters/resources.ts` — same
- [x] 2.3 Remove `// @ts-nocheck` from `ts/characters/combat.ts` — same
- [x] 2.4 Remove `// @ts-nocheck` from `ts/characters/spells.ts` — same
- [x] 2.5 Remove `// @ts-nocheck` from `ts/characters/inventory.ts` — same
- [x] 2.6 Remove `// @ts-nocheck` from `ts/characters/sheet.ts` — same (largest of this group, may need sub-tasks)

## 3. Mid-size domain modules

- [x] 3.1 Remove `// @ts-nocheck` from `ts/compendium.ts` — fix errors, `npm run typecheck` clean
- [x] 3.2 Remove `// @ts-nocheck` from `ts/encounter.ts` — same
- [x] 3.3 Remove `// @ts-nocheck` from `ts/combat-tracker.ts` — same
- [x] 3.4 Remove `// @ts-nocheck` from `ts/factions.ts` — same
- [x] 3.5 Remove `// @ts-nocheck` from `ts/timeline.ts` — same
- [x] 3.6 Remove `// @ts-nocheck` from `ts/party.ts` — largest in-scope file; fix errors, tighten allowlist

## 4. Hardening and verification

- [x] 4.1 Reduce `any` hotspots in migrated files: type `(e:any)`/`(r:any)` callbacks, narrow remaining `any` where practical (no new `any` introduced)
- [x] 4.2 Tighten CI guard allowlist to exactly `ts/app.ts` + `ts/admin.ts`; verify a synthetic `// @ts-nocheck` in a migrated file fails CI
- [x] 4.3 `npm run typecheck` clean (all migrated files)
- [x] 4.4 `npm run test:unit` green; add/adjust vitest guard test if used
- [x] 4.5 `npm run build:vite` succeeds (5 bundles)
- [x] 4.6 `task test:e2e` green — no behavior change from typing fixes
- [x] 4.7 Run full local gates (`task ci`) and ensure no regressions
