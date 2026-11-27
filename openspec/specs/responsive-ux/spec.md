# responsive-ux Specification

## Purpose
Systematic responsive design approach with consistent breakpoints, improved mobile touch targets, and skeleton loading states.

## Requirements

### Requirement: CSS SHALL use a 4-tier standardized breakpoint system
The CSS SHALL define exactly four breakpoints as CSS custom properties on `:root`:
- `--bp-sm`: 576px (phones)
- `--bp-md`: 768px (tablets portrait)
- `--bp-lg`: 992px (tablets landscape / small desktop)
- `--bp-xl`: 1200px (desktop)

All responsive behavior SHALL use these four breakpoints. The existing 19 ad-hoc media query values SHALL be consolidated to match.

#### Scenario: Responsive layout uses consistent breakpoints
- **WHEN** a developer adds a new responsive rule
- **THEN** they MUST use one of the four standard breakpoints
- **THEN** the rule is applied at the expected viewport width

### Requirement: Interactive elements SHALL meet minimum touch target size
All interactive elements (buttons, links, form controls, sidebar items, bottom tab items) SHALL have a minimum touch target of 44×44 CSS pixels on viewports below `--bp-md` (768px). In session mode, this SHALL apply at all viewport sizes.

#### Scenario: Mobile touch targets are large enough
- **WHEN** the viewport width is below 768px
- **THEN** all buttons, nav items, and interactive controls have at least 44px height and 44px clickable area
- **THEN** tap targets do not overlap or crowd each other

#### Scenario: Session mode enforces large targets on desktop
- **WHEN** session mode is active on a desktop viewport
- **THEN** all interactive elements have a minimum 44px touch target
- **THEN** form controls have 18px minimum font size

### Requirement: Loading states SHALL use skeleton screens
When the app fetches data from the API, the UI SHALL display skeleton placeholders that match the content layout (character card grid, list rows) instead of a generic spinner overlay. Skeletons SHALL appear only after a 200ms delay and SHALL be replaced immediately when data arrives.

#### Scenario: Character grid shows skeletons during loading
- **WHEN** the characters view is loading
- **THEN** skeleton cards matching the character card grid layout are shown
- **THEN** real cards replace skeletons when the API response arrives

#### Scenario: Fast connections skip skeletons
- **WHEN** the API responds within 200ms
- **THEN** no skeletons are shown
- **THEN** content appears immediately without layout shift

### Requirement: Bottom tab bar SHALL be consistent across mobile views
The bottom tab bar SHALL show the active tab with a highlighted icon and label. The "More" tab SHALL dynamically show/hide additional navigation items based on the user's role (admin, DM, player). On viewports at or above `--bp-md` (768px), the bottom tab bar SHALL be hidden and the sidebar SHALL be shown instead.

#### Scenario: Mobile shows bottom tab bar
- **WHEN** the viewport width is below 768px
- **THEN** the bottom tab bar is visible
- **THEN** the sidebar is hidden
- **THEN** tapping a tab switches the view

#### Scenario: Desktop shows sidebar
- **WHEN** the viewport width is 768px or above
- **THEN** the sidebar is visible
- **THEN** the bottom tab bar is hidden
- **THEN** clicking a sidebar item switches the view
