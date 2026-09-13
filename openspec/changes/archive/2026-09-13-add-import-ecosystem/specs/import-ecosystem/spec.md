## ADDED Requirements

### Requirement: External character import
The system SHALL import characters from D&D Beyond and Foundry VTT actor payloads into the existing canonical character shape, populating abilities, class/level, HP, AC, and inventory where the source provides them.

#### Scenario: D&D Beyond character imported
- **WHEN** a user submits a D&D Beyond character payload
- **THEN** a Villum character is created with the mapped name, abilities, class/level, HP and AC

#### Scenario: Foundry actor imported
- **WHEN** a user submits a Foundry VTT actor payload
- **THEN** abilities, HP, AC and owned weapons are mapped into a Villum character

### Requirement: External compendium import
The system SHALL import 5e.tools and Foundry JSON source packs into a compendium schema using field mapping, reusing the existing compendium import pipeline for validation, dedup and logging.

#### Scenario: 5e.tools pack imported
- **WHEN** a user imports a 5e.tools JSON file into a schema
- **THEN** entries are created under that schema with names mapped from the source

#### Scenario: Foundry LevelDB pack rejected clearly
- **WHEN** a user supplies a Foundry LevelDB `.db` pack
- **THEN** the system reports that only JSON source packs are supported and imports nothing

### Requirement: Dry-run preview
Every external import SHALL support a dry-run that reports, without writing, how many rows would be created, skipped, or treated as duplicates.

#### Scenario: Preview before commit
- **WHEN** a user submits an import with dry-run enabled
- **THEN** a per-row preview is returned and no data is written

### Requirement: Duplicate policy
Compendium import SHALL require an explicit duplicate policy (`skip`, `overwrite`, `create-new`, `force`) and apply it using the existing name-based dedup.

#### Scenario: Duplicate skipped
- **WHEN** an imported entry's name already exists and the policy is `skip`
- **THEN** the existing entry is left unchanged and the row is reported skipped

### Requirement: Import logging and rollback
Each external import SHALL be recorded with its source, counts and the ids it created, and SHALL be reversible by deleting exactly those created rows.

#### Scenario: Rollback removes created entries
- **WHEN** a user rolls back a completed character or compendium import
- **THEN** the rows created by that import are deleted and the log is marked rolled back

### Requirement: Safe external fetch
When an import is fetched from a URL, the system SHALL reject loopback and private-network addresses and bound the response size.

#### Scenario: Private address rejected
- **WHEN** an import URL resolves to a loopback or private address
- **THEN** the fetch is refused with an error
