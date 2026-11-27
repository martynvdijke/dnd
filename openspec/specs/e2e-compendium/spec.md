# e2e-compendium Specification

## Purpose
Playwright e2e tests for compendium browsing, searching, and admin CRUD (schemas, entries, bulk ops).

## Requirements

### Requirement: Compendium browser e2e coverage
The Playwright e2e tests SHALL verify the compendium browsing experience for all authenticated users.

#### Scenario: Navigate to compendium browser
- **WHEN** authenticated user clicks "Compendium" in the nav
- **THEN** the compendium browser page SHALL load with a category picker showing available schema types

#### Scenario: Browse compendium by category
- **WHEN** user selects a compendium category (e.g., "Monsters", "Spells", "Races")
- **THEN** entries for that category SHALL be displayed in a list

#### Scenario: View compendium entry detail
- **WHEN** user clicks a compendium entry in the list
- **THEN** the entry detail view SHALL show all schema-defined fields

#### Scenario: Search compendium
- **WHEN** user types a search query in the compendium search box
- **THEN** matching entries SHALL be displayed across all categories

### Requirement: Compendium admin CRUD e2e coverage
The Playwright e2e tests SHALL verify the admin compendium management UI.

#### Scenario: Navigate to admin compendium
- **WHEN** authenticated admin user navigates to the admin section
- **THEN** the admin compendium management interface SHALL be accessible

#### Scenario: Create a new compendium entry
- **WHEN** admin clicks "Add Entry" and fills the schema-driven form
- **THEN** a new entry SHALL be created and appear in the list

#### Scenario: Edit a compendium entry
- **WHEN** admin clicks "Edit" on an entry and modifies fields
- **THEN** the entry SHALL be updated with the new values

#### Scenario: Delete a compendium entry
- **WHEN** admin clicks "Delete" on an entry and confirms
- **THEN** the entry SHALL be removed from the list

#### Scenario: Bulk select and delete entries
- **WHEN** admin selects multiple entries via checkboxes and clicks "Delete Selected"
- **THEN** all selected entries SHALL be deleted

#### Scenario: Bulk select and update entries
- **WHEN** admin selects multiple entries and clicks "Edit Selected"
- **THEN** a batch update modal SHALL appear allowing field value changes

### Requirement: Compendium schema management e2e coverage
The Playwright e2e tests SHALL verify the dynamic schema management interface.

#### Scenario: View compendium schemas
- **WHEN** admin navigates to schema management in the admin panel
- **THEN** all existing schemas SHALL be listed with their entry counts

#### Scenario: Create a new schema
- **WHEN** admin clicks "New Schema" and provides a name and field definitions
- **THEN** the schema SHALL be created and appear in the list

#### Scenario: Edit a schema
- **WHEN** admin edits an existing schema's name or field definitions
- **THEN** the schema SHALL be updated
