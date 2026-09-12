# Proposal: improve-e2e-reliability

## Why

The E2E suite is brittle and slow: 104 hardcoded `waitForTimeout` sleeps across `tests/` (e.g. `journal.spec.ts:39,59,62,93,125,131` six 300-500ms sleeps, `party.spec.ts:161` 1500ms, `oneshot.spec.ts:228` 2000ms, `search.spec.ts:10,36,42` 300-1000ms, plus 4 sleeps inside `tests/helpers.ts:10,94,104,112`), nine stale helper copies (7× `waitLoadingDone` in `campaign.spec.ts:6`, `campaign-extended.spec.ts:6`, `character-advanced.spec.ts:6`, `character.spec.ts:6`, `companions.spec.ts:6`, `compendium.spec.ts:6`, `selection.spec.ts:6` + `waitModalClosed` in `campaign.spec.ts:13`, `character.spec.ts:13`, `search.spec.ts:14`, `responsive.spec.ts:14`, `character-sheet.spec.ts:249`) diverging from canonical `helpers.ts:14` (which also awaits `window.api` + `__apiReady`) and `helpers.ts:65`, 152 `test.slow()` sites masking systemic slowness (~60 in `oneshot-content.spec.ts` alone, 5 in `search.spec.ts`), and a 1568-line `oneshot-content.spec.ts` single file. `playwright.config.ts` defines 3 projects but CI runs chromium only. Sleeps cause flakes and mask real readiness signals.

## What Changes

- **Deduplicate helpers** — delete 7 stale `waitLoadingDone` copies (`campaign.spec.ts:6`, `campaign-extended.spec.ts:6`, `character-advanced.spec.ts:6`, `character.spec.ts:6`, `companions.spec.ts:6`, `compendium.spec.ts:6`, `selection.spec.ts:6`) and 5 `waitModalClosed` redefinitions (`campaign.spec.ts:13`, `character.spec.ts:13`, `search.spec.ts:14`, `responsive.spec.ts:14`, `character-sheet.spec.ts:249`); all specs import canonical helpers from `tests/helpers.ts:14`/`helpers.ts:65`.
- **Replace sleeps with proper waits** — replace 104 `waitForTimeout` calls with deterministic waits: `expect(locator).toBeVisible()/toBeHidden()`, `page.waitForResponse` (for search/API), `page.waitForFunction` (for `__apiReady`/overlay), `locator.waitFor()` — categorized per file cluster; also fix 4 sleeps inside `tests/helpers.ts:10,94,104,112` (navbar toggler, bottom-sheet close) to use proper waits.
- **Fix slowness root cause and remove `test.slow()`** — investigate `search.spec.ts` (5 `test.slow()` at `:46,54,60,66,82`), `combat-animations`, `session-mode`, `oneshot`, `combat-tracker`, `smoke` slowness; fix underlying waits/timeouts (e.g. `LOGIN_TIMEOUT` 30s, `NAV_TIMEOUT` 10s per AGENTS.md); remove `test.slow()` where no longer needed, keeping it only where truly justified.
- **Split large spec** — split `tests/oneshot-content.spec.ts` (1568L, 62KB) by content type (e.g. `oneshot-content-spells.spec.ts`, `oneshot-content-monsters.spec.ts`, `oneshot-content-items.spec.ts`, etc.) — one file per content type, each <600L.
- **Guardrail against regression** — add lint-ish check forbidding new `waitForTimeout` (grep check in CI or `prek` hook) and document rule; flake rate must not increase — run full E2E suite twice per phase (`task test:e2e` rebuilds `villum-server`).

## Capabilities

### New Capabilities
- `e2e-test-reliability`: Reliable, deterministic E2E suite with deduplicated helpers, proper waits instead of sleeps, justified slow markers, modular spec files, and a guardrail forbidding new hardcoded sleeps.

### Modified Capabilities
- None

## Impact

- **Code**: `tests/helpers.ts` (fix 4 sleeps, keep `login()` helper, `LOGIN_TIMEOUT` 30s, `NAV_TIMEOUT` 10s), 7 files with stale `waitLoadingDone` + 5 with stale `waitModalClosed` (remove), all files with `waitForTimeout` (journal, party, oneshot, search, combat-tracker, shops, etc.), `tests/oneshot-content*.spec.ts` (split), `playwright.config.ts` (no functional change; CI still chromium), `prek`/`Taskfile`/`ci.yaml` (new grep guardrail).
- **Tests**: `task test:e2e` must stay green; flake rate not increased (run suite twice per phase); `test.slow()` only where justified (reduce 152 → justified set, ~60 in `oneshot-content` are primary target).
- **CI**: New grep/hook check for `waitForTimeout`; `flaky-tests.log` (retries: 2) should trend down.
- **Docs**: AGENTS.md E2E rules remain (every `data-testid` referenced in `tests/`, `login()` helper, timeouts, `test.slow()` usage).
