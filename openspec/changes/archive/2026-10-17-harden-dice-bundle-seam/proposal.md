# Proposal: harden-dice-bundle-seam

## Why

`dice/bundle.js` is a gitignored esbuild artifact embedded at Go compile time via `//go:embed`. The existing `dice-bundle-build` spec guarantees the build pipeline and that a missing bundle fails fast, but nothing detects a present-but-stale bundle: if `dice/bundle.js` exists from an older `@dice-roller/rpg-dice-roller` version, the Go embed silently ships it. The Go<->JS contract (`__diceRoll`/`__diceRoller` names and signatures) is also undocumented and unasserted, so a JS-side rename breaks the Go side only at runtime.

## What Changes

- The `build:dice` esbuild step SHALL embed a verifiable build stamp (content hash and/or the resolved `@dice-roller/rpg-dice-roller` version) into or alongside `dice/bundle.js` that the Go side can read without executing the full bundle.
- `dice/engine.go` (engine init / `newEngine`) SHALL verify the build stamp at startup against an expected value; a mismatch or unreadable stamp SHALL fail fast with a clear error and degrade gracefully (no panic; dice endpoints return 503, consistent with `backend-error-resilience`).
- The Go<->JS contract for `globalThis.__diceRoll` and `globalThis.__diceRoller` (names, call signatures, JSON return shape) SHALL be documented and asserted by a Go test that fails if the bundle stops exposing them.
- The `@dice-roller/rpg-dice-roller` dependency version SHALL be pinned/locked (exact version in `package.json` and `package-lock.json`/`npm ci`) so a bump forces a rebuild, and the build-stamp check SHALL catch the case where it does not.

## Capabilities

### New Capabilities

<!-- No new capability; this change hardens an existing one -->

### Modified Capabilities

- `dice-bundle-build`: Add stale-bundle detection (build stamp + verification at engine init), Go<->JS contract documentation and assertion, and dependency pinning requirements. Existing requirements on `build:dice`, gitignored bundle, CI/Docker/Taskfile wiring, missing-bundle failure, and smoke test remain unchanged.

## Impact

- Affected code: `dice/bundle_stub.go`, `dice/engine.go`, `dice/roller-entry.js`, `package.json` / `package-lock.json`, esbuild invocation (`build:dice` npm script and Taskfile mirror), new contract test (e.g. `dice/contract_test.go` or `dice/engine_test.go` addition).
- No new runtime dependency; `goja` and `@dice-roller/rpg-dice-roller` stay. No API contract change beyond the already-specified 503 degradation on dice-engine unavailability.
- Risk if not done: stale embeds ship silently after dependency upgrades or partial rebuilds, and contract drift is caught only in production.
