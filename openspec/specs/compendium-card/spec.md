# compendium-card Specification

## Purpose
Reusable UI component to render a rich detail card for any linked compendium entity.

## Requirements

### Requirement: Reusable compendium card component
The system SHALL provide a server-rendered HTML partial that renders a rich detail card for any compendium entity type (spell, equipment, monster) fetched via HTMX.

#### Scenario: Fetch spell card
- **WHEN** `GET /compendium/card/spell/:id` is called
- **THEN** the response SHALL be an HTML partial containing the spell's full compendium data:
  - Name, level, school (with school icon/color)
  - Casting time, range, components, duration
  - Full description (parsed for formatting)
  - Higher level description
  - Class list
  - Source and page reference
- **THEN** the response SHALL include `HX-Trigger: compendiumCardLoaded` after render

#### Scenario: Fetch equipment card
- **WHEN** `GET /compendium/card/equipment/:id` is called
- **THEN** the response SHALL be an HTML partial containing the equipment's full compendium data:
  - Name, category, item type, rarity
  - Cost, weight
  - Full description
  - Properties (weapon damage, armor class bonus, etc.)
  - Source reference

#### Scenario: Fetch monster card
- **WHEN** `GET /compendium/card/monster/:id` is called
- **THEN** the response SHALL be an HTML partial containing the monster's full compendium stat block:
  - Name, size, type, alignment
  - Armor Class, Hit Points
  - Speed
  - Ability scores (STR/DEX/CON/INT/WIS/CHA with modifiers)
  - Saving throws, skills
  - Damage vulnerabilities, resistances, immunities
  - Condition immunities
  - Senses, languages
  - Challenge Rating (with XP)
  - Special abilities (full description)
  - Actions (full description)
  - Legendary actions (if any)
  - Source reference

#### Scenario: Card not found
- **WHEN** the compendium entry does not exist for the given ID
- **THEN** the response SHALL be an HTML partial with a "Entry not found in compendium" message and a "Remove link" button

### Requirement: Compendium search endpoint with type filter
The search endpoint SHALL support filtering by entity type for contextual pickers.

#### Scenario: Search with type filter
- **WHEN** `GET /api/compendium/search?type=spell&q=fire` is called
- **THEN** the response SHALL only include results from `compendium_spells` matching "fire"
- **WHEN** `GET /api/compendium/search?type=equipment&q=plate` is called
- **THEN** the response SHALL only include results from `compendium_equipment` matching "plate"
- **WHEN** `GET /api/compendium/search?type=monster&q=dragon` is called
- **THEN** the response SHALL only include results from `compendium_monsters` matching "dragon"

#### Scenario: Search without type filter (backward compatible)
- **WHEN** `GET /api/compendium/search?q=fire` is called (no type param)
- **THEN** the response SHALL match the current behavior searching all compendium tables
