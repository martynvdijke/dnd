## ADDED Requirements

### Requirement: Party as first-class entity
The system SHALL have a `parties` table with columns: `id`, `user_id`, `name`, `description`, `created_at`, `updated_at`. Campaign's `party_name` field SHALL be migrated to this new table.

#### Scenario: Create party
- **WHEN** a user sends POST to `/api/parties` with name and description
- **THEN** a new party SHALL be created and the ID returned

#### Scenario: List parties
- **WHEN** a user sends GET to `/api/parties`
- **THEN** all parties owned by the user SHALL be returned

#### Scenario: Rename party
- **WHEN** a user sends PUT to `/api/parties/:id` with new name
- **THEN** the party name SHALL be updated

#### Scenario: Delete party
- **WHEN** a user sends DELETE to `/api/parties/:id`
- **THEN** the party SHALL be deleted along with its factions and uploads

### Requirement: Factions moved to Party
The existing `factions` table's `campaign_id` column SHALL be replaced with `party_id`. Factions SHALL be accessible via the Party.

#### Scenario: List factions under party
- **WHEN** a user sends GET to `/api/parties/:id/factions`
- **THEN** all factions belonging to that party SHALL be returned

#### Scenario: Create faction under party
- **WHEN** a user sends POST to `/api/parties/:id/factions` with faction data
- **THEN** the faction SHALL be created with `party_id` set

### Requirement: File uploads on Party
The system SHALL support uploading files (images, documents) to a Party. Uploads SHALL use the existing `uploads` table with `owner_type="party"` and `owner_id=<party_id>`.

#### Scenario: Upload file to party
- **WHEN** a user sends POST to `/api/parties/:id/upload` with a file
- **THEN** the file SHALL be saved and linked to the party

#### Scenario: List party uploads
- **WHEN** a user sends GET to `/api/parties/:id/uploads`
- **THEN** all uploads for the party SHALL be returned
