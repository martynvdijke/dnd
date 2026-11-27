# sqlite-driver-migration Specification

## Purpose
Migrate from mattn/go-sqlite3 (CGo) to modernc.org/sqlite (pure Go).

## Requirements

### Requirement: Migrate from mattn/go-sqlite3 to modernc.org/sqlite
The system SHALL replace the `github.com/mattn/go-sqlite3` driver with `modernc.org/sqlite` in `go.mod`.

#### Scenario: go.mod updated
- **WHEN** the dependency change is applied
- **THEN** `go.mod` SHALL list `modernc.org/sqlite` instead of `github.com/mattn/go-sqlite3`
- **AND** `go.sum` SHALL be updated

### Requirement: Update import path in db.go
The system SHALL update the blank import in `db/db.go` from `_ "github.com/mattn/go-sqlite3"` to `_ "modernc.org/sqlite"`.

#### Scenario: Import path updated
- **WHEN** the source code is updated
- **THEN** `db/db.go` SHALL import `_ "modernc.org/sqlite"`

### Requirement: Update DSN format for modernc driver
The system SHALL update the DSN (Data Source Name) format to match modernc's expected format. The modernc driver uses `file:path.db` URI format with query parameters, still compatible with `database/sql` standard.

#### Scenario: DSN format updated
- **WHEN** `db.Init()` is called
- **THEN** the connection string SHALL use format compatible with modernc.org/sqlite while preserving `_journal_mode=WAL` and `_busy_timeout=5000`

### Requirement: Verify CGO_ENABLED=0 build
The system SHALL verify that the application builds and runs with `CGO_ENABLED=0`.

#### Scenario: CGO_ENABLED=0 build succeeds
- **WHEN** `CGO_ENABLED=0 go build ./...` is run
- **THEN** the build SHALL succeed without errors

#### Scenario: CGO_ENABLED=0 binary runs
- **WHEN** a binary built with `CGO_ENABLED=0` starts
- **THEN** the database SHALL initialize successfully and pass all health checks

### Requirement: All existing tests pass with new driver
The system SHALL run the full test suite after migration and all tests SHALL pass.

#### Scenario: Tests pass with modernc driver
- **WHEN** `go test ./...` is run after migration
- **THEN** all tests SHALL pass with no regressions
