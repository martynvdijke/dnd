## 1. Shared helpers (land first)

- [x] 1.1 Extract `refreshChar(id?: string)` helper (wraps `setCurrentChar(await api('GET', /api/characters/${id}))`) into `ts/lib/refresh.ts` or `ts/lib/state.ts`; replace 2-3 representative call sites and verify
- [x] 1.2 Extract `renderError(e: unknown)` helper (wraps `toast` error display) into `ts/lib/dom.ts` or `ts/lib/errors.ts`; replace representative `catch(e){toast(e.message,true)}` sites
- [x] 1.3 Replace remaining 23× `setCurrentChar(await api(...))` and ~40× `catch/toast` patterns with helpers across `ts/app.ts` and `ts/admin.ts`
- [x] 1.4 Verify: `npm run typecheck`, `npm run test:unit`, `npm run build:vite` clean

## 2. app.ts — core and character seams

- [x] 2.1 Extract Sort (L51) + Campaign/Character switching (L112) + New Character (L843) + Import/Export (L877) + Delete/Logout/Portrait (L954-980) into `ts/app/<domain>.ts` modules; wire `expose()` + `init.ts` imports
- [x] 2.2 Extract Features/Proficiencies/Details (L373-464) + Conditions/Feats/Companions (L1146-1464) + Character Sheet (L129) + Roll/Combat Actions (L131); reconcile any `app.ts` ↔ `ts/characters/*` duplication (keep `ts/characters/*` canonical)
- [x] 2.3 Extract Race Colors (L1017) + Multi-Class (L1074) + Calendar (L1141) + Notes TipTap (L1534)
- [x] 2.4 Verify after each sub-phase: `npm run typecheck`, `npm run test:unit`, `task test:e2e` green before next extraction

## 3. app.ts — large domain seams

- [x] 3.1 Extract Shops (L1646-2312, ~666 lines) into `ts/app/shops.ts` (or `ts/app/shop/*`); move D3/Chart.js imports with seam if applicable
- [x] 3.2 Extract Random Gen/Comparison (L2479-2554) into `ts/app/random-gen.ts`
- [x] 3.3 Extract Wiki/Campaign Graph (L2602-2816) into `ts/app/wiki.ts` (+ `ts/app/campaign-graph.ts` if warranted)
- [x] 3.4 Extract Polymorphic File Uploads (L3517) into `ts/app/uploads.ts`
- [x] 3.5 Verify after each sub-phase: `npm run typecheck`, `npm run test:unit`, `task test:e2e`

## 4. app.ts — remaining seams

- [x] 4.1 Extract One-Shot Tree/Items/Shops/Monsters/NPCs (L2816-3517, ~700 lines) into `ts/app/oneshot.ts` (or `ts/app/oneshot/*`)
- [x] 4.2 Extract Campaign Dashboard/Party Inventory/Session Planner (L3553-3700) + Encounter Difficulty/Treasure (L3796-3920) + D3 Force Graph/Graph/Analytics (L553-752)
- [x] 4.3 Verify `ts/app.ts` is < 800 lines (`wc -l ts/app.ts`); `npm run build:vite` still produces 5 bundles

## 5. admin.ts — domain seams

- [x] 5.1 Extract E-ink (L119) + AI Features (L144) + API Tokens (L194) into `ts/admin/<domain>.ts`; wire `expose()` + `init.ts`
- [x] 5.2 Extract Unified Compendium (L325) + Global Search (L365) + Entry Browser (L446) + Bulk Ops (L571)
- [x] 5.3 Extract Entry CRUD (L692) + Schema-Aware Editor (L786) + Import Integration (L922)
- [x] 5.4 Extract Logs (L945-1146) + Users (L1147) + Schemas (L1263)
- [x] 5.5 Extract Backup/Email/Push/Umami/OTel/AI settings (L1382-1593) + Import Wizard (L1716-2090) + Utils/PDF/Events/Camera (L2091-2508)
- [x] 5.6 Verify `ts/admin.ts` is < 800 lines; `npm run typecheck`, `npm run build:vite` clean

## 6. Final verification

- [x] 6.1 `npm run typecheck` clean (nocheck pragmas retained on moved code)
- [x] 6.2 `npm run test:unit` green
- [x] 6.3 `npm run build:vite` succeeds — 5 IIFE bundles (app, admin, pwa, setup, login)
- [x] 6.4 `task test:e2e` green — full Playwright suite with no regressions from the split
- [x] 6.5 Run full local gates (`task ci`) and ensure no regressions; confirm `app.ts` and `admin.ts` line counts under 800
