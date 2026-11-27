# dm-screen Specification

## Purpose
TBD - created by archiving change add-dm-tooling. Update Purpose after archive.
## Requirements
### Requirement: Campaign DM screen aggregate

The system SHALL provide `GET /api/campaigns/:id/dm-screen` restricted to the campaign's DM. The response SHALL include `campaign` (`id`, `name`, `party_name`, `dm_notes`), `session` (the most recent session plan with status `planned`, `ready`, or `in-progress`, or null), `combat` (`active`, `current_turn`, `entries`), `party` (each with `id`, `name`, `hp_current`, `hp_max`, `ac`, `conditions`), and `ambience` (`track`, `action`).

#### Scenario: DM reads the screen

- **WHEN** the campaign DM requests `/api/campaigns/:id/dm-screen`
- **THEN** the response contains `campaign`, `session`, `combat`, `party`, and `ambience` keys

#### Scenario: Non-DM is rejected

- **WHEN** a non-DM member requests `/api/campaigns/:id/dm-screen`
- **THEN** the response status is `403`

### Requirement: DM screen surface

The SPA SHALL expose a DM Screen modal launched from the campaign dashboard that renders the aggregate response and offers a manual refresh. The modal SHALL refresh automatically while open and SHALL stop refreshing when closed.

#### Scenario: Open the DM screen

- **WHEN** the DM opens the DM Screen from the campaign dashboard
- **THEN** the modal shows the party, the current combat turn, the active session plan, and ambience controls
