# web-push-notifications Specification

## Purpose
TBD - created by archiving change add-web-push-notifications. Update Purpose after archive.
## Requirements
### Requirement: Push subscription lifecycle
The system SHALL allow an authenticated user to register a push subscription (endpoint, keys, expiration) and SHALL store it keyed by the unique endpoint. Re-subscribing the same endpoint SHALL update the existing row. An authenticated user SHALL be able to delete their subscription.

#### Scenario: Successful subscription
- **WHEN** an authenticated user POSTs a valid subscription payload
- **THEN** the system stores it and responds 201

#### Scenario: Duplicate endpoint updates in place
- **WHEN** a subscription arrives whose endpoint already exists
- **THEN** the existing row is updated rather than duplicated

#### Scenario: Unauthenticated subscription is rejected
- **WHEN** the subscribe endpoint is called without a session
- **THEN** it responds 401

### Requirement: VAPID configuration
Administrators SHALL configure VAPID keys and subject through the admin settings. Keys SHALL auto-generate if absent when first saved. The system SHALL provide a test-push action targeting the administrator's own subscriptions. Private keys SHALL NOT appear in logs or non-admin responses.

#### Scenario: Admin saves keys and receives test push
- **WHEN** an admin generates keys and clicks "send test push" with a subscribed browser
- **THEN** that browser displays the test notification

#### Scenario: Non-admin cannot read private key
- **WHEN** a non-admin requests the push settings
- **THEN** the response is 403

### Requirement: Recap published notification
When a campaign recap is published/sent, the system SHALL send a push notification to all campaign members with active subscriptions who have not muted the campaign. The notification SHALL deep-link to the recap. Delivery failures with 404/410 SHALL delete the subscription.

#### Scenario: Members receive recap push
- **WHEN** a recap is published for a campaign with subscribed, unmuted members
- **THEN** each such member's subscriptions receive a notification linking to the recap

#### Scenario: Muted member receives nothing
- **WHEN** a member has muted the campaign
- **THEN** no notification is sent to any of their subscriptions

#### Scenario: Dead subscription is pruned
- **WHEN** the push service responds 404 or 410 for an endpoint
- **THEN** that subscription row is deleted

### Requirement: Session reminder notification
The system SHALL run a background scheduler that sends session reminders to campaign members with active subscriptions, covering two event sources with separate lead settings: external calendar feed events (minute-precision, default 60 minutes before start) and local calendar sessions (`event_type = 'session'`, day-granular "Session coming up" reminder within a configurable number of days, default 1). Each event SHALL trigger at most one reminder per member.

#### Scenario: External feed reminder sent once before event
- **WHEN** the scheduled time (a timed, non-all-day feed event's start minus the minute lead) is reached
- **THEN** subscribed, unmuted members receive one reminder notification linking to the event

#### Scenario: Local session reminder sent within day lead
- **WHEN** a local `session` event's date falls within the configured day lead
- **THEN** subscribed, unmuted members receive one "Session coming up" reminder

#### Scenario: All-day feed events do not get minute reminders
- **WHEN** a feed event is marked all-day or has no parseable start time
- **THEN** it is skipped by the minute-precision scan

#### Scenario: No duplicate reminders
- **WHEN** the scheduler runs again after a reminder was already sent for an event
- **THEN** no additional reminder is sent for that event

### Requirement: Graceful unsupported environments
Browsers or contexts without push support SHALL NOT surface errors; subscription controls SHALL be hidden or disabled via feature detection.

#### Scenario: Unsupported browser
- **WHEN** the app loads in a browser without PushManager support
- **THEN** no permission prompt appears and no errors are thrown
