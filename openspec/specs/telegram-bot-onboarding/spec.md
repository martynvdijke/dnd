# telegram-bot-onboarding Specification

## Purpose
TBD - created by archiving change telegram-group-commands. Update Purpose after archive.
## Requirements
### Requirement: Welcome on start

When `/start` is sent without a linking code, the bot SHALL reply with a short welcome that explains what the bot does, the steps to link an account, and what is available after linking, and SHALL attach the navigation keyboard.

#### Scenario: Start without a code welcomes

- **WHEN** a user sends `/start` without a code
- **THEN** the reply welcomes them, lists the linking steps, and includes navigation buttons

#### Scenario: Start with a code still links

- **WHEN** a user sends `/start <code>` with a valid code
- **THEN** the account is linked as before and the welcome is not sent

### Requirement: Friendly linking instructions

When an unlinked user runs a command that needs an account, the bot SHALL explain how to generate and redeem a link code, referencing the web UI location, in friendly, complete steps.

#### Scenario: Unlinked user gets instructions

- **WHEN** an unlinked user sends a command that requires linking
- **THEN** the reply explains the three linking steps and where to generate the code

#### Scenario: Instructions do not leak data

- **WHEN** an unlinked user runs any command
- **THEN** no account, character, or campaign data is included in the reply

### Requirement: Post-link guidance

After a successful link, the bot SHALL suggest claiming or creating a character and point at the help command.

#### Scenario: Successful link suggests next steps

- **WHEN** a user redeems a valid link code
- **THEN** the reply confirms the link and suggests `/claim`, `/create`, and `/help`

### Requirement: Group-added welcome

The bot SHALL subscribe to chat member updates and, when it is added to a group or supergroup, SHALL send a one-time welcome explaining that a DM can connect the group to a campaign and which commands become available. The welcome SHALL NOT be sent for ordinary member updates or when the bot was already a member.

#### Scenario: Added to a group welcomes

- **WHEN** the bot's membership changes from left, kicked, or absent to member, administrator, or creator in a group
- **THEN** the bot sends the group welcome with the campaign-connection explanation

#### Scenario: Ordinary member update is silent

- **WHEN** a chat member update does not add the bot to a group
- **THEN** the bot sends nothing

#### Scenario: Connected group is acknowledged

- **WHEN** the bot is added to a group whose chat is already bound to an enabled campaign
- **THEN** the welcome names the campaign and the available group commands
