## ADDED Requirements

### Requirement: User roles (normal, dm, admin)
The system SHALL support three user roles: `normal`, `dm`, and `admin`. The `normal` role has no access to one-shot features. The `dm` role can create and manage one-shots, see all users' characters, and link characters to one-shots. The `admin` role inherits all DM privileges plus system configuration access.

#### Scenario: DM creates one-shot
- **WHEN** a user with role `dm` makes a POST to `/oneshot-adventures`
- **THEN** the system creates the one-shot and returns 201

#### Scenario: Normal user cannot create one-shot
- **WHEN** a user with role `normal` makes a POST to `/oneshot-adventures`
- **THEN** the system returns 403 Forbidden

#### Scenario: Normal user cannot view one-shots
- **WHEN** a user with role `normal` makes a GET to `/oneshot-adventures`
- **THEN** the system returns 403 Forbidden

#### Scenario: Admin can view list of all one-shots
- **WHEN** a user with role `admin` makes a GET to `/oneshot-adventures`
- **THEN** the system returns all one-shots across all users

#### Scenario: Admin inherits DM privileges for one-shot operations
- **WHEN** a user with role `admin` makes a POST to `/oneshot-adventures`
- **THEN** the system creates the one-shot and returns 201

### Requirement: DMRequired() middleware
The system SHALL provide a `DMRequired()` middleware function that checks the authenticated user's role is `dm` or `admin`. Routes gated by this middleware SHALL reject non-DM users with 403 Forbidden.

#### Scenario: DMRequired allows DM
- **WHEN** a user with role `dm` accesses a route with DMRequired middleware
- **THEN** the request proceeds

#### Scenario: DMRequired allows admin
- **WHEN** a user with role `admin` accesses a route with DMRequired middleware
- **THEN** the request proceeds

#### Scenario: DMRequired rejects normal user
- **WHEN** a user with role `normal` accesses a route with DMRequired middleware
- **THEN** the system returns 403 Forbidden

### Requirement: DM can see all user characters
A user with role `dm` or `admin` SHALL be able to list all characters across all users for the purpose of linking them to one-shots.

#### Scenario: DM lists all characters
- **WHEN** a user with role `dm` makes a GET to `/characters/all`
- **THEN** the system returns all characters from all users
