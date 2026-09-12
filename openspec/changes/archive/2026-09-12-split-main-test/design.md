# Design: split-main-test

## Context

`main_test.go` is the legacy integration suite: 6578 lines, 124 `Test*` funcs, 134 total funcs, package `main` at repo root. Helpers occupy lines 23-511 (`TestMain` at :25, global `testRouter *gin.Engine` at :23, `buildRouter` at :49 ~400 lines duplicating `main.go` wiring, `testClient` with `req/get/post/put/del`, `setupAdmin`, `login`, `readJSON`). It covers auth, characters, dice, compendium, campaigns, locations/NPCs, sessions/quests/journal, party, stats, graph, rest/level-up, import/export, admin users, seed data, proficiencies, features/spells, death saves, spell filtering, and authorization in one file. Per-package tests already exist in `handlers/` (25+ files, e.g. `handlers/compendium_admin_test.go` 34KB); `telemetry_test.go` (294L, 9 funcs) shows the target small-file pattern. `main.go:27 main()` wires routes inline, so `buildRouter` is a drift-prone copy. `split-go-god-files` also plans to extract `setupRouter()` from `main.go`.

Stakeholders: backend contributors, reviewers, CI. Constraint: behavior-preserving, tests green at every step, minimize churn.

## Goals / Non-Goals

**Goals:**
- Partition `main_test.go` into reviewable domain files (<800L each, target <600L) preserving `package main` to avoid import churn.
- Extract shared helpers into a single `testutil_test.go` (or `main_testutil_test.go`) at repo root.
- Single-source router construction: tests call the same `setupRouter()`/`newRouter()` that `main.go` uses.
- Preserve all 124 tests; test count before/after must match exactly.

**Non-Goals:**
- Moving integration tests to `handlers/testutil` package or `tests/` — deferred; keeping `package main` at root minimizes risk (alternative evaluated but rejected for this change).
- Changing test logic, assertions, or fixtures beyond file moves and helper imports.
- Splitting `handlers/` per-package tests (already modular) or `telemetry_test.go`.
- Changing coverage floors or adding new coverage in this change (covered by `close-coverage-holes`).

## Decisions

### D1: Keep `package main` at repo root; split into `main_*_test.go` + `testutil_test.go`

Rationale: `main_test.go` is `package main` so it can call unexported `main.go` symbols. Moving to `handlers` or a new `integration` package would require exporting internals or adding test-only interfaces — larger churn and circular import risk. File-level split within `package main` is mechanical (`git mv` + helper extraction) and compiler-verified. Reference: `telemetry_test.go` already lives at root as `package main` in small-file form.

Alternatives: Move integration tests to `handlers/testutil` or new `integration_test` package — rejected: requires exposing `main.go` wiring via exported API, rewiring imports across 124 tests, higher review cost. Deferred as optional follow-up after split proves stable.

### D2: Shared helpers in `testutil_test.go` (root) — not `handlers/testutil/testutil.go`

`handlers/testutil/testutil.go` exists for per-package handler tests (fake ent client, Gin context helpers). Root integration tests need `testRouter`, `buildRouter`, `testClient`, `setupAdmin`, `login`, `readJSON` — which operate on the full Gin engine. Keep them at root to avoid cross-package test dependency and to preserve `package main` visibility. If helpers overlap, factor common bits (e.g., `readJSON`) into `handlers/testutil` later as a thin re-export.

Alternatives: Consolidate into `handlers/testutil` — rejected for this change due to `package main` visibility and the need to keep root tests self-contained.

### D3: Eliminate `buildRouter` drift via `main.go:setupRouter()` extraction (shared with `split-go-god-files`)

Refactor `main.go` to expose `setupRouter(db, cfg) *gin.Engine` (or `newRouter()`) that mounts all route groups. `buildRouter` in tests becomes a thin wrapper calling `setupRouter` with a test DB/config, or is deleted entirely. This is the same extraction `split-go-god-files` needs (its D7). Coordinate: either this change lands the extraction first (small PR) or depends on that change's `setupRouter` — document dependency and sequence.

Alternatives: Keep duplicating wiring in `buildRouter` — rejected: 400-line copy is the primary drift risk the audit flagged.

### D4: Domain file boundaries — 7 files mirroring natural test clusters

Proposed partition (tune after `grep -n "^func Test"` grouping):
- `main_auth_test.go` — auth, authorization, `canViewCharacter`, session, API token flows.
- `main_characters_test.go` — characters CRUD, proficiencies, features/spells, death saves, spell filtering, DM notes, import/export.
- `main_campaign_test.go` — campaigns, locations, NPCs, sessions, quests, journal, party, graph, rest/level-up.
- `main_compendium_test.go` — compendium CRUD, search, import, admin compendium.
- `main_dice_test.go` — dice, rolls, stats.
- `main_admin_test.go` — admin users, seed data, OTel settings, cleanup, resync.
- `main_misc_test.go` — remaining edge cases, telemetry-adjacent, utilities.

Each file target <800L. Helpers and `TestMain` stay in `testutil_test.go`.

Alternatives: Finer split (10+ files) — deferred; 7 files balance reviewability and `git log --follow` traceability.

### D5: Verification — test-count invariant + `go vet` + `task test` per phase

After each file move, assert `go test -list Test | wc -l` == 124 and `go test ./...` green. Use `git mv` so Git tracks renames. Sequence one domain file per commit to keep diffs <800L.

## Risks / Trade-offs

- [Large diff hides behavior change] → Mitigation: pure file moves only; one domain file per commit/phase; `go vet` + `task test` green between phases; reviewers see <800L diffs.
- [Router extraction breaks `main.go` or tests] → Mitigation: extract `setupRouter` as a small additive function; keep `main()` thin; add `TestSetupRouter` smoke test; coordinate with `split-go-god-files` to avoid merge conflict (agree on function name/signature first).
- [Helper extraction causes import cycle] → Mitigation: helpers stay `package main` at root; no cross-package test imports.
- [Merge conflicts with in-flight branches] → Mitigation: announce file-move commits; prefer `git mv`; sequence after or in lockstep with `split-go-god-files`.
- [Coverage miscount after split] → Mitigation: file moves don't change coverage; verify `go test -coverprofile` floors still pass (total ≥20%, handlers ≥30%, middleware ≥25%).

## Migration Plan

1. **Phase 1 — Router extraction** — refactor `main.go` to expose `setupRouter()`/`newRouter()`; `buildRouter` in tests calls it; `task test` green; coordinate with `split-go-god-files` (dependency).
2. **Phase 2 — Helper extraction** — create `testutil_test.go` with `TestMain`, `testRouter`, `buildRouter` wrapper, `testClient`, `setupAdmin`, `login`, `readJSON`; `main_test.go` imports helpers (same package, no import needed — just move defs); `task test` green; test count 124.
3. **Phase 3 — Domain splits (one file per commit)** — `main_auth_test.go` → `main_characters_test.go` → `main_campaign_test.go` → `main_compendium_test.go` → `main_dice_test.go` → `main_admin_test.go` → `main_misc_test.go`; delete moved tests from `main_test.go` each step; assert test count after each.
4. **Phase 4 — Cleanup** — delete now-empty `main_test.go` (or leave as shim re-export if needed); final `task ci` including `task test:e2e` (rebuilds `villum-server`).

Rollback: each phase is a standalone commit; revert the phase commit. Helpers are additive.

## Open Questions

- Exact `setupRouter` signature — confirm `db` and config params with `split-go-god-files` owners before implementing.
- Whether to keep `main_test.go` as a 10-line shim for `git log --follow` history — tentative: delete after split, rely on `git log --follow main_auth_test.go`.
- Final file count — confirm via `grep -n "^func Test"` grouping before cutting files; adjust boundaries if any file exceeds 800L.

## Verification (2026-09)

**1.1 Signature agreed with split-go-god-files:** `func setupRouter(mediaPath string) (*gin.Engine, func())` in `router.go:20` (package `main`). Returns engine + shutdownTelemetry closure. Consumed by `main.go:76` (`r, shutdown := setupRouter(mediaPath)`) and `main_test.go:buildRouter` thin wrapper. Coordinated via `app.go:initMedia`/`registerSchedulers` extraction; `split-go-god-files` owns `router.go`/`app.go`, this change consumes it.
