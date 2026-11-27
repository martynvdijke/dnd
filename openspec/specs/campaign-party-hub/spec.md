# campaign-party-hub Specification

## Purpose
Expanded party view with campaign-level locations, NPCs, quests, sessions, graph, and analytics tabs.

## Requirements

### Requirement: Party view with campaign sub-tabs

The party view (`#partyView`) SHALL include a sub-tab bar with the following tabs: Overview, Locations, NPCs, Sessions, Quests, Graph, Analytics. Each sub-tab SHALL render its content into `#partyContent` when selected. The Overview tab SHALL be the default and show the existing party/campaign management UI (character cards, party inventory, etc.).

#### Scenario: Party view shows sub-tab bar

- **WHEN** the party view is opened
- **THEN** the view SHALL display a sub-tab bar with: Overview, Locations, NPCs, Sessions, Quests, Graph, Analytics

#### Scenario: Overview tab shows campaign management

- **WHEN** the party view opens and the Overview tab is selected
- **THEN** it SHALL show the existing campaign/party management content (character cards, buttons)

#### Scenario: Sub-tabs navigate without page reload

- **WHEN** the user clicks a sub-tab (e.g., Locations)
- **THEN** the content SHALL switch to that tab's rendered content without a full page reload

### Requirement: Campaign locations view

The Locations sub-tab in the party view SHALL display all locations linked to campaign member characters. Locations SHALL be grouped by campaign and displayed with their linked character names. The view SHALL include a map (Leaflet) if any location has coordinates, or a list view otherwise.

#### Scenario: Locations rendered from campaign members

- **WHEN** the user selects the Locations sub-tab
- **THEN** it SHALL fetch locations for all characters in the current campaign and render them grouped by character

#### Scenario: Map view for geolocated locations

- **WHEN** any campaign location has latitude/longitude coordinates
- **THEN** the Locations tab SHALL render a Leaflet map with pins for those locations

### Requirement: Campaign NPCs view

The NPCs sub-tab in the party view SHALL display all NPCs linked to campaign member characters. NPCs SHALL be listed grouped by character, with the ability to filter/search across all campaign NPCs. Each NPC entry SHALL show name, relationship, and a link to view details.

#### Scenario: NPCs rendered from campaign members

- **WHEN** the user selects the NPCs sub-tab
- **THEN** it SHALL fetch NPCs for all characters in the current campaign and render them grouped by character

#### Scenario: Campaign-wide NPC search

- **WHEN** the user types in a search box in the NPCs tab
- **THEN** NPCs from all campaign members SHALL be filtered by name

### Requirement: Campaign sessions view

The Sessions sub-tab in the party view SHALL display session logs for the campaign. Sessions from all campaign member characters SHALL be shown in reverse chronological order. Each session entry SHALL show the date, session title/notes, and which character it belongs to.

#### Scenario: Sessions rendered chronologically

- **WHEN** the user selects the Sessions sub-tab
- **THEN** it SHALL fetch sessions for all campaign member characters and render them in reverse chronological order

### Requirement: Campaign quests view

The Quests sub-tab in the party view SHALL display quests for the campaign. Quests from all campaign member characters SHALL be shown grouped by status (active, completed, failed). Each quest SHALL show title, description, and which character it's linked to.

#### Scenario: Quests rendered grouped by status

- **WHEN** the user selects the Quests sub-tab
- **THEN** it SHALL fetch quests for all campaign member characters and render them grouped by status

### Requirement: Campaign graph view

The Graph sub-tab in the party view SHALL display a campaign-wide relationship graph (D3 force-directed layout) showing connections between campaign characters, NPCs, and locations across all campaign members.

#### Scenario: Campaign graph includes all entities

- **WHEN** the user selects the Graph sub-tab
- **THEN** it SHALL render a D3 force-directed graph showing characters, NPCs, and locations from the campaign with their connections

### Requirement: Campaign analytics view

The Analytics sub-tab in the party view SHALL display campaign-wide statistics aggregated across all campaign member characters, including total sessions played, quests completed, XP earned, and other aggregate metrics.

#### Scenario: Analytics rendered from campaign data

- **WHEN** the user selects the Analytics sub-tab
- **THEN** it SHALL display campaign-wide stats (total sessions, quests completed, etc.) aggregated from all campaign member characters
