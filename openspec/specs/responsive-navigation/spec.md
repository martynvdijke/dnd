# responsive-navigation Specification

## Purpose
Adaptive navigation shell — bottom tab bar on mobile, collapsible sidebar on desktop, campaign-first routing for players in active campaigns.

## Requirements

### Requirement: Bottom Tab Bar on Mobile
The system SHALL display a fixed bottom tab bar on viewports ≤768px wide with exactly 4 primary navigation items: Home, Campaign, Dice, and More. The bottom tab bar SHALL replace the current top navbar on mobile viewports.

#### Scenario: Mobile viewport shows bottom tab bar
- **WHEN** a user loads the app on a viewport ≤768px wide
- **THEN** a fixed bottom tab bar is visible with icons and labels for Home, Campaign, Dice, and More

#### Scenario: Top navbar is hidden on mobile
- **WHEN** a user loads the app on a viewport ≤768px wide
- **THEN** the top navbar SHALL NOT be visible

#### Scenario: Bottom tab bar respects device safe area
- **WHEN** the app is displayed on a device with a bottom safe area inset (e.g., iPhone notch)
- **THEN** the bottom tab bar SHALL include padding equal to `env(safe-area-inset-bottom)` to avoid overlap with device chrome

### Requirement: Collapsible Sidebar on Desktop
The system SHALL display a collapsible sidebar navigation on viewports >768px wide. The sidebar SHALL contain all navigation items currently in the top navbar and SHALL be toggleable via a hamburger menu button.

#### Scenario: Desktop viewport shows sidebar
- **WHEN** a user loads the app on a viewport >768px wide
- **THEN** a sidebar navigation is visible on the left side of the screen containing all navigation items

#### Scenario: Sidebar toggle
- **WHEN** a user clicks the hamburger menu button on a desktop viewport
- **THEN** the sidebar SHALL toggle between expanded and collapsed states

#### Scenario: Sidebar collapsed state shows icons only
- **WHEN** the sidebar is in collapsed state
- **THEN** only navigation icons (without labels) SHALL be visible, and the main content area SHALL expand to fill the freed space

### Requirement: More Menu Overflow
The system SHALL display secondary navigation items (Characters, Compendium, Encounters, One-Shots, Combat, Timeline, Factions, Shops, Admin) inside the "More" menu. On mobile, the More menu SHALL open as a bottom sheet. On desktop, secondary items SHALL be listed in an expandable sidebar section.

#### Scenario: More menu on mobile
- **WHEN** a user taps the "More" tab on a mobile viewport
- **THEN** a bottom sheet SHALL slide up displaying all secondary navigation items

#### Scenario: More menu on desktop
- **WHEN** a user clicks the "More" item in the desktop sidebar
- **THEN** an expandable section SHALL reveal all secondary navigation items within the sidebar

#### Scenario: Conditional nav items respect user role
- **WHEN** a user without the DM role views the More menu
- **THEN** the One-Shots, Combat, and Admin items SHALL NOT be displayed

### Requirement: Campaign-First Routing
The system SHALL navigate to the campaign dashboard view on initial load when the logged-in player belongs to at least one active campaign. If the player has no active campaigns, the system SHALL navigate to the characters list view.

#### Scenario: Player with active campaign lands on dashboard
- **WHEN** a player who belongs to an active campaign loads the app
- **THEN** the campaign dashboard view (partyView) SHALL be displayed as the initial view

#### Scenario: Player without active campaign lands on characters
- **WHEN** a player who does not belong to any active campaign loads the app
- **THEN** the characters list view (charactersView) SHALL be displayed as the initial view

#### Scenario: DM user lands on characters list
- **WHEN** a user with the DM role loads the app
- **THEN** the characters list view SHALL be displayed as the initial view (DMs manage multiple campaigns and need the overview)

### Requirement: Active Tab State
The system SHALL visually indicate the currently active navigation tab. The active state SHALL update when the user navigates to a different view via any mechanism (tab tap, sidebar click, programmatic navigation).

#### Scenario: Tab highlights on navigation
- **WHEN** a user taps the "Campaign" tab
- **THEN** the Campaign tab SHALL display an active visual state (distinct color/weight) and all other tabs SHALL display an inactive state

#### Scenario: Programmatic navigation updates tab state
- **WHEN** the application programmatically switches views (e.g., after creating a character, navigating to the sheet)
- **THEN** the corresponding navigation tab SHALL reflect the active state
