## ADDED Requirements

### Requirement: Live dice relay
When a campaign member rolls dice for a character that belongs to a campaign, the system SHALL broadcast the roll result to every connected member of that campaign. Rolls not tied to a campaign character SHALL NOT be broadcast.

#### Scenario: Player roll reaches the table
- **WHEN** a player rolls `1d20+5` for a character in campaign C
- **THEN** every connected member of C receives a `dice_roll` event with the roller, character, expression, and total

#### Scenario: Non-campaign roll stays private
- **WHEN** a user rolls for a character with no campaign
- **THEN** no broadcast is emitted

### Requirement: Live combat sync
When a combat entry is created, updated, deleted, or the turn advances, the system SHALL notify the campaign's connected members so their view reflects the current turn without polling.

#### Scenario: DM advances the turn
- **WHEN** the DM triggers next turn for campaign C
- **THEN** connected members of C receive a `combat_update` event and refresh their combat view

#### Scenario: HP change propagates
- **WHEN** the DM updates a combat entry's HP in campaign C
- **THEN** connected members of C receive a `combat_update` event

### Requirement: Live handout reveal
When a knowledge entry becomes shared or is bulk-revealed, the system SHALL broadcast the entry to the campaign's connected members. Unshared entries SHALL NOT be broadcast to members.

#### Scenario: DM shares a rumor
- **WHEN** the DM sets `shared = true` on a knowledge entry in campaign C
- **THEN** connected members of C receive a `knowledge_reveal` event containing the entry

#### Scenario: Unshared entry never broadcast
- **WHEN** a knowledge entry remains `shared = false`
- **THEN** no `knowledge_reveal` event is emitted for it

### Requirement: Access-scoped delivery
Live events SHALL be delivered only to authenticated campaign members. A user who is not a member of the campaign SHALL NOT receive its events.

#### Scenario: Stranger receives nothing
- **WHEN** a user who is not a member of campaign C is connected to the WebSocket
- **THEN** they receive no `dice_roll`, `combat_update`, or `knowledge_reveal` event from C

### Requirement: Player table view
The system SHALL provide a player-facing Table view that shows the current character summary, a dice tray, live initiative, and revealed handouts, and SHALL hide DM-only controls while active.

#### Scenario: Player opens table view
- **WHEN** a player enables table view
- **THEN** they see their character, the dice tray, live initiative, and revealed handouts, with no DM editing controls

#### Scenario: Handout appears live
- **WHEN** the DM reveals a handout while a player is in table view
- **THEN** the handout appears for the player without a manual refresh
