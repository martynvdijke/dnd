# e2e-monster-management Specification

## Purpose
Playwright e2e tests for monster library, campaign monster roster, API import flow.

## Requirements

### Requirement: Monster library e2e coverage
The Playwright e2e tests SHALL verify the custom monster library CRUD and compendium integration.

#### Scenario: Navigate to monster library
- **WHEN** authenticated DM user navigates to the monster library section
- **THEN** the library SHALL show existing custom monsters and a "Create Monster" button

#### Scenario: Create a custom monster
- **WHEN** DM user fills the monster creation form with stats/abilities and submits
- **THEN** the new monster SHALL appear in the library listing

#### Scenario: Edit a custom monster
- **WHEN** DM user edits an existing library monster
- **THEN** the changes SHALL be saved and reflected in the listing

#### Scenario: Delete a custom monster
- **WHEN** DM user deletes a library monster and confirms
- **THEN** the monster SHALL be removed from the listing

#### Scenario: Clone from compendium
- **WHEN** DM user clicks "Clone from Compendium" and selects a monster
- **THEN** the creation form SHALL be pre-filled with the compendium monster data

### Requirement: Campaign monster roster e2e coverage
The Playwright e2e tests SHALL verify the campaign monster roster feature.

#### Scenario: View campaign monster roster
- **WHEN** DM user navigates to a campaign's monster roster section
- **THEN** the roster SHALL list monsters assigned to that campaign

#### Scenario: Add monster to campaign roster
- **WHEN** DM user clicks "Add Monster" and selects from compendium/library
- **THEN** the monster SHALL appear in the campaign's roster

#### Scenario: Remove monster from campaign roster
- **WHEN** DM user removes a monster from the roster and confirms
- **THEN** the monster SHALL be removed

### Requirement: Monster API import e2e coverage
The Playwright e2e tests SHALL verify importing monsters from dnd5eapi.co.

#### Scenario: Open API import modal
- **WHEN** DM user clicks "Import from API" in the compendium browser
- **THEN** an import modal SHALL appear with search input

#### Scenario: Search and import monster from API
- **WHEN** DM user searches for a monster name and clicks import
- **THEN** the monster SHALL be imported and appear in the compendium monsters list

#### Scenario: Handle duplicate monster import
- **WHEN** DM user tries to import a monster that already exists in the compendium
- **THEN** a duplicate confirmation dialog SHALL appear
