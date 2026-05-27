## ADDED Requirements

### Requirement: DM session planner view
The system SHALL provide a DM-side session planning view where DMs can prepare notes, planned encounters, player goals, and track session status.

#### Scenario: Session planner accessible from campaign
- **WHEN** a DM views a campaign
- **THEN** the system SHALL show a "Session Planner" link or tab

#### Scenario: Create new session plan
- **WHEN** a DM clicks "New Session Plan"
- **THEN** the system SHALL show a form with fields: session date, title, planned encounters (linked from encounter builder), NPCs that may appear, player character goals, DM notes, expected duration

#### Scenario: Session plan status tracking
- **WHEN** viewing session plans
- **THEN** each plan SHALL show its status: "Planned", "Ready", "In Progress", or "Completed"
- **WHEN** a DM clicks "Start Session"
- **THEN** the status SHALL change to "In Progress"
- **WHEN** a DM clicks "End Session"
- **THEN** the status SHALL change to "Completed"

#### Scenario: Link encounters to session plan
- **WHEN** editing a session plan
- **THEN** the DM SHALL be able to select from existing encounter templates to link to the session

#### Scenario: Session plans listed chronologically
- **WHEN** viewing the session planner
- **THEN** plans SHALL be listed in reverse chronological order (most recent first), with status badges
