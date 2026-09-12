# e2e-test-reliability Specification

## Purpose
TBD - created by archiving change improve-e2e-reliability. Update Purpose after archive.
## Requirements
### Requirement: Deduplicated E2E helpers

The system SHALL have no stale or duplicated `waitLoadingDone` or `waitModalClosed` definitions in `tests/`; all specs SHALL import the canonical helpers from `tests/helpers.ts` (`waitLoadingDone` at `:14` that awaits overlay hidden plus `window.api` and `__apiReady`, and `waitModalClosed` at `:65`).

#### Scenario: Stale helpers removed

- **WHEN** `tests/campaign.spec.ts`, `tests/campaign-extended.spec.ts`, `tests/character-advanced.spec.ts`, `tests/character.spec.ts`, `tests/companions.spec.ts`, `tests/compendium.spec.ts`, `tests/selection.spec.ts`, `tests/search.spec.ts`, `tests/responsive.spec.ts`, and `tests/character-sheet.spec.ts` are inspected
- **THEN** they contain no local `waitLoadingDone` or `waitModalClosed` definitions and instead import `waitLoadingDone`/`waitModalClosed` from `tests/helpers.ts` (7 `waitLoadingDone` + 5 `waitModalClosed` stale copies removed)

### Requirement: No hardcoded waitForTimeout sleeps

The system SHALL have no `waitForTimeout` hardcoded sleeps in `tests/` (including `tests/helpers.ts:10,94,104,112` and all 104 call sites across `tests/` such as `journal.spec.ts:39,59,62,93,125,131`, `party.spec.ts:161`, `oneshot.spec.ts:228`, `search.spec.ts:10,36,42`, `combat-tracker.spec.ts`, `shops.spec.ts:10,30,48`) except for explicitly allowlisted cases with a documented reason comment; each former sleep SHALL be replaced with a deterministic wait (`expect(locator).toBeVisible()/toBeHidden()`, `page.waitForResponse`, `page.waitForFunction`, or `locator.waitFor()`).

#### Scenario: Sleeps replaced with proper waits

- **WHEN** `rg -n "waitForTimeout" tests/` is run (excluding allowlisted lines with a reason comment)
- **THEN** no matches are found and the E2E suite still passes (`task test:e2e` green) with flake rate not increased (full suite run twice per phase shows no new flakes)

#### Scenario: Helpers use proper waits

- **WHEN** `tests/helpers.ts` is inspected
- **THEN** lines `10,94,104,112` (navbar toggler, bottom-sheet close) use proper waits (`expect(...).toBeVisible()/toBeHidden()` or `waitForFunction`) and contain no `waitForTimeout`

### Requirement: Search slowness fixed and test.slow justified

The system SHALL fix the root cause of slowness (152 `test.slow()` total — ~60 in `oneshot-content.spec.ts`, 5 in `search.spec.ts` at `:46,54,60,66,82`, 14 in `campaign-extended`, 10 in `responsive`, plus `combat-animations`, `session-mode`, `combat-tracker`, `smoke`) and remove `test.slow()` from tests where it is no longer needed; `test.slow()` SHALL remain only where genuinely justified (e.g., heavy generation/seed) with a comment explaining why.

#### Scenario: Search slowness resolved

- **WHEN** `search.spec.ts` is run after the fix
- **THEN** its tests pass without `test.slow()` (or with fewer `test.slow()` markers) and without `waitForTimeout` sleeps, using `waitForResponse`/`waitForFunction` for search readiness

#### Scenario: Slow markers are justified

- **WHEN** any remaining `test.slow()` call is inspected
- **THEN** it is accompanied by a comment justifying why the test is legitimately slow, and the total count of `test.slow()` is reduced from 152 (~60 in `oneshot-content.spec.ts`, 14 in `campaign-extended`, 10 in `responsive`, plus `search.spec.ts:46,54,60,66,82`, `combat-animations.spec.ts:32,60`, `session-mode.spec.ts:10,19,38`, `oneshot.spec.ts:111`, `combat-tracker.spec.ts:24,30,37,58,83`, `smoke.spec.ts:12`)

### Requirement: Split large oneshot spec

The system SHALL split `tests/oneshot-content.spec.ts` (1568 lines, 62KB) into multiple files by content type (e.g. `oneshot-content-spells.spec.ts`, `oneshot-content-monsters.spec.ts`, etc.), each below 600 lines, with no logic change and shared `login()`/`waitLoadingDone` imports.

#### Scenario: Large spec modularized

- **WHEN** `tests/oneshot-content*.spec.ts` files are listed
- **THEN** no single file exceeds 600 lines and the original 1568-line `oneshot-content.spec.ts` no longer exists as a monolith, with all content-type coverage preserved

### Requirement: Guardrail forbidding new waitForTimeout

The system SHALL have a CI or `prek` guardrail that fails on new `waitForTimeout` uses in `tests/` (e.g. `rg -n "waitForTimeout" tests/` check in `ci.yaml`/`Taskfile` or `prek` hook), documented in `CONTRIBUTING.md` or `AGENTS.md`, with an allowlist mechanism for legitimate polling (comment `// allow-waitForTimeout: reason`).

#### Scenario: New sleeps are blocked

- **WHEN** a contributor adds a new `waitForTimeout` call in `tests/` without an allowlist comment
- **THEN** CI (or `prek` pre-commit hook) fails with a message directing them to use `expect(locator).toBeVisible()`, `page.waitForResponse`, or `page.waitForFunction` instead

### Requirement: E2E suite remains green with no increased flake rate

The system SHALL keep `task test:e2e` (which rebuilds `villum-server` before running Playwright) green after each phase, with flake rate not increased (verified by running the full suite twice per phase and checking `flaky-tests.log` with `retries: 2`).

#### Scenario: Suite green twice per phase

- **WHEN** any phase of this change is completed
- **THEN** `task test:e2e` passes on two consecutive runs and `flaky-tests.log` shows no increase in flaky tests compared to before the phase
