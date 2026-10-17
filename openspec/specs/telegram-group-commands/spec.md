# telegram-group-commands Specification

## Purpose
TBD - created by archiving change telegram-group-commands. Update Purpose after archive.
## Requirements
### Requirement: Campaign context resolution

Campaign commands SHALL resolve their campaign from the chat context: a chat bound to a campaign through the campaign's Telegram settings with delivery enabled SHALL use that campaign; otherwise, in a private chat, the system SHALL fall back to the linked sender's campaigns using the existing single-campaign auto-select and multi-campaign picker behavior; when neither applies, the system SHALL explain how to connect a group or link an account instead of revealing any campaign data.

#### Scenario: Bound chat uses its campaign

- **WHEN** a member of a bound, enabled campaign chat runs a campaign command
- **THEN** the command renders data for that campaign

#### Scenario: Disabled binding is ignored

- **WHEN** a chat is bound to a campaign but delivery for that campaign is disabled
- **THEN** the command does not use that campaign and falls back to the remaining resolution rules

#### Scenario: Private chat falls back to membership

- **WHEN** a linked user in a private chat runs a campaign command
- **THEN** the command resolves the campaign from their memberships, auto-selecting a single campaign and offering a picker for several

#### Scenario: Unresolvable context explains the path

- **WHEN** a group chat is not bound to any enabled campaign and the sender is unlinked
- **THEN** the bot explains that a DM can connect the group from the campaign's Telegram settings and reveals no campaign data

### Requirement: Party items command

The bot SHALL provide an `/items` command that lists the resolved campaign's party inventory with each item's name, quantity, and notes, ordered by name, capped with a remainder note when the list is long.

#### Scenario: Party items are listed

- **WHEN** a member of a bound campaign chat sends `/items`
- **THEN** the reply lists each party item with quantity and any notes

#### Scenario: Empty inventory

- **WHEN** the resolved campaign has no party items
- **THEN** the reply says the party has no items yet

#### Scenario: Long inventory is capped

- **WHEN** the campaign has more items than the render cap
- **THEN** the reply shows the capped list and a note stating how many more items exist

### Requirement: Open quests command

The bot SHALL provide a `/quests` command that lists the open quests (status `available` or `active`) of the resolved campaign's party characters, grouped by character, including objectives and rewards when present, and SHALL distinguish quests with no party characters.

#### Scenario: Open quests are grouped by character

- **WHEN** a member of a bound campaign chat sends `/quests`
- **THEN** the reply lists each party character's open quests with objectives and rewards when present

#### Scenario: Completed quests are excluded

- **WHEN** a party character has a quest with status `complete`, `failed`, or `abandoned`
- **THEN** that quest does not appear in the open quests reply

#### Scenario: No open quests

- **WHEN** no party character has an open quest
- **THEN** the reply says there are no open quests

#### Scenario: No party characters

- **WHEN** the resolved campaign has no characters on its roster
- **THEN** the reply says the party has no characters yet

### Requirement: Location visits command

The bot SHALL provide a `/visits` command that lists the resolved campaign's party location links, newest first, showing the character, the location, the location type, and the relationship when it is not `visited`, capped at a bounded number of entries.

#### Scenario: Visits are listed newest first

- **WHEN** a member of a bound campaign chat sends `/visits`
- **THEN** the reply lists party location links ordered from most recent to oldest with character, location, and type

#### Scenario: Relationships are shown

- **WHEN** a location link relationship is not `visited`
- **THEN** the reply shows the relationship alongside the location

#### Scenario: No visits

- **WHEN** the campaign's party has no location links
- **THEN** the reply says no locations have been recorded yet

#### Scenario: Long visit history is capped

- **WHEN** the party has more location links than the render cap
- **THEN** the reply shows the capped list and a note stating how many more links exist

### Requirement: Context-aware statistics command

The bot SHALL render campaign statistics when `/stats` runs in a bound campaign chat and character statistics otherwise. Campaign statistics SHALL include character count, session count, quest totals with open and completed counts, NPC count, location count, and average party level rounded to one decimal, using the same semantics as the web analytics (NPCs and locations counted for campaign members).

#### Scenario: Campaign statistics in a bound chat

- **WHEN** a member of a bound campaign chat sends `/stats`
- **THEN** the reply shows the campaign's character, session, quest, NPC, and location counts plus the average party level

#### Scenario: Character statistics elsewhere

- **WHEN** a linked user with a claimed character sends `/stats` in a private chat
- **THEN** the reply shows that character's statistics as before

#### Scenario: Statistics respect campaign membership

- **WHEN** computing campaign statistics
- **THEN** NPCs and locations are counted for users who are members of the campaign and sessions are counted for the campaign's party characters

### Requirement: Group command output safety

Group command replies SHALL escape user-supplied text, SHALL be chunked to respect Telegram's message size limit, and SHALL be read-only.

#### Scenario: Hostile item name is escaped

- **WHEN** a party item or quest name contains HTML-significant characters
- **THEN** the reply renders the text literally without enabling markup

#### Scenario: Long reply is chunked

- **WHEN** a rendered group reply exceeds the message limit
- **THEN** it is delivered as multiple ordered messages without losing content
