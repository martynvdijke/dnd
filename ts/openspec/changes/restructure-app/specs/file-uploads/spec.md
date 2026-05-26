## ADDED Requirements

### Requirement: Universal upload system
The existing `uploads` table SHALL be extended with `owner_type` (VARCHAR) and `owner_id` (INT64) columns to allow polymorphic file attachment across entities (Party, Item, NPC, etc.).

#### Scenario: Upload file to entity
- **WHEN** a user sends a POST to `/api/upload` with `owner_type`, `owner_id`, and a file
- **THEN** the file SHALL be saved, a new upload record created, and the upload ID returned

#### Scenario: List uploads for entity
- **WHEN** a user sends GET to `/api/uploads?owner_type=X&owner_id=Y`
- **THEN** all uploads for that entity SHALL be returned

#### Scenario: Delete upload
- **WHEN** a user sends DELETE to `/api/uploads/:id`
- **THEN** the upload SHALL be deleted (file and record)

#### Scenario: Validate owner_type
- **WHEN** a user sends upload with invalid `owner_type`
- **THEN** the system SHALL return 400 Bad Request

### Requirement: Upload on Party
Parties SHALL support file uploads as specified in `party-management/spec.md`.

#### Scenario: Upload party document
- **WHEN** a user uploads a document to a party
- **THEN** the file SHALL be accessible via `/api/parties/:id/uploads`
