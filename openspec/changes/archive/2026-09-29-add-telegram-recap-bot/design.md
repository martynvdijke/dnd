## Context

Villum is a single Go binary (Gin + modernc SQLite + ent for some entities) with a TypeScript frontend. Relevant existing machinery:

- **Recaps**: `campaign_recaps` (raw SQL, `handlers/recaps.go`) with `id, campaign_id, title, content, session_start_date, session_end_date (both nullable), word_count, is_edited, is_sent, created_at`. `CreateCampaignRecap` and `MarkRecapAsSent` call `NotifyCampaignRecapPublished()` (`handlers/push.go`) — a fire-and-forget goroutine that fans out Web Push. That is the natural hook point for Telegram.
- **Settings**: `app_settings(key TEXT PRIMARY KEY, value TEXT)` via `appSetting`/`setAppSetting` (`handlers/push.go`). `crypto/` provides AES-256-GCM (`AI_ENCRYPTION_KEY`) and already encrypts `ai_endpoints.api_key`.
- **Schedulers**: no cron lib or event bus. In-process goroutines registered in `app.go registerSchedulers` — `StartBackupScheduler` (1h), `StartPushReminderScheduler` (1min ticker), `StartCleanupTask` (15min).
- **Routing**: `router.go` mounts `RegisterPublicRoutes(r)` on the engine with no auth/CSRF (precedent: `POST /api/login`, `POST /invite/:token`, public TRMNL GETs), then `auth`/`dm`/`admin` groups with `AuthRequired` + `APITokenRequired` + `CSRFRequired`.
- **Campaign auth**: `isCampaignMember(c, campaignID int64)` / `isCampaignDM(c, campaignID int64)`; `campaign_members(id, campaign_id, user_id, role)`; `users(id, username, password, display_name, role, email, created_at)`. All IDs are `int64`.
- **Telegram transcript**: the "session" a group schedules is `campaign_calendar_events` where `event_type='session'` — columns `id, campaign_id, title, description, event_date (TEXT '2006-01-02'), event_type, color, created_at`. It is **date-only with no start/end time**. The ent `Session` entity is a *character journal* (`character_id`, `session_date`, `xp_earned`, …), not a schedule and not used here.

## Goals / Non-Goals

**Goals:**
- Post a campaign recap to Telegram on explicit send, and automatically after the session it covers
- Let linked users ask the bot for recaps, scoped to campaigns they belong to
- Account linking that proves both the Villum identity and the Telegram identity
- Work for self-hosted instances with no public URL
- Never double-post a recap to the same target

**Non-Goals:**
- Two-way conversational gameplay (dice rolls, combat) over Telegram — WebSocket covers in-app realtime
- Migrating existing Web Push or email recap delivery
- Rich media recaps (images/maps) in the bot
- A general notification-plugin framework

## Decisions

### 1. Long-polling is the default inbound transport; webhook is opt-in
`StartTelegramPoller` runs a `getUpdates` loop in the existing scheduler/goroutine style with a persisted `telegram.update_offset`. Webhook mode is enabled only when `BASE_URL` is a public HTTPS URL and `telegram_webhook_secret` is set, registering `POST /api/telegram/webhook` in `RegisterPublicRoutes` and calling `setWebhook` on startup.

**Why:** many self-hosted Villum instances sit behind NAT/Tailscale and have no public URL, so webhook-only would simply not work for them; long-polling needs no inbound route. Both transports funnel into one `HandleUpdate` dispatcher. **Alternative considered:** webhook-only — rejected (breaks the self-hosted default). **Trade-off:** long-polling requires a single poller — see Risks.

### 2. No Telegram SDK; `net/http` against the Bot API
A small `telegram/client.go` wraps `sendMessage`, `sendDocument`, `getUpdates`, `setWebhook`, `deleteWebhook`, `getChat` with the existing outbound style (`http.Client` with timeout, `User-Agent villum/<Version>`), matching `handlers/ai.go`.

**Why:** the surface we use is ~6 methods; a dependency is not justified. **Alternative considered:** `go-telegram-bot-api` — acceptable, but adds dependency weight for little gain here.

### 3. Account linking: one-time code, hashed, 15-minute TTL, single-use
An authenticated user requests a code (`POST /api/telegram/link-code`); the server stores only `SHA256(code)` in `telegram_link_codes(user_id, expires_at, used_at)` and returns the code once with a `t.me/<bot>?start=<code>` deep link. The user sends `/start <code>`; the bot hashes, matches an unexpired unused row, inserts `telegram_identities(user_id, telegram_user_id UNIQUE, telegram_chat_id, dm_enabled)`, and marks the code used.

**Why:** codes are low-entropy, so hash at rest; single-use + short TTL limits exposure. **Alternatives:** OAuth Login Widget (domain verification overhead), admin-manual mapping (toil) — both rejected. Revocation via `DELETE /api/telegram/unlink` and the `/unlink` command.

### 4. Targets: per-campaign chat, plus optional per-user DM (DM opt-in default off)
`campaign_telegram_settings(campaign_id PK, chat_id, chat_type, is_enabled, auto_post_enabled, title_cache, bound_at, bound_by_user_id)`. Only a campaign DM may set it. Delivery fan-out = the campaign chat (if enabled) plus every linked member with `dm_enabled = 1`. DMs are never silently redirected to the group.

**Why:** groups are where the table wants recaps; DMs serve players who want a personal copy. **Trade-off:** N DMs per recap costs rate-limit budget, hence opt-in.

### 5. Auto-post keys off the recap's own `session_end_date`, not a session-end event
The app has **no true session-end signal**: calendar sessions are date-only with no end time. Therefore auto-post is defined as: for a campaign with `auto_post_enabled`, a recap is posted once `session_end_date` is set and has passed by a configurable grace delay (default 30 minutes), and it has not already been delivered. Recaps with a null `session_end_date` are manual-send only.

**Why:** `session_end_date` is the one field that already means "the session this recap covers is over". **Explicit limitation:** because it is day-granular, auto-post is approximate. The stronger fix — adding an end time or an explicit "session over" action to `campaign_calendar_events` — is deliberately out of scope here and listed under Risks as the follow-up. **Alternative considered:** reusing `scanLocalSessionReminders`' calendar scan — rejected, it only knows session *starts*.

### 6. Idempotency and delivery tracking via `telegram_deliveries`
`telegram_deliveries(id, recap_id, campaign_id, target_type, target_chat_id, telegram_message_id, kind, status, attempt_count, last_error, created_at, sent_at)` with `UNIQUE(recap_id, target_chat_id)`. Sending inserts-or-ignores first, then dispatches; failures retry up to 3 attempts with backoff on the scheduler tick.

**Why:** auto-post and manual send can race on the same recap; a unique constraint is the cheapest correct guard. **Alternative:** a boolean on `campaign_recaps` — rejected, it cannot express per-target state and would conflate with the existing `is_sent` (which stays as the human-visible receipt).

### 7. Plain-text messages, chunked to 4096 characters
Recaps are freeform prose. Messages are sent with no `parse_mode`, split on paragraph/line boundaries into parts of ≤4096 chars; recaps beyond ~3 parts fall back to `sendDocument` with the body as a text file.

**Why:** MarkdownV2 requires escaping `_ * [ ] ( ) ~ > # + - = | { } . !` and a single unescaped character in a player-written recap breaks the whole send. Plain text is boring and unbreakable. **Alternative:** MarkdownV2 with an escape helper — available later if formatting is wanted.

### 8. Authorization resolves `telegram_user_id → user_id → campaign membership`
Every inbound command maps the Telegram sender to a Villum user via `telegram_identities`, then filters to campaigns where `isCampaignMember`. Unlinked senders get generic help text and no campaign data; unauthorized requests get the same "not found" response as a genuinely missing campaign, so IDs cannot be probed.

## Risks / Trade-offs

- **No real session-end signal** → auto-post is day-granular via `session_end_date` + grace; null end date means manual only. Follow-up: add an end time / "session over" action to calendar sessions. This is the single biggest fidelity gap in the change.
- **Telegram 4096-character cap** → chunking rule 7; long recaps degrade to a file attachment.
- **Multi-replica long-polling duplicates** → document single-replica assumption for polling; if more replicas are expected, use webhook mode or add an advisory-lock/leader lease. Telegram deduplicates webhook updates by `update_id`.
- **Cross-campaign disclosure** → membership enforced on every inbound lookup; indistinguishable "not found" for unauthorized IDs.
- **Rate limits (429)** → sends are throttled and `retry_after` is honored; failed sends are retried via `telegram_deliveries` rather than dropped.
- **Token exposure** → token encrypted with `crypto/` (not plaintext like VAPID) and never logged or returned in full; env override supported for bootstrap.
- **Group privacy mode** → the bot only needs to *send* for delivery; receiving group commands requires privacy mode off or `@botname` commands. Documented in admin help, not worked around.

## Migration Plan

1. Ship the migration plus the `telegram/` package, routes and settings surfaces (feature inactive while no token is configured).
2. Admin saves the bot token and transport mode; sends a test message.
3. Users link; DMs bind a campaign chat and enable auto-post.
4. Rollback: clearing the token disables the integration; orphaned tables are harmless and the recap flow is untouched.

## Resolved Questions

- **Transport default** → long-polling primary, webhook opt-in (§1).
- **Auto-post timing** → opt-in per campaign, grace delay after `session_end_date`, default 30 minutes (§5).
- **DM fan-out default** → off; explicit per-user opt-in (§4).
- **Message formatting** → plain text, no parse mode (§7).

## Open Questions

- Should unlinked members of a group-bound campaign be pinged to link (nice onboarding vs. spam)?
- Configurable grace delay in admin settings, or fixed at 30 minutes for v1?
