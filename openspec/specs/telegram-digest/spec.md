# telegram-digest Specification

## Purpose
TBD - created by archiving change add-telegram-digest-and-cadence. Update Purpose after archive.
## Requirements
### Requirement: Announcements from Telegram

A DM or admin SHALL be able to post an announcement to the campaign chat bound to their currently resolved campaign with `/announce <text>`. The command SHALL resolve the caller's campaign with the same helper the rest of the bot uses, SHALL require a DM or admin role, SHALL reject empty text, and SHALL send the body HTML-escaped through the existing escaped-reply path. The bot SHALL always answer the caller: a confirmation when the announcement is posted, and a clear rejection when the chat is not bound, the campaign is not enabled, or the caller lacks the DM role.

#### Scenario: A DM announces to the bound campaign chat

- **WHEN** a DM sends `/announce Session moved to Thursday`
- **THEN** the bot posts the escaped announcement text to the campaign's bound chat and replies confirming it was posted

#### Scenario: Empty announcement is rejected

- **WHEN** a DM sends `/announce` with no text
- **THEN** the bot replies that an announcement needs text and posts nothing

#### Scenario: Rejected when the chat is not bound

- **WHEN** a DM sends `/announce` for a campaign with no bound Telegram chat
- **THEN** the bot replies that no chat is bound and posts nothing

#### Scenario: Rejected when the campaign is not enabled

- **WHEN** a DM sends `/announce` for a campaign whose Telegram posting is disabled
- **THEN** the bot replies that posting is disabled and posts nothing

#### Scenario: Non-DM is rejected

- **WHEN** a member without the DM role sends `/announce`
- **THEN** the bot replies that the command is DM-only and posts nothing

#### Scenario: HTML in the announcement is escaped

- **WHEN** a DM sends an announcement containing `<` or `&`
- **THEN** the posted text renders those characters literally rather than as markup

### Requirement: Weekly digest subscriptions

A linked user SHALL be able to subscribe to or unsubscribe from a weekly summary for a campaign with `/digest` and `/digest off`. Subscriptions SHALL be persisted per `(user, campaign)` with a frequency, a last-sent timestamp, and a created timestamp, and a user SHALL only be able to subscribe to or unsubscribe from campaigns they belong to. The scheduler SHALL send each due subscription at most once per week.

#### Scenario: Subscribe to a campaign digest

- **WHEN** a linked member of a campaign sends `/digest`
- **THEN** a subscription for that user and campaign is stored with weekly frequency and the bot confirms it

#### Scenario: Unsubscribe from a campaign digest

- **WHEN** a subscribed member sends `/digest off`
- **THEN** the subscription is removed and the bot confirms it

#### Scenario: Membership is required

- **WHEN** a user who does not belong to a campaign attempts to subscribe to it
- **THEN** the subscription is not stored and the bot replies that the user is not a member

#### Scenario: A due digest is sent

- **WHEN** the scheduler runs and a subscription has no last-sent timestamp or one older than a week
- **THEN** the weekly summary is sent to that user and the last-sent timestamp is recorded

#### Scenario: Not-yet-due digest is skipped

- **WHEN** the scheduler runs and a subscription was last sent less than a week ago
- **THEN** no digest is sent for that subscription
