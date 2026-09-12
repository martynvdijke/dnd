# Proposal: split-main-test

## Why

`main_test.go` is a 6578-line monolith (124 Test functions, 134 total funcs in package `main` at repo root) that duplicates ~400 lines of route wiring from `main.go`, mixes integration concerns across every domain, and makes review, navigation, and coverage tracking unreliable. Per-package unit tests already live in `handlers/` — the root file is legacy debt that blocks modular test evolution and risks router drift.

## What Changes

- Split `main_test.go` (6578L) into domain-focused files at repo root (package `main` preserved): `main_auth_test.go`, `main_characters_test.go`, `main_campaign_test.go`, `main_compendium_test.go`, `main_dice_test.go`, `main_admin_test.go`, `main_misc_test.go` plus `testutil_test.go` (or `main_testutil_test.go`) holding shared helpers.
- Extract shared helpers (`TestMain`, `testRouter`, `buildRouter`, `testClient` with `req/get/post/put/del`, `setupAdmin`, `login`, `readJSON`) from lines 23-511 into the shared file; keep `package main` to minimize churn.
- Eliminate `buildRouter` drift: refactor `main.go` to expose `newRouter()` / `setupRouter()` (or `buildAppRouter()`) that both `main()` and tests call; coordinate with `split-go-god-files` change which also touches `main.go` wiring (dependency noted in design).
- Preserve all 124 tests — test count before/after must match; no test logic changes, only file moves and import of shared helpers.
- No API or behavior change; pure file reorganization with router reuse.

## Capabilities

### New Capabilities
- `integration-test-structure`: Domain-partitioned integration test suite at repo root with shared test utilities and single-source router construction.

### Modified Capabilities
- None

## Impact

- **Code**: `main_test.go` → 7-8 files at repo root (`main_*_test.go` + `testutil_test.go`); `main.go` gains exported `setupRouter`/`newRouter` helper (small refactor, no behavior change); optional coordination with `split-go-god-files` for router extraction.
- **Tests**: `go test` / `task test` must stay green; test count invariant (124 Test funcs); `telemetry_test.go` (294L) remains as reference small-file pattern.
- **CI/Coverage**: No coverage floor change; floors remain total ≥20%, `handlers/` ≥30%, `middleware/` ≥25%, vitest ≥20%.
- **Risk**: File moves hide logic changes → mitigated by pure-move phases, `go vet`, and test-count assertion.
