# campaign-monster-list Specification

## Purpose
Per-campaign curated monster list (subset of compendium + custom) for quick encounter building.

## Requirements

### Requirement: Campaign monster roster table
The system SHALL have a `campaign_monster_roster` table linking campaigns to selected compendium monsters for curated quick-access lists.

#### Scenario: Table exists after migration
- **WHEN** migrations run
- **THEN** the `campaign_monster_roster` table is created with columns: `id`, `campaign_id`, `compendium_monster_id`, `label`, `sort_order`, `created_at`
- **AND** foreign key constraints reference `campaigns(id)` and `compendium_monsters(id)` with CASCADE delete

### Requirement: Add monster to campaign roster
The system SHALL allow adding compendium monsters to a campaign's curated roster.

#### Scenario: Add from compendium picker
- **WHEN** a DM views a campaign and opens the monster roster section
- **THEN** they see an "Add Monster" button that opens the compendium picker
- **AND** selecting a monster from the compendium adds it to the campaign roster
- **AND** the roster section updates to show the new entry

#### Scenario: Add from monster detail
- **WHEN** a DM views a compendium monster detail
- **AND** they are viewing it in the context of a campaign (campaign_id is known)
- **THEN** an "Add to Campaign Roster" button is shown
- **AND** clicking it adds the monster to that campaign's roster

### Requirement: Campaign roster display
The system SHALL display the campaign's monster roster as a filterable list.

#### Scenario: Roster section in campaign view
- **WHEN** a DM views a campaign detail page
- **THEN** the "Monster Roster" section shows all monsters added to the campaign
- **AND** each entry shows name, type, CR, and a custom label (if set)
- **AND** entries are sorted by `sort_order` then name
- **AND** the user can search/filter within the roster

#### Scenario: Remove from roster
- **WHEN** a DM clicks "Remove" on a roster entry
- **THEN** a confirmation is shown
- **AND** on confirmation, the entry is deleted from `campaign_monster_roster`
- **AND** the compendium monster itself is not affected

### Requirement: Roster used in encounter monster picker
The campaign monster roster SHALL be available as a filterable source when adding monsters to encounters in that campaign.

#### Scenario: Campaign monsters tab in picker
- **WHEN** a DM adds monsters to an encounter in a campaign with a roster
- **THEN** the monster picker shows a "Campaign Roster" tab listing rostered monsters
- **AND** selecting one adds it to the encounter with full stat data
- **AND** the tab is disabled/hidden if the campaign has no roster entries
