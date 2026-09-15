## ADDED Requirements

### Requirement: World Overview view

The system SHALL provide a top-level World Overview view, reachable from the campaign navigation (top nav, sidebar, and the mobile "more" menu) and addressable by a URL hash, that presents the campaign's places and their timeline.

#### Scenario: Open the world view
- **WHEN** a user with an active campaign opens World from the navigation
- **THEN** the world view is shown with its map and place directory, and the URL hash reflects the view

#### Scenario: No campaign selected
- **WHEN** World is opened without an active campaign
- **THEN** the view is not entered and the user is prompted to select a campaign

### Requirement: Place directory

The world view SHALL list places with their name, type, and description, grouped by place type and by parent place where a parent-child relationship exists. Selecting a place SHALL show its detail.

#### Scenario: Groups reflect hierarchy
- **WHEN** a campaign has a region containing cities
- **THEN** the directory groups those cities under their parent region

#### Scenario: Select a place
- **WHEN** the user selects a place in the directory
- **THEN** the detail panel shows that place's name, type, description, and coordinates

### Requirement: Place map

The world view SHALL render a map of all places that have coordinates, with one marker per place, and selecting a marker SHALL select the corresponding place. Places without coordinates SHALL still appear in the directory.

#### Scenario: Places with coordinates are mapped
- **WHEN** the campaign has places with latitude/longitude
- **THEN** each is shown as a marker and the map fits to their bounds

#### Scenario: Places without coordinates are listed only
- **WHEN** a place has no coordinates
- **THEN** it appears in the directory but not as a map marker

### Requirement: Place-to-timeline linkage

The timeline create and edit forms SHALL offer a place selector, and submitting SHALL persist the selected place as the event's `linked_entity_type = "location"` and `linked_entity_id`. A place's detail SHALL list timeline events linked to it, and each such event SHALL link back to the place.

#### Scenario: Link an event to a place
- **WHEN** the DM creates or edits a timeline event and selects a place
- **THEN** the event is saved with that place's identifier as its linked entity

#### Scenario: Place shows its events
- **WHEN** a place has linked timeline events
- **THEN** the place detail lists those events and selecting one navigates to the event

#### Scenario: Event shows its place
- **WHEN** a timeline event is linked to a place
- **THEN** the timeline entry shows a link that opens the world view at that place

#### Scenario: Unlink a place
- **WHEN** the DM clears the place selector and saves
- **THEN** the event no longer has a linked place

### Requirement: World data visibility

The world view SHALL only expose places and events the requesting user is permitted to see within the active campaign, and unauthenticated requests SHALL be rejected.

#### Scenario: Only campaign-relevant places
- **WHEN** a campaign member opens World
- **THEN** only places linked to that campaign's characters, plus campaign timeline events, are shown

#### Scenario: Unauthenticated access
- **WHEN** a request for world data is made without a valid session
- **THEN** the request is rejected
