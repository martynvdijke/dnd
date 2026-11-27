# sheet-tab-priority Specification

## Purpose
Reorder character sheet tabs by gameplay priority; remove campaign-level tabs.

## Requirements

### Requirement: Tab bar reordered by priority

The character sheet tab bar SHALL display tabs in the following priority order: Stats, Combat, Spells, Inventory, Features, Feats, Companions, Crafting, Journal, Notes, Details, Dice, Party. The lowest-priority (campaign-level) tabs — Locations, NPCs, Sessions, Quests, Graph, Analytics — SHALL NOT appear in the character sheet tab bar.

#### Scenario: New tab order rendered

- **WHEN** a character sheet is opened
- **THEN** the tab bar SHALL show tabs in the order: stats, combat, spells, inventory, features, feats, companions, crafting, journal, notes, details, dice, party

#### Scenario: Campaign tabs removed from sheet

- **WHEN** a character sheet is opened
- **THEN** the tab bar SHALL NOT include locations, npcs, sessions, quests, graph, or analytics tabs

### Requirement: Party tab in character sheet

The character sheet tab bar SHALL include a "Party" tab as the last item. When clicked, it SHALL navigate to the party view (`showView('party')`).

#### Scenario: Party tab navigates to party view

- **WHEN** the user clicks the "Party" tab in the character sheet tab bar
- **THEN** the view SHALL switch to the party view (`#partyView`)

### Requirement: Shared sections array

The `sections` array defining tab order SHALL be defined in a single shared module (e.g., `ts/lib/tabs.ts`) and imported by both `ts/app.ts` and `ts/lib/shortcuts.ts`. Both files SHALL read from the same source.

#### Scenario: app.ts imports shared sections

- **WHEN** `ts/app.ts` renders the tab bar
- **THEN** it SHALL import the `sections` array from the shared module

#### Scenario: shortcuts.ts imports shared sections

- **WHEN** `ts/lib/shortcuts.ts` handles keyboard navigation
- **THEN** it SHALL import the `sections` array from the shared module

### Requirement: Keyboard shortcuts updated

The keyboard shortcut system SHALL be updated to match the new tab count and order. Keys 1-9 SHALL cycle through the first 9 tabs in the new order.

#### Scenario: Shortcut keys match new tab order

- **WHEN** the user presses key 1 on the character sheet
- **THEN** it SHALL switch to the first tab (stats)

### Requirement: Tab sections moved from sheet to party HTML

The HTML sections for locations, NPCs, sessions, quests, graph, and analytics SHALL be removed from `#sheetView` in `static/app.html` and rendered dynamically in the party view instead.

#### Scenario: Location section removed from sheet HTML

- **WHEN** `static/app.html` is loaded
- **THEN** `#sheetView` SHALL NOT contain `#locationsSection`, `#npcsSection`, `#sessionsSection`, `#questsSection`, `#graphSection`, or `#analyticsSection`
