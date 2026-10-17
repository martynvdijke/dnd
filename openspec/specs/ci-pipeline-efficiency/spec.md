# Spec: ci-pipeline-efficiency

## Purpose
Defines the pull-request CI pipeline shape - a shared composite setup action, module and npm caching, reusable OpenTelemetry export, least-privilege permissions and cancellation of superseded runs - so feedback stays fast and cheap.

## Requirements

### Requirement: Shared composite setup action replaces duplicated steps

A composite action (e.g. `.github/actions/setup/action.yml`) SHALL encapsulate the repeated sequence: checkout, setup-go with `cache: true`, `go mod download`, setup-node with npm cache, `npm ci`, and `npm run build:vite` (or `build:ts` where noted). Every `ci.yaml` job that previously repeated these steps SHALL use the composite action instead.

#### Scenario: Jobs use composite action

- **WHEN** `.github/workflows/ci.yaml` jobs `lint-typecheck`, `unit-test`, and `e2e` are inspected
- **THEN** each uses `uses: ./.github/actions/setup` (or the chosen path) instead of inline Checkout/setup-go/setup-node/npm ci/build:vite steps

#### Scenario: Composite action exists

- **WHEN** the composite action path is inspected
- **THEN** `action.yml` defines a `composite` action with inputs for `go-version`/`node-version` and runs the setup steps

### Requirement: Go module caching is enabled

The Go setup in the composite action SHALL enable module caching (`cache: true` on `actions/setup-go`), so subsequent jobs and runs reuse downloaded modules keyed by `go.sum`.

#### Scenario: Cache enabled

- **WHEN** the composite action is read
- **THEN** the `actions/setup-go` step includes `with: cache: true` (and `go-version`)

### Requirement: OTel export is a reusable workflow

The OTel export job SHALL be extracted to a reusable workflow (e.g. `.github/workflows/otel.yaml` with `on: workflow_call`) and SHALL be called from `ci.yaml`, `release.yaml`, `renovate.yml`, and `stale-branches.yml` via `uses: ./.github/workflows/otel.yaml`, removing the four copy-pasted `otel:` jobs.

#### Scenario: Reusable OTel workflow exists

- **WHEN** `.github/workflows/otel.yaml` is inspected
- **THEN** it declares `on: workflow_call` and a single job running `corentinmusard/otel-cicd-action@v4`

#### Scenario: Callers use reusable workflow

- **WHEN** `ci.yaml`, `release.yaml`, `renovate.yml`, and `stale-branches.yml` are inspected
- **THEN** none contains an inline `otel:` job with `otel-cicd-action`; each has a job with `uses: ./.github/workflows/otel.yaml`

### Requirement: CI permissions are least-privilege for PRs

`ci.yaml` top-level `permissions` SHALL be `contents: read` (not `write`) for PR runs; any job requiring additional scope SHALL declare it at job level. `release.yaml` retains `contents: write` where semantic-release needs it.

#### Scenario: PR permissions are read-only

- **WHEN** `.github/workflows/ci.yaml` top-level `permissions` is read
- **THEN** `contents: write` is absent and `contents: read` is present

### Requirement: Superseded PR runs are cancelled

`ci.yaml` SHALL declare `concurrency` with `cancel-in-progress: true` (scoped per workflow/ref, not cancelling `main`), so a new push to a PR branch cancels the prior in-flight run.

#### Scenario: Concurrency cancels superseded runs

- **WHEN** `.github/workflows/ci.yaml` is read
- **THEN** it contains `concurrency: { group: ..., cancel-in-progress: true }` (with `cancel-in-progress` conditional on non-main ref if implemented as an expression)

### Requirement: Playwright browser-matrix decision is documented

CI SHALL continue to run `--project=chromium` only. `playwright.config.ts` MAY keep `firefox` and `mobile-chrome` projects for local debugging; the decision not to expand the CI matrix in this change SHALL be noted in the workflow comment or design.

#### Scenario: CI runs chromium only

- **WHEN** the `e2e` job's Playwright invocation is read
- **THEN** it passes `--project=chromium` and does not add additional browser projects
