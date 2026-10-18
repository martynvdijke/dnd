## 1. Schema and migrations

- [x] 1.1 Migration 068 `add_telegram_digest_subscriptions.go`: create `telegram_digest_subscriptions(user_id INTEGER NOT NULL, campaign_id INTEGER NOT NULL, frequency TEXT NOT NULL DEFAULT 'weekly', last_sent_at TEXT, created_at TEXT NOT NULL, PRIMARY KEY(user_id, campaign_id))` plus `CREATE INDEX idx_telegram_digest_subscriptions_campaign ON telegram_digest_subscriptions(campaign_id)`
- [x] 1.2 Migration 069 `add_telegram_auto_post_schedule.go`: four plain `ALTER TABLE campaign_telegram_settings ADD COLUMN` statements (`auto_post_cadence_minutes INTEGER NOT NULL DEFAULT 0`, `quiet_start_hour INTEGER`, `quiet_end_hour INTEGER`, `auto_post_last_at TEXT`); raw-managed table, so these stay out of `db/migrations/safe_alters.go`
- [x] 1.3 Register 068 and 069 in `db/migrations/registry.go` (main is at 67) and add `telegram_digest_subscriptions` to `rawOnlyTables` in `db/schema_consistency_test.go`
- [x] 1.4 Regenerate the hard-coded SHA-256 fingerprint in `db/migrations_registry_test.go` and confirm `go test ./db/...` passes

## 2. Digest and announce backend

- [x] 2.1 `telegram/announce.go`: `runAnnounce` resolves the caller's campaign via `resolveMemberCampaign`, requires DM/admin via `campaignRole`, requires non-empty text, rejects with a clear reply when the chat is unbound, the campaign is not enabled, or the caller is not a DM, and sends the HTML-escaped body through the existing `sendEscapedReply`-style path
- [x] 2.2 `telegram/digest.go`: subscribe/upsert and unsubscribe per `(user, campaign)` with `frequency='weekly'`, `last_sent_at`, `created_at`; membership-checked so a user can only touch campaigns they belong to; `scanDigests(now)` selects subscriptions whose `last_sent_at` is null or older than seven days and sends each at most once per week, writing the timestamp before or atomically with the send
- [x] 2.3 Register `/announce` (category `catCampaign`) and `/digest` (category `catNotifications`, `/digest off` to unsubscribe) in `telegram/commands.go` with usage/description, and wire `scanDigests` into the one-minute ticker in `telegram/scheduler.go`

## 3. Auto-post cadence and quiet hours

- [x] 3.1 `telegram/settings.go`: extend `CampaignTelegramSettings` with `AutoPostCadenceMinutes int`, `QuietStartHour *int`, `QuietEndHour *int`, `AutoPostLastAt *string`; update the read and upsert helpers and add an update helper for the cadence/quiet fields
- [x] 3.2 `telegram/scheduler.go`: in `scanAutoPost`, skip a campaign when fewer than `auto_post_cadence_minutes` have elapsed since `auto_post_last_at` (stored UTC) or when the current UTC hour is inside the quiet window (start inclusive, end exclusive, wrapping midnight); advance `auto_post_last_at` only for the automatic path
- [x] 3.3 Confirm manual/command-triggered `DeliverRecap` calls and `retryFailed` remain unpaced and do not advance cadence
- [x] 3.4 `handlers/routes_telegram.go`: `GET` returns the four new fields alongside the existing ones; `PUT` accepts `auto_post_cadence_minutes` (0–1440), `quiet_start_hour`/`quiet_end_hour` (0–23, both set or both null, null clears), rejects out-of-range or half-set pairs with 400, keeps DM-only enforcement, and merges cleanly with the unchanged `auto_save_media` handling
- [x] 3.5 Handler tests: cadence and quiet-hour round-trip, 400 on out-of-range cadence and half-set quiet hours, `auto_save_media` still round-trips

## 4. Frontend

- [x] 4.1 `ts/lib/telegram.ts`: extend the campaign Telegram settings types/client with `auto_post_cadence_minutes`, `quiet_start_hour`, `quiet_end_hour`, `auto_post_last_at`
- [x] 4.2 `ts/campaign/telegram.ts`: add cadence input and quiet-hour start/end inputs labelled "(UTC)", save them through the existing PUT flow, and show `auto_post_last_at` read-only next to them
- [x] 4.3 Add the new `data-testid`s for the cadence and quiet-hour controls and reference every one of them in `tests/`

## 5. Tests

- [x] 5.1 Go unit tests: `/announce` (posted, empty text, unbound chat, disabled campaign, non-DM), digest subscribe/unsubscribe/membership denial/due-and-sent/not-yet-due
- [x] 5.2 Go unit tests for pacing and quiet hours: skipped before the window, posts after it, cadence 0 unpaced, quiet window boundary suppressed, unset quiet hours not suppressed, manual delivery unpaced
- [x] 5.3 Vitest unit tests for the new `ts/lib/telegram.ts` and `ts/campaign/telegram.ts` behaviour; keep the vitest coverage floors
- [x] 5.4 Playwright `tests/telegram.spec.ts`: set cadence and quiet hours through the UI and assert the persisted values; announce and digest happy paths via the Telegram mock

## 6. Verification and delivery

- [ ] 6.1 `openspec validate add-telegram-digest-and-cadence`; `task ci` green (Go + vitest + e2e + coverage gates)
- [ ] 6.2 Per the workflow contract: branch, Conventional Commits, local gates, push, open the PR, wait for `gh pr checks --watch`, squash-merge green
- [ ] 6.3 Post-merge follow-up PR: archive the change (`openspec archive add-telegram-digest-and-cadence`, `git add -f`) matching the repo convention
