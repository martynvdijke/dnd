## 1. Schema and token store

- [x] 1.1 Migration 065: create `telegram_auth_tokens(id, token_hash TEXT UNIQUE, tg_user_id INTEGER, chat_id INTEGER, tg_username TEXT, email TEXT, expires_at TEXT, used_at TEXT, created_at TEXT)` with indexes on `token_hash` and `tg_user_id`; register it in the migrations registry
- [x] 1.2 Add `telegram_auth_tokens` to the raw-managed table allowlist in `db/schema_consistency_test.go`
- [x] 1.3 `telegram/auth.go`: `InsertAuthToken(tgUserID, chatID int64, username, email, tokenHash, expiresAt string) error` and `ConsumeAuthToken(tokenHash string) (authToken, bool)` that atomically marks `used_at` and rejects expired rows (single-use); plus `lookupIdentityByTelegramID` reuse for the already-linked check

## 2. Command and email flow

- [x] 2.1 `authFlow{Email, UpdatedAt}` store keyed by chat ID with a 10-minute TTL: `getAuthFlow`/`putAuthFlow`/`authFlowActive`/`abortAuthFlow`
- [x] 2.2 Register a `/login` command (aliases `/auth`, `/signin`, category bot): redirect already-linked users to `/unlink`, refuse when `BaseURL` is unset, otherwise start the flow and prompt for an email
- [x] 2.3 `feedAuthFlow`: treat the next non-command message as the address; re-ask in place on malformed input; on valid input issue a token and hand it to the injected sender; end the flow
- [x] 2.4 Email validation helper (single `@`, non-empty local-part and domain, no `@` in domain start) and an enumeration-safe "📧 Check your inbox." acknowledgement used for every well-formed address
- [x] 2.5 Per-Telegram-user rate limiter: at most 3 link requests per 15 minutes; reply that the request was rate limited without sending further mail
- [x] 2.6 `HandleUpdate`: check the auth flow before the create flow; abort the auth flow on `/cancel` or any other command
- [x] 2.7 Add the `AuthLinkSender func(email, token string) error` injectable and `SetAuthLinkSender` (default a no-op error)

## 3. Public redemption handler

- [x] 3.1 New `handlers/telegram_auth.go`: `HandleTelegramAuthLink` reads `?token=`, hashes it, and consumes the token row
- [x] 3.2 Resolve the account: match `users` by `lower(email) = lower(:email)` and `email <> ''`; otherwise, when `TELEGRAM_SELF_REGISTER` is on, create a `player` with a uniquified username from the local-part, a random unusable password, the email, and the Telegram display name; refuse when the flag is off
- [x] 3.3 Create a real web session (`middleware.Store.Create`) and set the `session` cookie exactly as `HandleLogin` does
- [x] 3.4 Upsert the Telegram identity and send the originating chat a confirmation DM ("✅ Linked as <username>"); refusals reveal no account data
- [x] 3.5 Respond with a success page carrying `Referrer-Policy: no-referrer`
- [x] 3.6 Register the route as public (unauthenticated) in the router / `handlers/routes_telegram.go`
- [x] 3.7 `TELEGRAM_SELF_REGISTER` env flag, default on

## 4. Onboarding copy and wiring

- [x] 4.1 Update `welcomeText()` and `linkingHelpText()` to present `/login` alongside the web code flow
- [x] 4.2 Wire `telegram.SetAuthLinkSender` in `app.go` to a handler that builds `publicBaseURL() + "/telegram/auth?token=…"` and sends via the existing email settings/sender

## 5. Tests

- [x] 5.1 `telegram` unit tests: email validation, flow re-ask/abort/expiry, token store single-use and expiry, rate limiting, injected sender capture, already-linked redirect, base-URL-unset refusal, enumeration-safe reply
- [x] 5.2 `handlers` tests: valid token links + creates session + creates player; existing email links without creating; expired/used token rejected; self-registration off refused; identity upserted; `Referrer-Policy: no-referrer` present
- [x] 5.3 `db/schema_consistency_test.go` passes with the new table registered
- [x] 5.4 Coverage floors (Go total ≥20%, `handlers/` ≥40%, `middleware/` ≥40%) still met

## 6. Verification and delivery

- [x] 6.1 `openspec validate telegram-email-auth`
- [ ] 6.2 `task ci` green (Go + vitest + e2e + coverage gates); PR opened and merged per the workflow contract
- [ ] 6.3 Post-merge: archive the change in a follow-up PR (`openspec archive telegram-email-auth`, `git add -f`), matching the repo convention
