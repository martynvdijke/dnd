# session-mode Specification

## Purpose
Toggleable live-play mode that simplifies the UI to combat essentials with large touch targets. Auto-activates on initiative roll or DM session start.

## Requirements

### Requirement: Session Mode Toggle
The system SHALL provide a manual toggle to activate and deactivate Session Mode. The toggle SHALL be accessible from the navigation area on both mobile and desktop viewports.

#### Scenario: User activates Session Mode manually
- **WHEN** a user clicks/taps the Session Mode toggle while in normal mode
- **THEN** the UI SHALL transition to Session Mode: non-essential sections are hidden, the `body.session-mode` CSS class is applied, and a toast notification "Session Mode activated" is displayed with an "Undo" button

#### Scenario: User deactivates Session Mode manually
- **WHEN** a user clicks/taps the Session Mode toggle while in Session Mode
- **THEN** the UI SHALL return to normal mode: all previously visible sections are restored, the `body.session-mode` CSS class is removed

#### Scenario: Undo from toast notification
- **WHEN** Session Mode is auto-activated and the user taps "Undo" on the toast notification within 5 seconds
- **THEN** Session Mode SHALL deactivate and return to normal mode

### Requirement: Session Mode Auto-Activation
The system SHALL automatically activate Session Mode when specific in-game events occur. Auto-activation SHALL display a toast notification with an undo option.

#### Scenario: DM starts a session
- **WHEN** the DM starts a session plan and the WebSocket event `session.started` is received by a player client
- **THEN** Session Mode SHALL activate automatically on that player's device

#### Scenario: Initiative is rolled in combat
- **WHEN** initiative is rolled in the combat tracker and the WebSocket event `combat.initiative_rolled` is received
- **THEN** Session Mode SHALL activate automatically on all connected clients in that campaign

#### Scenario: Auto-activation can be disabled
- **WHEN** a user has disabled auto-activation in their preferences
- **THEN** Session Mode SHALL NOT auto-activate on WebSocket events, and the user MUST toggle it manually

### Requirement: Session Mode UI Simplification
When Session Mode is active, the system SHALL display only combat-essential UI elements with enlarged touch targets and high-contrast text.

#### Scenario: Session Mode hides non-essential navigation
- **WHEN** Session Mode is active
- **THEN** the sidebar (desktop) and bottom tab bar (mobile) SHALL be hidden, and a minimal top bar with the Session Mode toggle and campaign name SHALL be displayed instead

#### Scenario: Session Mode character view shows essentials only
- **WHEN** Session Mode is active and the user is viewing their character
- **THEN** only the following SHALL be visible: HP bar (large format), conditions list, quick damage/heal buttons, initiative position indicator, and concentration spell indicator (if applicable)

#### Scenario: Session Mode enlarges touch targets
- **WHEN** Session Mode is active
- **THEN** all interactive elements SHALL have a minimum touch target of 48x48px, and body text SHALL be at least 18px

#### Scenario: Session Mode FAB shows combat actions only
- **WHEN** Session Mode is active
- **THEN** the FAB SHALL display only two actions: "Roll Dice" and "End Turn"

### Requirement: Session Mode State Persistence
The system SHALL persist Session Mode state within the current browser tab session using `sessionStorage`. Session Mode state SHALL survive page refreshes but SHALL NOT persist across tab closures.

#### Scenario: Session Mode survives page refresh
- **WHEN** Session Mode is active and the user refreshes the page
- **THEN** Session Mode SHALL remain active after the page reloads

#### Scenario: Session Mode clears on tab close
- **WHEN** Session Mode is active and the user closes the browser tab
- **THEN** opening the app in a new tab SHALL start in normal mode

### Requirement: Session Mode Visual Theming
The system SHALL apply distinct visual theming when Session Mode is active to provide clear visual differentiation from normal mode.

#### Scenario: High contrast in Session Mode
- **WHEN** Session Mode is active
- **THEN** the background SHALL use a darker theme, text SHALL use higher contrast colors, and the HP bar SHALL use accent colors that are distinguishable at a glance

#### Scenario: Session Mode respects dark/light theme preference
- **WHEN** Session Mode is active
- **THEN** the session mode theme SHALL be a variant of the user's current theme preference (dark session theme for dark mode, light session theme for light mode)
