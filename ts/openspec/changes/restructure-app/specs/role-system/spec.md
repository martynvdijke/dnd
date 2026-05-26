## ADDED Requirements

### Requirement: Role-based access control
The system SHALL support three user roles: `normal`, `dm`, `admin`. The `role` field already exists on the `users` table and SHALL be extended to accept "normal", "dm", "admin".

#### Scenario: Default role on user creation
- **WHEN** a new user is created via admin
- **THEN** the user SHALL have role "normal" by default

### Requirement: DM middleware
The system SHALL provide a `DMRequired()` middleware function that restricts endpoints to users with role "dm" or "admin".

#### Scenario: DM access granted
- **WHEN** a user with role "dm" requests a DM-only endpoint
- **THEN** the system SHALL allow the request

#### Scenario: Normal user denied DM endpoint
- **WHEN** a user with role "normal" requests a DM-only endpoint
- **THEN** the system SHALL return 403 Forbidden

#### Scenario: Admin bypasses DM restriction
- **WHEN** a user with role "admin" requests a DM-only endpoint
- **THEN** the system SHALL allow the request

### Requirement: DM-only one-shot access
Only users with role "dm" or "admin" SHALL be able to create, update, delete one-shot adventures.

#### Scenario: DM creates one-shot
- **WHEN** a DM user sends POST to `/api/oneshot-adventures`
- **THEN** the system SHALL create the one-shot

#### Scenario: Normal user cannot create one-shot
- **WHEN** a normal user sends POST to `/api/oneshot-adventures`
- **THEN** the system SHALL return 403 Forbidden

### Requirement: DM can see all users' characters
A DM user SHALL be able to list all characters across all users for PC linking to one-shots.

#### Scenario: DM lists all characters
- **WHEN** a DM user requests `/api/characters` with `?all=true`
- **THEN** the system SHALL return all characters from all users

#### Scenario: Normal user cannot see other users' characters
- **WHEN** a normal user requests `/api/characters?all=true`
- **THEN** the system SHALL return only their own characters
