# character-sheet-ux Specification

## Purpose
Progressive disclosure for character sheets — collapsible sections with counts, persistent quick-action row, session-mode-aware layout.

## Requirements

### Requirement: Accordion Layout for Character Sections
The system SHALL render character sheet sections as a vertically scrolling accordion using native HTML `<details>/<summary>` elements. Each section SHALL be independently collapsible. The current horizontal tab bar SHALL be replaced by this accordion layout.

#### Scenario: Character sheet renders as accordion
- **WHEN** a user opens a character sheet
- **THEN** all character sections (Stats, Combat, Spells, Inventory, Features, Feats, Companions, Crafting, Locations, NPCs, Sessions, Quests, Journal, Analytics, Details, Notes, Dice) SHALL be rendered as collapsible accordion sections in a vertical list

#### Scenario: Section collapse and expand
- **WHEN** a user clicks/taps a section header
- **THEN** the section SHALL toggle between expanded (content visible) and collapsed (content hidden) states

#### Scenario: Multiple sections can be open simultaneously
- **WHEN** a user expands one section while another is already expanded
- **THEN** both sections SHALL remain expanded (accordion is non-exclusive)

### Requirement: Section Header Summary Counts
Each accordion section header SHALL display a summary count or key metric relevant to that section, allowing users to gauge content without expanding.

#### Scenario: Stats section shows ability count
- **WHEN** the character sheet is rendered
- **THEN** the Stats section header SHALL display the character's level and class (e.g., "Stats — Level 5 Fighter")

#### Scenario: Spells section shows prepared count
- **WHEN** the character has prepared spells
- **THEN** the Spells section header SHALL display the count of prepared spells (e.g., "Spells (8 prepared)")

#### Scenario: Inventory section shows item count and weight
- **WHEN** the character has inventory items
- **THEN** the Inventory section header SHALL display item count and total weight (e.g., "Inventory (23 items, 45 lb)")

#### Scenario: Combat section shows HP and AC
- **WHEN** the character sheet is rendered
- **THEN** the Combat section header SHALL display current HP, max HP, and AC (e.g., "Combat — HP 45/52, AC 16")

### Requirement: Sticky Quick-Action Bar
The system SHALL display a persistent quick-action bar above the accordion sections when viewing a character sheet. The bar SHALL remain visible (sticky) as the user scrolls through sections.

#### Scenario: Quick-action bar is visible
- **WHEN** a user views a character sheet
- **THEN** a quick-action bar SHALL be displayed above the accordion sections containing buttons for: Roll Initiative, Short Rest, Long Rest, and Level Up

#### Scenario: Quick-action bar stays visible on scroll
- **WHEN** a user scrolls down through the character sheet sections
- **THEN** the quick-action bar SHALL remain fixed at the top of the sheet view area using `position: sticky`

#### Scenario: Quick-action buttons trigger existing actions
- **WHEN** a user clicks "Roll Initiative" in the quick-action bar
- **THEN** the system SHALL execute the same initiative roll action as the existing dice roller

#### Scenario: Quick-action bar adapts to session mode
- **WHEN** Session Mode is active
- **THEN** the quick-action bar SHALL show only: Damage, Heal, and Conditions (replacing the between-session actions)

### Requirement: Session Mode Character Sheet Filtering
When Session Mode is active, the character sheet SHALL display only combat-relevant sections. All other sections SHALL be hidden.

#### Scenario: Only combat section visible in Session Mode
- **WHEN** Session Mode is active and the user views their character sheet
- **THEN** only the Combat section SHALL be visible and expanded; all other sections SHALL be hidden via CSS

#### Scenario: Exiting Session Mode restores all sections
- **WHEN** Session Mode is deactivated while viewing a character sheet
- **THEN** all accordion sections SHALL become visible again with their previous expand/collapse states restored

### Requirement: HTMX-Loaded Accordion Sections
Sections that currently load content via HTMX (Spells, Features, Feats, Companions, Crafting, Notes) SHALL continue to use HTMX for content loading within the accordion pattern.

#### Scenario: HTMX section loads on first expand
- **WHEN** a user expands an HTMX-loaded section (e.g., Spells) for the first time
- **THEN** the section content SHALL be fetched via the existing `hx-get` endpoint and rendered inside the accordion body

#### Scenario: HTMX section caches after first load
- **WHEN** a user collapses and re-expands an HTMX-loaded section
- **THEN** the previously loaded content SHALL be displayed immediately without a new network request
