# e2e-compendium-tests Specification

## Purpose
End-to-end tests in `tests/compendium.spec.ts` guard the compendium view's legacy tabs and dynamic schema tabs against regressions.

## Requirements

### Requirement: Legacy compendium tabs render after login

After login, the Compendium view SHALL render its legacy tabs (Races, Classes, Spells, Feats, Backgrounds, Equipment, Monsters) with content loading for each tab.

#### Scenario: Legacy tabs render

- **WHEN** a logged-in user navigates to the Compendium view
- **THEN** the legacy tabs render and are navigable

### Requirement: Dynamic schema tabs render entries

Schema entries created via the admin API SHALL appear as a dynamic schema tab below the legacy tabs, with entry cards showing a name and up to 3 key-value previews. When more than 20 entries exist, a "View All N entries" button SHALL appear; when no schema entries exist, the dynamic schema section SHALL be hidden. Active tab tracking SHALL highlight the currently selected tab.

#### Scenario: Dynamic schema tab appears with entries

- **WHEN** schema entries exist for a schema
- **THEN** the dynamic schema tab appears below the legacy tabs and its entry cards render name and up to 3 key-value previews

#### Scenario: View All button for large schemas

- **WHEN** a schema's `entry_count > 20`
- **THEN** a "View All N entries" button appears

#### Scenario: Empty schema section is hidden

- **WHEN** no schema entries exist
- **THEN** the dynamic schema section is hidden and only the legacy tabs are shown

#### Scenario: Active tab tracking

- **WHEN** the user clicks between legacy and schema tabs
- **THEN** the active tab is highlighted correctly
