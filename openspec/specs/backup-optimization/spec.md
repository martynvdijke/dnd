# backup-optimization Specification

## Purpose
Database backups use SQLite `VACUUM INTO` for safer, faster, more compact backup files instead of copying the live database.

## Requirements

### Requirement: Migrate backup to VACUUM INTO
The system SHALL use `VACUUM INTO '<backup_path>'` instead of file copy for creating database backups, producing a consistent, compacted snapshot.

#### Scenario: Backup uses VACUUM INTO
- **WHEN** the backup scheduler triggers a backup
- **THEN** `VACUUM INTO '<backup_file_path>'` SHALL be executed
- **AND** the backup file SHALL be a valid, consistent SQLite database

#### Scenario: Manual backup trigger uses VACUUM INTO
- **WHEN** a user triggers a manual backup via the API
- **THEN** `VACUUM INTO` SHALL be used instead of file copy

### Requirement: Fall back to file copy on VACUUM INTO failure
The system SHALL fall back to the original file-copy backup method if `VACUUM INTO` fails (e.g., disk space too low).

#### Scenario: Fallback on VACUUM INTO error
- **WHEN** `VACUUM INTO` returns an error
- **THEN** the system SHALL fall back to the file-copy backup method
- **AND** the fallback SHALL be logged as a warning

### Requirement: Log backup size savings
The system SHALL log the size of the VACUUM INTO backup compared to the live database file.

#### Scenario: Backup size logged
- **WHEN** a VACUUM INTO backup completes
- **THEN** the log SHALL contain the backup file size and the percentage reduction from the live DB file size
