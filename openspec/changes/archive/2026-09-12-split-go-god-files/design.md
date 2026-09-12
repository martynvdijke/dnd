## Context

The Go backend is a Gin application using `modernc.org/sqlite` and ent. Handlers live in `handlers/` (single package `handlers`), middleware in `middleware/`, migrations in `db/migrations.go`. An audit on `main` found five files above 2,000 lines (largest 3,816L) and several functions above 170L, with pervasive duplication: ~20 copies of the `ShouldBindJSON`+400 pattern, ~770 `c.JSON` error responses (277× 500, 152× 400, 77× 403), `userID,_ := c.Get("user_id")` 15+ times in one file, campaign membership re-implemented per handler, and three compendium import functions sharing ~70% logic. `main()` wires init, five schedulers, and four route groups inline. `db/migrations.go` bundles schema, seed, `Migrate()`, `ApplySafeAlters()`, and `rebuildNPCItemLinksIfNeeded` in one 2,092L file. `handlers/htmx.go:1539 HtmxRegisterRoutes` (276L) is an imperative route table that should be data-driven.

Stakeholders: backend contributors, reviewers, CI coverage gates. Constraint: behavior-preserving throughout; no API contract changes; tests green at every step.

## Goals / Non-Goals

**Goals:**
- Reduce every file to a reviewable size (target <600L, hard cap <800L for the largest topical slice) without changing package structure or public APIs.
- Eliminate the duplicated error/binding/auth stanzas via shared helpers, migrated incrementally.
- Split `db/migrations.go` into per-migration files with an explicit registry.
- Consolidate the three compendium import functions into one parameterized implementation.
- Make route registration data-driven and extract `main.go` wiring for testability.

**Non-Goals:**
- Splitting package `handlers` into many subpackages (file-level split first; subpackage only where clearly warranted, e.g., an optional `handlers/oneshot` subpackage after the file split proves stable).
- Changing ent schemas, HTTP contracts, or frontend behavior.
- Reformatting business logic beyond what file moves require; refactors stay mechanical.
- Adding new features or changing error semantics.

## Decisions

### D1: File-level split within package `handlers` (keep `package handlers`)

Rationale: changing the package name forces import churn across `main.go`, tests, and middleware. File-level splits keep `git blame` traceable (via `git log --follow`) and let the compiler verify no API drift. A subpackage is only introduced where the boundary is unambiguous (candidate: `oneshot` → `handlers/oneshot` after the initial file split, as a separate PR).

Alternatives: immediate subpackage per domain (`handlers/campaign`, `handlers/compendium`) — rejected: large diff, circular dependency risk between campaign/character/compendium handlers, and higher review cost. Deferred until file splits are merged and green.

### D2: Shared response helpers in `handlers/respond.go` before any god-file split

Create `BindOr400(c, &req) bool`, `WriteError(c, status, err)`, `WriteNotFound(c, msg)`, `WriteJSON(c, status, payload)` (thin wrappers over `c.JSON`/`c.AbortWithStatusJSON` that centralize the `gin.H{"error": err.Error()}` shape). Migrate call sites incrementally, one file per commit, so diffs stay reviewable and `grep` can verify zero remaining raw stanzas at the end.

Alternatives: helpers in `middleware/` — rejected: they are handler-specific (Gin context) and used only inside `handlers/`.

### D3: Campaign auth helpers as small pure functions, not new middleware

Extract `MustGetUserID(c) (int64, bool)`, `ParseCampaignID(c) (int64, bool)`, `IsCampaignMember(ctx, db, campaignID, userID) (bool, error)` into `handlers/auth_helpers.go` (or `handlers/campaign_auth.go`). Handlers call them explicitly; no implicit context mutation. This preserves testability (helpers unit-tested with a mock `*gin.Context` or a fake ent client) and avoids changing the middleware chain.

Alternatives: Gin middleware that injects `campaign_id`/`is_member` into context — rejected: adds hidden coupling, complicates handler tests, and requires auditing every route's middleware order.

### D4: Per-migration files with a registry in `db/migrations/`

Split `db/migrations.go` into `db/migrations/001_initial.go`, `002_*.go`, … plus `registry.go` that orders and runs them. `ApplySafeAlters()` and `rebuildNPCItemLinksIfNeeded()` become standalone files. `Migrate()` iterates the registry. Seed data stays with its migration or in `seed.go`. Behavior identical; `go test ./db` covers ordering.

Alternatives: keep one file and use `//go:embed` SQL files — useful but orthogonal; the registry already enables that later without blocking the split.

### D5: Consolidate three compendium import functions into one parameterized implementation

`ImportCompendiumEntries`, `ImportCompendiumEntriesWithMapping`, `ImportCompendiumBatchJSON` share ~70% logic and triplicate `LastInsertId()` handling. New internal `importCompendiumEntries(ctx, db, opts ImportOpts) (ImportResult, error)` with `ImportOpts{Mapping, BatchMode, …}`; existing exported functions become thin wrappers (or are deprecated behind the new one) so callers need not change immediately. Dedup eliminates the repeated error-ignore at `logID,_ = result.LastInsertId()`.

Alternatives: three separate refactors — rejected: leaves duplication; single parameterized path is the only way to prevent regression.

### D6: Data-driven route tables

`HtmxRegisterRoutes` (276L) and `compendium_htmx.go` registration become a `[]Route{Method, Path, Handler, Middleware}` slice iterated by a registrar. No behavior change; the table is diff-friendly and enables exhaustive route tests.

### D7: Extract `main.go` wiring into `setupRouter()` / `initMedia()` / `registerSchedulers()`

`main()` retains only flag parsing, DB open, and `Run()`. Router construction, scheduler registration (five schedulers), and route-group mounting move to testable functions in `app.go` / `router.go`. Enables `TestSetupRouter` without starting the server.

## Risks / Trade-offs

- [Large diff hides behavior change] → Mitigation: one god file per phase; each phase is a pure file move + helper wiring with `go vet` and `task test` green before the next phase; reviewers see <800L diffs.
- [Coverage floor breach during churn] → Mitigation: helpers are small and tested; file moves do not change coverage calculation; check `go test -coverprofile` after each phase (floors: total ≥20%, `handlers/` ≥30%, `middleware/` ≥25%).
- [Merge conflicts with in-flight feature branches] → Mitigation: sequence this change after or in lockstep with active feature branches; announce file-move commits; prefer `git mv` so Git tracks renames.
- [Circular imports if subpackages introduced too early] → Mitigation: keep `package handlers` for the main split; subpackage extraction is a separate, later PR with explicit dependency audit.
- [Registry ordering bug in migrations] → Mitigation: registry is an ordered slice with a test asserting `len(registry) == count(files)` and that `Migrate()` on a fresh DB yields the same schema hash as before.
- [Route table regression] → Mitigation: snapshot test of registered routes (`router.Routes()`) before/after the data-driven rewrite.

## Migration Plan

**Order (behavior-preserving; tests green between steps):**

1. **Phase 1 — Helpers** — add `handlers/respond.go` and `handlers/auth_helpers.go` with unit tests; migrate `campaign.go` call sites first (highest duplication), then `characters.go`; verify with `go build`, `go vet`, `task test`.
2. **Phase 2 — `campaign.go` split** — slice into `campaign_core.go`, `campaign_locations.go`, `campaign_npcs.go`, `campaign_sessions.go`, `campaign_graph.go`, `campaign_party.go` (or equivalent topical files); `GetCampaignGraphData` extracted and decomposed into helper queries; `go test ./handlers -run Campaign` green.
3. **Phase 3 — `oneshot.go` split** — slice into `oneshot_crud.go`, `oneshot_generation.go`, `oneshot_htmx.go`, `oneshot_mapping.go`; keep `package handlers`; verify `go test ./handlers -run Oneshot`.
4. **Phase 4 — `characters.go` split** — `characters_crud.go`, `characters_import.go` (`ImportCharacterJSON` + `importCharacters`), `characters_text.go` (`characterToText`); verify.
5. **Phase 5 — `db/migrations.go` split** — create `db/migrations/` with per-migration files + `registry.go`; keep `db/migrations.go` as a shim re-exporting `Migrate()` until callers updated; verify `go test ./db`.
6. **Phase 6 — `compendium_admin.go` + import consolidation** — split schemas vs. import; introduce `importCompendiumEntries(opts)` and thin wrappers; dedup `LastInsertId()` handling; verify `go test ./handlers -run Compendium`.
7. **Phase 7 — `htmx.go` / `compendium_htmx.go` routes + `main.go` wiring** — data-driven route tables; extract `setupRouter`/`initMedia`/`registerSchedulers`; add `TestSetupRouter`; final `task ci`.

Rollback: each phase is a standalone commit/PR; revert the phase commit. Helpers are additive, so reverting a split phase only reconstitutes the original file.

## Open Questions

- Should `oneshot` become a subpackage (`handlers/oneshot`) after the file split? Tentative: yes, but only as a follow-up PR once file splits are merged and coverage is stable.
- Exact file boundaries for `campaign.go` — confirm with `grep -n "^func.*Campaign"` and ent query grouping before finalizing slices.
