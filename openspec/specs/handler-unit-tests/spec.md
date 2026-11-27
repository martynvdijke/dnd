# handler-unit-tests Specification

## Purpose
Per-handler isolated Go tests with dedicated DB, mock auth, table-driven cases for every handler group.

## Requirements

### Requirement: Characters handler unit tests

The system SHALL have per-handler unit tests for every endpoint in `handlers/characters.go`.

#### Scenario: Create character with valid data returns 201
- **WHEN** POST /api/characters is called with valid name, race, class, and ability scores
- **THEN** response code SHALL be 201 and response body SHALL contain the created character with id, name, race, class fields

#### Scenario: Create character with missing name returns 400
- **WHEN** POST /api/characters is called without a name field
- **THEN** response code SHALL be 400 with an error message

#### Scenario: Create character with negative ability scores returns 400
- **WHEN** POST /api/characters is called with str=-5
- **THEN** response code SHALL be 400

#### Scenario: Get character returns full data
- **WHEN** GET /api/characters/:id is called for an existing character
- **THEN** response code SHALL be 200 and body SHALL contain name, race, class, level, hp, abilities, modifiers, proficiencies

#### Scenario: Get non-existent character returns 404
- **WHEN** GET /api/characters/99999 is called
- **THEN** response code SHALL be 404

#### Scenario: Update character validates at minimum required fields
- **WHEN** PUT /api/characters/:id omits required fields (race, class, level, abilities)
- **THEN** response code SHALL be 400

#### Scenario: Delete character cascades to related data
- **WHEN** DELETE /api/characters/:id is called for a character with inventory items, spells, and features
- **THEN** response code SHALL be 200 and a subsequent GET SHALL return 404

#### Scenario: Update currency validates non-negative values
- **WHEN** PUT /api/characters/:id/currency is called with negative gp value
- **THEN** response code SHALL be 400

### Requirement: Combat handler unit tests

The system SHALL have per-handler unit tests for every endpoint in `handlers/combat.go`.

#### Scenario: Create combat entry validates required fields
- **WHEN** POST /api/combat is called without name or hp_max
- **THEN** response code SHALL be 400

#### Scenario: Roll initiative for character returns valid range
- **WHEN** POST /api/combat/initiative is called with a valid character_id
- **THEN** response code SHALL be 200 and total SHALL be between 1 and (20+dex_mod)

#### Scenario: Next turn with no entries returns 200 with empty state
- **WHEN** POST /api/combat/next-turn is called with no combat entries
- **THEN** response code SHALL be 200

#### Scenario: Update combat entry validates HP bounds
- **WHEN** PUT /api/combat/:id sets hp_current below 0
- **THEN** response SHALL allow 0 (unconscious) but not negative

#### Scenario: Delete non-existent entry returns 200 or 404
- **WHEN** DELETE /api/combat/99999 is called
- **THEN** response code SHALL be 200 (idempotent)

### Requirement: Compendium handler unit tests

The system SHALL have per-handler unit tests for compendium endpoints.

#### Scenario: List races returns array with system/source fields
- **WHEN** GET /api/compendium/races is called
- **THEN** response code SHALL be 200 and each entry SHALL have system and source fields

#### Scenario: Search compendium requires non-empty query
- **WHEN** GET /api/compendium/search is called without q parameter
- **THEN** response code SHALL be 400

#### Scenario: Filter spells by level returns correct count
- **WHEN** GET /api/compendium/spells?level=0 is called
- **THEN** response code SHALL be 200 and all returned spells SHALL have level 0

#### Scenario: Filter spells by school returns correct school
- **WHEN** GET /api/compendium/spells?school=Evocation is called
- **THEN** response code SHALL be 200 and all returned spells SHALL have school "Evocation"

### Requirement: Conditions handler unit tests

The system SHALL have per-handler unit tests for conditions endpoints.

#### Scenario: Create condition validates duration
- **WHEN** POST /api/conditions is called with negative duration
- **THEN** response code SHALL be 400

#### Scenario: Tick conditions decrements duration correctly
- **WHEN** POST /api/conditions/tick is called with count=N
- **THEN** conditions with remaining_duration <= N SHALL be removed

#### Scenario: Get condition types returns known types
- **WHEN** GET /api/conditions/types is called
- **THEN** response code SHALL be 200 and body SHALL contain at least: blinded, charmed, deafened, frightened, poisoned, prone, restrained, stunned, unconscious

### Requirement: Campaign handler unit tests

The system SHALL have per-handler unit tests for campaign endpoints.

#### Scenario: Create campaign validates name
- **WHEN** POST /api/campaigns is called with empty name
- **THEN** response code SHALL be 400

#### Scenario: Non-owner non-admin cannot update campaign
- **WHEN** PUT /api/campaigns/:id is called by a user who is neither owner nor admin
- **THEN** response code SHALL be 403

#### Scenario: Add member validates user exists
- **WHEN** POST /api/campaigns/:id/members is called with non-existent username
- **THEN** response code SHALL be 404

#### Scenario: Delete campaign unassigns characters
- **WHEN** DELETE /api/campaigns/:id is called for a campaign with assigned characters
- **THEN** all character campaign_ids SHALL be set to NULL

### Requirement: Encounters handler unit tests

The system SHALL have per-handler unit tests for encounter endpoints.

#### Scenario: Calculate XP with empty party returns 0
- **WHEN** POST /api/encounters/calculate-xp is called with empty party_levels
- **THEN** response code SHALL be 200 with total_xp=0

#### Scenario: Calculate XP with known monsters returns correct total
- **WHEN** POST /api/encounters/calculate-xp is called with 3 goblins (CR 1/4, 50 XP each) against 4 level-1 characters
- **THEN** total_xp SHALL be 150

### Requirement: Dice handler unit tests

The system SHALL have per-handler unit tests for dice endpoints.

#### Scenario: Roll with invalid expression returns 400
- **WHEN** POST /api/roll is called with empty expression
- **THEN** response code SHALL be 400

#### Scenario: Roll with advantage returns two rolls
- **WHEN** POST /api/roll is called with expression="1d20" and advantage="advantage"
- **THEN** response SHALL contain 2 rolls in the breakdown and total SHALL equal the higher roll

#### Scenario: Roll with disadvantage returns two rolls
- **WHEN** POST /api/roll is called with expression="1d20" and advantage="disadvantage"
- **THEN** response SHALL contain 2 rolls and total SHALL equal the lower roll

### Requirement: Generators handler unit tests

The system SHALL have per-handler unit tests for generator endpoints.

#### Scenario: Generate NPC returns all required fields
- **WHEN** GET /api/generate/npc is called
- **THEN** response SHALL contain name, race, class, personality, appearance fields

#### Scenario: Generate name with race filter respects filter
- **WHEN** GET /api/generate/name?race=dwarf is called
- **THEN** response SHALL be 200

### Requirement: Notes handler unit tests

The system SHALL have per-handler unit tests for notes endpoints.

#### Scenario: Create note validates character_id
- **WHEN** POST /api/notes is called without character_id
- **THEN** response code SHALL be 400

#### Scenario: Create note with dm visibility includes DM only icon
- **WHEN** POST /api/notes is called with visibility="dm"
- **THEN** response body SHALL indicate DM-only visibility

### Requirement: Maps handler unit tests

The system SHALL have per-handler unit tests for maps endpoints.

#### Scenario: Activate non-existent map returns 200
- **WHEN** POST /api/campaigns/:id/maps/99999/activate is called
- **THEN** response code SHALL be 200 (idempotent)

#### Scenario: Get active map with no maps returns null
- **WHEN** GET /api/campaigns/:id/maps/active is called with no maps
- **THEN** response SHALL contain null map field

### Requirement: Factions handler unit tests

The system SHALL have per-handler unit tests for faction endpoints.

#### Scenario: Set reputation validates standing range
- **WHEN** POST /api/faction-reputation is called with standing > 100 or < -100
- **THEN** response code SHALL be 400

#### Scenario: Create faction validates name
- **WHEN** POST /api/factions is called with empty name
- **THEN** response code SHALL be 400

### Requirement: Share handler unit tests

The system SHALL have per-handler unit tests for share endpoints.

#### Scenario: Create share link returns token and URL
- **WHEN** POST /api/share is called with valid entity_type and entity_id
- **THEN** response SHALL contain non-empty token and url fields

#### Scenario: Get shared character via public route returns correct data
- **WHEN** GET /api/share/:token is called with valid character share token
- **THEN** response SHALL contain character name, race, class, level

#### Scenario: Invalid token returns 404
- **WHEN** GET /api/share/invalidtoken is called
- **THEN** response code SHALL be 404

### Requirement: Homebrew handler unit tests

The system SHALL have per-handler unit tests for homebrew endpoints.

#### Scenario: Create homebrew entry validates type
- **WHEN** GET /api/homebrew/invalid is called
- **THEN** response code SHALL be 400

#### Scenario: Create homebrew race sets source to homebrew
- **WHEN** POST /api/homebrew/races is called
- **THEN** response source field SHALL be "homebrew"
