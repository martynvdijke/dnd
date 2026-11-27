# character-selection Specification

## Purpose
TBD - created by archiving change character-selection-ownership. Update Purpose after archive.
## Requirements
### Requirement: Campaign-first selection after login

For campaigns the user plays in, the system SHALL present a campaign selection step before character selection. After login, the user SHALL first choose a campaign. Selecting a campaign the user runs (owner, or `my_role` equal to `dm`) SHALL enter that campaign's Party View directly without requiring a character. Selecting a campaign the user only plays in SHALL show that campaign's characters. A user with no campaigns SHALL see an empty state that includes an action to create a campaign.

#### Scenario: User logs in with campaigns

- **WHEN** a logged-in user opens the app and has at least one campaign available
- **THEN** the system SHALL show a campaign picker listing the user's available campaigns

#### Scenario: User has no campaigns

- **WHEN** a logged-in user has no campaigns available
- **THEN** the system SHALL show an empty state with guidance to create or join a campaign and an action to create a campaign

#### Scenario: DM selects a campaign they run

- **WHEN** the user selects a campaign where they are the owner or have the `dm` role
- **THEN** the system enters the campaign's Party View without showing a character picker

#### Scenario: Player selects a campaign

- **WHEN** the user selects a campaign they only play in
- **THEN** the system SHALL show the characters belonging to that campaign, both owned and shared

#### Scenario: Create campaign from the empty state

- **WHEN** a user with no campaigns uses the create-campaign action
- **THEN** the campaign is created, selected as DM, and the user lands in its Party View

### Requirement: Selection is persisted for the session

The system SHALL remember the currently selected campaign and character so that navigation between views keeps the selection. The selection SHALL be stored client-side and SHALL be cleared on logout.

#### Scenario: Selection survives navigation

- **WHEN** the user has selected a campaign and a character and navigates to another view
- **THEN** the selection SHALL remain active and scoped to that campaign

#### Scenario: Selection cleared on logout

- **WHEN** the user logs out
- **THEN** the stored campaign and character selection SHALL be cleared

### Requirement: Character available only via a campaign

A character SHALL be selectable in the context of any campaign it belongs to. If a character belongs to no campaign, the system SHALL make it available under an "Unassigned" group when a campaign is selected. A character SHALL appear in the character picker of every campaign it belongs to.

#### Scenario: Character without campaign

- **WHEN** a user owns a character that belongs to no campaign
- **THEN** the system SHALL surface that character under an unassigned grouping so it is not lost

#### Scenario: Character in a campaign appears in that campaign's picker

- **WHEN** a character belongs to a campaign
- **THEN** the character SHALL appear in that campaign's character list

#### Scenario: Character in multiple campaigns appears in each picker

- **WHEN** a character belongs to two campaigns
- **THEN** the character SHALL appear in each campaign's character list

### Requirement: Party membership supports multiple characters

A campaign party SHALL support one or more characters, and the selection flow SHALL allow switching between any character that is part of the campaign.

#### Scenario: Multiple party members

- **WHEN** a campaign has more than one character in its party
- **THEN** the user SHALL be able to select any of those characters as the active character

### Requirement: DM campaigns support character-less operation

A user running a campaign SHALL be able to operate with no active character selected. The character picker SHALL offer a "Continue as DM" action for campaigns the user runs, the campaign-scoped views SHALL remain usable without an active character, and character-only views SHALL show an explicit empty state directing the user to pick a character from the Party View instead of rendering blank or failing.

#### Scenario: Continue as DM from the character picker

- **WHEN** a DM opens the character picker for a campaign they run
- **THEN** they can continue to the Party View without selecting a character

#### Scenario: Character-only view without an active character

- **WHEN** a DM with no active character opens a character-only view
- **THEN** the view shows guidance to choose a character from the Party View and does not fail

#### Scenario: Character-required actions without an active character

- **WHEN** an action requires an active character and none is selected
- **THEN** the action is blocked with a clear message rather than crashing

### Requirement: Switching campaigns follows the same role logic

Switching campaigns through the campaign-context controls SHALL apply the same role-aware entry as the initial selection: campaigns the user runs lead to the Party View, campaigns the user plays in lead to the character picker with a way back to the campaign list.

#### Scenario: Switch to a campaign run by the user

- **WHEN** the user switches to a campaign they run
- **THEN** the Party View opens without an intermediate character step

#### Scenario: Switch back from a DM campaign to a player campaign

- **WHEN** the user switches from a campaign they run to one they play in
- **THEN** the character picker for that campaign is shown
