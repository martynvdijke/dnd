## ADDED Requirements

### Requirement: Upload files attached to parties
The system SHALL support file uploads attached to parties via the same polymorphic owner_type/owner_id upload system.

#### Scenario: Upload file to party
- **WHEN** a DM uploads a file with owner_type="party" and owner_id=<party_id>
- **THEN** the system stores the file and creates an upload record linked to that party

#### Scenario: List uploads for party
- **WHEN** a DM makes a GET to `/parties/:id/uploads`
- **THEN** the system returns uploads with owner_type="party" and owner_id=<party_id>

#### Scenario: Delete party upload
- **WHEN** a DM makes a DELETE to `/uploads/:id`
- **THEN** the system deletes the upload record and its file
