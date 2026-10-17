# telegram-bot-runtime Specification

## Purpose
TBD - created by archiving change heat-style-telegram-bot. Update Purpose after archive.
## Requirements
### Requirement: Supervised bot runtime

The system SHALL run the Telegram bot through a supervised runtime that re-reads its stored settings periodically and after an admin saves settings, starting, stopping, or restarting the bot client as the token changes, without restarting the server process.

#### Scenario: Token added starts the bot

- **WHEN** an admin saves a bot token while no client is running
- **THEN** the bot client starts within the supervision interval and begins receiving updates

#### Scenario: Token cleared stops the bot

- **WHEN** an admin clears the bot token or disables the bot
- **THEN** the client stops consuming updates and sending is unavailable until a token is stored again

#### Scenario: Changed token restarts the bot

- **WHEN** an admin replaces the stored bot token
- **THEN** the old client is stopped and a new client with the new token starts within the supervision interval

#### Scenario: Unchanged settings keep the client running

- **WHEN** the supervisor re-reads settings that did not change
- **THEN** the running client and its in-flight long poll are left untouched

### Requirement: Transport modes

The system SHALL support `off`, `polling`, `webhook`, and `auto` transport modes. Long-polling SHALL be driven by the bot library; webhook updates SHALL be delivered through the existing webhook endpoint into the same running client. Webhook mode SHALL be available only when a public HTTPS base URL and a webhook secret are configured, and an unconfigured webhook SHALL NOT be registered.

#### Scenario: Polling consumes updates

- **WHEN** the mode resolves to `polling` with a stored token
- **THEN** the running client long-polls for updates and any registered webhook is removed

#### Scenario: Webhook mode receives updates

- **WHEN** the mode resolves to `webhook` and a valid update is posted to the webhook endpoint with the correct secret
- **THEN** the update is handed to the running client and its command is executed

#### Scenario: Auto resolves from the base URL

- **WHEN** the mode is `auto` and the configured base URL is public HTTPS
- **THEN** the effective mode is `webhook`; otherwise the effective mode is `polling`

#### Scenario: Update offset is tracked

- **WHEN** an update is processed, from polling or webhook
- **THEN** the highest processed update id is recorded and reported in the admin settings payload

### Requirement: Single Telegram settings surface

All Telegram settings, including the bot token, webhook secret, mode, grace minutes, update offset, and per-campaign delivery settings, SHALL be read and written through one typed settings surface. Secrets SHALL remain encrypted at rest, SHALL NOT be logged, and SHALL NOT be returned in full to any client. Environment overrides SHALL keep precedence over stored values.

#### Scenario: No direct settings SQL outside the surface

- **WHEN** settings are read by handlers, the scheduler, or delivery routing
- **THEN** they are read through the settings surface rather than ad-hoc queries

#### Scenario: Campaign settings through accessors

- **WHEN** campaign Telegram delivery settings are read or written by the web UI, scheduler, or delivery path
- **THEN** the same accessor pair is used and returns a typed result

#### Scenario: Secrets stay encrypted

- **WHEN** the bot token or webhook secret is stored
- **THEN** it is encrypted at rest and any client response masks it

### Requirement: Command registry and native menu

The bot SHALL maintain a single command registry that drives dispatch, the `/help` text, and the Telegram native command menu registered on every client start. Commands SHALL carry a category, usage text, and optional aliases.

#### Scenario: Help is generated from the registry

- **WHEN** a linked user sends `/help`
- **THEN** the reply lists the registered commands grouped by category with their usage

#### Scenario: Native menu is registered

- **WHEN** the bot client starts with a stored token
- **THEN** the command menu derived from the registry is registered with Telegram

#### Scenario: Alias resolves to its command

- **WHEN** a user sends a registered alias
- **THEN** the aliased command runs

#### Scenario: Unknown command gets help

- **WHEN** a linked user sends an unregistered command
- **THEN** the bot replies with the help text instead of silently ignoring it

#### Scenario: Command mention is stripped

- **WHEN** a command arrives as `/help@TheBotName`
- **THEN** it dispatches to `/help`

### Requirement: Inline keyboards and callbacks

Bot views SHALL be able to attach inline keyboards, and callback taps SHALL be answered and routed by their namespaced callback data. Callback handling SHALL enforce the same access rules as the equivalent command.

#### Scenario: Navigation keyboard runs a command

- **WHEN** a user taps a navigation button
- **THEN** the callback is answered and the corresponding command runs with its keyboard attached

#### Scenario: Callback needs no permission

- **WHEN** a user taps a button for a view they are allowed to see
- **THEN** the view is sent to that chat

#### Scenario: Tapping for foreign data is denied

- **WHEN** a callback would expose or modify data the user cannot access
- **THEN** the bot replies that the action is unavailable and changes nothing

### Requirement: Notification and status commands

The bot SHALL let a linked user toggle whether direct-message recap copies are delivered to them, and SHALL report bot and link status on request.

#### Scenario: Subscribe enables DMs

- **WHEN** a linked user sends `/subscribe`
- **THEN** their identity is marked as DM-enabled and the reply confirms it

#### Scenario: Unsubscribe disables DMs

- **WHEN** a linked user sends `/unsubscribe`
- **THEN** their identity is marked as not DM-enabled and the reply confirms it

#### Scenario: Status shows link and claim state

- **WHEN** a linked user sends `/status`
- **THEN** the reply reports the linked account, transport mode, and claimed character if any

#### Scenario: Unlinked user is guided to link

- **WHEN** an unlinked Telegram user sends any command other than the linking entry points
- **THEN** the bot replies with linking instructions and no account data

### Requirement: Sender availability independent of transport

Sending helpers SHALL work whenever a bot token is stored, even if the transport is stopped, by using the running client or a short-lived client built from the stored settings.

#### Scenario: Admin test message with transport off

- **WHEN** an admin sends a Telegram test message while no client is running but a token is stored
- **THEN** the message is delivered and reported as sent

#### Scenario: Delivery retries use the same path

- **WHEN** the recap delivery path sends a message while the supervisor is between client restarts
- **THEN** the send still uses the stored token and records success or failure as before
