# character-ownership Specification

## Purpose
TBD - created by archiving change character-selection-ownership. Update Purpose after archive.
## Requirements
### Requirement: Ownership determination

The system SHALL classify every character as either owned or shared for the current user. A character is owned when its `user_id` matches the current user, or when the user holds an ownership-granting membership (e.g., DM role) for the character's campaign.

#### Scenario: Character owned by user

- **WHEN** a character's `user_id` equals the current user's id
- **THEN** the system SHALL treat the character as owned

#### Scenario: Character owned by another user

- **WHEN** a character's `user_id` differs from the current user's id and the user has no ownership-granting campaign membership
- **THEN** the system SHALL treat the character as shared

### Requirement: Shared characters are read-only

For a shared character, the system SHALL allow viewing notes, journals, feats, and reference information (race, class, level, background, portrait, appearance, backstory) but SHALL NOT allow editing them or the character's sheet.

#### Scenario: Viewing a shared character

- **WHEN** a user opens a shared character
- **THEN** the system SHALL display the character's notes, journals, feats, and reference info in read-only form

#### Scenario: Editing a shared character is blocked

- **WHEN** a user attempts to modify a shared character or any of its notes, journals, feats, or sheet fields
- **THEN** the system SHALL reject the write and the UI SHALL not expose edit controls

### Requirement: Owned characters have full access

For an owned character, the system SHALL provide full access: editing the sheet, managing feats, adjusting ability scores and HP, tracking resources and conditions, and all existing character functionality.

#### Scenario: Editing an owned character

- **WHEN** a user edits an owned character
- **THEN** the system SHALL persist the changes and expose the full edit UI

#### Scenario: Scoring and tracking on owned characters

- **WHEN** a user views an owned character
- **THEN** the system SHALL show ability score editing, HP/resource tracking, and feat management controls

### Requirement: Server-side write enforcement

The system SHALL enforce ownership on the server for all character write endpoints, independent of UI state, so that a shared character can never be modified through the API.

#### Scenario: API write to a shared character

- **WHEN** a request attempts to write to a character the user does not own
- **THEN** the server SHALL return an error response (e.g., 403 Forbidden) and SHALL NOT apply the change

### Requirement: Party view respects ownership

The party view SHALL render shared characters without exposing DM-only or edit affordances to users who do not own them.

#### Scenario: Shared character in party view

- **WHEN** a shared character is rendered in the party view for a non-owner
- **THEN** the party view SHALL display the character's read-only info and SHALL not show edit or DM-only controls
