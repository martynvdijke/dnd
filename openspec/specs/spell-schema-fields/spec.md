# spell-schema-fields Specification

## Purpose
The spell compendium schema carries extended fields (including publisher and homebrew origin) beyond the basic spell attributes.

## Requirements

### Requirement: Spell compendium has publisher field
The spell schema SHALL include a `publisher` field.

#### Scenario: Publisher stored and returned
- **WHEN** a spell entry is created or edited with a publisher value
- **THEN** the publisher SHALL be stored in the database
- **AND** returned in spell API responses

### Requirement: Spell schema fields auto-generate in admin UI
The admin UI SHALL render form fields for publisher based on schema field definitions.

#### Scenario: Admin form shows new fields
- **WHEN** an admin creates or edits a spell entry
- **THEN** the form SHALL include an input for publisher (text)
