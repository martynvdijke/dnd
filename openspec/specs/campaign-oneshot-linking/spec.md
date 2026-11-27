# campaign-oneshot-linking Specification

## Purpose
Link multiple one-shot adventures to a campaign, with ordering, status tracking, and campaign-level overview.

## Requirements

### Requirement: Campaign can own multiple one-shot adventures

A campaign SHALL support linking multiple one-shot adventures. Each one-shot adventure SHALL be orderable within the campaign. The campaign overview SHALL display all linked one-shots with their title, difficulty, estimated duration, and status.

#### Scenario: Link one-shot to campaign
- **WHEN** user creates a one-shot with `campaign_id` set to a valid campaign ID
- **THEN** the one-shot SHALL appear in that campaign's one-shot list

#### Scenario: List campaign one-shots
- **WHEN** user views a campaign overview
- **THEN** all one-shots linked to that campaign SHALL be listed in order with title, difficulty, and estimated minutes

#### Scenario: Reorder campaign one-shots
- **WHEN** user drags a one-shot to a new position in the campaign list
- **THEN** the one-shot's sort order SHALL be persisted and reflected in the campaign overview

### Requirement: Campaign overview shows linked one-shot status

The campaign overview SHALL display the completion/planning status of each linked one-shot (draft, planned, completed).

#### Scenario: One-shot status visible in campaign
- **WHEN** user views the campaign overview
- **THEN** each linked one-shot SHALL show its status badge

### Requirement: Quick navigation between campaign and its one-shots

The campaign overview SHALL provide buttons to open any linked one-shot's detail view directly.

#### Scenario: Navigate from campaign to one-shot
- **WHEN** user clicks a one-shot in the campaign overview
- **THEN** the app SHALL navigate to that one-shot's detail view
