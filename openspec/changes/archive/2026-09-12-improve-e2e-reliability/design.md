# Design: improve-e2e-reliability

## Context

Audit found: 104 `waitForTimeout` calls across `tests/` — `journal.spec.ts:39,59,62,93,125,131` (six 300-500ms), `party.spec.ts:161` (1500ms), `oneshot.spec.ts:228` (2000ms), `search.spec.ts:10,36,42` (300-1000ms), `combat-tracker.spec.ts` 5 sleeps, `shops.spec.ts:10,30,48`, plus 4 sleeps inside `tests/helpers.ts:10,94,104,112` (navbar toggler, bottom-sheet close). Seven stale `waitLoadingDone` copies in `campaign.spec.ts:6`, `campaign-extended.spec.ts:6`, `character-advanced.spec.ts:6`, `character.spec.ts:6`, `companions.spec.ts:6`, `compendium.spec.ts:6`, `selection.spec.ts:6` (overlay-only, swallows errors) diverge from canonical `helpers.ts:14` (also awaits `window.api` + `__apiReady`); `waitModalClosed` redefined in `campaign.spec.ts:13`, `character.spec.ts:13`, `search.spec.ts:14`, `responsive.spec.ts:14`, `character-sheet.spec.ts:249` vs `helpers.ts:65`. 152 `test.slow()` sites mask slowness: ~60 in `oneshot-content.spec.ts` alone, `search.spec.ts:46,54,60,66,82` (5 in one file), `campaign-extended` 14, `responsive` 10, `combat-animations.spec.ts:32,60`, `session-mode.spec.ts:10,19,38`, `oneshot.spec.ts:111`, `combat-tracker.spec.ts:24,30,37,58,83`, `smoke.spec.ts:12`. `oneshot-content.spec.ts` is 1568 lines (62KB). `playwright.config.ts` defines 3 projects but CI runs chromium only; retries: 2; CI captures `flaky-tests.log`. AGENTS.md requires: every `data-testid` in `ts/`/`static/*.html` referenced in `tests/`, E2E uses `login()` from `tests/helpers.ts`, `LOGIN_TIMEOUT` 30s, `NAV_TIMEOUT` 10s, `test.slow()` only for genuinely slow tests.

## Goals / Non-Goals

**Goals:**
- Eliminate all 104 `waitForTimeout` sleeps, replacing with deterministic waits (`expect(locator).toBeVisible()`, `page.waitForResponse`, `page.waitForFunction`, `locator.waitFor()`).
- Deduplicate helpers: delete 7 stale `waitLoadingDone` + 5 `waitModalClosed`; all specs import from `tests/helpers.ts`.
- Fix root cause of `search.spec.ts` slowness (5 `test.slow()`) and remove `test.slow()` where no longer needed.
- Split `oneshot-content.spec.ts` (1568L) by content type into <600L files.
- Add guardrail forbidding new `waitForTimeout` (grep in CI or `prek` hook) without increasing flake rate (run suite twice per phase).

**Non-Goals:**
- Enabling all 3 Playwright projects in CI (chromium-only remains; multi-browser is separate change).
- Changing app behavior or `data-testid` contracts (lint must stay green).
- Rewriting all E2E tests — scoped to sleeps, stale helpers, slow markers, and one large file split.
- Changing coverage floors (covered by `close-coverage-holes`).

## Decisions

### D1: One canonical helper import — delete stale copies

Delete `waitLoadingDone` in `campaign.spec.ts:6`, `campaign-extended.spec.ts:6`, `character-advanced.spec.ts:6`, `character.spec.ts:6`, `companions.spec.ts:6`, `compendium.spec.ts:6`, `selection.spec.ts:6` and `waitModalClosed` in `campaign.spec.ts:13`, `character.spec.ts:13`, `search.spec.ts:14`, `responsive.spec.ts:14`, `character-sheet.spec.ts:249`. All files `import { waitLoadingDone, waitModalClosed, login } from './helpers'`. Canonical `helpers.ts:14` awaits overlay hidden + `window.api` ready + `__apiReady` flag; stale copies only await overlay and swallow errors — so replacement fixes hidden flakes.

Alternatives: keep per-file helpers with local tweaks — rejected: divergence is the bug; single source of truth is required.

### D2: Replace `waitForTimeout` with categorized proper waits

Map each sleep to a readiness signal:
- **UI visibility** (`journal`, `shops`, `combat-tracker` 300-500ms sleeps) → `await expect(locator).toBeVisible()` / `toBeHidden()` / `locator.waitFor({state: 'visible'})`.
- **Network** (`search.spec.ts` 300-1000ms) → `await page.waitForResponse(resp => resp.url().includes('/api/search') && resp.status()===200)` or `waitForFunction` on results count.
- **Overlay/animation** (`party 1500ms`, `oneshot 2000ms`, `combat-animations`) → `waitForFunction(() => !document.querySelector('.loading-overlay'))` or `expect(overlay).toBeHidden()`.
- **Helpers internals** (`helpers.ts:10,94,104,112` navbar toggler/bottom-sheet) → `expect(toggler).toBeVisible()` + `expect(sheet).toBeHidden()` instead of `waitForTimeout`.

Categorize per file cluster and convert in phases (one cluster per commit) to isolate flake.

Alternatives: increase sleep durations — rejected: masks flake, slows suite. `page.waitForTimeout` is never the right fix.

### D3: Helpers.ts sleeps fixed first

Fix `tests/helpers.ts:10,94,104,112` before converting spec files, since helpers are used everywhere. Replace with `expect(...).toBeVisible()/toBeHidden()` and `page.waitForFunction` on `__apiReady`. This unblocks spec conversions and prevents double-fixing.

### D4: Investigate `search.spec.ts` slowness before removing `test.slow()`

`search.spec.ts` has 5 `test.slow()` but total is 152 (`oneshot-content` ~60, `campaign-extended` 14, `responsive` 10 etc.) — systemic. Likely root cause is `waitForTimeout` + missing `waitForResponse` for search API. Fix waits first, then measure: `playwright test --reporter=list` timing. Remove `test.slow()` only when test passes without it and timing is under default timeout. Keep `test.slow()` where genuinely needed (e.g., `oneshot` generation that hits AI or heavy DB seed) with a comment justifying.

Alternatives: blanket remove all `test.slow()` — rejected: some tests are legitimately slow; removal without fixing waits would cause timeout flakes.

### D5: Split `oneshot-content.spec.ts` by content type

Current 1568L file covers all oneshot content types. Split into e.g. `oneshot-content-spells.spec.ts`, `oneshot-content-monsters.spec.ts`, `oneshot-content-items.spec.ts`, `oneshot-content-npcs.spec.ts` (or by actual `describe` blocks). Each file <600L, shares `login()` + `waitLoadingDone` imports. No logic change, just `describe` moves.

Alternatives: keep one file and use `test.describe` grouping — rejected: file is already 62KB, review and sharding suffer.

### D6: Guardrail — grep check in CI (and optionally `prek` hook)

Add a CI step (and `prek` pre-commit hook if feasible) that fails on `waitForTimeout` in `tests/` (allowlist only for documented exceptions, e.g., polling with comment). Simplest: `rg -n "waitForTimeout" tests/` in `ci.yaml` or `Taskfile` `lint:e2e` task. This prevents reintroduction without heavy tooling.

Alternatives: Playwright lint plugin or custom ESLint rule — heavier; grep is sufficient and already used elsewhere.

## Risks / Trade-offs

- [Replacing sleeps with wrong wait causes new flakes] → Mitigation: map each sleep to the actual readiness signal (overlay, API response, locator); run `task test:e2e` twice per phase; keep `retries: 2` and check `flaky-tests.log`.
- [Helper deduplication breaks specs that relied on swallow-errors behavior] → Mitigation: canonical helper is stricter (also awaits `window.api`); if a spec relied on leniency, fix the spec's timing rather than re-adding a lenient helper.
- [Removing `test.slow()` causes timeout failures] → Mitigation: remove only after waits fixed and timing measured; keep `test.slow()` with justification comment where still needed.
- [Splitting `oneshot-content` loses coverage or duplicates setup] → Mitigation: each new file imports shared `login()`; run full suite after split; no logic change.
- [Grep guardrail false positives] → Mitigation: allowlist via `// allow-waitForTimeout: reason` comment or exclude `helpers.ts` if a single polling use remains; document rule.
- [CI chromium-only hides cross-browser flakes] → Mitigation: out of scope; note that `playwright.config.ts` 3 projects remain but CI runs chromium only — no change in this proposal.

## Migration Plan

1. **Phase 1 — Helpers dedup + helpers.ts sleeps** — delete 7 stale `waitLoadingDone` + 5 `waitModalClosed` (campaign, character, companions, compendium, selection, responsive, character-sheet); fix `helpers.ts:10,94,104,112` sleeps to proper waits; all specs import from `helpers.ts`; `task test:e2e` twice.2. **Phase 2 — Journal/shops/combat-tracker sleeps** — replace `journal.spec.ts` 6 sleeps, `shops.spec.ts:10,30,48`, `combat-tracker.spec.ts` 5 sleeps with `expect(...).toBeVisible()/Hidden()`; `task test:e2e` twice.
3. **Phase 3 — Search/party/oneshot sleeps + search slowness** — replace `search.spec.ts:10,36,42`, `party.spec.ts:161`, `oneshot.spec.ts:228` with `waitForResponse`/`waitForFunction`; investigate search slowness root cause; remove `search.spec.ts` `test.slow()` where fixed; `task test:e2e` twice.
4. **Phase 4 — Remaining sleeps + `test.slow()` cleanup** — replace remaining `waitForTimeout` across other specs (`combat-animations`, `session-mode`, etc.); audit all 152 `test.slow()` sites (~60 in `oneshot-content` alone) and remove where no longer needed (keep with comment where justified); `task test:e2e` twice.
5. **Phase 5 — Split `oneshot-content.spec.ts`** — split 1568L file by content type into <600L files; imports deduped; `task test:e2e` twice.
6. **Phase 6 — Guardrail** — add `rg -n "waitForTimeout" tests/` check to `ci.yaml`/`Taskfile` (and `prek` hook if feasible); document rule in `CONTRIBUTING.md`/`AGENTS.md`; final `task ci` green.

Rollback: each phase is a standalone commit; revert the phase. Guardrail is additive and reverted separately.

## Open Questions

- Are any `waitForTimeout` uses legitimate polling (e.g., debounce) that should be allowlisted vs. replaced with `waitForFunction`?
- Exact `oneshot-content` split boundaries — confirm `describe` blocks in the 1568L file before cutting.
- Which `test.slow()` sites are legitimately slow (AI generation, heavy seed) vs. masked waits — measure after Phase 3 before bulk removal.
