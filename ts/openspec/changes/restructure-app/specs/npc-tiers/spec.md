## ADDED Requirements

### Requirement: Two-tier NPC system
The existing `npcs` table SHALL gain a `npc_type` column ("simple"|"full"). Simple NPCs use existing fields plus `backstory` field. Full NPCs get additional character-sheet-like fields.

#### Scenario: Create simple NPC (backstory only)
- **WHEN** a user sends POST to `/api/npcs` with `npc_type="simple"`, `name`, and `backstory`
- **THEN** a simple NPC SHALL be created with the given backstory

#### Scenario: Create full NPC (character sheet)
- **WHEN** a user sends POST to `/api/npcs` with `npc_type="full"` and full stats
- **THEN** a full NPC SHALL be created with all character-sheet fields

#### Scenario: List NPCs shows type
- **WHEN** a user sends GET to `/api/npcs`
- **THEN** each NPC SHALL include its `npc_type` field

#### Scenario: Update NPC tier
- **WHEN** a user sends PUT to `/api/npcs/:id` with `npc_type` change
- **THEN** the NPC's tier SHALL be updated (existing data preserved where applicable)

### Requirement: Full NPC sheet fields
Full NPCs SHALL support additional fields: `armor_class`, `speed`, `hit_dice`, `skills` (JSON), `saving_throws`, `damage_vulnerabilities`, `damage_resistances`, `damage_immunities`, `condition_immunities`, `senses`, `languages`, `challenge_rating`, `xp`, `traits` (JSON), `actions` (JSON), `legendary_actions` (JSON), `spellcasting` (JSON).

#### Scenario: Full NPC with traits and actions
- **WHEN** a user creates a full NPC with traits and actions
- **THEN** the traits and actions SHALL be stored and returned
