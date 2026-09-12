# Spec: integration-test-structure

## ADDED Requirements

### Requirement: Domain-partitioned integration test files

The system SHALL provide domain-partitioned integration test files at the repository root in package `main`, replacing the single `main_test.go` monolith, with each file covering a single domain and remaining below 800 lines.

#### Scenario: Files exist and monolith is removed

- **WHEN** the change is complete
- **THEN** the repository contains domain files such as `main_auth_test.go`, `main_characters_test.go`, `main_campaign_test.go`, `main_compendium_test.go`, `main_dice_test.go`, `main_admin_test.go`, `main_misc_test.go` (or equivalent topical partition) and the original 6578-line `main_test.go` no longer exists as a monolith (or remains only as a minimal shim)

#### Scenario: File size bound

- **WHEN** any new domain test file is measured
- **THEN** its line count is below 800 lines (target below 600 lines)

### Requirement: Shared test utilities extracted

The system SHALL extract shared integration-test helpers (`TestMain`, global `testRouter`, `buildRouter` wrapper, `testClient` with `req/get/post/put/del`, `setupAdmin`, `login`, `readJSON`) into a single shared file at the repository root (e.g. `testutil_test.go` or `main_testutil_test.go`) in package `main`.

#### Scenario: Helpers are centralized

- **WHEN** a contributor searches for `TestMain` or `testClient` definitions
- **THEN** they are defined in exactly one shared file at the repo root and all domain test files use that single definition without duplication

### Requirement: Single-source router construction

The system SHALL provide a single router-construction function (e.g. `setupRouter` or `newRouter`) in `main.go` that both production startup and integration tests call, eliminating the duplicated ~400-line `buildRouter` wiring.

#### Scenario: Tests reuse production router

- **WHEN** integration tests build a test router
- **THEN** they call the same `setupRouter`/`newRouter` function that `main()` uses, and no separate ~400-line route-registration copy remains in test code

### Requirement: Test count invariant

The system SHALL preserve all 124 Test functions after the split, with no tests added, removed, or silently skipped.

#### Scenario: Test count matches before and after

- **WHEN** `go test -list Test` (or equivalent) is run before and after the change
- **THEN** the count of Test functions is identical (124) and `go test ./...` passes

### Requirement: Package and churn minimization

The system SHALL keep the split files in package `main` at the repository root (not moved to a new package) to minimize import churn and preserve access to unexported `main.go` symbols.

#### Scenario: Package remains main at root

- **WHEN** any split test file is inspected
- **THEN** its package clause is `package main` and it resides at the repository root alongside `main.go`

### Requirement: Coordination with split-go-god-files router extraction

The system SHALL coordinate the `setupRouter`/`newRouter` extraction with the `split-go-god-files` change to avoid conflicting refactors of `main.go` (documented dependency and agreed function signature).

#### Scenario: No conflicting main.go refactor

- **WHEN** both changes are applied
- **THEN** `main.go` exposes exactly one router-construction function used by both production and tests, and the two changes do not introduce duplicate or conflicting extractions
