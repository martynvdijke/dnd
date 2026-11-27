# query-index-optimization Specification

## Purpose
Composite covering indexes for hot query paths, analyzed via EXPLAIN QUERY PLAN.

## Requirements

### Requirement: Add composite index on characters (user_id, name, level)
The system SHALL create a composite index `idx_characters_user_name_level` on `characters(user_id, name, level)` to accelerate character listing queries.

#### Scenario: Index created via migration
- **WHEN** the database migration runs
- **THEN** the index `idx_characters_user_name_level` SHALL exist on `characters(user_id, name, level)`
- **AND** the migration SHALL use `CREATE INDEX IF NOT EXISTS`

### Requirement: Add composite index on characters (campaign_id, name)
The system SHALL create a composite index `idx_characters_campaign_name` on `characters(campaign_id, name)` for campaign roster queries.

#### Scenario: Campaign roster index created
- **WHEN** the database migration runs
- **THEN** the index `idx_characters_campaign_name` SHALL exist on `characters(campaign_id, name)`

### Requirement: Add composite index on spells (character_id, level, name)
The system SHALL create a composite index `idx_spells_char_level_name` on `spells(character_id, level, name)` for spellbook queries sorted by level.

#### Scenario: Spellbook index created
- **WHEN** the database migration runs
- **THEN** the index `idx_spells_char_level_name` SHALL exist on `spells(character_id, level, name)`

### Requirement: Add composite index on inventory (character_id, category, name)
The system SHALL create a composite index `idx_inventory_char_category_name` on `inventory(character_id, category, name)` for inventory grouped by category.

#### Scenario: Inventory index created
- **WHEN** the database migration runs
- **THEN** the index `idx_inventory_char_category_name` SHALL exist on `inventory(character_id, category, name)`

### Requirement: Add composite index on sessions (character_id, session_date)
The system SHALL create a composite index `idx_sessions_char_date` on `sessions(character_id, session_date)` for session history sorted by date.

#### Scenario: Session history index created
- **WHEN** the database migration runs
- **THEN** the index `idx_sessions_char_date` SHALL exist on `sessions(character_id, session_date)`

### Requirement: Add composite index on quests (character_id, status, updated_at)
The system SHALL create a composite index `idx_quests_char_status_updated` on `quests(character_id, status, updated_at)` for active quest queries.

#### Scenario: Quests index created
- **WHEN** the database migration runs
- **THEN** the index `idx_quests_char_status_updated` SHALL exist on `quests(character_id, status, updated_at)`

### Requirement: Add composite index on journal (character_id, entry_date)
The system SHALL create a composite index `idx_journal_char_entry_date` on `journal(character_id, entry_date)` for journal entries sorted by date.

#### Scenario: Journal index created
- **WHEN** the database migration runs
- **THEN** the index `idx_journal_char_entry_date` SHALL exist on `journal(character_id, entry_date)`

### Requirement: Add composite index on combat_entries (campaign_id, turn_order)
The system SHALL create a composite index `idx_combat_campaign_turn` on `combat_entries(campaign_id, turn_order)` for combat tracker queries.

#### Scenario: Combat tracker index created
- **WHEN** the database migration runs
- **THEN** the index `idx_combat_campaign_turn` SHALL exist on `combat_entries(campaign_id, turn_order)`

### Requirement: Add composite index on dice_rolls (user_id, timestamp)
The system SHALL create a composite index `idx_dice_rolls_user_timestamp` on `dice_rolls(user_id, timestamp)` for dice roll history queries.

#### Scenario: Dice rolls index created
- **WHEN** the database migration runs
- **THEN** the index `idx_dice_rolls_user_timestamp` SHALL exist on `dice_rolls(user_id, timestamp)`

### Requirement: Add composite index on campaigns (user_id, name)
The system SHALL create a composite index `idx_campaigns_user_name` on `campaigns(user_id, name)` for campaign listing queries.

#### Scenario: Campaigns index created
- **WHEN** the database migration runs
- **THEN** the index `idx_campaigns_user_name` SHALL exist on `campaigns(user_id, name)`

### Requirement: Verify indexes with EXPLAIN QUERY PLAN
The system SHALL verify that each new index is used by the query planner for its intended query pattern by running EXPLAIN QUERY PLAN during testing.

#### Scenario: EXPLAIN QUERY PLAN confirms index usage
- **WHEN** test runs EXPLAIN QUERY PLAN on each hot path query
- **THEN** the result SHALL show the new index in the execution plan
- **AND** the `detail` column SHALL contain the expected index name
