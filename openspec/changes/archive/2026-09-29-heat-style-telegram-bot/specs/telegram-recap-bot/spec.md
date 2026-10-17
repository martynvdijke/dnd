## MODIFIED Requirements

### Requirement: Telegram bot configuration

Administrators SHALL configure the Telegram bot token and transport mode through admin settings. The token SHALL be stored encrypted at rest and SHALL NOT be logged or returned in full to any client. The system SHALL support long-polling and webhook transports, with webhook available only when a public HTTPS base URL and a webhook secret are configured. Configuration SHALL be applied to the running bot without restarting the server process, through the single Telegram settings surface.

#### Scenario: Admin configures polling mode
- **WHEN** an admin saves a bot token with transport mode `polling`
- **THEN** the system starts consuming updates via long-polling and reports the mode as `polling`

#### Scenario: Configuration changes apply live
- **WHEN** an admin changes the token or transport mode
- **THEN** the running bot converges on the new configuration within the supervision interval without a server restart

#### Scenario: Webhook mode requires a public URL and secret
- **WHEN** transport mode is `webhook` but no public HTTPS base URL or webhook secret is configured
- **THEN** the system does not register the webhook and reports the configuration as invalid

#### Scenario: Token is never exposed
- **WHEN** any client requests the Telegram settings
- **THEN** the response omits the token or masks it, and the token never appears in logs

### Requirement: Message size handling

The system SHALL send bot messages using Telegram's HTML parse mode, escaping any user-supplied content, and SHALL NOT exceed Telegram's 4096-character message limit, splitting long content without corrupting it.

#### Scenario: Long recap is split
- **WHEN** a recap exceeds the message limit
- **THEN** it is delivered as multiple ordered parts or as a document, with no content lost

#### Scenario: User content is escaped
- **WHEN** rendered content contains HTML-significant characters
- **THEN** they appear literally in the message and do not break the message or enable markup
