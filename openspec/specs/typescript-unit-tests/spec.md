# typescript-unit-tests Specification

## Purpose
Vitest unit tests cover the TypeScript bundles' pure functions, running locally via `npm run test:unit` and enforced in CI and pre-commit.

## Requirements

### Requirement: Vitest unit test runner with happy-dom environment

The project SHALL use `vitest` with the `happy-dom` environment (configured in `vitest.config.ts`) as its TypeScript unit test runner, exposed as `npm run test:unit` and invocable as `npx vitest run`.

#### Scenario: Unit tests run on demand

- **WHEN** a developer runs `npm run test:unit`
- **THEN** vitest executes the unit test suite (`ts/**/*.test.ts`) under the happy-dom environment

#### Scenario: Unit tests run in CI and pre-commit

- **WHEN** CI runs the TypeScript unit-test step
- **THEN** vitest runs with coverage collection
- **WHEN** a commit touches TypeScript sources
- **THEN** the pre-commit `vitest` hook runs `npx vitest run` before the commit is accepted

### Requirement: Compendium render functions are unit tested

Pure functions exported from `ts/compendium.ts` (`renderSchemaEntries`, `entryPreview`) SHALL be covered by unit tests in `ts/compendium.test.ts`.

#### Scenario: renderSchemaEntries card markup

- **WHEN** `renderSchemaEntries` is called with an entry that has a name
- **THEN** it returns the correct HTML card markup for that entry

#### Scenario: View All link depends on entry count

- **WHEN** an entry has `entry_count > 20`
- **THEN** the rendered card includes a "View All N entries" link
- **WHEN** an entry has `entry_count <= 20`
- **THEN** the link is omitted

#### Scenario: Empty entries array renders nothing

- **WHEN** `renderSchemaEntries` is called with an empty entries array
- **THEN** it returns an empty string

#### Scenario: entryPreview key-value previews

- **WHEN** `entryPreview` is called with entry data
- **THEN** it returns up to 3 key-value pairs, skipping the `name` field
- **AND** it handles null/undefined data and empty data objects gracefully without throwing
