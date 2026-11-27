# accessibility Specification

## Purpose
TBD - created by archiving change add-player-experience. Update Purpose after archive.
## Requirements
### Requirement: Skip link and main landmark
The application shell SHALL provide a skip link as the first focusable element that moves focus to the main content, and the primary content SHALL be wrapped in a `main` landmark.

#### Scenario: Skip link targets main content
- **WHEN** a keyboard user focuses the first element and activates the skip link
- **THEN** focus moves to the main content region

### Requirement: Sheet tab semantics
The character sheet tab bar SHALL expose tab semantics: each tab SHALL have `role="tab"`, `aria-selected` reflecting the active tab, and `aria-controls` pointing at its panel; each section SHALL have `role="tabpanel"`. Tabs SHALL support roving tabindex and ArrowLeft, ArrowRight, Home, and End keyboard navigation.

#### Scenario: Active tab is announced
- **WHEN** the sheet renders with the Stats tab active
- **THEN** the Stats tab has `aria-selected="true"` and every other tab has `aria-selected="false"`

#### Scenario: Arrow keys move between tabs
- **WHEN** a tab has focus and the user presses ArrowRight
- **THEN** focus and selection move to the next tab

#### Scenario: Home and End jump to the ends
- **WHEN** a tab has focus and the user presses End
- **THEN** focus and selection move to the last tab

### Requirement: Notification live region
Toast notifications SHALL be announced by assistive technology through a polite live region.

#### Scenario: Toast is announced politely
- **WHEN** a toast is shown
- **THEN** it is rendered inside a container with `aria-live="polite"` and the toast itself has `role="status"`

### Requirement: Active navigation state
The active navigation item SHALL be marked with `aria-current="page"` in both the sidebar and the bottom tab bar.

#### Scenario: Current view is marked
- **WHEN** the characters view is active
- **THEN** the characters navigation item has `aria-current="page"` and other items do not
