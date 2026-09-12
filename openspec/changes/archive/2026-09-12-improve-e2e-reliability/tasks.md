# Tasks: improve-e2e-reliability

## 1. Deduplicate helpers and fix helpers.ts sleeps

- [x] 1.1 Delete 7 stale `waitLoadingDone` copies in `tests/campaign.spec.ts:6`, `tests/campaign-extended.spec.ts:6`, `tests/character-advanced.spec.ts:6`, `tests/character.spec.ts:6`, `tests/companions.spec.ts:6`, `tests/compendium.spec.ts:6`, `tests/selection.spec.ts:6`; delete 5 `waitModalClosed` redefinitions in `tests/campaign.spec.ts:13`, `tests/character.spec.ts:13`, `tests/search.spec.ts:14`, `tests/responsive.spec.ts:14`, `tests/character-sheet.spec.ts:249`; update imports to `import { waitLoadingDone, waitModalClosed, login } from './helpers'`
- [x] 1.2 Replace 4 sleeps in `tests/helpers.ts:10,94,104,112` (navbar toggler, bottom-sheet close) with proper waits (`expect(toggler).toBeVisible()`, `expect(sheet).toBeHidden()`, `page.waitForFunction(() => window.__apiReady)`, overlay hidden checks)
- [x] 1.3 Run `task test:e2e` twice (rebuilds `villum-server`); verify no new flakes in `flaky-tests.log` (retries: 2); ensure `LOGIN_TIMEOUT` 30s / `NAV_TIMEOUT` 10s respected

## 2. Replace journal / shops / combat-tracker sleeps

- [x] 2.1 Replace 6 sleeps in `tests/journal.spec.ts:39,59,62,93,125,131` (300-500ms) with `expect(locator).toBeVisible()/toBeHidden()` or `locator.waitFor({state:'visible'})`
- [x] 2.2 Replace sleeps in `tests/shops.spec.ts:10,30,48` with `expect(...).toBeVisible()` / `page.waitForResponse` for shop API
- [x] 2.3 Replace 5 sleeps in `tests/combat-tracker.spec.ts` with `expect(overlay).toBeHidden()` / `waitForFunction` on tracker readiness
- [x] 2.4 Run `task test:e2e` twice; verify no increased flake rate; check `data-testid` lint still passes (every `data-testid` in `ts/`/`static/*.html` referenced in `tests/`)

## 3. Replace search / party / oneshot sleeps and fix search slowness

- [x] 3.1 Replace sleeps in `tests/search.spec.ts:10,36,42` (300-1000ms) with `page.waitForResponse(resp => resp.url().includes('/api/search') && resp.status()===200)` or `waitForFunction` on results count; verify `login()` helper still used
- [x] 3.2 Replace `tests/party.spec.ts:161` (1500ms) and `tests/oneshot.spec.ts:228` (2000ms) with `waitForFunction` on overlay/`__apiReady` or `expect(...).toBeHidden()`
- [x] 3.3 Investigate `search.spec.ts` 5 `test.slow()` root cause (timing with `waitForResponse`); remove `test.slow()` markers where tests now pass within default timeout; keep with justification comment where still needed
- [x] 3.4 Run `task test:e2e` twice; measure `search.spec.ts` timing (`--reporter=list`); confirm flake rate not increased

## 4. Replace remaining sleeps and audit test.slow()

- [x] 4.1 Replace remaining `waitForTimeout` across other specs (`combat-animations.spec.ts`, `session-mode.spec.ts:10,19,38`, `oneshot.spec.ts:111`, `smoke.spec.ts:12`, etc.) with `expect(...).toBeVisible()`, `waitForResponse`, or `waitForFunction` as appropriate
- [x] 4.2 Audit all 152 `test.slow()` sites (~60 in `oneshot-content.spec.ts`, 14 in `campaign-extended.spec.ts`, 10 in `responsive.spec.ts`, plus `search.spec.ts:46,54,60,66,82`, `combat-animations.spec.ts:32,60`, `session-mode.spec.ts:10,19,38`, `oneshot.spec.ts:111`, `combat-tracker.spec.ts:24,30,37,58,83`, `smoke.spec.ts:12`); remove where no longer needed, keep only justified with comment
- [x] 4.3 Run `task test:e2e` twice; verify total `waitForTimeout` count is 0 (except allowlisted); `rg -n "waitForTimeout" tests/` should be clean

## 5. Split oneshot-content.spec.ts

- [x] 5.1 Split `tests/oneshot-content.spec.ts` (1568L, 62KB) by content type (e.g. `oneshot-content-spells.spec.ts`, `oneshot-content-monsters.spec.ts`, `oneshot-content-items.spec.ts`, etc.); each file <600L; share `login()`/`waitLoadingDone` imports; no logic change
- [x] 5.2 Delete original `oneshot-content.spec.ts` (or leave shim if needed); run `task test:e2e` twice; verify all content-type tests still pass and no `data-testid` regressions

## 6. Guardrail and final gates

- [x] 6.1 Add grep guardrail forbidding new `waitForTimeout` — `rg -n "waitForTimeout" tests/` check in `.github/workflows/ci.yaml` or `Taskfile` `lint:e2e` task (and `prek` hook if feasible); allowlist via `// allow-waitForTimeout: reason` comment; document rule in `CONTRIBUTING.md` or `AGENTS.md`
- [x] 6.2 Run `task test` (Go) + `npm run test:unit` + `npm run typecheck` — must pass
- [x] 6.3 Final `task test:e2e` twice and full `task ci` must pass; confirm `flaky-tests.log` trends down vs. before; verify `openspec validate` if available and `openspec status` shows all artifacts done
