# character-spell-linking Specification

## Purpose
TBD - created by archiving change spell-picker-filters-admin-autosave. Update Purpose after archive.
## Requirements
### Requirement: Character spell picker filters
The character spell-link picker MUST allow users to filter compendium spells by name, class, and level, and MUST apply all selected filters together.

#### Scenario: Search by name, class, and level
- **WHEN** a user enters a spell name and selects a class and level range
- **THEN** the picker shows only spells matching all three criteria

#### Scenario: Filter by class without a name
- **WHEN** a user selects a class and leaves the name search empty
- **THEN** the picker returns spells available to that class without requiring a name query

#### Scenario: Filter by level
- **WHEN** a user selects cantrip, 1st–3rd, 4th–6th, or 7th–9th
- **THEN** the picker returns only spells whose level is within the selected range

#### Scenario: Pagination preserves filters
- **WHEN** a user moves to another picker page after applying filters
- **THEN** the request and displayed results retain the name, class, and level filters

### Requirement: Spell-only linking
The character spell-link operation MUST create links only from compendium spell records and MUST reject equipment records or invalid references.

#### Scenario: Link a compendium spell
- **WHEN** a user selects a valid compendium spell
- **THEN** the character spellbook receives the spell data and its compendium spell reference

#### Scenario: Equipment cannot be linked as a spell
- **WHEN** a spell-link request contains an equipment reference
- **THEN** the request is rejected and no spellbook row is created

#### Scenario: Generic entry must resolve to a spell
- **WHEN** a spell-link request contains a generic compendium entry reference
- **THEN** the request succeeds only if that entry represents a spell; otherwise it is rejected
