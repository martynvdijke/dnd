# trmnl-combat-status Specification

## Purpose
TBD - created by archiving change add-trmnl-combat-status-display. Update Purpose after archive.
## Requirements
### Requirement: Combat status polling endpoint
The system SHALL expose a public read-only endpoint `GET /api/trmnl/combat?campaign_id=...` that returns the combat tracker state for a campaign. Like the other TRMNL polling endpoints it SHALL be public (no credentials required) and read-only. Entries SHALL be ordered by initiative roll descending, then turn order ascending, matching the web combat tracker.

#### Scenario: Public read-only access

- WHEN the endpoint is polled without any credentials
- THEN it responds 200 with combat data and performs no writes

#### Scenario: Campaign with no combat

- WHEN the endpoint is polled for a campaign with no combat entries, or without a `campaign_id`
- THEN it responds 200 with an empty entry list rather than an error

### Requirement: Combat payload contents
Each entry in the combat payload SHALL include name, type, total initiative (roll + modifier), current HP, max HP, AC, active-turn flag, and conditions resolved against the built-in 5e condition types (`ConditionTypes`). Tokens in the stored `condition_ids` CSV that do not match a known condition SHALL be skipped without failing the request.

#### Scenario: Conditions are resolved to names

- WHEN a combat entry's `condition_ids` contains known condition names
- THEN the payload contains their canonical names from the built-in condition types

#### Scenario: Unknown condition token is skipped

- WHEN a combat entry's `condition_ids` contains a token that is not a known condition
- THEN that token is omitted from the payload and the request still succeeds

### Requirement: Combat display mode in TRMNL plugin
The TRMNL plugin SHALL offer a `combat` display mode via a new polling URL entry and Liquid blocks in all four layout templates. The active combatant SHALL be rendered visually distinct from other entries. When the payload has no entries, templates SHALL render an explicit idle message instead of a blank screen.

#### Scenario: Active combatant is distinct

- WHEN the combat template renders an entry with the active flag set
- THEN that row is styled distinctly (inverted/bold) from non-active rows

#### Scenario: Idle screen

- WHEN the polled payload contains zero entries
- THEN the template renders an explicit "No active combat" message

#### Scenario: Overflow on small layouts

- WHEN more entries exist than the half/quadrant layouts can render
- THEN the template caps the list and indicates the remaining count
