# go-coverage-tracking Specification

## Purpose
Go test coverage is generated in CI, compared against configured thresholds, and published as an artifact for inspection.

## Requirements

### Requirement: CI runs Go tests with a coverage profile

The CI unit-test job SHALL run the Go test suite with coverage instrumentation (`go test -coverprofile=coverage.out -tags sqlite_fts5 ./...` via `scripts/ci/test-go.sh`) and SHALL upload the resulting `coverage.out` as a `go-coverage` artifact for inspection.

#### Scenario: Coverage profile is produced and uploaded

- **WHEN** the CI unit-test job runs the Go test suite
- **THEN** the tests execute with `-coverprofile=coverage.out` so a coverage profile is generated
- **AND** the profile is uploaded as the `go-coverage` artifact

#### Scenario: No external coverage service is required

- **WHEN** coverage results are produced
- **THEN** the total is derived locally with `go tool cover -func` tooling and no external service (e.g. Codecov) is needed

### Requirement: Coverage below the configured threshold fails CI

The CI job SHALL parse total Go coverage and fail the job when it falls below the configured threshold. The coverage check SHALL NOT mask test results: test failures fail the job directly, and the coverage gate applies to the overall job result.

#### Scenario: Coverage below threshold fails the job

- **WHEN** total Go coverage is below the configured threshold (initially 30%; later revisions enforce per-package floors with the total staying at or above 20%)
- **THEN** the CI job fails

#### Scenario: Test failures are not masked by the coverage gate

- **WHEN** Go tests fail outright
- **THEN** the job fails on the test results and the coverage gate does not replace or hide the test failure
