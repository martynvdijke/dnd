## 1. Documentation reconciliation

- [x] 1.1 Update `AGENTS.md:37` coverage floor line from `handlers/ >=30%, middleware/ >=25%; vitest >=20%` to `handlers/ >=40%, middleware/ >=40% (total >=20%); vitest >=35% statements/lines/functions, >=30% branches`
- [x] 1.2 Verify `openspec/specs/local-quality-gates/spec.md` stale values are captured by the delta in `openspec/changes/reconcile-coverage-gates/specs/local-quality-gates/spec.md` (no direct edit to archived spec needed before archive)
- [x] 1.3 Ensure `openspec/specs/critical-path-test-coverage/spec.md` existing "Raised coverage floors" requirement is preserved and the new consistency/ratcheting deltas are in `openspec/changes/reconcile-coverage-gates/specs/critical-path-test-coverage/spec.md`

## 2. Gate enforcement ratchet

- [x] 2.1 Ratchet `vitest.config.ts` coverage thresholds from 20/20/20/20 to `statements: 35, branches: 30, functions: 35, lines: 35`
- [x] 2.2 Confirm `scripts/ci/test-go.sh` already enforces `handlers >=40%, middleware >=40%, total >=20%` (no change); confirm `.github/workflows/ci.yaml` enforcement unchanged (only the threshold comment updated to match)
- [x] 2.3 Verify `scripts/ci/test-vitest.sh` (`npx vitest run --coverage`) will fail when vitest reports below the new thresholds (vitest itself enforces via config)

## 3. Validation

- [x] 3.1 Run `npm run test:unit` locally and confirm vitest coverage passes the new 35/30/35/35 floors (measured overall ~43.55% lines, so headroom of ~8 points)
- [x] 3.2 Run `go test -tags sqlite_fts5 -v -coverprofile=coverage.out -covermode=atomic -coverpkg=./handlers/...,./middleware/... ./...` / `scripts/ci/test-go.sh` and confirm handlers/middleware remain >=40%
- [x] 3.3 Run `task ci` (full parity suite) to confirm all gates pass; fix `vitest.config.ts` or docs if any gate fails unexpectedly
- [x] 3.4 Run `openspec validate reconcile-coverage-gates --strict` and fix until it passes
