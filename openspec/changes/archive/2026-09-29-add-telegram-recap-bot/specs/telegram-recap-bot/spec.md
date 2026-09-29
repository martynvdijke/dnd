## ADDED Requirements

### Requirement: Telegram bot configuration
Administrators SHALL configure the Telegram bot token and transport mode through admin settings. The token SHALL be stored encrypted at rest and SHALL NOT be logged or returned in full to any client. The system SHALL support long-polling and webhook transports, with webhook available only when a public HTTPS base URL and a webhook secret are configured.

#### Scenario: Admin configures polling mode
- **WHEN** an admin saves a bot token with transport mode `polling`
- **THEN** the system starts consuming updates via long-polling and reports the mode as `polling`

#### Scenario: Webhook mode requires a public URL and secret
- **WHEN** transport mode is `webhook` but no public HTTPS base URL or webhook secret is configured
- **THEN** the system does not register the webhook and reports the configuration as invalid

#### Scenario: Token is never exposed
- **WHEN** any client requests the Telegram settings
- **THEN** the response omits the token or masks it, and the token never appears in logs

### Requirement: Telegram account linking
An authenticated user SHALL be able to generate a single-use, time-limited linking code and redeem it by sending `/start <code>` to the bot. The system SHALL store only a hash of the code. A successful redemption SHALL bind the Telegram user to the Villum user. Users SHALL be able to revoke the link from the web UI and via `/unlink`.

#### Scenario: Successful linking
- **WHEN** an authenticated user generates a code and sends `/start <code>` to the bot before it expires
- **THEN** the Telegram identity is bound to that Villum user and the code can no longer be used

#### Scenario: Expired or reused code is rejected
- **WHEN** a code is redeemed after its expiry or after it was already used
- **THEN** the binding does not occur and the bot replies that the code is invalid or expired

#### Scenario: Unlinking revokes access
- **WHEN** a linked user unlinks
- **THEN** subsequent bot commands from that Telegram identity are treated as unlinked

### Requirement: Manual recap delivery to Telegram
When a campaign recap is published or marked sent, the system SHALL post it to the campaign's linked Telegram chat when delivery for that campaign is enabled, in addition to existing Web Push behavior. Delivery SHALL NOT block the HTTP response.

#### Scenario: Enabled campaign chat receives the recap
- **WHEN** a campaign DM sends a recap for a campaign with an enabled Telegram chat
- **THEN** the recap is posted to that chat

#### Scenario: Disabled campaign chat receives nothing
- **WHEN** a recap is sent for a campaign whose Telegram delivery is disabled or unbound
- **THEN** nothing is posted to Telegram

#### Scenario: Members with DM opt-in receive a copy
- **WHEN** a recap is sent and campaign members have opted in to Telegram DMs
- **THEN** each such linked member receives the recap as a direct message

### Requirement: Automatic recap delivery after a session
The system SHALL run a background scheduler that posts a recap automatically once the session it covers has ended, when auto-post is enabled for the campaign. Because the system has no session end-time signal, "ended" SHALL be interpreted as the recap's `session_end_date` having passed by a grace delay. Recaps without a `session_end_date` SHALL NOT be auto-posted.

#### Scenario: Auto-post after the grace delay
- **WHEN** a recap for an auto-post-enabled campaign has a `session_end_date` older than the grace delay and has not been delivered
- **THEN** the scheduler posts it to Telegram exactly once

#### Scenario: Recap without an end date stays manual
- **WHEN** a recap has no `session_end_date`
- **THEN** the scheduler never auto-posts it

#### Scenario: Auto-post disabled per campaign
- **WHEN** auto-post is disabled for a campaign
- **THEN** no recap for that campaign is posted automatically

### Requirement: Inbound recap queries
Linked users SHALL be able to query recaps from Telegram using chat commands, including a most-recent-recap command and a per-campaign recap command. Responses SHALL be limited to campaigns the sender belongs to.

#### Scenario: Member queries a campaign recap
- **WHEN** a linked member sends `/recap` for a campaign they belong to
- **THEN** the bot replies with recaps from that campaign only

#### Scenario: Non-member is denied
- **WHEN** a linked user requests a recap for a campaign they do not belong to
- **THEN** the bot replies with the same "not found" response as for a nonexistent campaign and reveals no recap data

#### Scenario: Unlinked sender gets help
- **WHEN** an unlinked Telegram user sends a command
- **THEN** the bot replies with linking instructions and no campaign data

### Requirement: Delivery idempotency
The system SHALL record each delivery per recap and target chat, and SHALL NOT deliver the same recap to the same target more than once, including when manual send and auto-post occur close together.

#### Scenario: Duplicate delivery suppressed
- **WHEN** a recap is delivered to a target chat and a subsequent send or scheduler tick targets the same recap and chat
- **THEN** no second message is sent to that chat

#### Scenario: Failed delivery is retried then recorded
- **WHEN** a send fails transiently
- **THEN** the system retries a bounded number of times and records the final status with the last error

### Requirement: Message size handling
The system SHALL send recap content as plain text and SHALL NOT exceed Telegram's 4096-character message limit, splitting long content without corrupting it.

#### Scenario: Long recap is split
- **WHEN** a recap exceeds the message limit
- **THEN** it is delivered as multiple ordered parts or as a document, with no content lost

### Requirement: Webhook authenticity
When webhook transport is used, the system SHALL reject requests whose secret token header does not match the configured webhook secret, without logging the presented value.

#### Scenario: Forged webhook is rejected
- **WHEN** a request arrives at the webhook endpoint with a missing or incorrect secret token header
- **THEN** it is rejected with 403 and no update is processed
