## ADDED Requirements

### Requirement: Campaign dashboard view
The system SHALL provide a consolidated campaign dashboard showing recent activity, upcoming events, and pending items in a single view.

#### Scenario: Dashboard accessible from campaign
- **WHEN** a DM views a campaign
- **THEN** the system SHALL show a "Dashboard" tab as the default view

#### Scenario: Dashboard shows upcoming calendar events
- **WHEN** the dashboard is rendered
- **THEN** the system SHALL show the next 5 upcoming calendar events with date, title, and type

#### Scenario: Dashboard shows recent recaps
- **WHEN** the dashboard is rendered
- **THEN** the system SHALL show the 3 most recent session recaps with title, date, and word count

#### Scenario: Dashboard shows pending combat
- **WHEN** the dashboard is rendered AND there are active combat entries
- **THEN** the system SHALL show a "Combat in Progress" card with the active combat count

#### Scenario: Dashboard shows recent dice rolls
- **WHEN** the dashboard is rendered
- **THEN** the system SHALL show the 5 most recent dice rolls across all party members

#### Scenario: Dashboard shows character quick-stats
- **WHEN** the dashboard is rendered
- **THEN** the system SHALL show a row of party member cards with name, class, HP bar, and status

#### Scenario: Dashboard layout uses responsive card grid
- **WHEN** the dashboard is rendered on desktop
- **THEN** cards SHALL display in a 2-3 column grid
- **WHEN** on mobile
- **THEN** cards SHALL stack vertically
