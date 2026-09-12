## Why

Five handler and migration files have grown past 1,700 lines (up to 3,816 lines in `handlers/oneshot.go` with 136 functions) and several functions exceed 170–300 lines. The sprawl mixes unrelated concerns (ent↔model mapping, CRUD, generation, HTMX wiring, migrations, seed data) in single files, hides duplication (~770 `c.JSON` error stanzas, ~20 `ShouldBindJSON+400` copies, three near-identical compendium import functions), and makes reviews, ownership, and coverage tracking unreliable. Splitting is overdue before further feature work compounds the cost.

## What Changes

- **Shared handler response helpers** — new `handlers/respond.go` with `BindOr400`, `WriteError`, `WriteNotFound`, `WriteJSON` (and status-specific wrappers) to replace duplicated `ShouldBindJSON`/`c.JSON(500/400/403, …)` stanzas; migrate call sites incrementally file-by-file.
- **Campaign auth/context helpers** — extracted helpers (`MustGetUserID`, `ParseCampaignID`, `RequireCampaignMember` / `isCampaignMember`) in `handlers/auth_helpers.go` or `middleware/campaign.go` to centralize `c.Get("user_id")` and membership checks duplicated ~15× in `campaign.go`.
- **God-file splits (file-level, package `handlers` preserved)** — one focused file per former god file, no import churn:
  - `handlers/oneshot.go` (3,816L) → `oneshot_crud.go`, `oneshot_generation.go`, `oneshot_htmx.go`, `oneshot_mapping.go` (or equivalent topical slices).
  - `handlers/campaign.go` (2,172L) → `campaign_core.go`, `campaign_locations.go`, `campaign_npcs.go`, `campaign_sessions.go`, `campaign_graph.go`, etc.
  - `handlers/characters.go` (2,112L) → `characters_crud.go`, `characters_import.go`, `characters_text.go`.
  - `handlers/compendium_admin.go` (2,083L) → `compendium_admin_schemas.go`, `compendium_admin_import.go` (consolidated import path).
  - `db/migrations.go` (2,092L) → `db/migrations/*.go` per migration with a registry; `ApplySafeAlters` and `rebuildNPCItemLinksIfNeeded` isolated.
  - `handlers/htmx.go` / `compendium_htmx.go` route tables made data-driven.
- **Compendium import consolidation** — the three near-identical functions `ImportCompendiumEntries` / `ImportCompendiumEntriesWithMapping` / `ImportCompendiumBatchJSON` (~70% shared logic, `logID,_ = result.LastInsertId()` triplicated) become one parameterized implementation with thin wrappers.
- **Main wiring extraction** — `main.go:27 main()` (173L) split into `setupRouter()` and `initMedia()` / scheduler registrars; no behavior change.

## Capabilities

### New Capabilities
- `backend-modularity`: Backend file and function modularity — focused files, shared response/auth helpers, data-driven route tables, per-migration files, and a single parameterized compendium import path, all behavior-preserving with no API changes.

### Modified Capabilities
- None

## Impact

- **Code**: `handlers/*.go`, `db/migrations.go` → `db/migrations/*.go`, `main.go`; new `handlers/respond.go`, `handlers/auth_helpers.go` (or `middleware/campaign.go`). Pure file moves and helper extraction within package `handlers` (no package renames except where clearly warranted, e.g., `oneshot` subpackage as optional follow-up).
- **APIs**: None — behavior-preserving; all HTTP contracts, status codes, and JSON shapes unchanged.
- **Tests/CI**: Existing `go test` suites must stay green at every phase; coverage floors enforced — Go total ≥20%, `handlers/` ≥30%, `middleware/` ≥25%. Risk is churn-induced flake, mitigated by one-god-file-per-phase sequencing.
- **Dependencies**: None new.
