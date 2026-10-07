## ADDED Requirements

### Requirement: Email sign-in command

The bot SHALL provide a `/login` command (with `/auth` and `/signin` aliases) that starts email sign-in. When the sending Telegram user is already linked, the bot SHALL reply that the account is already linked, name the linked Villum user, and point at `/unlink` instead of starting the flow.

#### Scenario: Login starts the email flow

- **WHEN** an unlinked Telegram user sends `/login`
- **THEN** the bot asks for an email address and waits for the reply

#### Scenario: Already-linked user is redirected

- **WHEN** a linked Telegram user sends `/login`
- **THEN** the bot replies that the account is already linked, names the Villum user, and suggests `/unlink`, and no email flow starts

### Requirement: Email capture flow

While an email flow is active for a chat, the bot SHALL interpret the next non-command message as the email address, validate it, re-ask in place when it is malformed, and accept `/cancel` or any other command to abort the flow. The flow SHALL expire after a bounded period of inactivity.

#### Scenario: Valid address is accepted

- **WHEN** the user replies to the email prompt with a well-formed address
- **THEN** the bot sends the sign-in link and ends the flow

#### Scenario: Malformed address is re-asked

- **WHEN** the user replies with a value that is not a well-formed email address
- **THEN** the bot re-asks for the address without sending mail and keeps the flow active

#### Scenario: Flow can be aborted

- **WHEN** the user sends `/cancel` or another command during the email flow
- **THEN** the flow is discarded and nothing is sent

#### Scenario: Stale flow expires

- **WHEN** the user replies long after the flow expired
- **THEN** the message is handled as a normal message rather than an address

### Requirement: Magic-link token issuance

On a valid address the bot SHALL issue a single-use token that is 32 random bytes, stored only as a SHA-256 hash with its expiry, and associated with the originating Telegram user, chat, username and email. The token SHALL expire after 15 minutes and SHALL be usable at most once. The link sent by email SHALL point at the configured public base URL with the raw token.

#### Scenario: Token is stored hashed

- **WHEN** a valid address is submitted
- **THEN** a token row is stored holding a hash of the token, the Telegram user, chat, username, email and an expiry, and the raw token is never persisted

#### Scenario: Token is single use

- **WHEN** a token has already been redeemed
- **THEN** a second redemption of the same token is refused

#### Scenario: Expired token is refused

- **WHEN** a token is presented after its 15-minute expiry
- **THEN** the redemption is refused

### Requirement: Link redemption

The bot SHALL expose a public redemption endpoint that accepts a token, and on a valid token SHALL link the Telegram identity to the resolved Villum account and confirm the link to the originating chat by direct message. On an invalid, used or expired token the endpoint SHALL refuse without linking and SHALL reveal no account data. The response SHALL set `Referrer-Policy: no-referrer`.

#### Scenario: Valid token links and confirms

- **WHEN** a valid token is opened
- **THEN** the Telegram identity is linked to the resolved account and a confirmation is sent to the originating chat

#### Scenario: Invalid token is refused

- **WHEN** an unknown, used or expired token is opened
- **THEN** the link is refused, nothing is linked, and no account data is shown

#### Scenario: Redemption does not leak the token

- **WHEN** the redemption page is returned
- **THEN** the response carries a `Referrer-Policy: no-referrer` header

### Requirement: Passwordless web session

A successful redemption SHALL create a real Villum web session for the resolved account and set the session cookie, so opening the emailed link signs the player into the web app without a password.

#### Scenario: Opening the link signs the player in

- **WHEN** a valid token is opened in a browser
- **THEN** the browser receives a session cookie for the resolved account

### Requirement: Self-registration of new players

When no account matches the address and self-registration is enabled, the bot SHALL create a new account with role `player`, a unique username derived from the email local-part, an unusable random password, the submitted email, and a display name taken from Telegram. When self-registration is disabled, an unmatched address SHALL be refused and no account SHALL be created.

#### Scenario: Unknown address creates a player

- **WHEN** a valid token is redeemed for an address with no matching account and self-registration is enabled
- **THEN** a new `player` account is created with a unique username from the local-part, the email set, and the Telegram display name

#### Scenario: Self-registration disabled refuses

- **WHEN** a valid token is redeemed for an address with no matching account and self-registration is disabled
- **THEN** the redemption is refused and no account is created

### Requirement: Email match resolution

The bot SHALL match an existing account by case-insensitive email, ignoring accounts whose email is empty, and SHALL link the first such match without creating a new account.

#### Scenario: Existing account is linked

- **WHEN** a valid token is redeemed and a non-empty account email matches case-insensitively
- **THEN** that account is linked and no new account is created

### Requirement: Enumeration safety

For any well-formed address the bot SHALL send the same acknowledgement text whether or not an account exists and whether or not the message is actually delivered, and SHALL NOT disclose account existence, creation or delivery failures to the sender.

#### Scenario: Same reply for known and unknown addresses

- **WHEN** a user submits a well-formed address that matches an existing account and then one that does not
- **THEN** both replies are the identical acknowledgement text

#### Scenario: Delivery failure is not disclosed

- **WHEN** sending the link fails
- **THEN** the bot still replies with the same acknowledgement and does not report the failure to the sender

### Requirement: Abuse rate limiting

The bot SHALL limit email-link requests to three per Telegram user per 15 minutes and SHALL reply without sending further mail once the limit is reached.

#### Scenario: Fourth request is throttled

- **WHEN** a Telegram user has requested three links within 15 minutes and requests another
- **THEN** no further mail is sent and the bot replies that the request was rate limited

### Requirement: Availability and fallback

When the public base URL is not configured the bot SHALL report that email sign-in is unavailable instead of starting the flow. The existing web-generated code flow SHALL remain available for linking.

#### Scenario: Missing base URL is reported

- **WHEN** a user starts `/login` and no public base URL is configured
- **THEN** the bot replies that email sign-in is unavailable and starts no flow

#### Scenario: Code flow still links

- **WHEN** a user redeems a web-generated code with `/start <code>`
- **THEN** the account is linked exactly as before
