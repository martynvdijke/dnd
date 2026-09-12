# backend-modularity Specification

## Purpose
TBD - created by archiving change split-go-god-files. Update Purpose after archive.
## Requirements
### Requirement: Shared handler response helpers

The system SHALL provide shared helpers in `handlers/respond.go` (e.g., `BindOr400`, `WriteError`, `WriteNotFound`, `WriteJSON`) that centralize JSON binding and error-response formatting, and handler call sites SHALL migrate to these helpers incrementally so that raw `ShouldBindJSON`+400 and raw `c.JSON(500/400/403, gin.H{"error": …})` stanzas are eliminated from migrated files.

#### Scenario: Binding helper returns 400 on invalid payload
- **WHEN** a handler calls `BindOr400` with a malformed JSON body
- **THEN** the helper writes a 400 response with an error payload and the handler returns without further processing

#### Scenario: Error helpers preserve existing status and shape
- **WHEN** a handler calls `WriteError` or `WriteNotFound` with an error
- **THEN** the response status and JSON `{"error": …}` shape are identical to the pre-refactor behavior

### Requirement: Campaign auth context helpers

The system SHALL provide helpers to centralize `user_id` extraction from `gin.Context` and campaign membership checks (e.g., `MustGetUserID`, `ParseCampaignID`, `IsCampaignMember`), and campaign handlers SHALL use these helpers instead of re-implementing `c.Get("user_id")` and `isCampaignMember` per handler.

#### Scenario: Membership helper authorizes a member
- **WHEN** `IsCampaignMember` is called for a user who is owner or a row in `campaign_members` for that campaign
- **THEN** it returns true and the handler proceeds

#### Scenario: Helpers reject missing or non-member callers
- **WHEN** `MustGetUserID` finds no `user_id` in context or `IsCampaignMember` finds no membership
- **THEN** the helper signals absence and the handler returns 401 or 403 as before

### Requirement: God files split into focused files (package handlers preserved)

Each god file SHALL be split into topical focused files within package `handlers` (no package rename), with no change to exported symbols or HTTP behavior: `handlers/oneshot.go` (3,816L), `handlers/campaign.go` (2,172L), `handlers/characters.go` (2,112L), `handlers/compendium_admin.go` (2,083L), and `handlers/htmx.go` / `compendium_htmx.go` route tables, each resulting file targeting <800L.

#### Scenario: File split preserves behavior
- **WHEN** a god file is split into its topical files and the test suite runs
- **THEN** all existing handler tests pass and no HTTP contract changes are observed

#### Scenario: Package remains handlers
- **WHEN** the split is applied
- **THEN** imports in `main.go`, tests, and other packages require no changes because the package name is unchanged

### Requirement: Campaign graph handler decomposed

`GetCampaignGraphData` (304L, 8 ent queries + manual dedup) SHALL be decomposed into smaller helpers (e.g., per-entity query functions and a dedup helper) within the campaign file slice, preserving the exact JSON response shape and status codes.

#### Scenario: Graph endpoint returns identical data after decomposition
- **WHEN** `GET /api/campaigns/:id/graph` is called before and after the refactor with the same DB state
- **THEN** the response bodies are byte-identical and all edge-case tests still pass

### Requirement: Data-driven HTMX route registration

`HtmxRegisterRoutes` (276L) and the compendium HTMX route registration SHALL be rewritten as a data-driven table (slice of route descriptors) iterated by a registrar, with no change to registered methods, paths, or middleware.

#### Scenario: Route table snapshot unchanged
- **WHEN** `router.Routes()` is snapshotted before and after the rewrite
- **THEN** the set of registered routes (method + path + handler identity) is identical

### Requirement: Migrations split into per-migration files with registry

`db/migrations.go` (2,092L) SHALL be split into per-migration files under `db/migrations/` with an ordered registry; `Migrate()`, `ApplySafeAlters()`, and `rebuildNPCItemLinksIfNeeded` SHALL be isolated into their own files, preserving schema, seed data, and migration ordering.

#### Scenario: Fresh DB migrates to identical schema
- **WHEN** `Migrate()` runs on an empty DB before and after the split
- **THEN** the resulting schema (and `SELECT sql FROM sqlite_master` hash) is identical and `go test ./db` passes

#### Scenario: Registry ordering is tested
- **WHEN** the registry is inspected
- **THEN** its length equals the number of migration files and its order matches the intended sequence

### Requirement: Single parameterized compendium import implementation

The three near-identical functions `ImportCompendiumEntries` / `ImportCompendiumEntriesWithMapping` / `ImportCompendiumBatchJSON` SHALL be consolidated into one parameterized internal function (e.g., `importCompendiumEntries(opts ImportOpts)`) with thin wrappers, eliminating duplicated `LastInsertId()` and ~70% shared logic, with no change to import semantics or API responses.

#### Scenario: All three import paths produce identical results after consolidation
- **WHEN** each import path is exercised with the same input before and after consolidation
- **THEN** the DB state, returned counts, and error handling are identical

#### Scenario: LastInsertId handling is centralized
- **WHEN** the import implementation is inspected
- **THEN** `LastInsertId()` is handled in exactly one place with proper error checking

### Requirement: Main wiring extracted

`main()` (173L) SHALL be decomposed so that router construction, scheduler registration, and route-group mounting are extracted into testable functions (e.g., `setupRouter()`, `initMedia()`, `registerSchedulers()`), with `main()` retaining only flag parsing, DB open, and `Run()`.

#### Scenario: Router setup is testable without starting the server
- **WHEN** `setupRouter()` is called in a test
- **THEN** it returns a fully-wired `*gin.Engine` whose routes match the pre-refactor server without side effects

### Requirement: Behavior-preserving and coverage-safe

The entire refactor SHALL be behavior-preserving (no API contract changes) and SHALL keep coverage floors green after every phase: Go total ≥20%, `handlers/` ≥30%, `middleware/` ≥25%; each phase SHALL be verified with `go build`, `go vet`, and `task test` before proceeding.

#### Scenario: Coverage floors hold after each phase
- **WHEN** `go test -coverprofile` is run after any phase
- **THEN** total, `handlers/`, and `middleware/` coverages remain at or above their floors

#### Scenario: No API drift
- **WHEN** the full test suite and a route snapshot are compared before and after the change
- **THEN** no HTTP status, JSON shape, or route has changed
