# Villum UX/UI & Feature Analysis — November 2026

**Scope:** full audit of the web app's features, UI/UX, and user flow.
**Method:** five evidence lanes — (1) full codebase map, (2) live desktop walkthrough (1440×900, 16 views), (3) live mobile pass (390×844), (4) live deep-flow walkthrough (character creation, dice, combat, compendium, one-shot, admin), (5) two-round visual design review against BRANDKIT.md. Every P0 below was either reproduced live or verified in code with file:line evidence.

**Evidence artifacts** (scratch, not committed): `/tmp/opencode/shots/` — 89 screenshots (`d-*` desktop, `m-*` mobile, `df-*` deep-flow), `mobile-observations.md`, `deepflow-observations.md`, `design-critique.md` (817 lines, two rounds).

---

## Executive summary

Villum is **feature-rich and functionally deep** — the one-shot builder, combat tracker, compendium, and 15-tab character sheet rival or beat dedicated tools in scope. But the app has three systemic problems:

1. **The brand is a skin over a Bootstrap app, not a design system.** 28 custom properties exist, all colour/radius/shadow — no semantic colour layer, no type scale, no spacing scale. Every visual defect (Bootstrap blue leaking into the admin CTA, six competing primary-button styles, SRD badges at 1.96:1 contrast, stock white modals on parchment) traces back to this single gap.
2. **Core loops have silent breakages.** Roll Initiative is a no-op for hand-added combatants. Three admin tabs are blank from one misbalanced `<div>`. Combat and Encounters views load blank via direct URL. Six FAB buttons call functions that don't exist. None of these produce an error — the UI claims success.
3. **The first-run and mobile experiences undermine trust.** A flaky campaign-select modal gates the app inconsistently; creating a campaign doesn't select it; mobile loses the entire top navbar (search, logout, theme, language, brand — gone) with no replacement; `/admin` is a dead-end page outside the app shell.

The single highest-leverage fix is **building the missing system layer** (semantic tokens, type scale, spacing scale, page-header/empty-state/destructive-dialog components). Nearly every P1 visual finding is a symptom of its absence.

---

## P0 — Broken or harmful (fix first)

### Silent functional bugs

| # | Bug | Evidence |
|---|-----|----------|
| 1 | **Roll Initiative is a silent no-op for hand-added combatants.** The add-combatant flow stores `campaign_id` NULL, but the roll query filters `campaign_id=?` → matches nothing → INIT stays "-" while a success toast fires. | `ts/combat-tracker.ts:249-261` (POST payload omits campaign_id), `handlers/combat.go:179-199` (`AND campaign_id=?` when request carries one), toast at `combat-tracker.ts:82,206`. Reproduced live. |
| 2 | **Admin: 3 of 15 tabs blank + TRMNL unreachable + Logs tab blank.** One misbalanced div: `#adminImport` opens `admin.html:752` and closes `:990`, swallowing `#adminEInk` (:854), `#adminSettings` (:869), `#adminTrmnl` (:933) and the real Application Logs card (:946). `#adminLogs` is referenced by `ts/admin.ts:93,105` but doesn't exist. Bonus: two tabs show active at once (stale class after Settings→Import). | Reproduced live; DOM verified. |
| 3 | **`#/combat` and `#/encounters` load blank** on direct URL, reload, AND in-SPA hash change — all view divs stay `display:none`; only the mobile More-sheet buttons activate them. Reproduces at desktop too. Routing bug, not CSS. | Reproduced live by two independent agents. |
| 4 | **Six dead FAB buttons** calling nonexistent functions: End Turn→`nextTurn` (real: `advanceCombatTurn`, `combat-tracker.ts:209`) ×2, New Encounter→`showNewEncounter` (real: `showCreateEncounter`, `encounter.ts:59`), New Event→`showNewTimelineEvent`, New Adventure→`showNewOneShot`, New Faction→`showNewFaction`. All targets absent from the built bundle. | Verified in app.js bundle + source. |
| 5 | **`campaign_id=""` is a pattern, not a one-off.** Combat is proven broken; the one-shot form ships a hidden `campaign_id value=""`. Audit every form that scopes by campaign. | `ts/combat-tracker.ts:249`, one-shot form DOM. |
| 6 | **No validation on character creation:** empty submit silently persists an "Unnamed" Level-1 character and drops you into its 15-tab sheet. Race/Class are free-text with no datalist. | Reproduced live. |
| 7 | **Dice quick-buttons set the expression but don't roll**; the 3D animation makes the result box lag ~1s with no spinner. | Reproduced live (two agents). |
| 8 | **Compendium filter toolbar renders twice** — two search inputs, eight selects, two count rows, stacked. | Reproduced live, visually confirmed. |
| 9 | **PWA install path is dead code:** `captureInstallPrompt` (`pwa.ts:22-40`) appends to nonexistent `#moreMenu` and is never called; the pwa bundle is also double-loaded (script tag `app.html:448` + import in app.js). | Code-verified. |

### Flow-level P0

| # | Issue | Evidence |
|---|-------|----------|
| 10 | **First-run campaign gate is flaky.** The "Select a Campaign" modal auto-opens, blocks some nav clicks, dismisses on Escape, later auto-closes; it is also invisible to the accessibility tree while shown. **Creating a campaign does not select it** — the roster modal opens, but selection state never lands, so the gate re-appears. | Reproduced across three walk runs. |
| 11 | **Mobile deletes the navbar with no replacement:** at 390px there is no search, no logout, no theme toggle, no language selector, no brand anywhere. The bottom sheet (nav-more) works well and is the only path to 12 destinations. | Live mobile pass. |
| 12 | **`/admin` is a dead-end page:** separate non-hash page, no back link, no bottom tab bar on mobile, no h1, browser-back is the only escape. | Live, both viewports. |
| 13 | **Campaign context breaks on fresh mobile load:** context button shows "No campaign" despite a selected campaign in localStorage; the picker is a full-page view with no back affordance containing a **Sign Out** button (destructive action inside a picker); re-picking dumps you on `#/party`, not where you were. | Live mobile pass. |
| 14 | **Shared Links is modal-as-nav:** its content renders into `#genericModal`; the Bootstrap backdrop then blocks all navigation until dismissed — "a page wearing a dialog's clothes." | Reproduced live; root cause of the original walkthrough blockage. |
| 15 | **Contrast fails on the most-clicked controls:** white-on-gold CTAs 2.81:1; gold-on-parchment 2.37:1 (BRANDKIT explicitly forbids gold as text); muted body 3.79:1; SRD badge cyan 1.96:1. The accessible colour (oxblood, 10:1) is used for secondary/destructive actions; the inaccessible one for primaries. | Measured, designer round 1+2. |
| 16 | **Red has collapsed as a signal:** Save/Print/Export/Export PDF/Delete on the character sheet are five identical red-outlined buttons; red-outline is used for 9 ordinary secondary actions app-wide while red-fill means destructive. At point of use, red means nothing. | Visual review + grep counts. |
| 17 | **Character sheet has no hierarchy:** 15 unranked tabs (10+5 rows) beside the 16-item sidebar = 36+ parallel destinations; "Character Folio" and the character name render as peers (~32px/25px); folio search is wedged inside the header. | Designer round 2, live. |
| 18 | **Character grid and Sessions have no empty state** (the grid renders an empty div); Combat has a correct one — three different empty-state systems, one broken, one absent. | Live + code (`ts/characters/list.ts:18-33`). |
| 19 | **Duplicate `data-testid`s across nav surfaces** (`nav-recaps` exists on both the hidden navbar `<a>` and sidebar `<button>`): automation/test-id clicks silently hit hidden elements; adding a view requires edits in 3 places (navbar `app.html:36-50`, sidebar `:89-104`, bottom sheet `navigation.ts:169-190`). | DOM + code. |

Minor but report-worthy: 401 on `GET /api/user/me` at each page-load cycle (benign retry, visible in network logs); duplicate h1 on Shops; dead views with no entry point (wiki — only via global search or `#/wiki` hash; `campaignOverview` — zero callers; `singleEncounter` — only via encounter detail; battlemap — only via a button inside the campaign dashboard modal); `4044`/`500` routes untested; README says v2.13.0, app runs v3.5.3.

---

## P1 — The design system gap (highest leverage)

These are not 30 separate bugs; they are symptoms of one absence. Fix the system, and most of them resolve together.

1. **Semantic tokens.** Add `--danger/--warning/--info/--success` + a type scale + a spacing scale to the 28 existing (colour/radius/shadow-only) custom properties. Today, Bootstrap defaults leak wherever no token exists: admin Import CTA is Bootstrap blue `#0d6efd`, Copilot has stray blue/grey buttons, combat rows use `table-active` grey with `bs-danger` red.
2. **One page-header component.** All 16 page titles are `<h1 class="h2">` (32px) — the same size as the section headings beneath them, producing the duplicate-titles seen on One-Shots, Timeline, Shops, Factions, Dice, and the "two display titles competing" on the character sheet. A page-header primitive (title + subtitle + actions slot) kills this class.
3. **One primary-button treatment.** `btn-primary` (108 sites) vs `btn-gold` (42) compete; five different treatments appear in the combat toolbar alone; Next Turn *looks disabled* and Roll Init *looks destructive*. Define: primary, secondary, ghost, destructive — and reserve red strictly for destructive.
4. **One empty-state component.** Combat's is correct (icon + copy + CTA); one view shows a broken orphaned icon at the screen edge; Characters/Sessions show nothing. Ship the Combat pattern as a component.
5. **One destructive dialog.** 55 `confirm()` + 9 `alert()` call sites render native browser dialogs in an app that otherwise uses styled modals — jarring and unstylable. (Account deletion's typed-confirm is the right pattern for the top of this ladder.)
6. **Sheet information architecture.** Group the 15 tabs into categories (e.g. Build / Play / Story / Meta) with clear visual rank; move the folio search out of the character header; differentiate sheet-level actions (Save) from per-character actions (Delete).
7. **Bring `/admin` into the shell.** Same nav chrome, an h1, a back link, tab state that can't show two actives. 15 client-side tabs + blank-tab bugs all live on an orphan page.
8. **Footer fix (smallest big win):** `body{min-height:100vh}` without `display:flex; flex-direction:column` + `main{flex:1}` leaves the footer floating mid-page on 9/16 desktop and 6 mobile views — every sparse page looks unfinished. Two lines of CSS.
9. **Typography discipline.** 20+ ad-hoc font sizes (0.65–3.5rem); Playfair Display is never used at display size. Let the display face do display work.

---

## P2 — Polish and debt

- **Performance:** app.js is a 1,338 kB minified monolith (379 kB gzip) with vite's >500 kB warning; three attempted dynamic imports are defeated by static imports elsewhere (leaflet — dynamic in `ts/party-subtabs.ts` but static in `ts/characters/locations.ts` + `ts/world.ts`; d3 — dynamic in `ts/app/graph-analytics.ts` but static in `ts/party-subtabs.ts`; `ts/navigation.ts` — dynamic in `ts/search.ts` but static in app/character-ops/oneshot/shops/…). Route-level code-splitting is the goal; today it's decorative.
- **Login page pulls Bootstrap + FontAwesome from CDNs** while the app self-hosts fonts — the one page a PWA/offline user needs most fails first.
- **i18n is a stub (en/nl, 26 keys, nav chrome only) but the language selector is prominent.** Either commit to real i18n or hide the selector; all content/forms/errors are hard-coded English.
- **Dark mode ignores `prefers-color-scheme`** (localStorage only); a third e-ink theme exists behind a cookie + admin toggle.
- **Duplication debt:** theme toggle ×2 (`ts/lib/theme.ts` vs private copy in `admin.ts:42-57`), API-token bootstrap ×2, logout ×2, switchCharacter ×2, two toast implementations, `validateSelection` name collision, `showAdminTab` string surgery with a no-op ternary, 4,246-line style.css + 414-line eink.css + heavy inline styles in TS templates (three styling idioms mixed).
- **A11y beyond contrast:** #genericModal lacks `aria-labelledby`/`aria-modal` and never moves/restores focus; dynamic content (party/oneshot/compendium) has essentially no ARIA; admin has no h1-h4 at all; mobile tap targets at/below the 44px floor (bottom tabs 43px, World map zoom 30×30). The good: keyboard shortcuts with a help dialog, skip link, aria-live toast region, ARIA tablist roving focus in the sheet.
- **Docs:** README version (v2.13.0) and manual-build instructions (CGO) are stale vs reality (v3.5.3, modernc + sqlite_fts5 per Taskfile).

---

## Feature-level observations

- **Maturity (best→worst):** one-shot builder (largest, most coherent surface) → compendium (live search 284→7 results with rich metadata; only chrome issues) → characters (breadth without hierarchy/validation) → combat tracker (deep feature set, broken core loop) → campaigns → world/battlemap (Leaflet map works; entry points buried) → settings (thin) → PWA/i18n (stubs presented as features).
- **Unreachable/dead features:** wiki view, campaignOverview, singleEncounter, battlemap, TRMNL admin panel, PWA install prompt — all shipped, all dark. Either surface them or remove them; every dark feature is trust erosion + maintenance cost.
- **The registry layer is a hidden strength:** 18 entity types with ownership/search/transfer/link flags backing universal search and entity links — most users will never know it exists. Consider surfacing cross-entity linking in the UI (it's the app's differentiator).
- **TRMNL/e-ink, mobile Flutter app, Telegram** are substantial satellite surfaces not covered by this audit's live passes.

---

## Recommended order of attack

| Phase | Items | Effort |
|-------|-------|--------|
| **0 — stop the bleeding** | P0 #1–#9 (bugs w/ file:line above) | S–M each, mostly mechanical |
| **1 — first-run & flow** | campaign-select fix + create-selects-campaign (#10); mobile chrome replacement (#11); /admin into shell (#12–13); Shared Links becomes a view (#14) | M |
| **2 — system build** | semantic tokens + type/spacing scale; page-header, empty-state, destructive-dialog, primary-button components; sheet IA; footer flex fix | M–L, highest leverage |
| **3 — trust polish** | contrast remediation (#15), red reserved for destructive (#16), validation on all create forms, dead-code removal or surfacing | S–M each |
| **4 — debt** | real code-splitting; login self-host; i18n real-or-hide; prefers-color-scheme; dedupe nav surfaces + theme/toast/logout duplicates; docs refresh | M |

---

*Generated 2026-11-18 from a five-lane audit (codebase map, desktop walkthrough, mobile pass, deep-flow walkthrough, two-round design review). Scratch artifacts under /tmp/opencode/shots/ will be deleted; this document is the durable record. Not committed — review and decide the PR.*
