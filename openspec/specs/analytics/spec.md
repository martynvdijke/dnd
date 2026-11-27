# analytics Specification

## Purpose
TBD - created by archiving change add-analytics-insights. Update Purpose after archive.
## Requirements
### Requirement: Accurate session cadence

The character stats endpoint SHALL report `sessions_per_month` as the session count divided by the number of whole months spanned by the character's first and last session dates, with a minimum divisor of one.

#### Scenario: Multiple months of play

- **WHEN** a character has 6 sessions spanning 3 months
- **THEN** `sessions_per_month` is 2

#### Scenario: Single session

- **WHEN** a character has 1 session
- **THEN** `sessions_per_month` is 1

#### Scenario: No sessions

- **WHEN** a character has no sessions
- **THEN** `sessions_per_month` is 0

### Requirement: Dice luck counts

The character stats endpoint SHALL report `natural_20s` and `natural_1s` as the number of stored d20 rolls whose result text contains a natural 20 or natural 1 respectively.

#### Scenario: Natural 20 recorded

- **WHEN** a character has a stored `1d20` roll with result text containing `[20]`
- **THEN** `natural_20s` is at least 1

#### Scenario: Natural 1 recorded

- **WHEN** a character has a stored `1d20` roll with result text containing `[1]`
- **THEN** `natural_1s` is at least 1

### Requirement: Combat analytics breakdowns

The combat-log stats endpoint SHALL include `damage_by_type` (total damage grouped by damage type, descending) and `top_healers` (total healing grouped by actor, descending, limited to five) alongside the existing totals.

#### Scenario: Damage by type

- **WHEN** a campaign's combat log contains fire and slashing damage
- **THEN** `damage_by_type` lists each type with its summed damage

#### Scenario: Healing by actor

- **WHEN** a campaign's combat log contains healing from multiple actors
- **THEN** `top_healers` lists the actors ordered by total healing

### Requirement: Analytics surface

The character analytics view SHALL display session cadence and dice luck, and the campaign dashboard SHALL display the combat analytics breakdowns.

#### Scenario: Character analytics

- **WHEN** a user opens the analytics view for a character
- **THEN** sessions per month and natural 20/1 counts are visible

#### Scenario: Campaign dashboard

- **WHEN** a user opens a campaign dashboard
- **THEN** a combat analytics card shows damage by type and top healers
