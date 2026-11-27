# wal-checkpoint-strategy Specification

## Purpose
Smarter WAL checkpoint management — checkpoint on write-heavy ops, less aggressive polling.

## Requirements

### Requirement: Passive WAL checkpoint after write-heavy operations
The system SHALL run `PRAGMA wal_checkpoint(PASSIVE)` after any write-heavy operation that affects more than 50 rows or writes more than 100KB of data.

#### Scenario: Checkpoint after batch import
- **WHEN** a compendium import operation completes and affects 50+ rows
- **THEN** `PRAGMA wal_checkpoint(PASSIVE)` SHALL be executed

#### Scenario: Checkpoint after character creation
- **WHEN** a new character is created with all sub-resources (inventory, spells, features)
- **THEN** `PRAGMA wal_checkpoint(PASSIVE)` SHALL be executed

### Requirement: Reduce periodic TRUNCATE checkpoint frequency
The system SHALL change the periodic checkpoint from hourly to every 4 hours.

#### Scenario: Periodic checkpoint interval changed
- **WHEN** the background cleanup task runs
- **THEN** `PRAGMA wal_checkpoint(TRUNCATE)` SHALL run every 4 hours instead of every 1 hour

### Requirement: Log WAL file size before and after checkpoint
The system SHALL log the WAL file size (in bytes or pages) before and after each TRUNCATE checkpoint.

#### Scenario: WAL size logged on checkpoint
- **WHEN** a TRUNCATE checkpoint completes
- **THEN** the log SHALL contain the WAL page count before and after the checkpoint
