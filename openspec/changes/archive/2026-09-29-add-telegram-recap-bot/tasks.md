## 1. Foundation

- [x] 1.1 Migration: create `telegram_identities`, `telegram_link_codes`, `campaign_telegram_settings`, `telegram_deliveries` (with `UNIQUE(recap_id, target_chat_id)`) via the existing migration mechanism; no changes to existing tables
- [x] 1.2 `telegram/config.go`: read/validate config from `app_settings` (`telegram_bot_token`, `telegram_webhook_secret`, `telegram_mode`, `telegram_update_offset`) with env override; reuse `crypto/` for token + secret (encrypted at rest, plaintext fallback with warning matching `ai_endpoints`)
- [x] 1.3 `telegram/client.go`: `net/http` Bot API wrapper — `sendMessage`, `sendDocument`, `getUpdates`, `setWebhook`, `deleteWebhook`, `getChat`; timeout, `User-Agent villum/<Version>`, 429 `retry_after` handling
- [x] 1.4 Helpers: `chunkMessage(text, 4096)` splitting on paragraph/line boundaries, and plain-text send policy (no `parse_mode`)
- [x] 1.5 `telegram/store.go`: raw-SQL CRUD for the four tables; link-code hash/consume/expire; identity lookup by `telegram_user_id`

## 2. Account linking

- [x] 2.1 `telegram/linking.go`: create code (8 chars, ambiguity-free alphabet, SHA-256 at rest, 15-min TTL, single-use), consume code, unlink identity, expire sweep
- [x] 2.2 `POST /api/telegram/link-code`, `GET /api/telegram/status`, `DELETE /api/telegram/unlink` (auth group), returning the code + `t.me/<bot>?start=<code>` deep link once
- [x] 2.3 `PUT /api/telegram/prefs` for the per-user `dm_enabled` toggle
- [x] 2.4 Rate-limit link-code creation per user and link attempts per `telegram_user_id` (link-code creation throttled to 5 per 15 min; separate redemption-attempt throttle not implemented)
- [x] 2.5 Handler tests: code issue/consume/expiry/single-use, unlink, auth rejection, brute-force throttling

## 3. Inbound transport

- [x] 3.1 `telegram/poller.go`: `StartTelegramPoller` `getUpdates` loop with persisted offset, backoff on error, registered from `app.go registerSchedulers` when mode is polling/auto
- [x] 3.2 `telegram/webhook.go`: public `POST /api/telegram/webhook` registered in `RegisterPublicRoutes`, constant-time `X-Telegram-Bot-Api-Secret-Token` check, `setWebhook`/`deleteWebhook` on startup mode change
- [x] 3.3 Single unified `HandleUpdate(ctx, update)` dispatcher shared by both transports
- [x] 3.4 Tests: offset advance, secret rejection (403 without leaking token), mode selection logic

## 4. Outbound delivery

- [x] 4.1 `telegram/routing.go`: `deliverRecap(recapID, kind)` — resolve campaign chat + opted-in DMs, insert-or-ignore `telegram_deliveries`, send, record `telegram_message_id`/status
- [x] 4.2 Wire the manual path into `handlers/recaps.go` `SendRecap`/`MarkRecapAsSent` alongside `NotifyCampaignRecapPublished` (fire-and-forget, non-blocking)
- [x] 4.3 Retry/backoff for `failed` rows (≤3 attempts) with per-target throttling and `retry_after` respect
- [x] 4.4 Tests: manual send reaches bound chat; duplicate send produces one delivery; chunking at/over 4096; `sendDocument` fallback; retry and terminal failure

## 5. Inbound queries

- [x] 5.1 `telegram/handlers.go`: `/help`, `/recap [campaign]`, `/lastrecap`, `/unlink`
- [x] 5.2 Resolve `telegram_user_id → user_id → member campaigns`; generic reply for unlinked senders; identical "not found" for unauthorized campaign IDs
- [x] 5.3 Campaign ambiguity handling: reply with a picker limited to member campaigns
- [x] 5.4 Tests: member can query own campaign; non-member cannot; unlinked gets help; no cross-campaign leakage

## 6. Automatic posting after a session

- [x] 6.1 `telegram/scheduler.go`: `StartTelegramAutoPostScheduler` (1-min ticker, backup-scheduler pattern) selecting recaps with `is_sent = 0`, non-null `session_end_date` past the grace delay, campaign `auto_post_enabled = 1`
- [x] 6.2 Idempotent dispatch through `deliverRecap(..., kind='auto')`; skip recaps modified within the grace window; global kill switch (limitation: `campaign_recaps` has no `updated_at`, so the modified-within-grace skip is not enforced; null `session_end_date` and disabled campaigns are excluded and delivery is idempotent)
- [x] 6.3 Grace delay default 30 minutes (admin-configurable if Open Question resolves yes)
- [x] 6.4 Tests: auto-posts exactly once per recap×target; null `session_end_date` never auto-posts; disabled campaign never auto-posts

## 7. Configuration surfaces

- [x] 7.1 Admin endpoints `GET/POST /api/admin/telegram-settings` (token, transport mode, test-send) in the `admin` group; follow the email/push settings pattern; never return the token in full or log it
- [x] 7.2 Admin UI section (admin bundle): token field, transport mode, test message, status (mode, last poll/update, webhook URL)
- [x] 7.3 Campaign Telegram section `GET/PUT /campaigns/:id/telegram` (auth group, `isCampaignDM` guard): `chat_id`, `is_enabled`, `auto_post_enabled`; bind helpers
- [x] 7.4 User "Link Telegram" + DM opt-in UI in account/settings
- [x] 7.5 Tests: admin-only access, campaign DM guard, token never echoed back, settings round-trip
- [x] 7.6 Ensure every new `data-testid` is referenced in `tests/` per repo lint rule

## 8. Verification

- [x] 8.1 `go vet ./... && go build ./...` clean
- [x] 8.2 `task test` green including new Go tests
- [x] 8.3 `npm run typecheck && npm run test:unit` green
- [x] 8.4 E2E (mocked Telegram Bot API): link flow, manual send, `/lastrecap` query, membership denial; all new `data-testid`s covered
- [ ] 8.5 Manual smoke (deferred, requires a real bot token): real bot token → link → recap send posts to group → auto-post after `session_end_date` → `/lastrecap` answers; verify no duplicate on repeated ticks
