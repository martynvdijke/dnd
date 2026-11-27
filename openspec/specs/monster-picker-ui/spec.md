# monster-picker-ui Specification

## Purpose
Unified monster picker component usable from one-shot monster sections and encounter forms, with search/filter and multi-select support.

## Requirements

### Requirement: Unified monster picker modal
The system SHALL provide a shared HTMX modal component for selecting and adding monsters to one-shots and encounters, replacing ad-hoc picker patterns.

#### Scenario: Picker opens in context
- **WHEN** a user clicks "From Compendium" or "From Library" in the one-shot monsters section, or "Add Monster" in an encounter
- **THEN** a modal opens titled "Pick a Monster"
- **AND** the modal accepts a context parameter (`oneshot_id`, `encounter_id`, or `campaign_id`)
- **AND** the appropriate "Add" action is performed when a monster is selected

#### Scenario: Three-tab layout
- **WHEN** the picker modal opens
- **THEN** it displays three tabs: "Compendium", "My Library", and (if campaign context) "Campaign Roster"
- **AND** the first tab (Compendium) is selected by default
- **AND** each tab shows its own content area and search bar

### Requirement: Compendium tab
The compendium tab SHALL display the full monster compendium with search and filter.

#### Scenario: Compendium tab functionality
- **WHEN** the Compendium tab is active
- **THEN** it shows a search input, CR filter, type filter, and a scrollable list of monsters
- **AND** the list shows each monster's name, type, size, CR, AC, and HP
- **AND** clicking a monster's row shows a quick stat preview
- **AND** clicking "Add" adds the monster to the target context (one-shot or encounter)

#### Scenario: Search and filter
- **WHEN** a user types in the search field or changes a filter
- **THEN** the monster list updates via HTMX without full page reload
- **AND** the search filters by monster name (partial match)
- **AND** the CR filter and type filter narrow results

### Requirement: Library tab
The library tab SHALL display the user's custom monsters from `monster_library`.

#### Scenario: Library tab functionality
- **WHEN** the Library tab is active
- **THEN** it shows the user's custom monsters in the same list format
- **AND** each entry shows an "Add" button
- **AND** an empty state with a "Create Monster" link if no library entries exist

### Requirement: Campaign roster tab
When a campaign context is provided, the picker SHALL show a campaign roster tab.

#### Scenario: Campaign roster tab
- **WHEN** the picker is opened from a campaign encounter
- **THEN** a "Campaign Roster" tab is available
- **AND** it lists monsters in the campaign's `campaign_monster_roster`
- **AND** the tab is hidden if the campaign has no roster entries

### Requirement: Add monster to one-shot
When the picker is opened in one-shot context, selecting a monster SHALL add it to the one-shot's monster list.

#### Scenario: Add to one-shot
- **WHEN** a user clicks "Add" on a monster in the picker with `oneshot_id` context
- **THEN** the system calls the existing import endpoint to add the monster to the one-shot
- **AND** a success indicator appears on the picked monster ("Added")
- **AND** the one-shot monsters section updates to show the new monster

### Requirement: Add monster to encounter
When the picker is opened in encounter context, selecting a monster SHALL add it to the encounter's monster list.

#### Scenario: Add to encounter
- **WHEN** a user clicks "Add" on a monster in the picker with `encounter_id` context
- **THEN** the system adds the monster to the encounter with count=1
- **AND** a success indicator appears on the picked monster ("Added")
- **AND** the encounter monsters section updates to show the new monster

### Requirement: Quick preview on hover/click
The picker SHALL show a quick stat preview when a monster row is clicked.

#### Scenario: Stat preview
- **WHEN** a user clicks on a monster row in the picker
- **THEN** a detail panel shows AC, HP, CR, six ability scores, saves, skills, and damage resistances
- **AND** the preview content is loaded via HTMX from the existing monster detail endpoint
- **AND** the "Add" button is available in both the list row and the preview panel
