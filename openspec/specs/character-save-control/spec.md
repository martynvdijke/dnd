# character-save-control Specification

## Purpose
Manual save button + auto-save debounce management for character sheets.

## Requirements

### Requirement: Manual save button

The character sheet header SHALL display a "Save" button with a floppy-disk icon alongside the existing Print, Export, Back, and Delete buttons. Clicking the button SHALL immediately persist all unsaved changes for the current character via `PUT /api/characters/:id`.

#### Scenario: Save button visible in sheet header

- **WHEN** a character sheet is open
- **THEN** the header button group SHALL include a "Save" button with the text "Save" and a `fa-floppy-disk` icon

#### Scenario: Manual save persists all changes

- **WHEN** the user modifies character fields and clicks the Save button
- **THEN** the system SHALL call `PUT /api/characters/:id` with the full `currentChar` object and the dirty flag SHALL be cleared on success

#### Scenario: Save button disabled while saving

- **WHEN** a save operation is in progress
- **THEN** the Save button SHALL be disabled and show "Saving..." text to prevent duplicate submissions

### Requirement: Dirty state indicator

The system SHALL track whether the character sheet has unsaved changes. When changes are pending (dirty state), the Save button SHALL show a distinct visual style (e.g., gold/primary color, subtle pulse animation). When saved, the button SHALL revert to default outline style. A dirty flag SHALL be set whenever any tracked character field is modified via input handlers, and cleared only after a successful save API response.

#### Scenario: Dirty state shown on field modification

- **WHEN** the user types in any character sheet input field
- **THEN** the Save button SHALL change from outline styling to filled/styled state indicating unsaved changes

#### Scenario: Dirty state cleared after save

- **WHEN** the save operation completes successfully
- **THEN** the Save button SHALL revert to default outline styling

#### Scenario: Dirty state on stepper interactions

- **WHEN** the user clicks a stepper (+/-) button or edits a stepper value inline
- **THEN** the dirty flag SHALL be set

### Requirement: Deferred periodic auto-save

The system SHALL perform periodic auto-saves at a configurable interval (default 12 seconds). Auto-save SHALL only fire if there are unsaved changes (dirty flag is set). The periodic save SHALL NOT fire while a manual save is in progress. The auto-save SHALL NOT trigger on every keystroke — input events SHALL only set the dirty flag.

#### Scenario: Auto-save fires on interval when dirty

- **WHEN** the dirty flag is set and the auto-save interval (default 12s) has elapsed since the last save
- **THEN** the system SHALL automatically call `PUT /api/characters/:id` and clear the dirty flag on success

#### Scenario: Auto-save does not fire if clean

- **WHEN** the dirty flag is not set
- **THEN** the auto-save SHALL NOT make any API call

#### Scenario: Auto-save skipped during manual save

- **WHEN** a manual save is in progress
- **THEN** the auto-save SHALL skip its cycle (not queue a duplicate save)

### Requirement: Save on tab switch

When the user switches character sheet tabs, any pending unsaved changes SHALL be flushed immediately via an API call before the tab renders.

#### Scenario: Unsaved changes flushed on tab switch

- **WHEN** the user clicks a different tab and the dirty flag is set
- **THEN** the system SHALL immediately save changes before rendering the new tab content

### Requirement: Shareable save module

The save functionality (manual save, dirty state, auto-save interval, tab-switch flush) SHALL be implemented in a shared module (e.g., `ts/lib/save.ts`) rather than duplicated across `app.ts` and `characters/combat.ts`. Both files SHALL import and use the shared module.

#### Scenario: app.ts uses shared save module

- **WHEN** `app.ts` needs to persist character changes
- **THEN** it SHALL call the shared save function from `ts/lib/save.ts`

#### Scenario: characters/combat.ts uses shared save module

- **WHEN** `characters/combat.ts` needs to persist character changes
- **THEN** it SHALL call the shared save function from `ts/lib/save.ts`
