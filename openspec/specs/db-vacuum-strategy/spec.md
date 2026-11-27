# db-vacuum-strategy Specification

## Purpose
Auto-vacuum and periodic VACUUM to control DB file size growth.

## Requirements

### Requirement: Enable auto_vacuum mode
The system SHALL enable `PRAGMA auto_vacuum=INCREMENTAL` to allow incremental free-page reuse without full VACUUM.

#### Scenario: auto_vacuum set on init
- **WHEN** the database connection is opened
- **THEN** `PRAGMA auto_vacuum` SHALL be set to `INCREMENTAL`

### Requirement: Periodic incremental vacuum
The system SHALL run `PRAGMA incremental_vacuum` weekly (every 168 hours) to reclaim free pages.

#### Scenario: Incremental vacuum on schedule
- **WHEN** 168 hours have elapsed since the last incremental vacuum
- **THEN** `PRAGMA incremental_vacuum` SHALL be executed
- **AND** the number of freed pages SHALL be logged

### Requirement: Full VACUUM on major schema migrations
The system SHALL run a full `VACUUM` after any schema migration that drops or alters columns.

#### Scenario: VACUUM after schema change
- **WHEN** a migration drops or alters a column in an existing table
- **THEN** `VACUUM` SHALL be executed after the migration completes

### Requirement: Log DB page statistics
The system SHALL log database page statistics (total pages, free pages, page size) on startup and after each VACUUM.

#### Scenario: Page stats logged on startup
- **WHEN** the database connection is initialized
- **THEN** `PRAGMA page_count`, `PRAGMA freelist_count`, and `PRAGMA page_size` SHALL be logged

#### Scenario: Page stats logged after VACUUM
- **WHEN** a VACUUM operation completes
- **THEN** the updated page_count and freelist_count SHALL be logged
