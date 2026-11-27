# pre-release-regression-suite Specification

## Purpose
A tagged Playwright regression suite (`tests/regression.spec.ts`) verifies core views and key flows before releases, runnable selectively with `npx playwright test --grep @regression`.

## Requirements

### Requirement: Tagged regression suite exists

A regression suite SHALL exist in `tests/regression.spec.ts` with each test tagged `@regression`, following the existing patterns from `tests/smoke.spec.ts`, so it can run selectively on demand or as part of CI without a separate workflow.

#### Scenario: Selective execution

- **WHEN** `npx playwright test --grep @regression` runs
- **THEN** only the regression-tagged tests execute

### Requirement: Core view navigation regression

The suite SHALL navigate each major view (Characters, Compendium, Dice, Encounters, and the other core views) and verify that key elements exist, not just the view heading.

#### Scenario: Core views render with key elements

- **WHEN** the regression suite navigates a core view
- **THEN** the view's key elements are verified to exist, not merely its heading

### Requirement: Compendium and API health regressions

The suite SHALL regress the compendium legacy tabs (each tab loads non-empty content), the schema entry lifecycle (create via admin API, verify in the user compendium view, delete, verify removed), and API response times (key endpoints respond under 2 seconds).

#### Scenario: Legacy compendium tabs load content

- **WHEN** the suite clicks each legacy compendium tab
- **THEN** each tab loads a non-empty list

#### Scenario: Schema entry lifecycle round-trip

- **WHEN** the suite creates a schema entry via the admin API
- **THEN** the entry is visible in the user compendium view
- **AND** after deletion the entry is removed from the view

#### Scenario: API response times stay fast

- **WHEN** the suite calls key API endpoints
- **THEN** each responds in under 2 seconds
