# e2e-test-standards Specification

## Purpose
Test selector conventions, timeout standards, and helper patterns.

## Requirements

### Requirement: Tests use data-testid selectors
All E2E tests SHALL use `[data-testid="..."]` locators for interactive elements instead of text-based selectors (`:has-text()`, text content matching) or fragile CSS class selectors. The following elements SHALL have `data-testid` attributes:
- Sidebar navigation buttons (`data-testid="nav-{name}"`)
- Modal trigger buttons (`data-testid="modal-{action}"`)
- Character cards (`data-testid="character-card-{id}"`)
- Tab containers (`data-testid="tab-{name}"`)
- Form submit buttons (`data-testid="btn-{action}"`)

#### Scenario: Navigation works with data-testid
- **WHEN** a test clicks `[data-testid="nav-compendium"]`
- **THEN** the compendium view renders
- **THEN** no strict-mode violation occurs (single matching element)

#### Scenario: Text-based selector removed
- **WHEN** the UI text changes (e.g., "Character Folio" → "Characters")
- **THEN** tests using `data-testid` continue to pass
- **THEN** tests using text selectors may fail (migration is required)

### Requirement: Tests use shared login helper
All E2E tests that require authentication SHALL use the shared `login()` helper from `tests/helpers.ts`. Inline login logic (filling username/password fields and clicking submit) SHALL NOT appear in individual spec files.

#### Scenario: Login helper is imported
- **WHEN** a test requires an authenticated session
- **THEN** the test imports `login` from `./helpers.js`
- **THEN** the test calls `await login(page)` before interacting with protected features

#### Scenario: Login helper succeeds
- **WHEN** `login(page)` is called
- **THEN** the page navigates to `/login`
- **THEN** credentials are filled and submitted
- **THEN** `waitLoadingDone` confirms the SPA has initialized
- **THEN** the user is redirected to `/` (or equivalent)

### Requirement: Slow tests are annotated
Every test that performs SPA initialization, async data loading, or runs on mobile viewport SHALL call `test.slow()` at the top of the test function or describe block.

#### Scenario: Slow annotation present
- **WHEN** a test calls `test.slow()`
- **THEN** Playwright triples the default timeout for that test
- **THEN** the test still passes within the extended timeout

### Requirement: Visual tests are isolated
Screenshot-based visual tests SHALL run in a separate spec file (`visual-ci.spec.ts`) with a fixed viewport size and CI-only execution guard. The `maxDiffPixelRatio` SHALL be set to 0.05 for all screenshot comparisons.

#### Scenario: Visual test runs in CI
- **WHEN** `CI=true` and the test runs
- **THEN** screenshots are captured at 1280x720 viewport
- **THEN** comparison uses `maxDiffPixelRatio: 0.05`
- **THEN** the test passes or fails based on pixel diff threshold
