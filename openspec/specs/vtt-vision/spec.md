# vtt-vision Specification

## Purpose
TBD - created by archiving change add-vtt-vision. Update Purpose after archive.
## Requirements
### Requirement: Wall storage and editing
The system SHALL store line-of-sight walls per campaign map as a JSON array of segments in normalized 0..1 coordinates. A campaign member SHALL be able to read the walls and a campaign DM SHALL be able to replace them. Malformed wall payloads SHALL be rejected with 400 and SHALL NOT overwrite stored walls.

#### Scenario: DM saves walls
- **WHEN** a DM sends a valid array of wall segments to `PUT /api/maps/:id/walls`
- **THEN** the walls are persisted and returned by `GET /api/maps/:id/walls`

#### Scenario: Non-DM cannot edit walls
- **WHEN** a non-DM campaign member sends walls to `PUT /api/maps/:id/walls`
- **THEN** the request is rejected with 403

#### Scenario: Malformed walls are rejected
- **WHEN** a DM sends a payload that is not a JSON array of segments
- **THEN** the request is rejected with 400 and the stored walls are unchanged

### Requirement: Line-of-sight visibility
The battlemap SHALL compute a visibility polygon for a selected token from the map's walls and the token's vision radius, and SHALL render the area outside that polygon as dimmed. The visibility computation SHALL be performed client-side and SHALL only run when vision is enabled.

#### Scenario: Vision reveals the visible region
- **WHEN** vision is enabled and a token is selected on a map with walls
- **THEN** the board shows a lit visibility polygon and dims the remainder

#### Scenario: Vision is opt-in
- **WHEN** vision is not enabled
- **THEN** no visibility polygon is computed or rendered

#### Scenario: No walls means full visibility
- **WHEN** vision is enabled on a map with no walls
- **THEN** the visibility region is bounded only by the token's vision radius

### Requirement: Token auras
Each battlemap token SHALL support an aura radius expressed in grid cells. A token with a positive aura radius SHALL render a translucent circle of that radius around it on the battlemap.

#### Scenario: Aura renders at the configured radius
- **WHEN** a token has an aura radius of 2 grid cells
- **THEN** a translucent circle two cells wide is drawn around the token

#### Scenario: Zero aura renders nothing
- **WHEN** a token has an aura radius of 0
- **THEN** no aura circle is drawn for that token

### Requirement: Token condition badges
The battlemap SHALL render the conditions of a token linked to a combat entry as visible badges on the token.

#### Scenario: Conditions appear on the token
- **WHEN** a token's linked combat entry has conditions
- **THEN** each condition is shown as a badge on the token

#### Scenario: No conditions means no badges
- **WHEN** a token has no conditions
- **THEN** no condition badges are drawn
