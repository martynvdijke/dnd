# Tasks: close-coverage-holes

## 1. P1 — Security-critical coverage (session, authz, apitoken)

- [x] 1.1 Add `middleware/store_test.go` — table-driven tests for `NewDBSessionStore`, `Create`, `Get`, `Delete`, `Cleanup` (Create/Get round-trip, Delete removes, Cleanup removes expired, missing session error); use temp SQLite via `handlers/testutil` or `middleware` test helper; deterministic expiry (no `time.Sleep`)
- [x] 1.2 Add `handlers/characters_authz_test.go` (or extend existing) — `canViewCharacter:905` matrix: owner allowed, DM allowed, party member allowed, stranger denied, anonymous denied; table-driven with fake ent context
- [x] 1.3 Add `middleware/apitoken_test.go` — valid token, invalid token, missing token, expired/revoked paths; verify `go test -coverprofile` shows `middleware/store.go` and `middleware/apitoken.go` covered
- [x] 1.4 Verify `task test` green; record `middleware/` coverage via `go test -coverprofile` + `awk` gate (target ≥40% after all phases)

## 2. P2 — HTMX layer and campaign gaps

- [x] 2.1 Add `handlers/campaign_htmx_test.go` — cover all 15 funcs at 0% in `handlers/campaign_htmx.go` (`HtmxCampaignEncountersSection:55`, `HtmxCreateEncounter:167`, `HtmxDeleteEncounter:188`, etc.); `httptest` + `handlers/testutil`; assert status, `Content-Type`, fragment markers; table-driven success/error
- [x] 2.2 Add `handlers/compendium_htmx_test.go` and `handlers/htmx_test.go` — cover `Htmx*` in `compendium_htmx.go` and `htmx.go`; same pattern; verify `go test -coverprofile` shows HTMX layer above 0%
- [x] 2.3 Add `handlers/campaign_npcs_test.go` — cover all 5 funcs in `handlers/campaign_npcs.go` (`:14,:39,:64,:82,:88`); `httptest` with campaign membership contexts
- [x] 2.4 Add `handlers/campaign_extra_test.go` (or extend `campaign_test.go`) — `ListLocations:52`, `UpdateLocation:108`, `SearchLocations:312`, `UnlinkNPC:470`, `ListQuests:632` plus `handlers/characters.go:207 ListAllCharacters`, `:558 UpdateCharacterDMNotes`, `:1285 ImportCharacterJSON`; verify `handlers/` coverage uplift

## 3. P3 — AI helpers, admin, and infrastructure

- [x] 3.1 Add `handlers/ai_helpers_test.go` — pure unit table-driven tests for `sanitizeError:598` and `truncateResponse:608` (export if needed via test seam; keep behavior-preserving)
- [x] 3.2 Add `handlers/ai_test.go` extensions — `HandleTextGeneration:345`, `HandleImageGeneration:478`, `SaveGeneratedImage:624`, `GetAIEnabled:713` with mocked AI provider (`http.Client` stub or injected interface); cover error/sanitize/truncate branches
- [x] 3.3 Add `handlers/admin_extra_test.go` — `HandleAdminResyncSearchIndex:192`, `GetOTelSettings:12`/`SaveOTelSettings:21`, `StartDBCleanupTask:10`; `httptest` with admin context
- [x] 3.4 Add `middleware/logging_test.go` extensions — cover `middleware/logging.go:337-446` app-logger block (`NewAppLogger`, `Handler`, `Buffer`, `Logger`, `Debug/Info/Warn/Error`, `InitAppLogger*`, `RequestLogger`) via buffer logger + `httptest`; assert log output and level filtering
- [x] 3.5 Verify `task test` green; capture `handlers/` and `middleware/` coverage percentages

## 4. Frontend — vitest pure-logic coverage

- [x] 4.1 Add `ts/dice.test.ts` — unit tests for pure logic in `ts/dice.ts` (parsers, roll helpers); no DOM mounting; `npm run test:unit` green
- [x] 4.2 Add `ts/bottom-sheet.test.ts` and `ts/search.test.ts` — pure state/logic in `ts/bottom-sheet.ts` (8.9%) and `ts/search.ts` (15.6%)
- [x] 4.3 Add `ts/compendium.test.ts` — pure helpers in `ts/compendium.ts` (14.3%)
- [x] 4.4 Add `ts/characters/combat.test.ts` and `ts/characters/sheet.test.ts` — pure reducers/formatters in `characters/combat.ts` (7.5%) and `characters/sheet.ts` (12.1%)
- [x] 4.5 Verify `npm run test:unit` and `npm run typecheck` pass; vitest coverage ≥20% all metrics (`npm run test:unit -- --coverage`)

## 5. Raise coverage floors and final gates

- [x] 5.1 Bump CI floors in `.github/workflows/ci.yaml:116-138` and `Taskfile` from `handlers ≥30%`/`middleware ≥25%` to `handlers ≥40%`/`middleware ≥40%` (keep total ≥20%, vitest ≥20%); commit separately after coverage confirmed
- [x] 5.2 Run `task test` (Go) + `npm run test:unit` — must pass with new floors
- [x] 5.3 Run `task test:e2e` (rebuilds `villum-server`) — must pass (no frontend behavior change, sanity)
- [x] 5.4 Full `task ci` must pass; document new floors in `CONTRIBUTING.md` if needed; verify every new `data-testid` (if any) is referenced in `tests/` per AGENTS.md lint
