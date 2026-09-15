# character-journal-notes Specification

## Purpose
TBD - created by archiving change session-journal-world-ai. Update Purpose after archive.
## Requirements
### Requirement: Journal tab loads and persists

The character sheet journal tab SHALL load the character's journal entries and SHALL allow creating, editing, and deleting entries with rich-text content that persists across reloads.

#### Scenario: Entries load on open
- **WHEN** the journal tab is shown
- **THEN** the character's existing journal entries are loaded and displayed

#### Scenario: Entry persists
- **WHEN** the user creates or edits a journal entry and saves
- **THEN** reopening the tab (or reloading) shows the saved entry

#### Scenario: Read-only for shared characters
- **WHEN** the character is shared with the user but not owned by them
- **THEN** journal entries are shown read-only with no create/edit/delete affordances

### Requirement: Notes tab loads and persists

The character sheet notes tab SHALL load the character's notes and allow creating, editing, and deleting them, and edits SHALL persist across reloads.

#### Scenario: Notes load on open
- **WHEN** the notes tab is shown
- **THEN** the character's existing notes are loaded and displayed

#### Scenario: Note persists
- **WHEN** the user creates or edits a note and saves
- **THEN** reopening the tab (or reloading) shows the saved note

### Requirement: Active tab renders on open and deep-link

The character sheet SHALL render the content of the initially active section on open, including when the sheet is opened directly at a journal or notes hash, without requiring the user to click the tab.

#### Scenario: Deep-link to notes
- **WHEN** the sheet is opened with the notes section active
- **THEN** the notes content is loaded and visible without an extra click

#### Scenario: No duplicate render
- **WHEN** a tab is opened that was already rendered
- **THEN** the content is not loaded twice or duplicated

### Requirement: Single notes implementation

The system SHALL implement the character notes tab through a single code path, with no orphaned duplicate renderer.

#### Scenario: No dead duplicate
- **WHEN** the character notes feature is built
- **THEN** only the active implementation is present and referenced, and no unused duplicate notes module ships
