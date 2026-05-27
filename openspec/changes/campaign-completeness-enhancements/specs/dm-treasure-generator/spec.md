## ADDED Requirements

### Requirement: Treasure generator
The system SHALL provide a treasure generation tool that produces loot from DMG treasure tables based on encounter CR and hoard type.

#### Scenario: Generate individual treasure
- **WHEN** a DM clicks "Generate Treasure" and selects "Individual"
- **THEN** the system SHALL show inputs for CR range (0-4, 5-10, 11-16, 17+) and number of rolls
- **WHEN** submitted
- **THEN** the system SHALL generate coin amounts per the DMG Individual Treasure tables for the selected CR range

#### Scenario: Generate hoard treasure
- **WHEN** a DM clicks "Generate Treasure" and selects "Hoard"
- **THEN** the system SHALL show inputs for CR range
- **WHEN** submitted
- **THEN** the system SHALL generate coins, gemstones, art objects, and magic items per the DMG Hoard Treasure tables

#### Scenario: Treasure result displayed with total value
- **WHEN** treasure is generated
- **THEN** the system SHALL display the items, coins, and their total gold piece value

#### Scenario: Add generated treasure to party inventory
- **WHEN** treasure is generated
- **THEN** the system SHALL provide an "Add to Party Loot" button that transfers the generated items/coins to the party inventory/treasury
