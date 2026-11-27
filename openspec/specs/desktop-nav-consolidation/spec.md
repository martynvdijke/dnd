# desktop-nav-consolidation Specification

## Purpose
Clean up desktop navigation to remove top-bar/sidebar redundancy.

## Requirements

### Requirement: Desktop navigation consolidates top bar and sidebar

On desktop screens (>= 992px width), the top navbar SHALL only show the brand name, search bar, user info, and theme toggle. Navigation links SHALL be removed from the top navbar since the sidebar provides navigation.

#### Scenario: Desktop top navbar hides nav links
- **WHEN** the viewport width is >= 992px and the sidebar is visible
- **THEN** the top navbar SHALL NOT display navigation link items (Characters, Party View, Compendium, etc.)

#### Scenario: Desktop top navbar shows nav links when sidebar collapsed
- **WHEN** the sidebar is collapsed on desktop
- **THEN** the top navbar SHALL show navigation link items as a fallback

#### Scenario: Mobile top navbar unchanged
- **WHEN** the viewport width is < 992px
- **THEN** the top navbar SHALL display navigation links in the collapsible toggler as before

### Requirement: Sidebar remains the primary desktop navigation

The sidebar SHALL remain the primary navigation method on desktop, with all existing navigation items: Characters, Party View, Compendium, Dice, Encounters, One-Shots, Combat, Timeline, Factions, Shops, Admin.

#### Scenario: Sidebar contains all nav items
- **WHEN** user views the sidebar on desktop
- **THEN** all navigation items SHALL be present in the sidebar

#### Scenario: Sidebar collapse still works
- **WHEN** user clicks the sidebar toggle
- **THEN** the sidebar SHALL collapse/expand, and the top navbar SHALL adjust its link visibility accordingly
