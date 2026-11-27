# campaign-create-edit Specification

## Purpose
TBD - created by archiving change campaign-management-roster. Update Purpose after archive.
## Requirements
### Requirement: Campaign creation includes character roster selection

The campaign creation flow SHALL let the user enter the campaign name, party name, description, and DM notes, and select one or more characters to attach to the campaign. The character selector SHALL support choosing multiple characters at once, drawn from the roster candidates endpoint. Creating the campaign SHALL persist the campaign fields and create a membership for every selected character in the new campaign.

#### Scenario: Create campaign with player-controlled characters

- **WHEN** a user creates a campaign and selects one or more of their own characters
- **THEN** the campaign is created and a membership in the new campaign is created for each selected character

#### Scenario: Create campaign with external characters

- **WHEN** a user creates a campaign and selects characters owned by other members of the campaign alongside their own
- **THEN** all selected characters (own and external) become members of the new campaign

#### Scenario: Create campaign with no characters

- **WHEN** a user creates a campaign without selecting any characters
- **THEN** the campaign is created with an empty roster and characters can be added later

#### Scenario: Selected characters keep their other campaigns

- **WHEN** a user creates a campaign and selects a character that already belongs to another campaign
- **THEN** the new membership is created and the character's other memberships are unchanged

### Requirement: Campaign edit manages the roster

The campaign management flow SHALL show the campaign's current roster and SHALL allow any member to add characters from the candidates list and remove characters from the roster. Characters already in the roster SHALL be pre-selected in the picker. Roster changes SHALL persist immediately via the roster endpoints and SHALL affect only the managed campaign's memberships.

#### Scenario: Add a character from the manage view

- **WHEN** a campaign member opens the campaign management view and selects a character not yet in the roster
- **THEN** a membership in this campaign is created and the character appears in the roster list

#### Scenario: Remove a character from the manage view

- **WHEN** a campaign member removes a character from the roster in the campaign management view
- **THEN** this campaign's membership is removed, the character disappears from the roster list, and its other memberships are unchanged

#### Scenario: Existing roster is pre-selected

- **WHEN** a campaign member opens the campaign management view
- **THEN** all characters currently in the campaign are shown as selected in the roster picker

### Requirement: Campaign form exposes DM notes

The campaign create and edit forms SHALL include a DM notes field. The value SHALL be persisted via the existing campaign create and update endpoints using the `dm_notes` property, and SHALL be shown pre-filled when editing an existing campaign.

#### Scenario: Save DM notes on create

- **WHEN** a user creates a campaign with DM notes entered
- **THEN** the notes are stored on the campaign

#### Scenario: Edit DM notes

- **WHEN** a user edits an existing campaign and changes the DM notes
- **THEN** the stored DM notes are updated and the form shows the saved value on the next open

### Requirement: Roster picker distinguishes player-controlled and external characters

The roster picker SHALL label candidates as player-controlled (owned by the current user) or external (owned by another campaign member), showing the owner's username for external characters. Both groups SHALL be selectable in the same picker.

#### Scenario: Own characters are labeled as player-controlled

- **WHEN** the roster picker lists a character owned by the current user
- **THEN** it is grouped/labeled as player-controlled

#### Scenario: Other members' characters are labeled as external

- **WHEN** the roster picker lists a character owned by another campaign member
- **THEN** it is grouped/labeled as external with the owner's username shown
