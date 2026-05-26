## ADDED Requirements

### Requirement: Party as first-class entity
The system SHALL have a `parties` table with fields: id, user_id, name, description, created_at. Parties are owned by a user (the DM). Party name SHALL be renamable.

#### Scenario: Create party
- **WHEN** a DM makes a POST to `/parties` with a name
- **THEN** the system creates the party and returns 201 with the party ID

#### Scenario: Update party name
- **WHEN** a DM makes a PUT to `/parties/:id` with a new name
- **THEN** the system updates the party name and returns 200

#### Scenario: Delete party
- **WHEN** a DM makes a DELETE to `/parties/:id`
- **THEN** the system deletes the party and returns 200

### Requirement: Factions moved to Party
Factions SHALL be associated with a Party instead of (or in addition to) a Campaign. The factions table SHALL gain a nullable `party_id` foreign key. The UI SHALL show factions under Party view.

#### Scenario: Create faction under party
- **WHEN** a DM makes a POST to `/parties/:id/factions` with faction details
- **THEN** the system creates the faction linked to that party

#### Scenario: List factions for party
- **WHEN** a DM makes a GET to `/parties/:id/factions`
- **THEN** the system returns factions linked to the party

### Requirement: Party file uploads
The system SHALL support file uploads attached to a Party entity. Uploads use the polymorphic owner_type/owner_id pattern.

#### Scenario: Upload file to party
- **WHEN** a DM uploads a file with owner_type="party" and owner_id=<party_id>
- **THEN** the system stores the upload and links it to the party
