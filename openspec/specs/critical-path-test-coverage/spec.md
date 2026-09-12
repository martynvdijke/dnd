# critical-path-test-coverage Specification

## Purpose
TBD - created by archiving change close-coverage-holes. Update Purpose after archive.
## Requirements
### Requirement: Session store lifecycle coverage

The system SHALL have unit/integration tests covering `middleware/store.go` session lifecycle (`NewDBSessionStore`, `Create`, `Get`, `Delete`, `Cleanup` at 68-160) including creation, retrieval, deletion, expiry cleanup, and error paths.

#### Scenario: Session store fully exercised

- **WHEN** `go test -coverprofile` with `coverpkg` including `middleware` is run
- **THEN** `middleware/store.go` functions `NewDBSessionStore`, `Create`, `Get`, `Delete`, `Cleanup` are covered and session lifecycle edge cases (expired session cleanup, missing session) are tested

### Requirement: Authorization helper coverage

The system SHALL have tests covering `handlers/characters.go:905 canViewCharacter` authorization logic for all role cases (owner, DM, party member, unauthorized, anonymous) and `middleware/apitoken.go` token validation paths.

#### Scenario: Authorization matrix tested

- **WHEN** `canViewCharacter` is exercised with different user roles and campaign membership
- **THEN** access is correctly allowed or denied for each role and `middleware/apitoken.go` is covered

### Requirement: HTMX handler layer coverage

The system SHALL have tests covering the HTMX handler layer, specifically `handlers/campaign_htmx.go` (including `HtmxCampaignEncountersSection:55`, `HtmxCreateEncounter:167`, `HtmxDeleteEncounter:188` and all 15 funcs at 0%), `handlers/compendium_htmx.go` `Htmx*` handlers, and `handlers/htmx.go` `Htmx*` handlers, using `handlers/testutil` patterns.

#### Scenario: HTMX handlers exercised

- **WHEN** `go test -coverprofile` is run
- **THEN** `handlers/campaign_htmx.go`, `handlers/compendium_htmx.go`, and `handlers/htmx.go` HTMX handlers show coverage above 0% and key success/error branches are tested (status codes, content-type, fragment presence)

### Requirement: Campaign and NPC handler gaps closed

The system SHALL have tests covering `handlers/campaign.go` gaps (`ListLocations:52`, `UpdateLocation:108`, `SearchLocations:312`, `UnlinkNPC:470`, `ListQuests:632`) and all 5 functions in `handlers/campaign_npcs.go` (`:14,:39,:64,:82,:88`).

#### Scenario: Campaign gaps covered

- **WHEN** `go test -coverprofile` is run
- **THEN** the listed `campaign.go` and `campaign_npcs.go` functions are covered and their success/error paths verified

### Requirement: AI helper and handler coverage

The system SHALL have tests covering `handlers/ai.go` pure helpers `sanitizeError:598` and `truncateResponse:608` as unit tests, and handlers `HandleTextGeneration:345`, `HandleImageGeneration:478`, `SaveGeneratedImage:624`, `GetAIEnabled:713` with mocked dependencies.

#### Scenario: AI helpers and handlers tested

- **WHEN** `go test` runs
- **THEN** `sanitizeError` and `truncateResponse` have table-driven unit tests and `HandleTextGeneration`/`HandleImageGeneration`/`SaveGeneratedImage`/`GetAIEnabled` are covered via mocked provider tests

### Requirement: Admin and infrastructure coverage

The system SHALL have tests covering `handlers/admin.go:192 HandleAdminResyncSearchIndex`, `handlers/admin_otel.go:12 GetOTelSettings` and `:21 SaveOTelSettings`, `handlers/cleanup.go:10 StartDBCleanupTask`, and `middleware/logging.go:337-446` app-logger block (`NewAppLogger`, `Handler`, `Buffer`, `Logger`, `Debug/Info/Warn/Error`, `InitAppLogger*`, `RequestLogger`).

#### Scenario: Admin and logger covered

- **WHEN** `go test -coverprofile` is run
- **THEN** the listed admin, cleanup, and `middleware/logging.go` logger functions are covered and their key branches tested

### Requirement: Frontend pure-logic coverage

The system SHALL have vitest unit tests for pure logic in the worst-covered TypeScript modules: `ts/dice.ts` (7.7%), `ts/bottom-sheet.ts` (8.9%), `ts/compendium.ts` (14.3%), `ts/search.ts` (15.6%), `ts/characters/combat.ts` (7.5%), `ts/characters/sheet.ts` (12.1%).

#### Scenario: Frontend pure logic tested

- **WHEN** `npm run test:unit` is run
- **THEN** the listed TS modules have unit tests for their pure functions and overall vitest coverage remains at or above 20% for all metrics (lines, branches, functions, statements)

### Requirement: Raised coverage floors

The system SHALL enforce raised coverage floors of `handlers/ ≥40%` and `middleware/ ≥40%` (up from 30%/25%) in CI (`ci.yaml` and `Taskfile` `awk` gates), while keeping total ≥20% and vitest ≥20% all metrics.

#### Scenario: Floors enforced in CI

- **WHEN** `task ci` or `task test` coverage gate runs
- **THEN** CI fails if `handlers/` coverage is below 40% or `middleware/` coverage is below 40%, and passes when both are at or above 40%
