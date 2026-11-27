# mobile-interactions Specification

## Purpose
Mobile-native interaction patterns — bottom sheets, context-aware FAB, card-based combat tracker, pull-to-refresh, swipe gestures.

## Requirements

### Requirement: Bottom Sheet Component
The system SHALL provide a reusable bottom sheet component that replaces Bootstrap modals on mobile viewports (≤768px). On desktop viewports, the same triggers SHALL open standard centered modals.

#### Scenario: Bottom sheet opens on mobile
- **WHEN** a user triggers a modal action (e.g., "New Character", item details) on a viewport ≤768px wide
- **THEN** a bottom sheet SHALL slide up from the bottom of the screen with a CSS transition, displaying the content that would otherwise appear in a Bootstrap modal

#### Scenario: Bottom sheet has drag handle
- **WHEN** a bottom sheet is displayed
- **THEN** a drag handle indicator SHALL be visible at the top center of the sheet

#### Scenario: Bottom sheet dismisses on backdrop tap
- **WHEN** a user taps the backdrop overlay behind the bottom sheet
- **THEN** the bottom sheet SHALL slide down and dismiss

#### Scenario: Bottom sheet dismisses on swipe down
- **WHEN** a user swipes down on the bottom sheet drag handle with sufficient velocity
- **THEN** the bottom sheet SHALL dismiss with a slide-down animation

#### Scenario: Bottom sheet maximum height
- **WHEN** a bottom sheet is displayed
- **THEN** the sheet SHALL NOT exceed 85vh in height, and content exceeding this height SHALL be scrollable within the sheet

#### Scenario: Desktop fallback to modal
- **WHEN** the same modal trigger is activated on a viewport >768px wide
- **THEN** a standard centered Bootstrap modal SHALL be displayed (existing behavior preserved)

### Requirement: Context-Aware FAB
The system SHALL dynamically change the FAB (Floating Action Button) actions based on the currently active view. The static 3-item FAB SHALL be replaced with a view-responsive FAB.

#### Scenario: FAB shows character actions on character sheet
- **WHEN** the user is viewing a character sheet
- **THEN** the FAB SHALL display actions: Roll Dice, Short Rest, Long Rest

#### Scenario: FAB shows list actions on characters list
- **WHEN** the user is viewing the characters list
- **THEN** the FAB SHALL display actions: New Character, Import Character

#### Scenario: FAB shows combat actions in combat tracker
- **WHEN** the user is viewing the combat tracker
- **THEN** the FAB SHALL display actions: Roll Initiative, End Turn

#### Scenario: FAB shows dice actions on dice view
- **WHEN** the user is viewing the dice roller
- **THEN** the FAB SHALL display actions: Roll d20, Roll Custom

#### Scenario: FAB updates on view change
- **WHEN** the user navigates from one view to another
- **THEN** the FAB actions SHALL update to reflect the new view's action set within the same render cycle

### Requirement: Card-Based Combat Tracker on Mobile
The system SHALL render the combat tracker as a vertical card list on mobile viewports instead of the current table layout.

#### Scenario: Combat tracker renders as cards on mobile
- **WHEN** a user views the combat tracker on a viewport ≤768px wide
- **THEN** each combatant SHALL be displayed as a card containing: name, initiative value, HP bar, AC, active conditions, and action buttons (damage, heal, conditions)

#### Scenario: Active turn is highlighted in card list
- **WHEN** it is a combatant's turn
- **THEN** that combatant's card SHALL have a distinct visual highlight (border color or background)

#### Scenario: Combat tracker preserves table on desktop
- **WHEN** a user views the combat tracker on a viewport >768px wide
- **THEN** the existing table layout SHALL be displayed (no change to desktop experience)

### Requirement: Pull-to-Refresh
The system SHALL support pull-to-refresh gesture on mobile viewports for views that display dynamic data.

#### Scenario: Pull-to-refresh on campaign dashboard
- **WHEN** a user pulls down from the top of the campaign dashboard on a mobile viewport
- **THEN** the view SHALL refresh its data (weather, quests, events, dice rolls) and display a loading indicator during the refresh

#### Scenario: Pull-to-refresh on character list
- **WHEN** a user pulls down from the top of the character list on a mobile viewport
- **THEN** the character list SHALL re-fetch and display updated data

#### Scenario: Pull-to-refresh does not trigger on scroll up
- **WHEN** a user scrolls up normally (not from the top of the view)
- **THEN** the pull-to-refresh gesture SHALL NOT activate

### Requirement: Swipe-to-Dismiss Notifications
The system SHALL support swipe-to-dismiss for toast notifications and dismissible alerts on touch devices.

#### Scenario: Swipe toast to dismiss
- **WHEN** a toast notification is displayed and the user swipes it horizontally
- **THEN** the toast SHALL dismiss with a fade-out animation

#### Scenario: Non-touch dismiss still works
- **WHEN** a toast notification is displayed on a non-touch device
- **THEN** the existing auto-dismiss timer and close button SHALL continue to function
