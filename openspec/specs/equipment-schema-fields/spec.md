# equipment-schema-fields Specification

## Purpose
Add item_type, item_rarity, publisher fields to equipment schema.

## Requirements

### Requirement: Equipment compendium has item_type field
The equipment schema SHALL include an `item_type` field for storing the specific item type (e.g., Scroll, Wand).

#### Scenario: Item type stored and returned
- **WHEN** an equipment entry is created or edited with an item_type value
- **THEN** the item_type SHALL be stored in the database
- **AND** returned in equipment API responses

### Requirement: Equipment compendium has item_rarity field
The equipment schema SHALL include an `item_rarity` select field with options: common, uncommon, rare, very rare, legendary, artifact.

#### Scenario: Item rarity stored and returned
- **WHEN** an equipment entry is created or edited with an item_rarity value
- **THEN** the item_rarity SHALL be stored in the database
- **AND** returned in equipment API responses

### Requirement: Equipment compendium has publisher field
The equipment schema SHALL include a `publisher` field.

#### Scenario: Publisher stored and returned
- **WHEN** an equipment entry is created or edited with a publisher value
- **THEN** the publisher SHALL be stored in the database
- **AND** returned in equipment API responses

### Requirement: Equipment schema fields auto-generate in admin UI
The admin UI SHALL render form fields for item_type, item_rarity, and publisher based on schema field definitions.

#### Scenario: Admin form shows new fields
- **WHEN** an admin creates or edits an equipment entry
- **THEN** the form SHALL include inputs for item_type (text), item_rarity (select), and publisher (text)
