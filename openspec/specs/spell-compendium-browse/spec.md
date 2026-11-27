# spell-compendium-browse Specification

## Purpose
Dedicated spell browsing page with search, class/level/school filtering, and sortable results.

## Requirements

### Requirement: Spell browser renders searchable list

The system SHALL provide a dedicated spell browsing interface within the compendium spells tab that displays spells from the `compendium_spells` table in a searchable, filterable list.

#### Scenario: Default load shows all spells
- **WHEN** a user navigates to the compendium spells tab
- **THEN** the system SHALL display a searchable list of all spells ordered by level then name
- **AND** the list SHALL show spell name, level, school, casting time, range, and duration for each spell

#### Scenario: Empty state when no spells exist
- **WHEN** the `compendium_spells` table is empty
- **THEN** the system SHALL display an empty state message indicating no spells are available

### Requirement: Real-time search by name

The system SHALL filter the spell list in real-time as the user types a search query, with a debounce of 300ms.

#### Scenario: Search matches spells
- **WHEN** a user types "fire" in the search input
- **THEN** the system SHALL display only spells whose name contains "fire"
- **AND** results SHALL update within 500ms of the user stopping typing

#### Scenario: Search with no matches
- **WHEN** a user types "zzznotaspell" in the search input
- **THEN** the system SHALL display a "No spells found" message

#### Scenario: Empty search restores full list
- **WHEN** a user clears the search input
- **THEN** the system SHALL restore the full unfiltered spell list

### Requirement: Filter by class

The system SHALL allow filtering the spell list by character class (e.g., Wizard, Cleric, Druid).

#### Scenario: Class filter applied
- **WHEN** a user selects "Wizard" from the class filter dropdown
- **THEN** the system SHALL display only spells that include "Wizard" in their classes field

#### Scenario: Multiple class filters
- **WHEN** a user selects both "Wizard" and "Cleric" from the class filter
- **THEN** the system SHALL display spells available to either class (OR logic)

### Requirement: Filter by school

The system SHALL allow filtering the spell list by school of magic (e.g., Evocation, Necromancy, Illusion).

#### Scenario: School filter applied
- **WHEN** a user selects "Evocation" from the school filter
- **THEN** the system SHALL display only spells with school "Evocation"

### Requirement: Filter by level

The system SHALL allow filtering the spell list by spell level, supporting single level and level range selection.

#### Scenario: Single level filter
- **WHEN** a user selects level "3" from the level filter
- **THEN** the system SHALL display only spells of level 3

#### Scenario: Level range filter
- **WHEN** a user selects the "1st-3rd" level range filter
- **THEN** the system SHALL display only spells with level between 1 and 3 (inclusive)

#### Scenario: Cantrip filter
- **WHEN** a user selects "Cantrip" from the level filter
- **THEN** the system SHALL display only spells with level 0

### Requirement: Combined filters work together

The system SHALL support combining any of name search, class filter, school filter, and level filter simultaneously.

#### Scenario: All filters combined
- **WHEN** a user searches "fire", selects "Wizard" class, "Evocation" school, and level range "1st-3rd"
- **THEN** the system SHALL display only spells matching all criteria

### Requirement: Pagination for large result sets

The system SHALL paginate results, showing up to 50 spells per page, with a "Load More" or page navigation.

#### Scenario: Results exceed one page
- **WHEN** more than 50 spells match the current filters
- **THEN** the system SHALL show the first 50 results
- **AND** provide a way to load or navigate to the next page

### Requirement: Import source visual indicator

The system SHALL display a visual indicator showing the import source for each spell (SRD, homebrew, imported, etc.).

#### Scenario: SRD spell indicated
- **WHEN** a spell has `source: "srd"` and `system: "dnd5e"`
- **THEN** the system SHALL show an "SRD" badge on the spell entry

#### Scenario: Homebrew spell indicated
- **WHEN** a spell has `source: "homebrew"` or a non-empty `publisher` field
- **THEN** the system SHALL show a "Homebrew" badge or the publisher name on the spell entry
