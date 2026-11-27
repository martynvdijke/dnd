# compendium-location-linking Specification

## Purpose
Link character locations from compendium or global location pool.

## Requirements

### Requirement: Global location search when linking

The location linking flow for characters SHALL allow searching across all locations (not just the current user's) when linking a location to a character. The location selector in the "Link Location" modal SHALL include a search input that queries locations via the API.

#### Scenario: Search all locations when linking

- **WHEN** the user opens the "Link Location" modal for a character
- **THEN** the location selector SHALL include a search input that finds locations across all users (DM-visible)

#### Scenario: Link location to character

- **WHEN** the user selects a location from search results and confirms
- **THEN** a new `CharacterLocation` record SHALL be created linking the character and the selected location with the chosen relationship and notes

### Requirement: Location search from existing linked data

The locations tab on the character sheet SHALL show all linked locations with their details. The "Link Location" button SHALL be accessible from the locations tab header and SHALL open the searchable location picker.

#### Scenario: Link button in locations tab

- **WHEN** the user is on the locations tab of a character sheet
- **THEN** a "Link Location" button SHALL be visible in the tab header, opening the global location search modal
