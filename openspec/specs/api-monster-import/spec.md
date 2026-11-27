# api-monster-import Specification

## Purpose
UI for searching, previewing, and importing monsters from the D&D 5e API (dnd5eapi.co) into the local compendium.

## Requirements

### Requirement: Monster search via D&D 5e API
The system SHALL provide a UI to search the D&D 5e API (dnd5eapi.co) for monsters, using the existing `FetchFromDnDApi(category="monsters")` endpoint.

#### Scenario: Search modal opens from compendium browser
- **WHEN** a user clicks "Import from API" button in the compendium browser
- **THEN** a modal is displayed with a search input and "Search" button
- **AND** the modal shows instructions about the source API

#### Scenario: Search returns monster results
- **WHEN** a user types a monster name and clicks Search
- **THEN** the system calls `GET /api/compendium/search?category=monsters&q=<query>`
- **AND** results are displayed as a list with monster name, type, and CR
- **AND** each result has an "Import" button
- **AND** a "No results found" message is shown if the API returns nothing

#### Scenario: Search errors are displayed
- **WHEN** the API request fails or times out
- **THEN** an error message is displayed in the modal
- **AND** the user can retry the search

### Requirement: Import monster from API to compendium
The system SHALL import a monster from the D&D 5e API into the local `compendium_monsters` table.

#### Scenario: Import single monster
- **WHEN** a user clicks "Import" on an API search result
- **THEN** the system fetches the full monster detail from dnd5eapi.co
- **AND** maps all fields (name, type, size, ac, hp, stats, cr, saves, skills, damage resistances/immunities, condition immunities, senses, languages, special abilities, actions, legendary actions) to the `compendium_monsters` table
- **AND** sets `source` to `"dnd5eapi"` and `is_full` to `1`
- **AND** the monster is immediately available in the compendium browser

#### Scenario: Duplicate monster name detection
- **WHEN** a monster with the same name already exists in `compendium_monsters`
- **THEN** the system shows a confirmation dialog: "A monster named 'X' already exists. Import anyway?"
- **AND** on confirmation, imports with the source set to `"dnd5eapi"` to differentiate

#### Scenario: Import confirmation
- **WHEN** the import succeeds
- **THEN** a success toast/notification is shown
- **AND** the modal updates to show the monster as imported

### Requirement: API import error handling
The system SHALL handle API import errors gracefully.

#### Scenario: API detail fetch fails
- **WHEN** fetching the monster detail from dnd5eapi.co fails
- **THEN** an error message is shown: "Failed to import: <error details>"
- **AND** no changes are made to the database

#### Scenario: Field mapping failure
- **WHEN** the API response has an unexpected structure for a required field
- **THEN** the import attempts to use sensible defaults for missing fields
- **AND** the monster is still imported with available data
- **AND** a warning note is logged about the missing fields
