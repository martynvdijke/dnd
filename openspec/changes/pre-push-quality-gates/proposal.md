# Proposal: pre-push-quality-gates

## Why

New features repeatedly fail at release time: the first time much of the code meets the real CI gates is after a push to `main`, when the release workflow runs the full check suite and a Gotify "Release Failed" notification is the feedback signal. Local tooling and CI enforce different things — coverage gates, `gofmt`/`go mod tidy` checks, `build:vite`, the `data-testid` lint, and the Docker build smoke test run nowhere locally; build tags drift between `fts5`, `sqlite_fts5`, and none across package.json, pre-commit, Taskfile, and CI; and `task test:e2e` runs Playwright against a potentially stale `./villum-server` binary. Failure feedback arrives with maximal latency, after the authoring context (human or AI) is gone.

## What Changes

- Add shared check scripts under `scripts/ci/` that encode every CI gate exactly once (tidy check, gofmt check, vet, build:vite + typecheck, data-testid lint, Go tests with coverage thresholds, vitest with coverage thresholds, server build, Playwright e2e, Docker build smoke test).
- Refactor `.github/workflows/ci.yaml` to call the shared scripts, so local and CI gates are the same commands by construction and cannot drift.
- Add a Taskfile `ci` task that runs the full local parity suite via the shared scripts.
- Extend `.pre-commit-config.yaml` with a **pre-push stage** hook (run via prek) that executes the full parity suite before any push. Pre-push is the gate; the existing fast pre-commit hooks stay as-is.
- Unify Go build tags on `sqlite_fts5` across package.json, Taskfile, and pre-commit config (or drop them entirely if verification shows modernc.org/sqlite ignores them).
- Fix the stale-binary trap: e2e always builds `./villum-server` (with the same tags as CI) before Playwright runs.
- Update CONTRIBUTING.md to point at the single parity command instead of a manual command list.

## Capabilities

### New Capabilities

- `local-quality-gates`: Local quality gates that mirror CI exactly — shared check scripts as the single source of truth, a pre-push hook (prek) running the full parity suite, consistent build tags, and a fresh server binary for e2e runs.

### Modified Capabilities

<!-- None — no existing spec requirements change. -->

## Impact

- **New files**: `scripts/ci/*.sh` check scripts.
- **Modified files**: `.github/workflows/ci.yaml`, `.pre-commit-config.yaml`, `Taskfile.yml`, `package.json` (test script tags), `CONTRIBUTING.md`.
- **Dependencies**: prek must be installed by developers/agents (single binary, drop-in pre-commit replacement); `prek install --hook-type pre-push` required once per clone.
- **Workflow**: every push runs the full suite locally (~30 min including e2e — accepted); `git push --no-verify` remains the documented escape hatch. No application code changes.
