# telegram-character-commands Specification

## Purpose
TBD - created by archiving change heat-style-telegram-bot. Update Purpose after archive.
## Requirements
### Requirement: Character claiming

A linked user SHALL be able to claim one character they are allowed to edit as their table character, either by sending `/claim` and tapping a candidate or by sending `/claim <id|name>`. A character SHALL be claimable by at most one Telegram user, a Telegram user SHALL hold at most one claim, and `/unclaim` SHALL release the claim. Claims SHALL be cleared when the Villum account is unlinked and SHALL be re-validated on every read, becoming unavailable if the linked user loses edit rights to the character.

#### Scenario: Claim from the candidate keyboard

- **WHEN** a linked user sends `/claim`
- **THEN** the bot lists the claimable characters the user may edit and binds the one the user taps

#### Scenario: Claim by argument

- **WHEN** a linked user sends `/claim <id>` or `/claim <name>`
- **THEN** the character is claimed when it is claimable and the reply confirms it

#### Scenario: Already claimed character is refused

- **WHEN** a user tries to claim a character already claimed by a different Telegram user
- **THEN** the claim is refused and the existing claim is untouched

#### Scenario: Unclaim releases the character

- **WHEN** a linked user sends `/unclaim`
- **THEN** the claim is removed and the reply confirms it

#### Scenario: Unlink clears the claim

- **WHEN** a linked user unlinks their Telegram account
- **THEN** the claim is removed and the character becomes claimable again

#### Scenario: Lost edit rights invalidate a claim

- **WHEN** a claimed character is no longer editable by the linked user
- **THEN** claim-dependent replies report the claim as unavailable and suggest `/unclaim`

#### Scenario: Unlinked user cannot claim

- **WHEN** an unlinked Telegram user sends `/claim`
- **THEN** the bot replies with linking instructions and nothing is bound

### Requirement: Character listing command

The bot SHALL list the linked user's characters with their name, race, class, level, hit points, and campaign, marking the claimed character. The listing SHALL be restricted to characters the user is allowed to see.

#### Scenario: Characters are listed

- **WHEN** a linked user with characters sends `/characters`
- **THEN** the reply lists them with race, class, level, hit points, and campaign

#### Scenario: Claimed character is marked

- **WHEN** the listing includes the claimed character
- **THEN** that entry is marked as claimed

#### Scenario: No characters

- **WHEN** a linked user with no characters sends `/characters`
- **THEN** the bot suggests creating a character with `/create`

#### Scenario: Other users' characters are absent

- **WHEN** a linked user is a member of a campaign containing another user's private character
- **THEN** that character does not appear in the listing unless the user may edit it

### Requirement: Character sheet command

The bot SHALL render a character sheet summary for the claimed character, or for a character given by id or name, including identity (race, class, subclass, level), ability scores with modifiers, hit points, armor class, initiative, speed, proficiency bonus, passive perception, and counters for features, spells, and inventory. The sheet's other-content section SHALL point at the tracking commands for inventory, spells, conditions, features and currency. Rendered content SHALL escape user-supplied text.

#### Scenario: Sheet defaults to the claim

- **WHEN** a linked user with a claimed character sends `/sheet`
- **THEN** the reply renders that character's sheet summary

#### Scenario: Sheet by name

- **WHEN** a linked user sends `/sheet <name>` for a character they may edit
- **THEN** the reply renders that character's sheet summary

#### Scenario: Foreign character is denied

- **WHEN** a linked user requests a sheet for a character they may not edit
- **THEN** the bot replies with the same "not found" response as for a nonexistent character and reveals no sheet data

#### Scenario: Unclaimed user is guided

- **WHEN** a linked user without a claim sends `/sheet` and owns several characters
- **THEN** the bot replies with a selection list instead of guessing a character

#### Scenario: Hostile text is escaped

- **WHEN** a character name or trait contains HTML-significant characters
- **THEN** the message renders the text literally without enabling markup

#### Scenario: Other section points at tracking

- **WHEN** the sheet renders its other-content counters
- **THEN** it names the tracking commands for inventory, spells, conditions, features and currency

### Requirement: Character stats command

The bot SHALL render a glanceable stats block for the claimed character, or for a character given by id or name: ability scores with modifiers, hit points, armor class, initiative, speed, proficiency bonus, and passive perception.

#### Scenario: Stats default to the claim

- **WHEN** a linked user with a claimed character sends `/stats`
- **THEN** the reply renders that character's stats block

#### Scenario: Stats by id

- **WHEN** a linked user sends `/stats <id>` for a character they may edit
- **THEN** the reply renders that character's stats block

#### Scenario: Foreign character is denied

- **WHEN** a linked user requests stats for a character they may not edit
- **THEN** the bot replies with the same "not found" response as for a nonexistent character and reveals no stat data

### Requirement: Campaign overview command

The bot SHALL render an overview restricted to campaigns the sender belongs to. Without an argument it SHALL list the sender's campaigns with their role and character count; with a campaign argument it SHALL show the campaign's members, party characters with classes and levels, and the latest recap.

#### Scenario: Campaign list

- **WHEN** a linked user sends `/overview`
- **THEN** the reply lists their campaigns with role and character count

#### Scenario: Campaign detail

- **WHEN** a linked member sends `/overview <campaign>`
- **THEN** the reply shows the party composition and the latest recap

#### Scenario: Non-member is denied

- **WHEN** a linked user requests an overview for a campaign they do not belong to
- **THEN** the bot replies with the same "not found" response as for a nonexistent campaign and reveals no campaign data

#### Scenario: Selecting a campaign from the keyboard

- **WHEN** a user taps a campaign button on the overview list
- **THEN** the detail view for that campaign is sent

### Requirement: Guided character creation command

The bot SHALL create a character through a guided flow on `/create`: name, race, class, level (1–20), and an optional campaign chosen from the campaigns the user belongs to. The created character SHALL be owned by the linked Villum user, SHALL be created through the server's character-creation path so validation, defaults, currency, and campaign membership behave exactly as in the web app, and SHALL be claimed automatically. `/cancel` and any other command SHALL abort an in-progress flow.

#### Scenario: Full flow creates and claims

- **WHEN** a linked user answers name, race, class, and level
- **THEN** the character is created with the same defaults as the web creation path and is claimed by the user

#### Scenario: Inline level picker

- **WHEN** the flow reaches the level step
- **THEN** the bot offers level choices and validates any typed level to the allowed range

#### Scenario: Campaign selection

- **WHEN** the flow reaches the campaign step and the user belongs to campaigns
- **THEN** the bot offers the campaigns plus a "none" choice and attaches the character to the chosen campaign

#### Scenario: Invalid input is rejected in place

- **WHEN** the user answers a step with an empty or invalid value
- **THEN** the bot re-asks that step without creating anything

#### Scenario: Flow can be aborted

- **WHEN** the user sends `/cancel` or starts another command mid-flow
- **THEN** the flow is discarded and no character is created

#### Scenario: Unlinked user cannot create

- **WHEN** an unlinked Telegram user sends `/create`
- **THEN** the bot replies with linking instructions and no character is created

#### Scenario: Creation failure is reported

- **WHEN** the server's creation path rejects the collected values
- **THEN** the bot reports the failure and creates nothing
