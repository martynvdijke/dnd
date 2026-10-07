## Context

Account linking is exercised from the bot (`telegram/`) but implemented through the web UI: a signed-in user calls `POST /api/telegram/link-code` (`handlers/routes_telegram.go` → `telegram.CreateLinkCode`), an eight-character code is shown, and the bot redeems it with `/start <code>` (`telegram/handlers.go` → `ConsumeLinkCode` → `UpsertIdentity`). Two problems follow: the round trip is hostile on mobile, and a player without an account has no way in.

The pieces needed for an email link already exist elsewhere in the codebase: `handlers/email.go` reads the `email_settings` row and sends mail (including port 465 TLS), `handlers/password_reset.go` is the template for single-use hashed tokens (32 random bytes, SHA-256 at rest, expiry, injectable sender, IP rate limit), and `publicBaseURL` in `handlers/invitations.go` centralises URL construction from the `BaseURL` package variable. The `telegram` package must not import `handlers` (import cycle), so anything that needs the database, sessions or SMTP is reached through an injected callback, exactly like `SetCharacterCreator`.

## Goals / Non-Goals

**Goals:**

- Let a linked-or-new player authenticate and link Telegram in one chat step via an emailed single-use link.
- Reuse that redemption as a genuine passwordless web login (real session cookie).
- Self-register unknown emails as `player` accounts (open registration) while remaining enumeration-safe.
- Keep the existing code flow working and make the bot's onboarding copy describe the new path.
- Bound abuse with a per-Telegram-user rate limit and a configurable off switch.

**Non-Goals:**

- Changing password authentication, sessions, or the `retire-password-fallback` direction.
- Matching accounts by anything other than a non-empty, case-insensitive email.
- Email verification of the address beyond control of the link (opening the link is the proof).
- Telegram-native OAuth / Login Widget, or linking multiple email addresses to one account.

## Decisions

1. **Email magic link is the auth primitive.** A redemption token is 32 random bytes, stored only as its SHA-256 hex (`hashToken`), single-use via a `used_at` timestamp, and expiring after 15 minutes. This mirrors `password_reset_tokens` so there is one token idiom to reason about. Alternative considered: a numeric OTP typed back into the chat — rejected because it adds a second flow and a guessable secret without removing the email dependency.

2. **Find-or-create by email, self-registration default on.** On redemption, look up `users` where `lower(email) = lower(:email)` and `email <> ''`. If found, link it. If not, and `TELEGRAM_SELF_REGISTER` is on (default), create a `player`: username derived from the email local-part and uniquified (`alice`, `alice2`, …), a random unusable password so password login cannot match, `email` set, and `display_name` taken from the Telegram name. If the flag is off, the redemption is refused. Alternative: invite-only — rejected by the product decision to opened self-registration.

3. **Redemption is a real web login.** On success the handler creates a session through the same path as `HandleLogin` (`middleware.Store.Create(...)`) and sets the `session` cookie, so the link signs the player into the web app, not just the bot. Alternative: a bot-only link with no session — rejected; the user asked for passwordless web login too.

4. **Injected SMTP seam, not a direct import.** `telegram` exposes `SetAuthLinkSender(func(email, token string) error)`; `app.go` wires it to a `handlers` function that composes `publicBaseURL` + `/telegram/auth?token=` and calls the existing email sender. This keeps `handlers` → `telegram` one-directional, matching `SetCharacterCreator`.

5. **Separate in-chat flow store.** A small `authFlows map[chatID]*authFlow` with the same 10-minute TTL is checked in `HandleUpdate` before the create flow, and any non-`/login` command or `/cancel` aborts it. Rationale: the create flow's state machine is character-specific; a tiny dedicated flow is clearer than generalising it. Alternative: generalise the flow framework — deferred until a third flow exists.

6. **Public, unauthenticated redemption route.** `GET /telegram/auth?token=…` is registered under `RegisterPublicRoutes` because the caller is not logged in yet; the token is the credential. The response page sets `Referrer-Policy: no-referrer` so the token cannot leak through an outbound link, and the confirmation is also delivered by DM to the originating chat.

7. **Enumeration safety and idempotent messaging.** The bot always replies "📧 Check your inbox." for a well-formed address, whether or not it matches, and SMTP failures only change internal logging. The token row stores `tg_user_id`, `chat_id`, `tg_username` and `email`, so the DM confirmation goes to the chat that started the flow even if the same account is later linked from elsewhere.

8. **Known limitation.** An existing account whose `email` field is empty cannot be matched, so opening a link with that address creates a separate player account. This is documented and mitigable (add the email in web settings, or use the code flow); matching empty emails would risk attaching accounts by an unverified value.

## Risks / Trade-offs

- **Email deliverability** is now on the login path. Mitigate by reusing the proven SMTP settings/sender, keeping the code flow as a fallback, and surfacing a clear unavailability reply when `BaseURL` or SMTP is not configured.
- **Duplicate accounts** for users who never set an email (decision 8). Acceptable for an open-registration campaign tool; documented, and reachable through the code flow.
- **Username collisions** from local-parts. Mitigated by deterministic uniquification at creation time.
- **Token leakage** through browser referrers. Mitigated with `Referrer-Policy: no-referrer` on the redemption response and a short TTL.
- **Abuse / mail bombing.** Mitigated by the per-Telegram-user limit (3 per 15 minutes) and single-use tokens; a determined attacker can still cause at most three mails per Telegram account.
- **Flow interference** with `/create`. Mitigated by checking the auth flow first and aborting on any other command.
