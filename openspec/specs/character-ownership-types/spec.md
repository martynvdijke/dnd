# character-ownership-types Specification

## Purpose
Distinguish between player-owned (full editing) and linked/view-only characters (limited details, party view).

## Requirements

### Requirement: Character type field

The Character model SHALL have a `character_type` field with possible values `"player"` and `"linked"`. All existing characters SHALL default to `"player"` via migration. New characters SHALL default to `"player"`. The `character_type` SHALL be returned in all character API responses.

#### Scenario: New character defaults to player type

- **WHEN** a user creates a new character via `POST /api/characters`
- **THEN** the character SHALL have `character_type = "player"` by default

#### Scenario: Existing characters migrate to player type

- **WHEN** the database migration runs
- **THEN** all existing records in the characters table SHALL get `character_type = "player"`

#### Scenario: Character type returned in API

- **WHEN** a client calls `GET /api/characters` or `GET /api/characters/:id`
- **THEN** the response SHALL include the `character_type` field

### Requirement: Edit permission enforcement

The system SHALL enforce edit permissions at the API layer. A character with `character_type = "player"` SHALL be editable only by its owner (`user_id == current user id`), admins, and the campaign DM. A character with `character_type = "linked"` SHALL NOT be editable by non-admin, non-DM users regardless of ownership. The API SHALL reject write operations (`PUT`, `POST`, `DELETE` on character or sub-resources) with HTTP 403 if the user lacks permission.

#### Scenario: Owner can edit player character

- **WHEN** a user sends `PUT /api/characters/:id` for their own `player` character
- **THEN** the request SHALL succeed (200)

#### Scenario: Non-owner cannot edit player character

- **WHEN** a user who is not the owner, admin, or DM sends `PUT /api/characters/:id` for a `player` character
- **THEN** the request SHALL fail with HTTP 403

#### Scenario: Linked character write rejection

- **WHEN** any non-admin user sends a write request for a `linked` character
- **THEN** the request SHALL fail with HTTP 403

### Requirement: View-only character display in party

In the party view, characters with `character_type = "linked"` or characters owned by another user SHALL display with limited details. The character card SHALL show name, race, class, level, HP, AC, and status. The card SHALL NOT be clickable to open the full character sheet for non-owners of `player` characters. For `linked` characters, a "View Stats" button SHALL open a read-only summary modal instead.

#### Scenario: Linked character card in party

- **WHEN** a party view renders a `linked` character
- **THEN** a "View Stats" button SHALL appear instead of the full click-to-open behavior

#### Scenario: View-only stats modal for linked characters

- **WHEN** the user clicks "View Stats" on a linked character
- **THEN** a modal SHALL show the character's name, race, class, level, HP, AC, stats, and skills (read-only, no edit controls)

### Requirement: Hide edit controls for non-editable characters

When a user opens a character they cannot edit, the character sheet SHALL render in read-only mode. All input fields SHALL be disabled or replaced with display-only text. Save buttons, steppers, add/remove buttons, and all mutation controls SHALL be hidden.

#### Scenario: Read-only sheet for non-editable character

- **WHEN** a non-owner user opens a `player` character or any user opens a `linked` character
- **THEN** the sheet SHALL render all fields as read-only text, with no input fields, buttons, or save controls visible
