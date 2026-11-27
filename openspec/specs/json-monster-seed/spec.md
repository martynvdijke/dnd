# json-monster-seed Specification

## Purpose
Expand monster seed data via a `data/monsters.json` file loaded at migration time, replacing the current hardcoded Go struct approach with a flexible JSON format supporting full stat blocks.

## Requirements

### Requirement: JSON monster seed file
The system SHALL seed the `compendium_monsters` table from a `data/monsters.json` file at migration time, replacing the current hardcoded Go struct approach in `db/seed_monsters.go`.

#### Scenario: Monsters seeded from JSON on fresh database
- **WHEN** the application runs migrations on a fresh database
- **THEN** `compendium_monsters` is populated with entries from `data/monsters.json`
- **AND** the `source` field is set to `"SRD"` for each seeded monster
- **AND** the seed runs only if `compendium_monsters` is empty (idempotent)

#### Scenario: Existing seed migration preserved
- **WHEN** the application starts with an already-seeded database
- **THEN** the seed process skips without modifying existing data

### Requirement: JSON format matches compendium schema
The `data/monsters.json` file SHALL contain an array of objects with fields matching the `compendium_monsters` table columns.

#### Scenario: All required fields present
- **WHEN** a monster entry is read from `data/monsters.json`
- **THEN** it contains `name`, `type`, `size`, `ac`, `hp`, `str`, `dex`, `con`, `int_`, `wis`, `cha`, `cr`, `saves`, `skills`, `damage_vulnerabilities`, `damage_resistances`, `damage_immunities`, `condition_immunities`, `senses`, `languages`, `special_abilities`, `actions`, `legendary_actions`, and `description`
- **AND** the `int_` field uses the underscore suffix to avoid SQL keyword conflict

#### Scenario: Full stat-block monsters
- **WHEN** a monster in the JSON has a full stat block (saves, skills, abilities, actions)
- **THEN** the `is_full` column is set to `1`
- **AND** all stat fields are populated in the database

### Requirement: Minimum 100 SRD monsters
The `data/monsters.json` file SHALL contain at least 100 distinct SRD monsters from the D&D 5e Systems Reference Document.

#### Scenario: Comprehensive monster roster
- **WHEN** the seed completes
- **THEN** the `compendium_monsters` table contains 100+ entries
- **AND** entries cover multiple CRs from 0 to 30
- **AND** entries cover multiple types (beast, humanoid, undead, dragon, etc.)
- **AND** all content is from the SRD 5.1 under CC-BY-4.0

### Requirement: Attribution preserved
The SRD content SHALL maintain proper CC-BY-4.0 attribution.

#### Scenario: License file references monster data
- **WHEN** a user views `data/README.md`
- **THEN** it covers the monster JSON file under the same SRD 5.1 attribution
- **AND** no copyrighted or trademarked content from non-SRD sources is included
