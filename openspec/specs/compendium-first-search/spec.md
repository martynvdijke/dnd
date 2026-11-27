# compendium-first-search Specification

## Purpose
Unified modal-based compendium search UI that appears when adding any entity to a one-shot, encounter, or adventure. Supports searching by name across all schema types. Selecting an entry links it by ID. Fallback custom creation available.

## Requirements

### Requirement: Compendium search modal is the default for adding entities

When a DM or user clicks "Add Monster/Spell/Item/Background/Feat/Race/Class" in any one-shot, encounter, adventure, or character sheet context, the system SHALL first present a compendium search interface rather than a custom creation form.

#### Scenario: Clicking "Add Monster" in an encounter shows compendium search
- **WHEN** the DM clicks "Add Monster" in an encounter view
- **THEN** a modal opens showing a search input and a scrollable list of compendium monsters
- **AND** the modal is titled "Add from Compendium"

#### Scenario: Clicking "Add Spell" in character sheet shows compendium search
- **WHEN** the user clicks "Add Spell" on their character sheet
- **THEN** a modal opens showing a search input with compendium spells
- **AND** "Create Custom Spell" is available as a fallback option

#### Scenario: All entity types are covered
- **WHEN** any "Add" action for monsters, spells, items, backgrounds, feats, races, or classes is triggered
- **THEN** the system SHALL show the compendium search modal appropriate for that entity type

### Requirement: Interactive search with type filtering

The compendium search modal SHALL provide real-time search as the user types, filtering results by entity type (schema) when applicable, and displaying relevant attributes.

#### Scenario: Search by name
- **WHEN** the user types "goblin" in the search input
- **THEN** the system queries `/api/admin/compendium-search?q=goblin` or equivalent endpoint
- **AND** displays matching results within 300ms of the user stopping typing
- **AND** results show the entity name, type badge, and key attributes (CR for monsters, level+school for spells, category+weight for items)

#### Scenario: Filter by entity type
- **WHEN** the modal is opened for a specific entity type (e.g., monsters)
- **THEN** the search query is scoped to that type
- **AND** results from other types are not shown

#### Scenario: No results found
- **WHEN** the search returns no results
- **THEN** the modal displays "No compendium entries found" with a "Create Custom" button

### Requirement: Select compendium entry links it by ID

When the user selects a compendium entry from the search results, the system SHALL create a link (foreign key reference) rather than duplicating the compendium data.

#### Scenario: Selecting a compendium monster for an encounter
- **WHEN** the user clicks a compendium monster in the search results
- **THEN** a POST request is sent to `/api/encounters/:eid/import/compendium` or equivalent with the compendium entry ID
- **AND** the encounter detail view refreshes showing the new monster
- **AND** a success toast is displayed

#### Scenario: Selecting a compendium spell for a character
- **WHEN** the user clicks a compendium spell in the search results
- **THEN** the spell is added to the character's known spells with a reference to the compendium entry
- **AND** the spell list refreshes

### Requirement: Fallback custom creation

The compendium search modal SHALL provide an explicit "Create Custom" option that opens the existing custom creation form, allowing the user to manually enter all entity details.

#### Scenario: Creating a custom monster from the compendium modal
- **WHEN** the user clicks "Create Custom Monster" in the compendium search modal
- **THEN** the compendium modal closes
- **AND** the existing custom monster creation form opens (same fields as before: name, AC, HP, CR, stats, etc.)
- **AND** the custom monster is added to the encounter/one-shot without a compendium link

#### Scenario: Custom creation preserves compendium search state
- **WHEN** the user finishes custom creation and closes the modal
- **THEN** the parent view refreshes to show the newly created entity alongside any compendium-linked entities

### Requirement: Reusable search component

The compendium search modal SHALL be implemented as a reusable function that accepts configuration for entity type, context (encounter/oneshot/character), and a callback for selection handling.

#### Scenario: Same modal function used for all add contexts
- **WHEN** "Add Monster" is clicked in an encounter, one-shot, or adventure
- **THEN** the same compendiumSearchModal() function is called with different parameters
- **AND** the modal behaves consistently across all contexts

### Requirement: DM-scoped search access

The compendium search endpoint SHALL be accessible to DMs (not just admins) so that DM-facing one-shot and encounter modals can use it.

#### Scenario: DM searches compendium from one-shot modal
- **WHEN** a DM opens the compendium search modal from a one-shot
- **THEN** the search request is made to a DM-accessible endpoint (either /api/dm/compendium-search or the admin endpoint with DM access added)
- **AND** results are returned successfully

### Requirement: Legacy admin compendium tab removed

The admin panel SHALL have only one compendium management interface: the unified schema-based compendium. The old legacy compendium tab SHALL be removed.

#### Scenario: Admin opens compendium management
- **WHEN** an admin clicks "Compendium" in admin settings
- **THEN** only the unified schema-based compendium UI is shown
- **AND** the legacy compendium_* table browser is no longer accessible

#### Scenario: Legacy migration status removed
- **WHEN** an admin views the unified compendium
- **THEN** there is no "Migrate Legacy" prompt or legacy migration status check
- **AND** the admin can manage all compendium entries through the schema interface
