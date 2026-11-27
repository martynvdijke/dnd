# visual-regression-tests Specification

## Purpose
Playwright screenshot tests guard the visual layout of critical pages against unintentional regressions.

## Requirements

### Requirement: Screenshot tests cover critical pages

`tests/visual.spec.ts` SHALL capture screenshots of the compendium page with legacy tabs loaded, the compendium page with a dynamic schema tab visible, the desktop navigation sidebar, and the character list grid, comparing them against reference screenshots stored in `tests/visual-refs/`.

#### Scenario: Reference screenshots are committed and compared

- **WHEN** the visual suite runs after login and navigation
- **THEN** each critical page is compared against its reference screenshot in `tests/visual-refs/`

#### Scenario: Schema tab screenshot uses real data

- **WHEN** the compendium schema-tab screenshot is captured
- **THEN** a schema entry is created via the admin API first so the dynamic tab is visible in the shot

### Requirement: Visual suite runs deterministically and selectively

The visual suite SHALL run in the main CI `chromium` project, tolerate minor rendering differences with `maxDiffPixels: 100`, retry once on failure, and be selectable with the `@visual` tag (`npx playwright test --grep @visual`). Reference snapshots SHALL be updatable with `npx playwright test --update-snapshots` when intentional UI changes happen.

#### Scenario: Minor rendering differences are tolerated

- **WHEN** a captured screenshot differs from its reference by at most 100 pixels
- **THEN** the comparison passes

#### Scenario: Selective execution

- **WHEN** a developer runs `npx playwright test --grep @visual`
- **THEN** only the visual regression tests run

#### Scenario: Intentional UI changes update references

- **WHEN** an intentional UI change alters a page's appearance
- **THEN** running the suite with `--update-snapshots` refreshes the committed reference screenshots
