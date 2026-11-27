# link-spells Specification

## Purpose
Search and link compendium spells from character sheet; render spell overview and spell card from reference.

## Requirements

### Requirement: Character sheet spell linking from compendium
The character sheet spell section SHALL allow the user to search the compendium for a spell and link it to the character instead of creating an inline spell.

#### Scenario: Search for compendium spell
- **WHEN** user clicks "Add Spell from Compendium" button on the character sheet spells section
- **THEN** a search modal/picker appears with a text input and a "Search" button
- **WHEN** user types a search term and clicks Search
- **THEN** the system queries `GET /api/compendium/search?type=spell&q=<term>` and displays matching spells as a list with name, level, and school
- **WHEN** user clicks a spell result
- **THEN** the spell is linked to the character (stored as `compendium_spell_id` reference) and appears in the spell list

#### Scenario: Render linked spell in spell list
- **WHEN** a character has a linked compendium spell
- **THEN** the spell row in the spell list SHALL display a "📖 Compendium" badge/indicator
- **WHEN** user clicks the linked spell row
- **THEN** the spell card is fetched from `GET /compendium/card/spell/:id` and displayed in a modal or expanded panel, showing full compendium data including description, components, duration, range, higher levels, and classes

#### Scenario: Unlink a linked spell
- **WHEN** user clicks "Unlink" on a linked spell
- **THEN** the `compendium_spell_id` is set to NULL on that spell row
- **WHEN** the spell was only a reference (no inline data), it SHALL be deleted from the character's spells
- **WHEN** the spell has both inline data and a link, only the link is removed; inline data remains

#### Scenario: Mixed inline and linked spells
- **WHEN** a character has both inline and compendium-linked spells
- **THEN** both types SHALL appear in the spell list sorted by level then name
- **THEN** linked spells SHALL be visually distinguishable from inline spells (badge or icon)

### Requirement: Backend API for compendium spell linking
The backend SHALL provide endpoints to link/unlink compendium spells from a character.

#### Scenario: Link spell endpoint
- **WHEN** `POST /api/characters/:id/spells/link` is called with `{"compendium_spell_id": 42}`
- **THEN** a new `spell` row is created with `character_id`, `compendium_spell_id`, `prepared: false`, `always_prepared: false`, and `name` copied from the compendium entry
- **THEN** the response SHALL include the created spell row with full data

#### Scenario: Unlink spell endpoint
- **WHEN** `DELETE /api/characters/:id/spells/:spellId/link` is called
- **THEN** `compendium_spell_id` is set to NULL on that spell row
- **WHEN** the spell row has no inline data (all other fields are empty/default), it SHALL be deleted
