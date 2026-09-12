# Tasks: split-main-test

## 1. Router extraction (dependency for drift fix)

- [x] 1.1 Agree `setupRouter`/`newRouter` signature with `split-go-god-files` owners and document in design (coordinate to avoid duplicate extraction)
- [x] 1.2 Refactor `main.go` to expose `setupRouter(db, cfg) *gin.Engine` (or `newRouter()`) that mounts all route groups; keep `main()` thin (flag parsing + `Run`)
- [x] 1.3 Update `main_test.go:buildRouter` to call the new `setupRouter` wrapper; delete duplicated ~400-line wiring; verify `task test` green and `go vet` clean
- [x] 1.4 Add smoke test `TestSetupRouter` asserting router non-nil and key routes registered (e.g. `router.Routes()` snapshot)

## 2. Shared helper extraction

- [x] 2.1 Create `testutil_test.go` (or `main_testutil_test.go`) at repo root in `package main` with `TestMain`, `testRouter`, `buildRouter` wrapper, `testClient` (`req/get/post/put/del`), `setupAdmin`, `login`, `readJSON` moved from `main_test.go:23-511`
- [x] 2.2 Remove helper definitions from `main_test.go`; verify no duplication (`grep -n "func.*setupAdmin\|type testClient"` shows single definition) and `task test` green
- [x] 2.3 Verify test-count invariant still 124 (`go test -list Test | wc -l` or `go test -run Test -count=1 -list`)

## 3. Domain file splits (one file per commit; keep package main)

- [x] 3.1 Create `main_auth_test.go` — move auth, session, token, authorization (`canViewCharacter`) tests; delete moved funcs from `main_test.go`; verify `task test` and test count
- [x] 3.2 Create `main_characters_test.go` — move characters CRUD, proficiencies, features/spells, death saves, spell filtering, DM notes, import/export tests; verify `task test` and test count
- [x] 3.3 Create `main_campaign_test.go` — move campaigns, locations, NPCs, sessions, quests, journal, party, graph, rest/level-up tests; verify `task test` and test count
- [x] 3.4 Create `main_compendium_test.go` — move compendium CRUD, search, import, admin compendium tests; verify `task test` and test count
- [x] 3.5 Create `main_dice_test.go` — move dice, rolls, stats tests; verify `task test` and test count
- [x] 3.6 Create `main_admin_test.go` — move admin users, seed data, OTel settings, cleanup, resync tests; verify `task test` and test count
- [x] 3.7 Create `main_misc_test.go` — move remaining tests; delete now-empty `main_test.go` (or leave minimal shim); final `task test` green with 124 tests; each file <800L (`wc -l main_*_test.go`)

## 4. Verification and gates

- [x] 4.1 Run `task test` (Go) and `go vet ./...` — must pass
- [x] 4.2 Run `npm run test:unit` and `npm run typecheck` — must pass (no frontend changes, sanity check)
- [x] 4.3 Verify coverage floors still pass: Go total ≥20%, `handlers/` ≥30%, `middleware/` ≥25%, vitest ≥20% all metrics (`task test` coverage profile `awk` checks)
- [x] 4.4 Run `task test:e2e` (rebuilds `villum-server` binary first) — must pass
- [x] 4.5 Full `task ci` must pass; confirm `git log --follow` tracks renames (`git mv` used) and no `data-testid` changes (so `tests/` lint unaffected)
