## ADDED Requirements

### Requirement: Remove calendar from campaign
The calendar feature SHALL be removed from the Campaign UI. The `campaign_calendar_events` table and data SHALL remain in the database but the calendar routes SHALL be removed or hidden from the campaign view.

#### Scenario: Calendar routes return 404
- **WHEN** any user makes a GET to `/calendar`
- **THEN** the system returns 404 Not Found

#### Scenario: Calendar tab removed from campaign UI
- **WHEN** a user views the campaign page
- **THEN** no calendar tab or calendar-related UI elements appear

### Requirement: Factions removed from campaign UI
Factions SHALL be removed from the Campaign view and moved to the Party view. The underlying table still supports campaign_id for migration period.

#### Scenario: Factions tab removed from campaign
- **WHEN** a user views the campaign page
- **THEN** no factions tab appears

#### Scenario: Factions appear in party view
- **WHEN** a DM views a party page
- **THEN** a factions section is visible with party-linked factions

### Requirement: Shops removed from campaign UI
Shops SHALL be moved from the Campaign view to the One-Shot view. The underlying shop table gains oneshot_adventure_id.

#### Scenario: Shops tab removed from campaign
- **WHEN** a user views the campaign page
- **THEN** no shops tab appears

#### Scenario: Shops appear in one-shot view
- **WHEN** a DM views a one-shot adventure
- **THEN** a shops section is visible with one-shot-linked shops
