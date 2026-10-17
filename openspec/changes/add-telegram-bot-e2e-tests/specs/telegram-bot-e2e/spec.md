# telegram-bot-e2e

## ADDED Requirements

### Requirement: Telegram login via email magic link

The bot SHALL support sign-in for an unlinked Telegram user via an emailed magic link, and the redemption SHALL create a web session visible in the browser.

#### Scenario: Unlinked user receives and redeems a magic link
- **GIVEN** a fresh unlinked Telegram user and email delivery to an SMTP mock
- **WHEN** the user sends `/login` then an email address, and the raw captured email is parsed for the `/telegram/auth?token=...` link and visited in the same browser context
- **THEN** the success page "You're signed in" is rendered, a `session` cookie exists, and a subsequent `/status` reply contains `Linked` and the Telegram username

### Requirement: Telegram character claim via inline button

The bot SHALL list claimable characters for `/claim` with an inline keyboard whose button `callback_data` is `claim:<id>`, and tapping it SHALL claim the character.

#### Scenario: Claim via callback query
- **GIVEN** a linked user and a seeded `player` character
- **WHEN** the user sends `/claim`, the test parses `reply_markup` inline keyboard `callback_data` `claim:<id>`, and posts a `callback_query` update with that data and required `message.date`
- **THEN** the bot replies containing the character name; a subsequent `/claim` replies `Currently claimed` with that name; `/unclaim` replies released

### Requirement: Telegram status for a linked user

The bot SHALL reply to `/status` for a linked user instead of remaining silent.

#### Scenario: Linked user requests status
- **GIVEN** a Telegram user linked via a link code
- **WHEN** the user sends `/status` via the webhook
- **THEN** the bot replies containing `Linked`
