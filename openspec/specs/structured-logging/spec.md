# structured-logging Specification

## Purpose
Central structured logging through AppLog replacing all raw log.Printf/fmt.Printf calls across backend.

## Requirements

### Requirement: All log output uses structured AppLog calls
The system SHALL route all diagnostic log output through the structured `AppLog` singleton defined in `middleware/logging.go`, using appropriate source tags and structured attributes. Raw `log.Printf`, `log.Println`, and `fmt.Printf` diagnostic calls (not output) SHALL be replaced.

#### Scenario: Handler logs use AppLog with source tag
- **WHEN** a handler in `handlers/` emits a diagnostic log message (e.g., error, warning, info)
- **THEN** it SHALL use `middleware.AppLog.Info("handler", ...)`, `.Error()`, `.Warn()`, or `.Debug()` with a descriptive source tag matching the handler name
- **AND** structured attributes SHALL be passed as key-value pairs rather than formatted into the message string

#### Scenario: DB package logs use AppLog
- **WHEN** the `db/` package logs database warnings, seed status, connection info, or errors
- **THEN** it SHALL use `middleware.AppLog` with source `"db"`, `"seed"`, or `"migration"` as appropriate

#### Scenario: Crypto package logs use AppLog
- **WHEN** the `crypto/` package logs encryption key status or warnings
- **THEN** it SHALL use `middleware.AppLog` with source `"crypto"`

#### Scenario: Middleware logs use AppLog
- **WHEN** middleware such as tracing.go or request logging emits diagnostic messages
- **THEN** it SHALL use `middleware.AppLog` with source `"tracing"` or `"http"` as appropriate

#### Scenario: Early boot logs preserved as log.Printf
- **WHEN** a log call occurs before `InitAppLogger()` has been called (e.g., `main.go` database init, telemetry init)
- **THEN** `log.Printf`/`log.Fatalf` MAY be used since `AppLog` is not yet initialized

#### Scenario: Output formatting preserved as fmt.Print
- **WHEN** the code produces user-facing output (e.g., character print sheet in `characters.go`)
- **THEN** `fmt.Fprintf`/`fmt.Printf` SHALL NOT be replaced — these are output, not diagnostics

### Requirement: LogBuffer size is configurable
The system SHALL allow configuring the in-memory log ring buffer size via environment variable `LOG_BUFFER_SIZE`, with a default of 20000 entries.

#### Scenario: Buffer uses LOG_BUFFER_SIZE env var
- **WHEN** the application starts
- **THEN** `InitAppLoggerFromEnv()` SHALL read `LOG_BUFFER_SIZE` env var
- **AND** if set to a positive integer, use it as the ring buffer capacity
- **AND** if unset or invalid, default to 20000

### Requirement: Log level respects LOG_LEVEL env var at startup
The system SHALL read the `LOG_LEVEL` environment variable at startup, falling back to `"warn"` if unset, and then SHALL be overridden by the persisted DB setting from `log_settings` table if present.

#### Scenario: LOG_LEVEL env var sets initial level
- **WHEN** `InitAppLoggerFromEnv()` is called
- **THEN** it SHALL read `LOG_LEVEL` env var and parse it as a valid level (debug, info, warn, error)
- **AND** if unset or invalid, default to `"warn"`

#### Scenario: DB log_settings overrides env var after init
- **WHEN** `InitLogSettings()` is called after `InitAppLoggerFromEnv()`
- **THEN** the persisted `min_level` from `log_settings` table SHALL override the env var value
