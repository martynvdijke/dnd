# compendium-spell-linking Specification

## Purpose
Characters link their known and prepared spells from the compendium spell list, keeping spell details in sync with the shared compendium.

## Requirements

### Requirement: Compendium spell reference on character spells

The `Spell` model SHALL have a new nullable field `compendium_spell_id` referencing the compendium spell list. When set, the spell's name, level, school, casting time, range, components, duration, and description SHALL be pre-filled from the compendium entry on creation. The spell data SHALL be stored as a copy.

#### Scenario: Add spell from compendium

- **WHEN** the user clicks "Link from Compendium" in the spells tab
- **THEN** a compendium spell picker SHALL open, showing spells filterable by level, school, and name

#### Scenario: New spell copies compendium data

- **WHEN** the user selects a compendium spell and confirms
- **THEN** a new spell SHALL be created with `compendium_spell_id` set, and all spell fields pre-filled from the compendium entry

#### Scenario: Manual spell creation unchanged

- **WHEN** the user clicks "Add Spell" (not from compendium)
- **THEN** the existing manual spell creation flow SHALL work as before, with `compendium_spell_id` set to null

### Requirement: Compendium link indicator in spell list

Spells linked to a compendium entry SHALL display a compendium badge/icon next to their name in the spell list. Hovering SHALL show the compendium source.

#### Scenario: Linked spell badge visible

- **WHEN** rendering a spell with `compendium_spell_id` set
- **THEN** a compendium badge/icon SHALL be shown next to the spell name

### Requirement: Unlink compendium reference

Linked spells SHALL have an "Unlink" option that removes the `compendium_spell_id` reference while keeping the spell data intact.

#### Scenario: Unlink keeps spell data

- **WHEN** the user clicks "Unlink" on a compendium-linked spell
- **THEN** the `compendium_spell_id` SHALL be set to null, but the spell SHALL remain with all its data unchanged
