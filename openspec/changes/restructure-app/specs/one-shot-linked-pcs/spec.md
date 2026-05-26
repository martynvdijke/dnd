## ADDED Requirements

### Requirement: Link player characters to one-shots
The system SHALL support linking player characters (from any user) to a one-shot adventure. Each link has: character_id, oneshot_adventure_id, role, notes. DM can see all characters across all users.

#### Scenario: Link character to one-shot
- **WHEN** a DM makes a POST to `/oneshot-adventures/:id/characters` with character_id and role
- **THEN** the system creates the link and returns 201

#### Scenario: List linked characters for one-shot
- **WHEN** a DM makes a GET to `/oneshot-adventures/:id/characters`
- **THEN** the system returns all linked characters with their user names

#### Scenario: Remove character from one-shot
- **WHEN** a DM makes a DELETE to `/oneshot-adventures/:id/characters/:charId`
- **THEN** the system removes the link and returns 200

#### Scenario: DM can browse all user characters
- **WHEN** a DM makes a GET to `/characters/all`
- **THEN** the system returns all characters owned by any user, including their owner names
