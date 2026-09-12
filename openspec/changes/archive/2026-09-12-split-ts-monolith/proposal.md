## Why

`ts/app.ts` (4067 lines) and `ts/admin.ts` (2730 lines) together hold 39% of all frontend code (17,402 lines total). Monoliths this large slow navigation, review, and type-safety work, and every change risks unrelated regressions.

## What Changes

- Split `ts/app.ts` into domain modules (target: `app.ts` < 800 lines orchestrator) following existing `// ─── Section ───` seams
- Split `ts/admin.ts` into `ts/admin/*` modules (target: `admin.ts` < 800 lines) following the same seam pattern
- Follow the established extraction precedent: module self-registers via `expose()` from `ts/lib/expose.ts`; `ts/init.ts` imports modules for side effects — no new framework
- Reconcile duplication where functions exist both in `app.ts` and `ts/characters/*` during extraction
- Extract cross-cutting helpers `refreshChar()` (replaces 23× `setCurrentChar(await api('GET', /api/characters/${id}))`) and `renderError()` (replaces ~40× `catch(e){toast(e.message,true)}`) during the split
- Keep behavior byte-identical; verify with `npm run typecheck` + `npm run test:unit` + `task test:e2e` after EACH extraction phase (one phase per domain cluster)
- Keep existing `// @ts-nocheck` pragmas on moved code where needed (nocheck removal is handled by `restore-ts-type-safety` separately)

## Capabilities

### New Capabilities

- `frontend-modularity`: Domain-modular frontend where `ts/app.ts` and `ts/admin.ts` are thin orchestrators importing self-registering domain modules, with shared helpers for character refresh and error rendering

### Modified Capabilities

- None

## Impact

- **Frontend**: `ts/app.ts` → `ts/app/*` or `ts/domains/*` modules; `ts/admin.ts` → `ts/admin/*` modules; `ts/init.ts` import graph; `vite.config.ts` entry points (if needed, but bundles stay 5 IIFEs: app, admin, pwa, setup, login)
- **Tests**: No new product behavior — existing vitest + Playwright suites are the regression net; no backend or migration changes
- **Risk if not done**: Continued monolith drag on velocity, merge conflicts on `app.ts`/`admin.ts`, and blocked type-safety cleanup
