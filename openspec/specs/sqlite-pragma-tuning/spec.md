# sqlite-pragma-tuning Specification

## Purpose
Optimized SQLite runtime configuration (synchronous, cache, mmap, temp_store).

## Requirements

### Requirement: Configure synchronous mode
The system SHALL set `PRAGMA synchronous=NORMAL` on database initialization for a balanced durability/performance tradeoff with WAL mode.

#### Scenario: synchronous pragma set on init
- **WHEN** the database connection is opened
- **THEN** `PRAGMA synchronous` SHALL be set to `NORMAL`
- **AND** the value SHALL be logged at startup

### Requirement: Configure page cache size
The system SHALL set `PRAGMA cache_size=-64000` (64MB) to keep hot data pages in memory for read performance.

#### Scenario: cache_size pragma set on init
- **WHEN** the database connection is opened
- **THEN** `PRAGMA cache_size` SHALL be set to `-64000`

### Requirement: Configure memory-mapped I/O
The system SHALL set `PRAGMA mmap_size=268435456` (256MB) to enable memory-mapped I/O for read queries.

#### Scenario: mmap_size pragma set on init
- **WHEN** the database connection is opened
- **THEN** `PRAGMA mmap_size` SHALL be set to `268435456`

#### Scenario: mmap_size configurable via env var
- **WHEN** environment variable `SQLITE_MMAP_SIZE` is set
- **THEN** the value SHALL be used instead of the default

### Requirement: Configure temp store to memory
The system SHALL set `PRAGMA temp_store=MEMORY` to avoid disk I/O for temporary tables and indices.

#### Scenario: temp_store pragma set on init
- **WHEN** the database connection is opened
- **THEN** `PRAGMA temp_store` SHALL be set to `MEMORY`

### Requirement: Periodic optimize call
The system SHALL call `PRAGMA optimize` periodically (every 60 minutes) to auto-analyze statistics for the query planner.

#### Scenario: optimize called on schedule
- **WHEN** 60 minutes have elapsed since the last optimize call
- **THEN** `PRAGMA optimize` SHALL be executed
- **AND** the execution SHALL be logged
