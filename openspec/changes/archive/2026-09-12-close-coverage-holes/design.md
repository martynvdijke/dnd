# Design: close-coverage-holes

## Context

`go test -coverprofile` (coverpkg `handlers+middleware`) shows 576 functions at 0.0%. Untested layers include: HTMX handlers (`handlers/campaign_htmx.go` 15 funcs at 0% — `HtmxCampaignEncountersSection:55`, `HtmxCreateEncounter:167`, `HtmxDeleteEncounter:188`, etc.; most `Htmx*` in `compendium_htmx.go`, `htmx.go`), AI (`handlers/ai.go:345 HandleTextGeneration`, `:478 HandleImageGeneration` 10%, `:598 sanitizeError`, `:608 truncateResponse`, `:624 SaveGeneratedImage`, `:713 GetAIEnabled`), `handlers/campaign.go` gaps (`ListLocations:52`, `UpdateLocation:108`, `SearchLocations:312`, `UnlinkNPC:470`, `ListQuests:632`), `handlers/campaign_npcs.go` all 5 funcs (`:14,:39,:64,:82,:88`), `handlers/characters.go:207 ListAllCharacters`, `:558 UpdateCharacterDMNotes`, `:905 canViewCharacter` (authorization), `:1285 ImportCharacterJSON`, `handlers/admin.go:192`, `admin_otel.go:12,21`, `cleanup.go:10`, plus `middleware/logging.go:337-446` (entire app-logger: `NewAppLogger`, `Handler`, `Buffer`, `Logger`, `Debug/Info/Warn/Error`, `InitAppLogger*`, `RequestLogger`) and `middleware/store.go:68-160` (session store `Cleanup`, `NewDBSessionStore`, `Create/Get/Delete/Cleanup`). CI enforces floors via `awk` on split profiles (`ci.yaml:116-138`): total ≥20%, `handlers/` ≥30%, `middleware/` ≥25%; vitest ≥20% all metrics. Recent measures ~35.8% handlers / 30.6% middleware — just above floors. `handlers/testutil/testutil.go` exists for per-package handler tests. Frontend vitest worst-covered: `dice.ts` 7.7%, `bottom-sheet.ts` 8.9%, `compendium.ts` 14.3%, `search.ts` 15.6%, `characters/combat.ts` 7.5%, `characters/sheet.ts` 12.1%.

## Goals / Non-Goals

**Goals:**
- Close 0% gaps in priority order: P1 security-critical (session store, `canViewCharacter`, apitoken), P2 HTMX layer, P3 pure helpers and admin — enough to raise `handlers` and `middleware` floors to ≥40%.
- Use existing `handlers/testutil` patterns (fake ent client, `httptest`, Gin test context) for handler tests; table-driven where possible.
- Add vitest unit tests for pure logic in worst-covered TS files (no DOM-heavy E2E duplication).
- Raise CI coverage floors to `handlers ≥40%`, `middleware ≥40%` once achieved; keep `task ci` green.

**Non-Goals:**
- Reaching 80%+ coverage in one change — scope is critical paths + easy wins, not exhaustive.
- Changing handler behavior or APIs — tests only; any testability seam (e.g., exporting `sanitizeError`) is minimal and behavior-preserving.
- E2E coverage for HTMX (covered by `improve-e2e-reliability`); this change is unit/integration via `go test` and `vitest`.
- Reformatting or splitting god files (covered by `split-go-god-files` / `split-main-test`).

## Decisions

### D1: Priority order P1 → P2 → P3; P1 gates the floor raise

P1 (security-critical) first because `middleware/store.go` session lifecycle and `canViewCharacter` authorization are correctness-critical and currently 0%. P2 (HTMX) is the largest 0% block (entire layer) but lower risk than auth. P3 (pure helpers like `sanitizeError`/`truncateResponse` and admin settings) are easy wins that lift percentages quickly. Floor raise (handlers/middleware ≥40%) is proposed only after P1+P2 land and `go test -coverprofile` confirms.

Alternatives: alphabetical or file-size order — rejected: would delay security coverage.

### D2: Handler tests use `handlers/testutil` + table-driven style

`handlers/testutil/testutil.go` already provides ent test DB and Gin helpers. Extend it only if needed (e.g., session store mock). HTMX handlers return HTML fragments — assert status, `Content-Type`, and fragment markers (or golden-file snippets) rather than full DOM. Campaign/NPC handlers tested via `httptest` with role-based contexts to exercise authz paths.

Alternatives: new test framework or `httptest` from scratch per file — rejected: inconsistent with 25+ existing per-package tests.

### D3: Pure helpers tested as unit tests; AI handlers with mocked dependencies

`sanitizeError` and `truncateResponse` are pure functions — direct unit tests with table cases. `HandleTextGeneration`/`HandleImageGeneration`/`SaveGeneratedImage`/`GetAIEnabled` require mocking AI provider (inject interface or stub `http.Client`); keep mock minimal and behavior-preserving.

Alternatives: integration test hitting real AI — rejected: flaky, cost, not needed for coverage.

### D4: Middleware `store.go` and `logging.go` — real DB + buffer logger

Session store tests use a real (temp) SQLite DB via `testutil` to exercise `Create/Get/Delete/Cleanup` and expiry. `logging.go` app-logger block (lines 337-446) tested via buffer logger and `RequestLogger` with `httptest`; assert log output and level filtering.

Alternatives: mock DB for store — rejected: store logic is DB-coupled; real temp DB is already used elsewhere.

### D5: Frontend vitest — pure logic only, not component rendering

For `dice.ts`, `bottom-sheet.ts`, `compendium.ts`, `search.ts`, `characters/combat.ts`, `characters/sheet.ts`, extract/test pure functions (parsers, formatters, state reducers) rather than mounting components. This avoids heavy DOM setup and aligns with vitest's fast unit path.

Alternatives: Playwright component tests — rejected: covered by E2E; vitest should stay pure.

### D6: Floor raise is a separate commit after coverage achieved

Bump `ci.yaml`/`Taskfile` thresholds from `handlers ≥30%`, `middleware ≥25%` to `≥40%` each only after `go test -coverprofile` + `awk` confirms new percentages. Keep total ≥20%, vitest ≥20%.

## Risks / Trade-offs

- [Coverage miscount due to `coverpkg` split profiles] → Mitigation: follow `ci.yaml:116-138` `awk` logic exactly; run `task test` coverage locally and compare.
- [HTMX tests brittle on HTML] → Mitigation: assert status + fragment markers, not exact HTML; golden files only for stable fragments.
- [AI mock drift] → Mitigation: mock at `http.Client` boundary; keep table cases focused on error/sanitize/truncate paths.
- [Session store tests flaky on timing] → Mitigation: use deterministic expiry (inject clock or set `created_at` explicitly); no `time.Sleep`.
- [Floor raise breaks CI for other branches] → Mitigation: land floor bump as last commit; announce; keep main green.
- [Frontend tests miss DOM logic] → Mitigation: scope is pure logic; DOM covered by E2E (`improve-e2e-reliability`).

## Migration Plan

1. **Phase 1 — P1 security-critical** — `middleware/store.go` (session lifecycle), `handlers/characters.go:905 canViewCharacter`, `middleware/apitoken.go`; table-driven authz cases; `task test` green; measure `middleware` %.
2. **Phase 2 — P2 HTMX + campaign_npcs** — `handlers/campaign_htmx.go` 15 funcs, `compendium_htmx.go`/`htmx.go`, `handlers/campaign.go` gaps (`ListLocations`, `UpdateLocation`, `SearchLocations`, `UnlinkNPC`, `ListQuests`), `campaign_npcs.go` 5 funcs; `task test` green; measure `handlers` %.
3. **Phase 3 — P3 helpers + admin** — `sanitizeError`, `truncateResponse`, `HandleTextGeneration`, `HandleImageGeneration`, `SaveGeneratedImage`, `GetAIEnabled`, `HandleAdminResyncSearchIndex`, `Get/SaveOTelSettings`, `StartDBCleanupTask`; `task test` green.
4. **Phase 4 — Frontend vitest** — unit tests for `dice.ts`, `bottom-sheet.ts`, `compendium.ts`, `search.ts`, `characters/combat.ts`, `characters/sheet.ts` pure logic; `npm run test:unit` green; vitest ≥20% all metrics.
5. **Phase 5 — Floor raise** — bump `handlers ≥40%`, `middleware ≥40%` in `ci.yaml` and `Taskfile`; final `task ci` green.

Rollback: each phase is additive tests only; revert commit. Floor bump reverted separately if needed.

## Open Questions

- Exact `handlers/testutil` helpers needed for session store — does it already mock DB or need a temp SQLite helper?
- Which pure functions in `dice.ts`/`compendium.ts` are exportable without refactor — audit exports before writing tests.
- Confirm `canViewCharacter` test matrix (owner, DM, party member, stranger, anonymous) with auth owners.
