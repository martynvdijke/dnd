# coverage-enforcement Specification

## Purpose
Per-package and TS coverage gates with documented thresholds.

## Requirements

### Requirement: Coverage thresholds are per-package
The CI pipeline SHALL enforce separate coverage thresholds for each Go package instead of a single total threshold. The initial thresholds SHALL be:
- `./handlers/...` — 40%
- `./middleware/...` — 60%
- Root package (`./...` excluding handlers/middleware) — 20%

#### Scenario: Package meets threshold
- **WHEN** `go test -coverprofile=coverage.out -coverpkg=./handlers/...,./middleware/... ./...` runs
- **THEN** `go tool cover -func=coverage.out` output is parsed per package prefix
- **THEN** each package above its threshold prints `PASS: ...`
- **THEN** CI continues

#### Scenario: Package below threshold
- **WHEN** a package's coverage drops below its threshold
- **THEN** the CI step prints `FAIL: handlers/ coverage 38.2% < 40% threshold`
- **THEN** CI exits with non-zero code

### Requirement: TypeScript coverage is measured
The CI pipeline SHALL run `vitest run --coverage` for TypeScript unit tests and enforce a minimum coverage threshold.

#### Scenario: TS coverage passes
- **WHEN** `vitest run --coverage` executes
- **THEN** coverage is reported for all TS files matched by `vitest.config.ts`
- **THEN** if coverage is at or above the threshold, CI passes

### Requirement: Coverage is a separate CI job
Coverage enforcement SHALL run in the `unit-test` CI job, not in the `e2e` job. Coverage failure SHALL NOT block the E2E tests.

#### Scenario: Coverage fails, E2E still runs
- **WHEN** the `unit-test` job fails due to coverage threshold
- **THEN** the `e2e` job is skipped
- **THEN** the overall workflow status is failure

### Requirement: Coverage threshold is documented
The coverage thresholds SHALL be defined in `ci.yaml` as explicit variables at the top of the file, with a comment explaining how to update them.

#### Scenario: Thresholds are visible
- **WHEN** a developer opens `.github/workflows/ci.yaml`
- **THEN** coverage thresholds are defined as top-level environment variables or as constants in the coverage step
- **THEN** each threshold has a comment explaining which package it covers and why that number was chosen
