# race-colors Specification

## Purpose
Assign and display custom colors for character races across the UI.

## Requirements

### Requirement: Race color storage

The system SHALL store a mapping of race names to hex color codes in a `race_colors` table.

#### Scenario: Store race color
- **WHEN** user assigns a color to a race name
- **THEN** the mapping is persisted in the `race_colors` table

#### Scenario: Update existing race color
- **WHEN** user changes the color for an existing race
- **THEN** the stored color is updated to the new value

#### Scenario: Remove race color
- **WHEN** user clears a race's color assignment
- **THEN** the mapping is removed from the `race_colors` table

### Requirement: Race color API

The system SHALL provide API endpoints to read and update the race color mapping.

#### Scenario: Get all race colors
- **WHEN** a GET request is made to `/api/race-colors`
- **THEN** the response returns a JSON object mapping race names to hex color strings (e.g., `{"Elf":"#8B5CF6","Dwarf":"#D97706"}`)

#### Scenario: Update race colors
- **WHEN** a PUT request is made to `/api/race-colors` with a JSON body of race-to-color mappings
- **THEN** the race colors are upserted in the database and the updated map is returned

### Requirement: Race color management UI

The system SHALL provide a user interface for viewing and editing race color assignments.

#### Scenario: View current race colors
- **WHEN** user navigates to the race colors settings page
- **THEN** all configured race colors are displayed as a list with race name and color swatch

#### Scenario: Discover races from characters
- **WHEN** user opens the race colors management UI
- **THEN** all unique race names from the user's existing characters are shown, even if no color is yet assigned

#### Scenario: Change a race color
- **WHEN** user clicks a color swatch or input for a race
- **THEN** a color picker appears allowing selection of a new color
- **WHEN** user selects a new color and saves
- **THEN** the color is stored and the display updates immediately

### Requirement: Race color display on campaign overview

The campaign overview dashboard SHALL display race colors as colored badges next to character names.

#### Scenario: Campaign overview shows race colors
- **WHEN** the campaign overview page loads
- **THEN** each character's race label is displayed with its assigned color as a background or dot indicator

#### Scenario: Uncolored race display
- **WHEN** a character's race has no color assigned in the race_colors table
- **THEN** the race label is displayed in the default muted text style (no color change)

### Requirement: Race color display in character lists

The system SHALL display race colors in character list views and companion list views.

#### Scenario: Character list shows race colors
- **WHEN** the character list view is loaded
- **THEN** each character's race is shown with its assigned color as a colored badge

#### Scenario: Companion list shows race colors
- **WHEN** the companion list view is loaded
- **THEN** each companion's race is shown with its assigned color as a colored badge
