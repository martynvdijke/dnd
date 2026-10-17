# Tasks: harden-dice-bundle-seam

## 1. Build stamp — esbuild step

- [x] 1.1 Update `build:dice` (npm script and Taskfile mirror) to inject a verifiable build stamp into `dice/bundle.js` (e.g., `globalThis.__diceBuildStamp = "<rpg-dice-roller-version>:<hash>"` via `--banner`/`--footer` or `roller-entry.js` wrapper), and document the stamp format
- [x] 1.2 Ensure the stamp encodes at least the resolved `@dice-roller/rpg-dice-roller` version and/or content hash and is readable without executing the full roller

## 2. Go verification at engine init

- [x] 2.1 Update `dice/engine.go` (`newEngine` / `NewPool`) to extract the build stamp from `BundleJS` and verify it against the expected value; on mismatch/missing/unreadable return a clear error (`expected vs actual`)
- [x] 2.2 Wire the verification error into the existing graceful-degradation path (no panic; dice endpoints return 503, non-dice endpoints unaffected per `backend-error-resilience`)

## 3. Contract documentation and test

- [x] 3.1 Document the Go<->JS contract (`globalThis.__diceRoll` and `globalThis.__diceRoller` names, call signatures, JSON return shape) alongside the dice package (`dice/roller-entry.js` header, `dice/engine.go` comment, or `dice/README.md`)
- [x] 3.2 Add a Go contract test (e.g., `dice/contract_test.go: TestDiceContract`) that loads `BundleJS` in goja and asserts both globals are present, callable, and that `__diceRoll("2d6")` returns the expected JSON shape; test fails clearly if the contract is broken

## 4. Dependency pinning

- [x] 4.1 Pin `@dice-roller/rpg-dice-roller` to an exact version in `package.json` (remove `^`/`~` range) and update `package-lock.json` (`npm install` / `npm ci`); verify CI uses `npm ci`
- [x] 4.2 Verify the stale-after-bump case is caught: bump pin without rebuilding, then confirm engine init fails due to stamp mismatch rather than silently using the old bundle

## 5. Validation

- [x] 5.1 Run `npm run build:dice` then `go test -tags sqlite_fts5 ./dice -run TestDiceContract` and `go test -tags sqlite_fts5 ./dice -run TestRoll` (or equivalent) — all pass on a fresh bundle
- [x] 5.2 Run `openspec validate harden-dice-bundle-seam` and fix until it passes
