# oneshot-mini-campaign Specification

## Purpose
One-shot mode with campaign-like features — timeline, encounters, factions, shops scoped to the one-shot.

## Requirements

### Requirement: One-shot can function as a mini-campaign

A one-shot adventure SHALL support a "mini campaign" mode where it has its own scoped timeline events, encounters, factions, and shops. These entities SHALL be filtered by `oneshot_adventure_id` rather than campaign ID.

#### Scenario: Toggle mini-campaign mode
- **WHEN** user enables "Mini Campaign" on a one-shot
- **THEN** the one-shot detail view SHALL show additional sections: Timeline, Encounters, Factions, Shops

#### Scenario: One-shot timeline events
- **WHEN** user adds a timeline event to a one-shot in mini-campaign mode
- **THEN** the event SHALL be scoped to that one-shot and visible only in its timeline

#### Scenario: One-shot encounters
- **WHEN** user views encounters in a one-shot mini-campaign
- **THEN** encounters SHALL be filtered to those linked to the one-shot's acts/scenes

#### Scenario: One-shot factions
- **WHEN** user adds a faction to a one-shot mini-campaign
- **THEN** the faction SHALL be scoped to that one-shot (linked via `oneshot_adventure_id`)

#### Scenario: One-shot shops
- **WHEN** user views shops in a one-shot mini-campaign
- **THEN** shops SHALL be filtered to those with matching `oneshot_adventure_id`

### Requirement: Shops null-safety fix

The shops tab SHALL safely handle null/undefined responses without crashing.

#### Scenario: Shops API returns null
- **WHEN** `GET /api/shops` returns null or undefined
- **THEN** the UI SHALL display "No shops" empty state instead of crashing

#### Scenario: Shops API returns empty array
- **WHEN** `GET /api/shops` returns an empty array
- **THEN** the UI SHALL display "No shops" empty state
