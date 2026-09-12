## 1. Foundation — shared helpers

- [x] 1.1 Create `handlers/respond.go` with `BindOr400(c, obj) bool`, `WriteError(c, status, err)`, `WriteNotFound(c, msg)`, `WriteJSON(c, status, payload)` wrappers (centralize `gin.H{"error": err.Error()}` shape); add unit tests for each helper (400 on bad bind, 500/403/404 shapes).
- [x] 1.2 Create `handlers/auth_helpers.go` (or `handlers/campaign_auth.go`) with `MustGetUserID(c) (int64, bool)`, `ParseCampaignID(c) (int64, bool)`, `IsCampaignMember(ctx, db, campaignID, userID) (bool, error)`; add unit tests (missing user_id → false, member/owner/admin → true).
- [x] 1.3 Migrate `handlers/campaign.go` call sites to the new helpers (replace ~15× `userID,_ := c.Get("user_id")`, per-handler campaignID parse + membership, and ~20 `ShouldBindJSON`+400 stanzas in campaign-adjacent handlers); run `go vet ./...` and `task test`.
- [x] 1.4 Migrate `handlers/characters.go` and one additional handler file to `BindOr400`/`WriteError`; grep to confirm reduction in raw `c.JSON(500`/`c.JSON(400` stanzas; run `go vet`, `task test`.
- [x] 1.5 Verification: `go build ./...` passes; `go vet ./...` clean; `task test` green; coverage floors hold (`go test -coverprofile` total ≥20%, handlers/ ≥30%, middleware/ ≥25%).

## 2. Split handlers/campaign.go (2,172L)

- [x] 2.1 Inventory `handlers/campaign.go` (`grep -n "^func"`) and define topical slices: `campaign_core.go` (campaigns CRUD), `campaign_locations.go`, `campaign_npcs.go`, `campaign_sessions.go`, `campaign_quests.go`, `campaign_graph.go`, `campaign_party.go` (adjust to actual func grouping); create files and `git mv`-style moves.
- [x] 2.2 Decompose `GetCampaignGraphData` (304L, 8 ent queries + manual dedup) into helpers: per-entity query funcs + `deduplicateGraphNodes/Edges`; keep exact JSON shape; add a golden-file or snapshot test for graph response.
- [x] 2.3 Wire `RequireCampaignMember`/`IsCampaignMember` throughout the new campaign files; remove per-handler re-implementations.
- [x] 2.4 Verification: `go build ./...` passes; `go vet ./...` clean; `task test` green (esp. `go test ./handlers -run Campaign`); no file exceeds 800L; coverage floors hold.

## 3. Split handlers/oneshot.go (3,816L)

- [x] 3.1 Inventory `handlers/oneshot.go` (136 funcs) and slice into `oneshot_crud.go` (adventures/acts/scenes/dialogs/NPCs/locations CRUD), `oneshot_generation.go`, `oneshot_htmx.go` (HTMX handlers), `oneshot_mapping.go` (ent↔model mapping); keep `package handlers`.
- [x] 3.2 Move ent↔model mapping helpers into `oneshot_mapping.go` with unit tests for representative mappers.
- [x] 3.3 Isolate HTMX handlers (the ~30 handlers) into `oneshot_htmx.go`; verify `go test ./handlers -run Oneshot` and HTMX route tests still pass.
- [x] 3.4 Verification: `go build ./...` passes; `go vet ./...` clean; `task test` green; no resulting file exceeds 800L; coverage floors hold.

## 4. Split handlers/characters.go (2,112L) and htmx route tables

- [x] 4.1 Split `handlers/characters.go` into `characters_crud.go`, `characters_import.go` (`ImportCharacterJSON` 172L + `importCharacters` 145L), `characters_text.go` (`characterToText` 172L); keep shared helpers in `characters_crud.go` or `characters_helpers.go`.
- [x] 4.2 Rewrite `handlers/htmx.go:1539 HtmxRegisterRoutes` (276L) and `handlers/compendium_htmx.go` registration as data-driven `[]Route{Method, Path, Handler, Middleware}` tables with a registrar loop; add a route-snapshot test (`router.Routes()` before/after identical).
- [x] 4.3 Verification: `go build ./...` passes; `go vet ./...` clean; `task test` green; route snapshot identical; coverage floors hold.

## 5. Split db/migrations.go (2,092L)

- [x] 5.1 Create `db/migrations/` with per-migration files (`001_initial.go`, `002_*.go`, …) and `registry.go` (ordered `[]Migration`); keep `db/migrations.go` as a shim re-exporting `Migrate()` initially to avoid import churn, then point callers at the registry.
- [x] 5.2 Isolate `ApplySafeAlters()` and `rebuildNPCItemLinksIfNeeded()` into `db/migrations/safe_alters.go` and `db/migrations/npc_item_links.go`; preserve exact SQL/seed behavior.
- [x] 5.3 Add a test asserting `len(registry) == count(migration files)` and that `Migrate()` on a fresh DB yields the same `sqlite_master` hash as before; run `go test ./db -count=1`.
- [x] 5.4 Verification: `go build ./...` passes; `go vet ./...` clean; `task test` green; no file exceeds 800L; coverage floors hold.

## 6. Split handlers/compendium_admin.go and consolidate imports

- [x] 6.1 Split `handlers/compendium_admin.go` (2,083L) into `compendium_admin_schemas.go` (admin schemas) and `compendium_admin_import.go` (batch import/mapping/export).
- [x] 6.2 Consolidate `ImportCompendiumEntries` (227L at :569), `ImportCompendiumEntriesWithMapping` (236L at :1589), `ImportCompendiumBatchJSON` (228L at :1829) into one parameterized `importCompendiumEntries(ctx, db, opts ImportOpts) (ImportResult, error)` with thin wrappers; centralize `LastInsertId()` error handling (single site); dedup ~70% shared logic.
- [x] 6.3 Migrate remaining raw `c.JSON`/`ShouldBindJSON` stanzas in compendium files to `BindOr400`/`WriteError`; grep to confirm no remaining `logID,_ = result.LastInsertId()` ignores.
- [x] 6.4 Verification: `go build ./...` passes; `go vet ./...` clean; `task test` green (esp. `go test ./handlers -run Compendium`); coverage floors hold.

## 7. Extract main.go wiring and final gates

- [x] 7.1 Extract `main.go:27 main()` (173L) into `setupRouter()`, `initMedia()`, `registerSchedulers()` (or `app.go`/`router.go`) keeping `main()` to flag parsing + DB open + `Run()`; make router construction testable.
- [x] 7.2 Add `TestSetupRouter` (assert routes/middleware match pre-refactor snapshot) and ensure `go test ./...` covers the new wiring.
- [x] 7.3 Final verification: `go build ./...` passes; `go vet ./...` clean; `task ci` passes (full gates incl. coverage floors: total ≥20%, handlers/ ≥30%, middleware/ ≥25%); no god file remains above 800L; `grep -R "c\.JSON(500"` shows only centralized helper usage in migrated files.
