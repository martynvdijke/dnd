# otel-log-correlation Specification

## Purpose
Trace ID injection into structured log entries for cross-system correlation.

## Requirements

### Requirement: Log entries include trace ID when available
The system SHALL inject the current OpenTelemetry trace ID into log entry attributes when a valid trace context exists.

#### Scenario: Trace ID present in context
- **WHEN** a log record is handled by `logHandler.Handle()`
- **AND** the request context contains a valid OTel span with a trace ID
- **THEN** the log entry SHALL include a `trace_id` attribute containing the hex-encoded trace ID string

#### Scenario: No trace context
- **WHEN** a log record is handled by `logHandler.Handle()`
- **AND** the request context does not contain a valid OTel span
- **THEN** the log entry SHALL NOT include a `trace_id` attribute

### Requirement: Trace ID appears in admin log viewer
The admin log viewer SHALL display the `trace_id` attribute when present, and SHALL make it clickable to copy to clipboard.

#### Scenario: Trace ID displayed in log entry
- **WHEN** a log entry has a `trace_id` attribute
- **THEN** the admin log viewer SHALL display it in the attributes column with a distinctive styling (e.g., monospace, with a copy button)

#### Scenario: Trace ID copied to clipboard
- **WHEN** the user clicks on a trace ID value in the log viewer
- **THEN** the trace ID SHALL be copied to the clipboard
- **AND** a brief "Copied!" tooltip SHALL appear

### Requirement: OTel log export passes trace ID
The OTel log export function in `telemetry.go` SHALL include the trace ID in OTLP log record attributes when available.

#### Scenario: OTel log export includes trace_id
- **WHEN** the OTel log export function sends a log record to the OTLP endpoint
- **AND** the request context contains a valid trace ID
- **THEN** the exported log record SHALL include an attribute `trace_id` with the hex trace ID string
