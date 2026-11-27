# campaign-overview Specification

## Purpose
Rich campaign dashboard showing timeline, encounters, factions, shops, combat tracker, and linked one-shots.

## Requirements

### Requirement: Campaign overview is a primary view

The campaign overview SHALL be a full-page view (not a modal) accessible from the Party View and sidebar nav. It SHALL display: campaign info, linked one-shots, timeline events, encounters, factions, shops, combat tracker status, and party members.

#### Scenario: Open campaign overview
- **WHEN** user clicks "Campaign Overview" for a campaign
- **THEN** the full-page campaign overview view SHALL open showing all sections

#### Scenario: Campaign overview has lazy-loaded sections
- **WHEN** the campaign overview loads
- **THEN** sections (timeline, encounters, factions, shops, combat) SHALL load lazily via HTMX to prevent slow initial load

### Requirement: Campaign overview includes combat tracker integration

The campaign overview SHALL show recent combat encounters for the campaign, with links to open the combat tracker for each.

#### Scenario: Recent combats in overview
- **WHEN** user views campaign overview
- **THEN** the 5 most recent combat encounters for the campaign SHALL be listed with name and round count

### Requirement: Campaign overview includes encounters section

The campaign overview SHALL display encounters linked to the campaign's one-shots or directly to the campaign.

#### Scenario: Campaign encounters listed
- **WHEN** user views the encounters section of campaign overview
- **THEN** encounters associated with the campaign SHALL be listed with name, difficulty, and status

### Requirement: Campaign overview shows factions

The campaign overview SHALL display factions associated with the campaign or its one-shots.

#### Scenario: Campaign factions listed
- **WHEN** user views the factions section of campaign overview
- **THEN** factions with `campaign_id` matching the current campaign SHALL be listed

### Requirement: Campaign overview shows shops

The campaign overview SHALL display shops associated with the campaign.

#### Scenario: Campaign shops listed
- **WHEN** user views the shops section of campaign overview
- **THEN** shops with `campaign_id` matching the current campaign SHALL be listed
