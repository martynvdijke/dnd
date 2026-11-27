# mobile-character-sheet Specification

## Purpose
Mobile-first responsive character sheet with card-based layout, large touch targets, and stepper controls for all numeric stats.

## Requirements

### Requirement: Character sheet renders in card-based layout on mobile
The system SHALL render the character sheet's stats, combat, and details sections as distinct visual cards with rounded corners, shadows, and section headings when the viewport is less than 768px wide.

#### Scenario: Mobile layout uses cards
- **WHEN** the character sheet is displayed on a viewport narrower than 768px
- **THEN** each section (ability scores, combat stats, skills, proficiencies, details fields) SHALL appear as a separate card with a heading, background fill, rounded corners, and padding

#### Scenario: Desktop layout uses existing grid
- **WHEN** the character sheet is displayed on a viewport 768px or wider
- **THEN** the existing Bootstrap grid layout SHALL be preserved without card wrappers

### Requirement: Touch targets are at least 44x44px
All interactive elements on the character sheet (buttons, inputs, clickable badges, rollable skill rows) SHALL have a minimum touch target of 44x44px.

#### Scenario: Stat values have minimum size
- **WHEN** an ability score value is displayed
- **THEN** the clickable number area SHALL be at least 44px tall

#### Scenario: Buttons meet minimum size
- **WHEN** any button is rendered in the stats, combat, or details sections
- **THEN** its tap target SHALL meet or exceed 44x44px

### Requirement: Tab bar is sticky on mobile
The character sheet tab navigation bar SHALL be sticky/stick to the top of the viewport on mobile scroll, remaining accessible at all times.

#### Scenario: Tab bar sticks on scroll
- **WHEN** the user scrolls down on the sheet view on a mobile viewport
- **THEN** the tab bar SHALL remain visible at the top of the visible area

### Requirement: Existing desktop rendering is unchanged
The sheet rendering for viewports 768px and wider SHALL produce the same HTML structure and visual appearance as before the change, with the exception of enhanced stepper controls.

#### Scenario: Desktop layout unchanged
- **WHEN** viewing the sheet on a desktop viewport
- **THEN** the column widths, spacing, and overall layout SHALL match the current design
