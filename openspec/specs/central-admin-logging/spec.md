# central-admin-logging Specification

## Purpose
Centralized application logging with admin UI viewer, structured log emission, OTel log export, and configurable verbosity.

## Requirements

### Requirement: Admin log viewer tab
The admin panel SHALL include a "Logs" tab that displays application log entries in a table with columns for timestamp, level, source, and message.

#### Scenario: Log tab visible to admin users
- **WHEN** an admin user navigates to the admin panel
- **THEN** the navigation tabs include a "Logs" button
- **AND** clicking it displays the log viewer

#### Scenario: Log entries displayed in table
- **WHEN** the Logs tab is active
- **THEN** log entries are displayed in a table sorted by timestamp descending
- **AND** each row shows the log level badge, timestamp, source name, and message text

### Requirement: Structured log emission
The system SHALL provide a structured logger that emits log entries with level, source, message, timestamp, and optional attributes (key-value pairs).

#### Scenario: Log entry with attributes
- **WHEN** code calls `AppLogger.Warn("email", "send failed", "error", err, "recipient", to)`
- **THEN** a log entry is created with level `warn`, source `email`, message `send failed`, and attributes `error` and `recipient`

#### Scenario: Log levels supported
- **WHEN** code emits a log entry
- **THEN** the level MUST be one of: `debug`, `info`, `warn`, `error`

### Requirement: Log query API
The system SHALL expose `GET /api/admin/logs` that returns log entries with optional filters.

#### Scenario: Query all logs
- **WHEN** an admin requests `GET /api/admin/logs`
- **THEN** the response is a JSON array of recent log entries with id, timestamp, level, source, message, and attributes

#### Scenario: Filter by level
- **WHEN** an admin requests `GET /api/admin/logs?level=error`
- **THEN** only log entries at `error` level or above are returned

#### Scenario: Filter by source
- **WHEN** an admin requests `GET /api/admin/logs?source=email`
- **THEN** only log entries from the `email` source are returned

#### Scenario: Limit results
- **WHEN** an admin requests `GET /api/admin/logs?limit=10`
- **THEN** at most 10 log entries are returned

### Requirement: Configurable log verbosity
The system SHALL allow the minimum log level to be configured and persisted. The default SHALL be `warn`.

#### Scenario: Get current log level
- **WHEN** an admin requests `GET /api/admin/log-level`
- **THEN** the response contains the current minimum log level and its effective value

#### Scenario: Update log level
- **WHEN** an admin requests `PUT /api/admin/log-level` with body `{"level": "debug"}`
- **THEN** the minimum log level is updated to `debug`
- **AND** the new value is persisted across server restarts

#### Scenario: Initial default level
- **WHEN** the server starts for the first time with no log level configured
- **THEN** the effective minimum log level is `warn`

### Requirement: Log level UI control
The Logs tab SHALL include a dropdown or selector to change the minimum log level.

#### Scenario: Change log level from UI
- **WHEN** an admin selects a new log level from the dropdown in the Logs tab
- **THEN** the UI sends `PUT /api/admin/log-level` with the selected level
- **AND** the log view updates to show only entries at or above the new level

### Requirement: OTel log export
Structured log entries SHALL be exported to the configured OTel endpoint when OTel is enabled.

#### Scenario: OTel log export enabled
- **WHEN** `OTEL_EXPORTER_OTLP_ENDPOINT` is set
- **AND** a log entry is emitted
- **THEN** the log entry is also exported as an OTel log record to the configured OTLP endpoint

#### Scenario: OTel log export disabled (no-op)
- **WHEN** `OTEL_EXPORTER_OTLP_ENDPOINT` is not set
- **THEN** log entries are not exported via OTel
- **AND** the app continues to function normally

### Requirement: Email settings log emission
The email/SMTP system SHALL emit structured log entries for key operations.

#### Scenario: Email sent successfully
- **WHEN** an email is sent successfully
- **THEN** a log entry at `info` level is created with source `email` and attributes including the recipient address

#### Scenario: Email send failure
- **WHEN** an email send fails
- **THEN** a log entry at `error` level is created with source `email` and attributes including the error and recipient

### Requirement: AI endpoint log emission
The AI endpoint system SHALL emit structured log entries for key operations.

#### Scenario: AI request started
- **WHEN** an AI text generation request starts
- **THEN** a log entry at `debug` level is created with source `ai` and the endpoint name

#### Scenario: AI request succeeded
- **WHEN** an AI request completes successfully
- **THEN** a log entry at `info` level is created with source `ai`, endpoint name, and response status

#### Scenario: AI request failed
- **WHEN** an AI request fails
- **THEN** a log entry at `error` level is created with source `ai`, endpoint name, and error details

### Requirement: Backup log emission
The backup system SHALL emit structured log entries for key operations.

#### Scenario: Backup created
- **WHEN** a backup is created successfully
- **THEN** a log entry at `info` level is created with source `backup` and the backup path

#### Scenario: Backup failed
- **WHEN** a backup fails
- **THEN** a log entry at `error` level is created with source `backup` and the error message

### Requirement: Ring buffer log storage
The system SHALL store recent log entries in an in-memory ring buffer for UI viewing.

#### Scenario: Ring buffer capacity
- **WHEN** the ring buffer reaches its maximum capacity (default 5000 entries)
- **THEN** the oldest entry is overwritten by each new entry
- **AND** the buffer continues to store the most recent entries

#### Scenario: Ring buffer configurable
- **WHEN** the `LOG_BUFFER_SIZE` environment variable is set to a positive integer
- **THEN** the ring buffer uses that value as its maximum capacity
