# local-quality-gates Specification

## Purpose
TBD - created by archiving change pre-push-quality-gates. Update Purpose after archive.
## Requirements
### Requirement: Shared check scripts are the single source of truth for quality gates
Every quality gate enforced in CI SHALL be implemented exactly once as a script under `scripts/ci/`, and CI workflow steps SHALL invoke these scripts instead of inline commands. The same scripts SHALL be runnable locally, individually and as a full suite.

#### Scenario: CI and local run the identical gate command
- **WHEN** a developer runs a script from `scripts/ci/` locally and CI runs the corresponding step
- **THEN** both execute the same command with the same arguments and the same pass/fail criteria

#### Scenario: New gate added
- **WHEN** a new quality gate is introduced
- **THEN** it is added as a script under `scripts/ci/` and referenced from CI, with no inline duplication of its command

### Requirement: Full CI parity suite runs locally before push
A pre-push git hook, managed by prek and defined in `.pre-commit-config.yaml`, SHALL run the complete parity suite: go mod tidy check, gofmt check, go vet, frontend build and typecheck, data-testid lint, Go tests with coverage thresholds, frontend unit tests with coverage thresholds, server build, Playwright e2e (chromium), and Docker build smoke test. A failing gate SHALL block the push.

#### Scenario: Push with all gates passing
- **WHEN** a developer runs `git push` with the hooks installed and all gates pass
- **THEN** the push proceeds normally

#### Scenario: Push with a failing gate
- **WHEN** a developer runs `git push` and any parity gate fails
- **THEN** the push is blocked and the failing gate's output is shown

#### Scenario: Hooks not yet installed
- **WHEN** a developer clones the repository fresh
- **THEN** documentation instructs running `prek install --hook-type pre-push`, and the full suite remains runnable on demand via a single `task ci` command

#### Scenario: Emergency bypass
- **WHEN** a developer must push without running the suite
- **THEN** `git push --no-verify` bypasses the hook and this escape hatch is documented

### Requirement: Coverage thresholds are enforced locally identically to CI
The Go coverage gates (total ≥20%, handlers ≥30%, middleware ≥25%, measured with `-coverpkg=./handlers/...,./middleware/... -covermode=atomic`) and the vitest coverage gates (≥20% statements, branches, functions, lines) SHALL be evaluated by the shared scripts on every local parity run.

#### Scenario: Coverage below threshold locally
- **WHEN** a local parity run measures coverage below any configured threshold
- **THEN** the corresponding script exits non-zero and reports which threshold failed

### Requirement: E2E tests always run against a freshly built server binary
The local e2e gate SHALL build `./villum-server` (with the same Go build tags CI uses) immediately before invoking Playwright, and the Taskfile `test:e2e` task SHALL declare the build as a dependency.

#### Scenario: Frontend changed since last binary build
- **WHEN** a developer runs the e2e gate after modifying files under `ts/` or `static/`
- **THEN** the server binary is rebuilt first and Playwright tests exercise the current code

### Requirement: Go build tags are consistent across all invocations
Every local and CI invocation of Go build/vet/test SHALL use the same build tags — either `sqlite_fts5` everywhere or no tags everywhere — as determined by verifying whether modernc.org/sqlite honors the tag.

#### Scenario: Tag usage audit
- **WHEN** package.json scripts, Taskfile tasks, pre-commit hooks, and CI workflows are inspected
- **THEN** all Go invocations use the identical tag set

### Requirement: Docker build smoke test adapts to environment availability
The parity suite SHALL run the Docker build smoke test when a Docker daemon is available, and SHALL skip it with a clearly printed warning when Docker is unavailable locally. CI SHALL always run it.

#### Scenario: Local run without Docker
- **WHEN** the parity suite runs on a machine without a Docker daemon
- **THEN** the Docker smoke test is skipped with a warning and all other gates still run and enforce
