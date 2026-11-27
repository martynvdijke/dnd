# mobile-responsive-layout Specification

## Purpose
Responsive grid system that adapts the character sheet from desktop (multi-column) to mobile (single-column card stack) with appropriate typography and spacing.

## Requirements

### Requirement: Ability score grid adapts to viewport
The 6 ability scores SHALL render in a responsive grid that adapts columns based on viewport width.

#### Scenario: Desktop shows 6 columns
- **WHEN** the viewport is 992px or wider
- **THEN** each ability score SHALL occupy 1/6 of the row (`col-lg-2`)

#### Scenario: Tablet shows 3 columns
- **WHEN** the viewport is between 768px and 991px
- **THEN** each ability score SHALL occupy 1/3 of the row (`col-md-4`)

#### Scenario: Mobile shows 2 columns
- **WHEN** the viewport is below 768px
- **THEN** each ability score SHALL occupy 1/2 of the row (`col-6`)

### Requirement: Combat stats stack vertically on mobile
The combat stat cards (AC, Initiative, Speed) and HP section SHALL stack in a single column on mobile viewports.

#### Scenario: Combat stats on mobile
- **WHEN** the viewport is below 576px
- **THEN** AC, Initiative, Speed, and the HP bar SHALL each occupy full width in a vertical stack

#### Scenario: Combat stats on desktop
- **WHEN** the viewport is 576px or wider
- **THEN** AC, Initiative, Speed SHALL display in a 3-column row, with HP bar below

### Requirement: Skills display in 2-column grid on mobile
The skill list (18 skills) SHALL display in a 2-column grid on mobile and 3-column on tablet.

#### Scenario: Skills on mobile
- **WHEN** the viewport is below 768px
- **THEN** skill rows SHALL display in a 2-column grid

#### Scenario: Skills on desktop
- **WHEN** the viewport is 768px or wider
- **THEN** skill rows SHALL display in a 3-column grid

### Requirement: Typography scales for readability
Font sizes in the stat display SHALL increase on mobile for readability while fitting within the viewport.

#### Scenario: Stat labels readable on mobile
- **WHEN** the viewport is below 768px
- **THEN** ability score labels SHALL have font-size of at least 0.75rem and values at least 1.25rem

#### Scenario: Value text does not overflow
- **WHEN** an ability score value is displayed on mobile
- **THEN** the number SHALL not overflow its container at any value between 1 and 30

### Requirement: Spacing is generous on mobile
Card padding, gaps between cards, and margins SHALL increase on mobile for comfortable touch interaction.

#### Scenario: Card padding on mobile
- **WHEN** the viewport is below 768px
- **THEN** stat cards SHALL have padding of at least 12px (p-3 equivalent)

#### Scenario: Gap between cards
- **WHEN** stat cards are displayed on mobile
- **THEN** the gap between cards SHALL be at least 12px
