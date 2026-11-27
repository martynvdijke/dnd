# race-schema-fields Specification

## Purpose
Add category, expansion, publisher fields to race schema.

## Requirements

### Requirement: Race compendium has category field
The race schema SHALL include a `category` field.

#### Scenario: Category stored and returned
- **WHEN** a race entry is created or edited with a category value
- **THEN** the category SHALL be stored in the database
- **AND** returned in race API responses

### Requirement: Race compendium has expansion field
The race schema SHALL include an `expansion` field.

#### Scenario: Expansion stored and returned
- **WHEN** a race entry is created or edited with an expansion value
- **THEN** the expansion SHALL be stored in the database
- **AND** returned in race API responses

### Requirement: Race compendium has publisher field
The race schema SHALL include a `publisher` field.

#### Scenario: Publisher stored and returned
- **WHEN** a race entry is created or edited with a publisher value
- **THEN** the publisher SHALL be stored in the database
- **AND** returned in race API responses

### Requirement: Race schema fields auto-generate in admin UI
The admin UI SHALL render form fields for category, expansion, and publisher based on schema field definitions.

#### Scenario: Admin form shows new fields
- **WHEN** an admin creates or edits a race entry
- **THEN** the form SHALL include inputs for category (text), expansion (text), and publisher (text)
