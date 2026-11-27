# compendium-global-search Specification

## Purpose
Unified search across all compendium content types (spells, equipment, monsters, races, classes, feats, backgrounds) from the main compendium view.

## Requirements

### Requirement: Global search bar in compendium view

The compendium view SHALL display a search input at the top of the compendium container, positioned above the tab navigation and all tab content panels.

#### Scenario: Search bar is visible on compendium load
- **WHEN** the user opens the compendium view
- **THEN** a search input with placeholder text "Search compendium..." is visible above the tabs
- **AND** the search bar has `hx-get="/htmx/compendium/search"` and `hx-trigger="keyup changed delay:300ms"` attributes

#### Scenario: Search bar stays visible during tab navigation
- **WHEN** the user switches between compendium tabs
- **THEN** the search bar remains visible at the top of the compendium view

### Requirement: Cross-type search across all compendium content

The search SHALL query spells, equipment, monsters, races, classes, feats, and backgrounds simultaneously and return results grouped by type.

#### Scenario: Search finds matches in multiple types
- **WHEN** the user types "fire" in the global search bar
- **THEN** results are returned from spells (e.g., "Fire Bolt"), equipment (e.g., "Fire Breath Potion"), and monsters (e.g., "Fire Elemental"), grouped by type

#### Scenario: Search with no matches shows empty state
- **WHEN** the user types a query that matches nothing
- **THEN** a "No results found" message is displayed

### Requirement: Results grouped by content type

Search results SHALL be displayed in a results panel, grouped by content type with type headings. Each result SHALL show its name and relevant metadata.

#### Scenario: Results display with type headings
- **WHEN** search results are returned
- **THEN** results are rendered under headings like "Spells", "Monsters", "Equipment", "Races", "Classes", "Feats", "Backgrounds"

#### Scenario: Results show type-specific metadata
- **WHEN** a spell appears in search results
- **THEN** the spell result shows its level and school as metadata badges
- **WHEN** a monster appears in search results
- **THEN** the monster result shows its CR and type as metadata badges
- **WHEN** an equipment item appears in search results
- **THEN** the equipment result shows its category as metadata

### Requirement: Result click navigates to tab

Clicking a search result SHALL switch the compendium to the corresponding tab via `loadCompendiumTab()`.

#### Scenario: Clicking a spell result opens the Spells tab
- **WHEN** the user clicks a spell result
- **THEN** the compendium switches to the Spells tab

#### Scenario: Clicking a monster result opens the Monsters tab
- **WHEN** the user clicks a monster result
- **THEN** the compendium switches to the Monsters tab

### Requirement: Search includes backgrounds

The search endpoint SHALL include `compendium_backgrounds` in its search results.

#### Scenario: Backgrounds appear in search results
- **WHEN** the user searches for a term matching a background name
- **THEN** the background appears in the results under a "Backgrounds" heading

### Requirement: Search includes classes

The search endpoint SHALL include `compendium_classes` in its search results.

#### Scenario: Classes appear in search results
- **WHEN** the user searches for a term matching a class name
- **THEN** the class appears in the results under a "Classes" heading

### Requirement: Results have a maximum per type

To prevent overwhelming results, each type SHALL return at most 10 results.

#### Scenario: More than 10 matches for a type
- **WHEN** a query matches more than 10 items of a single type
- **THEN** only the first 10 are shown, with no pagination indicator (it's a quick search, not a browser)
