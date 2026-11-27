# e-ink-ui Specification

## Purpose
The e-ink mode adapts DnD for tabletop RPG sessions on e-ink displays — providing high-contrast character sheets, initiative trackers, and reference screens readable from across the table.

## Requirements

### Requirement: E-ink mode toggle
The application SHALL provide a mechanism to enable e-ink mode via URL parameter (`?eink=1`) and optionally persist via cookie or admin setting.

#### Scenario: Enable via URL parameter
- **WHEN** a user appends `?eink=1` to any page URL
- **THEN** the page SHALL render in e-ink mode for that session
- **AND** a cookie SHALL be set to persist the preference

#### Scenario: Enable via admin settings
- **WHEN** an admin toggles "E-ink mode" in settings and saves
- **THEN** e-ink mode SHALL be enabled for all users site-wide

### Requirement: High contrast palette
All e-ink mode pages SHALL use strict black-on-white with no gradients, shadows, or semi-transparency.

#### Scenario: Colors
- **WHEN** e-ink mode is active
- **THEN** all text SHALL be `#000000` on `#ffffff` background
- **THEN** secondary/muted text SHALL use `#333333` on `#ffffff`
- **THEN** table borders and dividers SHALL use `#cccccc` solid lines

### Requirement: No motion or CSS effects

While e-ink mode is active the application SHALL disable all motion and decorative CSS effects.

#### Scenario: Motion and CSS effects are disabled
- **WHEN** e-ink mode is active
- **THEN** all CSS transitions, animations, keyframes SHALL be disabled
- **THEN** no `box-shadow`, `text-shadow`, `backdrop-filter`, `gradient`, or `filter` SHALL render
- **THEN** automatic page refresh SHALL be disabled

### Requirement: Large touch targets

While e-ink mode is active the application SHALL render large, well-spaced touch targets.

#### Scenario: Touch target sizes
- **WHEN** e-ink mode is active
- **THEN** all buttons, links, and interactive elements SHALL be minimum 48×48px
- **THEN** initiative tracker controls SHALL be minimum 56×56px
- **THEN** touch targets SHALL have 8px minimum spacing

### Requirement: Character sheet view
Character sheets SHALL be optimized for e-ink readability.

#### Scenario: Stat display
- **WHEN** viewing a character sheet in e-ink mode
- **THEN** stat values (STR, DEX, CON, etc.) SHALL display as bold minimum 24px numbers
- **THEN** stat names SHALL display as minimum 12px uppercase labels below values
- **THEN** HP, AC, and other key combat stats SHALL be prominently displayed with borders
- **THEN** skill modifiers SHALL use clear +/− prefix rather than color

#### Scenario: Layout simplification
- **WHEN** viewing a character sheet in e-ink mode
- **THEN** the sheet SHALL use a single-column or two-column grid layout
- **THEN** all sections (stats, skills, equipment, spells) SHALL have solid border separators
- **THEN** spell slot indicators SHALL use [filled] [empty] squares, not colored dots

### Requirement: Initiative tracker
The initiative tracker SHALL be simplified for tabletop e-ink use.

#### Scenario: Initiative order
- **WHEN** viewing the initiative tracker in e-ink mode
- **THEN** combatant entries SHALL display with large (minimum 20px) names
- **THEN** initiative values SHALL be bold minimum 24px
- **THEN** the active combatant SHALL be indicated by a solid black left border (not background color)
- **THEN** turn controls (next/prev) SHALL be minimum 56×56px buttons

### Requirement: Dice roller
The dice roller interface SHALL be high-contrast and touch-friendly.

#### Scenario: Dice roller readability
- **WHEN** using the dice roller in e-ink mode
- **THEN** dice buttons SHALL be minimum 56×56px with visible solid borders
- **THEN** roll results SHALL display as large (minimum 28px) bold numbers
- **THEN** dice type labels (d4, d6, d8, d10, d12, d20) SHALL be text, not icon-based
