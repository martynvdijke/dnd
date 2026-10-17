## Why

Linking Telegram to Villum currently means leaving the chat: open the web UI, sign in, open settings, generate an eight-character code, then come back and send `/start <code>`. Players on a phone routinely can't finish that loop, and brand-new players have no account at all, so the bot is unusable for them until a DM provisions one by hand. An emailed single-use link lets a player authenticate and link in one step, doubles as passwordless web sign-in, and can self-register new players — removing the DM from the critical path to the table's bot.

## What Changes

- Add `/login` (aliases `/auth`, `/signin`) which starts a short in-chat flow asking for the player's email, reusing the same guided-flow pattern as `/create`.
- Validate the address and reply with an enumeration-safe "📧 Check your inbox." regardless of whether the address matches an account.
- Issue a single-use token (32 random bytes, SHA-256 at rest, 15-minute expiry) stored in a new `telegram_auth_tokens` table alongside the originating Telegram user, chat, username and email, and email it a link to `BASE_URL/telegram/auth?token=…` using the existing SMTP settings.
- Add a public `GET /telegram/auth?token=…` that redeems the token: links the Telegram identity to the matching Villum account (case-insensitive email, non-empty), opens a real web session (passwordless web login), and — when self-registration is enabled — creates a new `player` account (unique username from the email local-part, unusable random password, email set) when no account matches.
- DM the originating chat to confirm the link and show a success page with `Referrer-Policy: no-referrer`.
- Enforce a per-Telegram-user rate limit (3 emails per 15 minutes), gate self-registration behind `TELEGRAM_SELF_REGISTER` (default on), and reply clearly when `BaseURL` is unset so email sign-in is unavailable.
- Keep the existing web eight-character code flow as a fallback and mention the email path in the welcome and linking instructions.

## Capabilities

### New Capabilities
- `telegram-email-auth`: bot-initiated email sign-in, single-use magic-link tokens, public token redemption, find-or-create account resolution, passwordless web session, and the surrounding rate limiting, enumeration safety and availability rules.

### Modified Capabilities
- `telegram-bot-onboarding`: the welcome and unlinked-linking instructions now present `/login` (email sign-in) alongside the existing code flow.

## Impact

- Backend: new `telegram/auth.go`; new `handlers/telegram_auth.go`; `telegram/commands.go` (command registry), `telegram/bot.go` (flow dispatch), `telegram/store.go`, `handlers/routes_telegram.go` (public route), `app.go` (sender wiring).
- Schema: migration 065 adding `telegram_auth_tokens` (+ `token_hash` and `tg_user_id` indexes); register it in the migrations registry and the `db/schema_consistency_test.go` raw-managed allowlist.
- Email: reuses `handlers/email.go` settings and sender; adds an injected `AuthLinkSender` wired from `app.go` (no `telegram` → `handlers` import).
- Config: `TELEGRAM_SELF_REGISTER` (default on). No breaking API changes; the existing code flow keeps working.
- Tests: `telegram` unit tests (flow, validation, rate limit, injected sender, token consume), `handlers` tests (link + session + player creation, expired/used rejection, self-registration off, existing-account link), schema consistency coverage; e2e optional.
