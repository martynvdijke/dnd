# query-instrumentation Specification

## Purpose
EXPLAIN QUERY PLAN analysis and slow query logging for hot paths.

## Requirements

### Requirement: Slow query logging
The system SHALL log SQL queries that exceed a configurable duration threshold (default 100ms), including the query text and execution time.

#### Scenario: Slow query logged
- **WHEN** a SQL query executes in more than 100ms
- **THEN** a warning log entry SHALL contain the query text and duration in milliseconds

#### Scenario: Threshold configurable
- **WHEN** environment variable `SQLITE_SLOW_QUERY_MS` is set
- **THEN** the slow query threshold SHALL use that value instead of the default 100ms

### Requirement: EXPLAIN QUERY PLAN on slow queries
The system SHALL run `EXPLAIN QUERY PLAN` on any query that exceeds the slow query threshold and log the query plan for debugging.

#### Scenario: EXPLAIN QUERY PLAN logged for slow queries
- **WHEN** a query exceeds the slow query threshold
- **THEN** EXPLAIN QUERY PLAN SHALL be executed for that query
- **AND** the execution plan SHALL be logged at debug level

### Requirement: Extend OTel tracing with query attributes
The existing `TraceQuery` function SHALL be extended to include the SQL query text and row count as span attributes.

#### Scenario: Query text in trace span
- **WHEN** a traced SQL query executes
- **THEN** the trace span SHALL include `db.query` attribute with the query text

#### Scenario: Row count in trace span
- **WHEN** a traced SQL query completes
- **THEN** the trace span SHALL include `db.rows_affected` attribute with the count

### Requirement: Query duration histogram metric
The system SHALL expose a Prometheus/OTel histogram metric for SQL query durations, bucketed by operation type.

#### Scenario: Query duration metric registered
- **WHEN** the application starts
- **THEN** a histogram metric `db_query_duration_ms` SHALL be registered
- **AND** it SHALL have `operation` and `table` labels

#### Scenario: Query duration recorded
- **WHEN** a SQL query completes
- **THEN** its duration SHALL be recorded in the `db_query_duration_ms` histogram
