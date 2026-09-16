# Design: harden-dice-bundle-seam

## Context

`dice/bundle_stub.go` embeds `dice/bundle.js` as a string (`var BundleJS string`). `dice/engine.go` creates a `goja` runtime, polyfills `crypto`/`require`/`globalThis`, runs `vm.RunString(bundleJS)`, then looks up `__diceRoll` (and `__diceRoller`) as callables. `dice/bundle.js` is produced by `npm run build:dice`:

```
esbuild dice/roller-entry.js --bundle --format=iife --global-name=__diceBundle --platform=browser --external:crypto --outfile=dice/bundle.js
```

from `@dice-roller/rpg-dice-roller ^5.5.1`. The bundle is gitignored.

`openspec/specs/dice-bundle-build/spec.md` already guarantees: the esbuild script, gitignored bundle, CI rebuild before Go, Dockerfile rebuild, Taskfile wiring, missing-bundle failure (`//go:embed` / `__diceRoll not found`), and a smoke test. `openspec/specs/backend-error-resilience/spec.md` already requires that dice-engine init failure does not panic and dice endpoints return 503. This change hardens the residual seam between Go and JS that neither spec covers: a present-but-stale bundle and an undocumented/untested `__diceRoll`/`__diceRoller` contract.

## Goals / Non-Goals

**Goals:**

- Detect stale embeds at startup with a clear error, without rebuilding on every Go invocation.
- Document and test-assert the Go<->JS contract so JS-side drift is caught in CI, not production.
- Pin the dice-roller version so upgrades are intentional and force a rebuild.

**Non-Goals:**

- Porting dice evaluation to pure Go or replacing `goja` — the JS library and emulator stay.
- Rebuilding `dice/bundle.js` on every `go build`/`go test` unconditionally — the stamp check is the cheaper guard.
- Changing any dice API contract beyond the already-specified 503 degradation.

## Decisions

### Decision: Build stamp over unconditional rebuild

- **Choice:** Embed a verifiable stamp (content hash and/or resolved dependency version) in or alongside the bundle; Go verifies it at `newEngine` time.
- **Why:** Unconditional `build:dice` on every Go build is correct but slower and couples Go tooling to Node/esbuild availability. A stamp is a single string compare at init and reuses the existing CI/Docker/Taskfile rebuild steps — the cheapest check that closes the seam.
- **Alternatives considered:** (1) Always rebuild before Go — rejected as heavier and redundant with existing pipeline wiring. (2) Embed `package.json` version directly without a stamp — rejected because it does not detect hand-edited or partially rebuilt bundles where the hash changed but the version did not.
- **Implementation notes:** Simplest is `globalThis.__diceBuildStamp = "<version>:<hash>"` injected by the esbuild step (e.g., via `--banner`/`--footer` or a small wrapper around `roller-entry.js`), or a `dice/bundle.meta.json` co-embedded file. Either satisfies "Go can read it without executing the roller." Prefer the single-file header/global approach to avoid a second embed.

### Decision: Verify at engine init, degrade per backend-error-resilience

- **Choice:** Validation lives in `newEngine` (or `NewPool`) before `vm.RunString` or immediately after extracting the stamp; on mismatch return `fmt.Errorf("dice bundle stale: expected %s, got %s", ...)`.
- **Why:** This is the sole seam where all callers route through. One guard there covers every dice path. Surfacing as engine unavailability reuses the existing graceful-degradation contract (no panic, 503 on dice endpoints).
- **Alternatives considered:** Build-time `go:generate` check — rejected because the embed is a compile-time string; runtime verification is the only place that sees both the expected and actual stamp in a single binary.

### Decision: Contract test in Go against goja

- **Choice:** A Go test (e.g., `dice/contract_test.go: TestDiceContract`) that loads `BundleJS` in a fresh `goja` runtime and asserts `__diceRoll` and `__diceRoller` are present and callable, and that `__diceRoll("2d6")` returns JSON with `notation`/`total`/`output`.
- **Why:** The existing smoke test proves a roll works end-to-end; it does not assert the names/signatures that `engine.go` depends on. A dedicated contract test fails with a precise message when the JS side renames or reshapes the export.
- **Alternatives considered:** TypeScript-only test — rejected because the Go side is the consumer that needs the guarantee.

### Decision: Pin dependency, keep goja and rpg-dice-roller

- **Choice:** Change `^5.5.1` to pinned `5.5.1` (or current exact) in `package.json` and commit the updated `package-lock.json`; CI uses `npm ci`.
- **Why:** A range allows `npm install` to silently advance the library without a rebuild. Pinning makes upgrades explicit; the stamp check is the backstop if someone forgets to rebuild.
- **Why not replace:** Porting to Go or swapping the emulator is out of scope and far larger than the seam fix.

## Risks / Trade-offs

- [Stale stamp not updated by esbuild step] → The esbuild wrapper must be the sole producer of `bundle.js`; add a `postbuild` check in CI that the stamp exists.
- [Stamp format drifts] → Document the exact prefix/format in the contract doc and parse leniently (version equality is the hard check; hash is advisory).
- [Extra init cost] → One string extraction and compare; negligible vs `goja` startup.
- [Pinned version delays security patches] → Pinning intentionally trades auto-upgrade for determinism; Dependabot/Renovate can still bump the pin via PR, which will rebuild via CI.

## Migration Plan

1. Update `package.json` pin and `package-lock.json` (`npm install --save-exact`).
2. Update `build:dice` to inject the stamp; update `dice/roller-entry.js` / esbuild banner/footer.
3. Update `dice/engine.go` to extract and verify the stamp at init.
4. Add `dice/contract_test.go` and document the contract.
5. CI/Docker/Taskfile already rebuild before Go — no pipeline changes beyond verifying `npm ci` is used.
6. Rollback: revert the pin and stamp check; existing missing-bundle behavior remains.

## Open Questions

- Exact stamp payload: `version` alone vs `version:sha256` — recommend both; version is the primary mismatch signal, hash catches local edits.
- Single global `__diceBuildStamp` vs `bundle.meta.json` — either passes validation; prefer the single-file global to keep `//go:embed` simple.

## Overlap with Existing Specs

- `dice-bundle-build` covers build pipeline and absence detection; this change is additive — it covers presence-with-staleness and contract assertion without modifying existing requirements, hence `## ADDED Requirements` only.
- `backend-error-resilience` already mandates graceful degradation (no panic, 503 on dice endpoints); this change reuses that behavior for the new stale-bundle failure mode rather than defining a new error handling strategy.
