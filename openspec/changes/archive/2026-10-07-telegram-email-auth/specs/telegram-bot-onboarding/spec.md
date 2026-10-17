## MODIFIED Requirements

### Requirement: Welcome on start

When `/start` is sent without a linking code, the bot SHALL reply with a short welcome that explains what the bot does, the steps to link an account — including the in-chat email sign-in path via `/login` alongside the web code flow — and what is available after linking, and SHALL attach the navigation keyboard.

#### Scenario: Start without a code welcomes

- **WHEN** a user sends `/start` without a code
- **THEN** the reply welcomes them, lists the linking steps including `/login`, and includes navigation buttons

#### Scenario: Start with a code still links

- **WHEN** a user sends `/start <code>` with a valid code
- **THEN** the account is linked as before and the welcome is not sent

### Requirement: Friendly linking instructions

When an unlinked user runs a command that needs an account, the bot SHALL explain how to link, in friendly, complete steps, covering both the in-chat email sign-in with `/login` and the web code flow that generates and redeems a link code.

#### Scenario: Unlinked user gets instructions

- **WHEN** an unlinked user sends a command that requires linking
- **THEN** the reply explains email sign-in with `/login` and the web code flow

#### Scenario: Instructions do not leak data

- **WHEN** an unlinked user runs any command
- **THEN** no account, character, or campaign data is included in the reply
