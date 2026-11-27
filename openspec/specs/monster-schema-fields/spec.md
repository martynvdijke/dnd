# monster-schema-fields Specification

## Purpose
Add alignment, expansion, publisher fields to monster schema.

## Requirements

### Requirement: Monster compendium has alignment field
The monster schema SHALL include an `alignment` field that stores the creature's moral and ethical alignment.

#### Scenario: Alignment stored and returned
- **WHEN** a monster entry is created or edited with an alignment value
- **THEN** the alignment SHALL be stored in the database
- **AND** returned in monster list and detail API responses

### Requirement: Monster compendium has expansion field
The monster schema SHALL include an `expansion` field that stores a sourcebook/expansion identifier.

#### Scenario: Expansion stored and returned
- **WHEN** a monster entry is created or edited with an expansion value
- **THEN** the expansion SHALL be stored in the database
- **AND** returned in monster API responses

### Requirement: Monster compendium has publisher field
The monster schema SHALL include a `publisher` field that stores the publisher name.

#### Scenario: Publisher stored and returned
- **WHEN** a monster entry is created or edited with a publisher value
- **THEN** the publisher SHALL be stored in the database
- **AND** returned in monster API responses

### Requirement: Monster schema fields are auto-generated in admin UI
The admin UI SHALL render form fields for alignment, expansion, and publisher based on schema field definitions.

#### Scenario: Admin form shows new fields
- **WHEN** an admin creates or edits a monster entry
- **THEN** the form SHALL include inputs for alignment (select), expansion (text), and publisher (text)
