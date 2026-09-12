# Proposal: close-coverage-holes

## Why

Coverage analysis shows 576 functions at 0.0% in `handlers+middleware` (coverpkg), with entire critical layers untested: the HTMX handler layer (15 funcs at 0% in `handlers/campaign_htmx.go`, most of `compendium_htmx.go`/`htmx.go`), `handlers/ai.go` (`HandleTextGeneration:345`, `GetAIEnabled:713`, `sanitizeError:598`, `truncateResponse:608`), `handlers/campaign_npcs.go` (all 5 funcs 0%), `canViewCharacter:905` authorization, session store (`middleware/store.go:68-160`), and app-logger (`middleware/logging.go:337-446`). This leaves security-critical paths unverified and blocks raising coverage floors beyond the current low bars (30.6% handlers / ~35.8% middleware).

## What Changes

- **P1 — Security-critical coverage**: add table-driven tests for `middleware/store.go` session lifecycle (`NewDBSessionStore`, `Create/Get/Delete/Cleanup`), `handlers/characters.go:905 canViewCharacter` authorization, and `middleware/apitoken.go`; verify session cleanup and authz edge cases.
- **P2 — HTMX layer**: add tests for `handlers/campaign_htmx.go` (`HtmxCampaignEncountersSection:55`, `HtmxCreateEncounter:167`, `HtmxDeleteEncounter:188`, etc.), `compendium_htmx.go`, and `htmx.go` `Htmx*` handlers using existing `handlers/testutil` patterns (fake ent client, `httptest`, Gin context); cover `handlers/campaign.go` gaps (`ListLocations:52`, `UpdateLocation:108`, `SearchLocations:312`, `UnlinkNPC:470`, `ListQuests:632`) and `handlers/campaign_npcs.go` all 5 funcs.
- **P3 — Easy wins and admin**: add unit tests for pure helpers `sanitizeError:598` / `truncateResponse:608`, `HandleTextGeneration:345` / `HandleImageGeneration:478` / `SaveGeneratedImage:624` / `GetAIEnabled:713`, `handlers/admin.go:192 HandleAdminResyncSearchIndex`, `handlers/admin_otel.go:12,21`, `handlers/cleanup.go:10`, plus remaining gaps.
- **Frontend**: add vitest unit tests for worst-covered TS modules — `ts/dice.ts` (7.7%), `ts/bottom-sheet.ts` (8.9%), `ts/compendium.ts` (14.3%), `ts/search.ts` (15.6%), `ts/characters/combat.ts` (7.5%), `ts/characters/sheet.ts` (12.1%) — focusing on pure logic.
- **Floors**: once achieved, propose new CI floors `handlers/ ≥40%`, `middleware/ ≥40%` (up from 30%/25%) in `ci.yaml:116-138` and `Taskfile` coverage gates; keep total ≥20% and vitest ≥20%.

## Capabilities

### New Capabilities
- `critical-path-test-coverage`: Prioritized closure of 0%-coverage critical paths across handlers, middleware, and frontend — security-critical first, then HTMX, then helpers/admin — with raised coverage floors.

### Modified Capabilities
- None

## Impact

- **Code**: `handlers/*_test.go` (new table-driven tests, using `handlers/testutil/testutil.go`), `middleware/*_test.go` (store, logging, apitoken), `ts/**/*.test.ts` (new vitest suites for dice, bottom-sheet, compendium, search, characters/combat, characters/sheet); possibly `Taskfile`/`ci.yaml` floor thresholds.
- **APIs**: None — tests only; no handler behavior changes (except minor testability seams if needed, e.g., exporting pure helpers).
- **CI**: Coverage gates tighten to `handlers ≥40%`, `middleware ≥40%`; `task test` and `task ci` must stay green. All `data-testid` added in tests must be referenced in `tests/` per AGENTS.md lint.
