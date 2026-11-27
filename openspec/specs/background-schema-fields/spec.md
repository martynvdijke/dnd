# background-schema-fields Specification

## Purpose
Add category, data_list, data_bonds, data_flaws, data_ideals, data_equipment, data_starting_gold, data_personality_traits, publisher fields to background schema.

## Requirements

### Requirement: Background compendium has category field
The background schema SHALL include a `category` field.

#### Scenario: Category stored and returned
- **WHEN** a background entry is created or edited with a category value
- **THEN** the category SHALL be stored in the database
- **AND** returned in background API responses

### Requirement: Background compendium has data_list field
The background schema SHALL include a `data_list` boolean field.

#### Scenario: Data list stored and returned
- **WHEN** a background entry is created or edited with a data_list value
- **THEN** the data_list SHALL be stored in the database
- **AND** returned in background API responses

### Requirement: Background compendium has data_bonds field
The background schema SHALL include a `data_bonds` text field for storing bonds as JSON.

#### Scenario: Data bonds stored and returned
- **WHEN** a background entry is created or edited with a data_bonds value
- **THEN** the data_bonds SHALL be stored in the database
- **AND** returned in background API responses

### Requirement: Background compendium has data_flaws field
The background schema SHALL include a `data_flaws` text field for storing flaws as JSON.

#### Scenario: Data flaws stored and returned
- **WHEN** a background entry is created or edited with a data_flaws value
- **THEN** the data_flaws SHALL be stored in the database
- **AND** returned in background API responses

### Requirement: Background compendium has data_ideals field
The background schema SHALL include a `data_ideals` text field for storing ideals as JSON.

#### Scenario: Data ideals stored and returned
- **WHEN** a background entry is created or edited with a data_ideals value
- **THEN** the data_ideals SHALL be stored in the database
- **AND** returned in background API responses

### Requirement: Background compendium has data_equipment field
The background schema SHALL include a `data_equipment` text field for storing equipment as JSON.

#### Scenario: Data equipment stored and returned
- **WHEN** a background entry is created or edited with a data_equipment value
- **THEN** the data_equipment SHALL be stored in the database
- **AND** returned in background API responses

### Requirement: Background compendium has data_starting_gold field
The background schema SHALL include a `data_starting_gold` integer field.

#### Scenario: Data starting gold stored and returned
- **WHEN** a background entry is created or edited with a data_starting_gold value
- **THEN** the data_starting_gold SHALL be stored in the database
- **AND** returned in background API responses

### Requirement: Background compendium has data_personality_traits field
The background schema SHALL include a `data_personality_traits` text field for storing personality traits as JSON.

#### Scenario: Data personality traits stored and returned
- **WHEN** a background entry is created or edited with a data_personality_traits value
- **THEN** the data_personality_traits SHALL be stored in the database
- **AND** returned in background API responses

### Requirement: Background compendium has publisher field
The background schema SHALL include a `publisher` field.

#### Scenario: Publisher stored and returned
- **WHEN** a background entry is created or edited with a publisher value
- **THEN** the publisher SHALL be stored in the database
- **AND** returned in background API responses

### Requirement: Background schema fields auto-generate in admin UI
The admin UI SHALL render form fields for all new background fields based on schema field definitions.

#### Scenario: Admin form shows new fields
- **WHEN** an admin creates or edits a background entry
- **THEN** the form SHALL include inputs for all new background fields
