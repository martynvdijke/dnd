# ci-architecture Specification

## Purpose
CI pipeline job structure, cost controls, and failure isolation.

## Requirements

### Requirement: CI runs three parallel job tracks
The CI pipeline SHALL split into three jobs that can run in parallel: `lint-typecheck`, `unit-test`, and `e2e`. The `e2e` job MUST wait for `lint-typecheck` and `unit-test` to pass before starting.

#### Scenario: All jobs pass
- **WHEN** a push/PR triggers CI
- **THEN** `lint-typecheck` and `unit-test` start in parallel
- **THEN** `e2e` starts only after both preceding jobs succeed
- **THEN** the overall workflow status is success

#### Scenario: Lint fails, e2e is skipped
- **WHEN** `lint-typecheck` fails
- **THEN** `unit-test` continues independently
- **THEN** `e2e` is skipped (not started)
- **THEN** the overall workflow status is failure

### Requirement: Build steps are separated
The `build:ts` npm script SHALL be split into `build:vite` (all Vite builds) and `typecheck` (tsc --noEmit). CI MAY run `build:vite` in the `e2e` job and `typecheck` in the `lint-typecheck` job.

#### Scenario: Vite build succeeds
- **WHEN** `npm run build:vite` executes
- **THEN** all entry points (app, pwa, admin, setup, login) produce IIFE JS files in `static/js/`
- **THEN** the exit code is zero

#### Scenario: TypeScript type check fails
- **WHEN** `npm run typecheck` executes against TypeScript files
- **THEN** any type errors produce a non-zero exit code
- **THEN** the `lint-typecheck` job fails

### Requirement: Docker build is smoke-tested in CI
CI SHALL include a step that runs `docker build .` (no push) to verify the Dockerfile compiles correctly.

#### Scenario: Docker build succeeds
- **WHEN** `docker build .` runs in CI
- **THEN** the Docker image builds without errors
- **THEN** the exit code is zero

#### Scenario: Docker build fails
- **WHEN** `docker build .` runs with a missing build tag or broken Dockerfile
- **THEN** a non-zero exit code causes the CI job to fail

### Requirement: go mod tidy runs before vet
The CI pipeline SHALL run `go mod tidy` before `go vet` to ensure `go.sum` is consistent with `go.mod`.

#### Scenario: Stale dependency removed
- **WHEN** a dependency is removed from `go.mod` but `go.sum` was not updated
- **THEN** `go mod tidy` produces a non-zero exit code
- **THEN** CI fails with clear indication of the issue

### Requirement: Flaky tests are tracked
CI SHALL log any test that fails on the first run but passes on retry to a `flaky-tests.log` artifact stored alongside the Playwright report.

#### Scenario: Flaky test detected
- **WHEN** a Playwright test fails on first attempt and passes on retry
- **THEN** the test name, file, and worker number are appended to `flaky-tests.log`
- **THEN** the log is uploaded as a CI artifact
- **THEN** the overall workflow status is still success (retry passed)
