# link-monsters Specification

## Purpose
Search and link compendium monsters when adding monsters to one-shot acts.

## Requirements

### Requirement: One-shot act monster linking from compendium
The one-shot act editor SHALL allow the DM to search the compendium for monsters and link them to an act instead of creating inline monster entries.

#### Scenario: Search for compendium monster in act editor
- **WHEN** DM clicks "Add Monster from Compendium" button in the one-shot act monster section
- **THEN** a search modal/picker appears with text input, CR filter, and type filter
- **WHEN** DM types a search term or applies filters
- **THEN** the system queries `GET /api/compendium/search?type=monster&q=<term>&cr=<cr>&monster_type=<type>` and displays matching monsters with name, CR, type, AC, HP
- **WHEN** DM clicks a monster result
- **THEN** the monster is linked to the act (stored as `compendium_monster_id` reference) and appears in the act's monster list

#### Scenario: Render linked monster in act view
- **WHEN** an act has a linked compendium monster
- **THEN** the monster row SHALL display a "📖 Compendium" badge/indicator
- **WHEN** DM clicks the linked monster row
- **THEN** the monster card is fetched from `GET /compendium/card/monster/:id` and displayed as a stat-block modal, showing full compendium data including AC, HP, abilities, saves, skills, resistances, immunities, actions, legendary actions

#### Scenario: Unlink a linked monster
- **WHEN** DM clicks "Unlink" on a linked act monster
- **THEN** the `compendium_monster_id` is set to NULL
- **WHEN** the monster row has no other custom data, it SHALL be deleted

### Requirement: Backend API for compendium monster linking
The backend SHALL provide endpoints to link/unlink compendium monsters from one-shot acts.

#### Scenario: Link monster to act endpoint
- **WHEN** `POST /api/oneshots/:adventureId/acts/:actId/monsters/link` is called with `{"compendium_monster_id": 42}`
- **THEN** a new `oneshot_monster` row is created with `act_id`, `adventure_id`, `compendium_monster_id`, and core stats copied from the compendium entry
- **THEN** the response SHALL include the created monster row with stat block

#### Scenario: Link monster to scene endpoint
- **WHEN** `POST /api/oneshots/:adventureId/scenes/:sceneId/monsters/link` is called with `{"compendium_monster_id": 42}`
- **THEN** a new `oneshot_monster` row is created with `scene_id`, `adventure_id`, `compendium_monster_id`, and core stats copied

#### Scenario: Unlink monster endpoint
- **WHEN** `DELETE /api/oneshots/monsters/:monsterId/link` is called
- **THEN** `compendium_monster_id` is set to NULL on that monster row
- **WHEN** the monster row has no other custom data beyond the reference, it SHALL be deleted
