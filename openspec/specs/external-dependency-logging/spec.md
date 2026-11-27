# external-dependency-logging Specification

## Purpose
Comprehensive structured logging for all external dependency handlers (AI, email, backup, cleanup, WebSocket, crafting) visible in the admin Logs tab.

## Requirements

### Requirement: AI handler emits structured logs with source "ai"

The system SHALL emit structured log entries from `handlers/ai.go` using `middleware.AppLog` with the source attribute set to `ai`. Log entries SHALL capture endpoint ID, model, and operation type. API keys and full prompt content SHALL NOT appear in log attributes. Log entries SHALL use appropriate levels: `debug` for request start (with prompt_length), `info` for successful completion (with response_time_ms and finish_reason), `error` for HTTP/parse failures (with status and truncated response preview).

#### Scenario: AI text generation success is logged at info level
- **WHEN** a text generation request completes successfully with HTTP 200
- **THEN** a log entry appears with level `info`, source `ai`, message indicating success, and attributes including `endpoint_id`, `model`, `response_time_ms`, and `finish_reason`

#### Scenario: AI text generation HTTP failure is logged at error level
- **WHEN** a text generation request returns a non-2xx HTTP status
- **THEN** a log entry appears with level `error`, source `ai`, and attributes including `status` and `response_preview` (truncated to 200 chars)

#### Scenario: AI image generation success is logged at info level
- **WHEN** an image generation request completes successfully
- **THEN** a log entry appears with level `info`, source `ai`, and attributes including `endpoint_id`, `model`, `response_time_ms`

#### Scenario: AI test endpoint results are logged
- **WHEN** an admin tests an AI endpoint
- **THEN** a log entry appears with source `ai`, level `debug` at start, and level `info` (success) or `warn` (failure) on completion

### Requirement: Email handler emits structured logs with source "email"

The system SHALL emit structured log entries from `handlers/email.go` using `middleware.AppLog` with the source attribute set to `email`. Log entries SHALL capture the send type (test, campaign_highlights), recipient, and result.

#### Scenario: Test email send success is logged
- **WHEN** a test email is sent successfully
- **THEN** a log entry appears with level `info`, source `email`, message "Test email sent", and attribute `to` with the recipient address

#### Scenario: Email send failure is logged at error level
- **WHEN** an email send fails (SMTP connection error, auth failure, etc.)
- **THEN** a log entry appears with level `error`, source `email`, and attribute `error` with the failure description

#### Scenario: Campaign highlights email send is logged
- **WHEN** campaign highlights are emailed successfully
- **THEN** a log entry appears with level `info`, source `email`, and attributes `campaign_id` and `recipient_count`

### Requirement: Backup handler emits structured logs with source "backup"

The system SHALL emit structured log entries from `handlers/backup.go` using `middleware.AppLog` with the source attribute set to `backup`.

#### Scenario: Backup creation is logged
- **WHEN** an automatic backup is created successfully
- **THEN** a log entry appears with level `info`, source `backup`, and attribute `path`

#### Scenario: Backup failure is logged at error level
- **WHEN** an automatic backup fails
- **THEN** a log entry appears with level `error`, source `backup`, and attribute `error`

#### Scenario: Backup pruning is logged
- **WHEN** old backups are pruned
- **THEN** a log entry appears with level `info`, source `backup`, and attributes `removed` and `keep_count`

### Requirement: Cleanup handler emits structured logs with source "cleanup"

The system SHALL emit structured log entries from `handlers/cleanup.go` using `middleware.AppLog` with the source attribute set to `cleanup`.

#### Scenario: Cleanup sweep results are logged
- **WHEN** a cleanup sweep completes
- **THEN** log entries appear with level `info`, source `cleanup`, and attributes describing the sweep type and count of removed items

#### Scenario: Cleanup failure is logged
- **WHEN** a cleanup operation fails
- **THEN** a log entry appears with level `error`, source `cleanup`, and attribute `error`

### Requirement: WebSocket handler emits structured logs with source "ws"

The system SHALL emit structured log entries from `handlers/ws.go` using `middleware.AppLog` with the source attribute set to `ws`.

#### Scenario: WebSocket upgrade failure is logged
- **WHEN** a WebSocket upgrade fails
- **THEN** a log entry appears with level `warn`, source `ws`, and attribute `error`

### Requirement: Crafting handler emits structured logs with source "crafting"

The system SHALL emit structured log entries from `handlers/crafting.go` using `middleware.AppLog` with the source attribute set to `crafting`.

#### Scenario: Recipe seeding failure is logged
- **WHEN** a recipe seeding operation fails
- **THEN** a log entry appears with level `error`, source `crafting`, and attributes including `recipe_name` and `error`

### Requirement: Admin UI populates source filter dynamically

The system SHALL provide a `GET /api/admin/log-sources` endpoint that returns the distinct set of source names observed in the ring buffer. The admin Logs tab SHALL call this endpoint on tab load and populate the source filter dropdown dynamically.

#### Scenario: Source filter is populated from observed sources
- **WHEN** the admin opens the Logs tab
- **THEN** the source filter dropdown is populated with options from `GET /api/admin/log-sources` plus an "All" option at the top

#### Scenario: Select a source filter shows only matching logs
- **WHEN** the admin selects a source from the filter dropdown
- **THEN** only log entries with that source are displayed

### Requirement: Log entry detail expansion in admin UI

The admin Logs tab SHALL support clicking a log row to expand/collapse a detail row showing full attribute key-value pairs. The attribute data SHALL already be present in the fetched logs (no additional API call needed).

#### Scenario: Click log row reveals attribute details
- **WHEN** the admin clicks a log row
- **THEN** a detail row is inserted below showing all attributes as key-value pairs in a readable layout

#### Scenario: Click expanded row collapses details
- **WHEN** the admin clicks an already-expanded log row
- **THEN** the detail row is removed

### Requirement: Auto-scroll to latest logs

The admin Logs tab SHALL auto-scroll to the bottom when new entries arrive during auto-refresh, but only if the user is already scrolled to the bottom (within 50px). If the user has scrolled up to inspect entries, auto-scroll SHALL be suppressed.

#### Scenario: Auto-scroll activates when at bottom
- **WHEN** new log entries arrive via auto-refresh and the log table is scrolled to within 50px of the bottom
- **THEN** the viewport auto-scrolls to show the new entries

#### Scenario: Auto-scroll is suppressed when scrolled up
- **WHEN** new log entries arrive via auto-refresh and the user has scrolled up more than 50px from the bottom
- **THEN** the viewport does NOT auto-scroll

### Requirement: AI endpoint calls are logged
The system SHALL log all outbound HTTP calls to AI endpoints with method, URL, status code, duration, and truncated error response body.

#### Scenario: Successful AI text generation is logged
- **WHEN** a handler in `handlers/ai.go` makes a POST request to an AI endpoint for text generation
- **THEN** it SHALL log an info-level entry with source `"ai"` containing attributes: `endpoint`, `model`, `status_code`, `duration_ms`
- **AND** the response status code and duration SHALL be captured

#### Scenario: Failed AI text generation is logged
- **WHEN** an AI text generation request fails (network error, non-2xx status)
- **THEN** it SHALL log an error-level entry with source `"ai"` containing attributes: `endpoint`, `status_code`, `duration_ms`, `error`

#### Scenario: AI image generation is logged
- **WHEN** a handler makes a POST request to an AI image generation endpoint
- **THEN** it SHALL log at info level (success) or error level (failure) with source `"ai"` and attributes: `endpoint`, `status_code`, `duration_ms`

### Requirement: Email sending is logged
The system SHALL log all outbound email operations including success, failure, and test email attempts.

#### Scenario: Successful email sent is logged
- **WHEN** an email is successfully sent via the email handler
- **THEN** it SHALL log an info-level entry with source `"email"` containing attributes: `recipient`, `subject`, `duration_ms`

#### Scenario: Email send failure is logged
- **WHEN** an email send attempt fails
- **THEN** it SHALL log an error-level entry with source `"email"` containing attributes: `recipient`, `error`

### Requirement: Backup operations are logged
The system SHALL log backup creation, pruning, and failure events.

#### Scenario: Backup created is logged
- **WHEN** an auto or manual backup is created successfully
- **THEN** it SHALL log an info-level entry with source `"backup"` containing attributes: `path`, `duration_ms`

#### Scenario: Backup failure is logged
- **WHEN** a backup operation fails
- **THEN** it SHALL log an error-level entry with source `"backup"` containing attributes: `error`

#### Scenario: Backup pruning is logged
- **WHEN** old backups are pruned
- **THEN** it SHALL log an info-level entry with source `"backup"` containing attributes: `removed_count`, `keep_count`

### Requirement: D&D 5e API lookups are logged
The system SHALL log outbound requests to the external D&D 5e API (`/compendium/api/:category`).

#### Scenario: Successful D&D API lookup is logged
- **WHEN** a request is made to the external D&D 5e API
- **THEN** it SHALL log an info-level entry with source `"dnd-api"` containing attributes: `category`, `status_code`, `duration_ms`

#### Scenario: Failed D&D API lookup is logged
- **WHEN** a D&D API request fails or returns an error status
- **THEN** it SHALL log an error-level entry with source `"dnd-api"` containing attributes: `category`, `error`, `status_code`

### Requirement: Cleanup tasks are logged
The system SHALL log all background cleanup task executions with affected record counts.

#### Scenario: Cleanup task removes records
- **WHEN** a background cleanup task removes old records (dice rolls, rest logs, NPC links)
- **THEN** it SHALL log an info-level entry with source `"cleanup"` containing attributes: `task`, `removed_count`

#### Scenario: Cleanup task encounters an error
- **WHEN** a cleanup task query fails
- **THEN** it SHALL log an error-level entry with source `"cleanup"` containing attributes: `task`, `error`

### Requirement: WebSocket connections are logged
The system SHALL log WebSocket connection upgrades and errors.

#### Scenario: WebSocket upgrade failure is logged
- **WHEN** a WebSocket upgrade attempt fails
- **THEN** it SHALL log an error-level entry with source `"ws"` containing attribute: `error`

#### Scenario: WebSocket connection established is logged (debug level)
- **WHEN** a WebSocket connection is successfully upgraded
- **THEN** it SHALL log a debug-level entry with source `"ws"` containing attribute: `remote_addr`

### Requirement: Database errors are logged via AppLog
The system SHALL replace all `log.Printf` database error messages with structured AppLog calls.

#### Scenario: DB connection info is logged
- **WHEN** the database connection is established and PRAGMA info is printed
- **THEN** it SHALL log an info-level entry with source `"db"` containing attributes: `page_count`, `page_size`, `mmap_size`

#### Scenario: DB migration progress is logged
- **WHEN** a database migration runs
- **THEN** it SHALL log info-level entries with source `"migration"` containing attributes: `version`, `statement` (at debug level for ALTER statements)

#### Scenario: Seed operations are logged
- **WHEN** compendium data is seeded
- **THEN** it SHALL log info-level entries with source `"seed"` containing attributes: `entity_type`, `count`
