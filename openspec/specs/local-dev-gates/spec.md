# local-dev-gates Specification

## Purpose
TBD - created by archiving change fix-taskfile-ci-parity. Update Purpose after archive.
## Requirements
### Requirement: Go build tags are identical between local commands and CI

Every local invocation of Go build, vet, and test SHALL use `-tags sqlite_fts5`, matching CI workflows (`ci.yaml` vet/test/build steps). The `package.json` `test` script and Taskfile `test` / `build` targets MUST use this tag.

#### Scenario: Local test tag parity

- **WHEN** `Taskfile.yml` `test` target and `package.json` `test` script are inspected
- **THEN** each runs `go test -tags sqlite_fts5 ./...` (or with coverage flags that include `-tags sqlite_fts5`)

#### Scenario: Local build tag parity

- **WHEN** `Taskfile.yml` `build` target is inspected
- **THEN** it runs `go build -tags sqlite_fts5` with ldflags (not bare `go build`)

#### Scenario: Vet tag parity

- **WHEN** any local vet invocation (Taskfile, prek hook, or script) is inspected
- **THEN** it uses `-tags sqlite_fts5` matching `ci.yaml`'s `go vet -tags sqlite_fts5 ./...`

### Requirement: Local build injects version ldflags matching Dockerfile

The Taskfile `build` target SHALL pass `-ldflags "-X main.Version=..."` derived from `git describe --tags --always` (fallback `0.0.0-dev`), so local binaries carry the same version wiring as `Dockerfile`'s `go build -ldflags "-X main.Version=${VERSION}"`.

#### Scenario: Build ldflags present

- **WHEN** `Taskfile.yml` `build` task is read
- **THEN** the `go build` command includes `-ldflags "-X main.Version=`

### Requirement: E2E task rebuilds the server binary before running Playwright

`task test:e2e` SHALL rebuild `./villum-server` with `-tags sqlite_fts5` immediately before `npx playwright test`, and SHALL clean `villum.db*` beforehand. Running e2e MUST NOT use a stale binary.

#### Scenario: E2E rebuilds binary

- **WHEN** `Taskfile.yml` `test:e2e` is inspected
- **THEN** it declares a dependency on `build` (or runs `go build -tags sqlite_fts5 -o villum-server .` before `npx playwright test`) and removes `villum.db*`

#### Scenario: Frontend change is reflected in e2e run

- **WHEN** a developer modifies `ts/` then runs `task test:e2e`
- **THEN** the server binary is rebuilt and Playwright exercises the current frontend assets

### Requirement: Taskfile provides a `ci` target composing the full gate suite in CI order

A `ci` task SHALL exist and SHALL run the same gates as CI in order: tidy/format/vet checks, `build:vite`, `typecheck`, `go test` with coverage, `vitest` coverage, `build`, and `test:e2e`. `AGENTS.md` documents `task ci` as the canonical full suite command.

#### Scenario: `task ci` exists

- **WHEN** `task --list` is run
- **THEN** a `ci` task is listed

#### Scenario: `task ci` gate order matches CI

- **WHEN** the `ci` task definition is read
- **THEN** it invokes vet, build:vite, typecheck, go test, vitest, build, and test:e2e in an order matching `ci.yaml` jobs (lint-typecheck → unit-test → e2e)

### Requirement: Duplicate npm script is removed

Only one of `watch:ts` / `dev:ts` SHALL remain in `package.json` (both were `vite build --watch`). The survivor SHALL be `watch:ts` or `dev:ts` consistently, and any docs referencing the removed name SHALL be updated.

#### Scenario: No duplicate scripts

- **WHEN** `package.json` `scripts` are listed
- **THEN** at most one of `watch:ts` / `dev:ts` exists

### Requirement: Taskfile mirrors documented npm gates where referenced

If `AGENTS.md` references `npm run test:unit`, the Taskfile SHALL provide a `test:unit` mirror (or AGENTS.md SHALL be updated to the canonical name) so every documented command has a runnable target.

#### Scenario: Documented command is runnable

- **WHEN** a developer runs each command listed under `AGENTS.md` "Local Gates" on-demand pieces
- **THEN** each command succeeds or has a Taskfile equivalent
