# Spec: dice-bundle-build

## ADDED Requirements

### Requirement: Bundle carries a verifiable build stamp readable by Go

The built `dice/bundle.js` SHALL carry a verifiable build stamp that encodes at least one of: the content hash of the bundle or the resolved `@dice-roller/rpg-dice-roller` version (or both). The stamp SHALL be exposed so the Go side can read it without relying on successful execution of the full dice-roller library (e.g., a header comment with a known prefix, a `globalThis.__diceBuildStamp` assignment, or a co-located manifest file that is itself embedded).

#### Scenario: Build stamp is present after build

- **WHEN** `npm run build:dice` (or `task build:dice`) is run
- **THEN** `dice/bundle.js` (or its co-located manifest) contains a build stamp with the `@dice-roller/rpg-dice-roller` version and/or a content hash, in a format documented in the Go<->JS contract

#### Scenario: Go can read the stamp without executing the bundle

- **WHEN** the Go code inspects the embedded bundle artifact at init time
- **THEN** it can extract the build stamp without evaluating `__diceRoll`/`__diceRoller`

### Requirement: Engine verifies the build stamp at startup and fails fast on staleness

`dice/engine.go` engine initialization (`newEngine` / `NewPool` path) SHALL verify the embedded build stamp against the expected value derived from the pinned dependency version and/or content hash. On mismatch, missing stamp, or unreadable stamp, initialization SHALL return a clear error that includes the expected vs actual stamp (or the missing-stamp reason). The error SHALL propagate as a dice-engine unavailability consistent with `backend-error-resilience` ("Dice engine failure degrades gracefully"): the process SHALL NOT panic, all non-dice endpoints SHALL continue to serve, and every dice endpoint SHALL return `503 Service Unavailable` with a JSON error when the engine is unavailable.

#### Scenario: Stale bundle fails fast with a clear error

- **WHEN** `dice/bundle.js` was built from a different `@dice-roller/rpg-dice-roller` version or has a content hash that does not match the expected stamp
- **THEN** dice engine initialization returns an error describing the stamp mismatch (expected vs actual) and the server starts without panicking

#### Scenario: Dice endpoints return 503 when stamp verification fails

- **WHEN** the build-stamp verification has failed and a dice endpoint is called
- **THEN** the endpoint returns `503` with a JSON error indicating the dice engine is unavailable, while non-dice endpoints return their normal responses

#### Scenario: Fresh bundle passes verification

- **WHEN** `npm run build:dice` has just been run for the currently pinned `@dice-roller/rpg-dice-roller` version and the Go binary is built/tested
- **THEN** engine initialization succeeds and a sample roll (e.g., `2d6`) executes without error

### Requirement: Go/JS dice contract is documented and asserted by a test

The Go<->JS contract SHALL be documented (names `globalThis.__diceRoll` and `globalThis.__diceRoller`, their call signatures, and the JSON shape returned by `__diceRoll`) and a Go test SHALL assert that the embedded bundle exposes the contract. The test SHALL fail with a clear message if either global is missing, not callable, or has an incompatible signature/return shape. The documentation SHALL live alongside the dice package (e.g., in `dice/roller-entry.js` header, `dice/engine.go` comment, or `dice/README.md`) and be referenced by the spec.

#### Scenario: Contract test catches a missing global

- **WHEN** `dice/bundle.js` is rebuilt without exposing `__diceRoll` (or `__diceRoller`)
- **THEN** the contract test fails, reporting which global is missing or not callable

#### Scenario: Contract test passes on a correct bundle

- **WHEN** the bundle is built via `npm run build:dice` and `go test -tags sqlite_fts5 ./dice -run TestDiceContract` (or equivalent contract test) is run
- **THEN** the test passes, confirming both globals are present, callable, and return the expected JSON shape for a sample expression

### Requirement: Dice-roller dependency is pinned and a bump forces rebuild

The `@dice-roller/rpg-dice-roller` dependency SHALL be pinned to an exact version in `package.json` (no `^`/`~` range) and locked via `package-lock.json`; CI SHALL install with `npm ci`. The build-stamp verification requirement above SHALL catch the case where the dependency version was bumped but `dice/bundle.js` was not rebuilt, by failing fast at engine init.

#### Scenario: Pinned version is exact

- **WHEN** `package.json` dependencies are inspected
- **THEN** `@dice-roller/rpg-dice-roller` is listed with an exact version (e.g., `5.5.1` without `^`/`~` prefix)

#### Scenario: Stale bundle after version bump is detected

- **WHEN** the pinned version in `package.json` is bumped but `npm run build:dice` is not re-run before `go test` / `go build`
- **THEN** engine initialization fails due to the build-stamp mismatch, rather than silently using the old bundle
