## Why

The Telegram bot can already auto-post recaps and save media, but it has no way to announce something to a table immediately and no way to pace or time its automatic posting. A DM who wants to push a session note has to open the web UI; a chatty campaign gets an automatic post the moment every recap lands, regardless of how recently one went out; and a bot that posts at 04:00 UTC is noise no one asked for. This change adds a command-driven announcement, an opt-in weekly digest, and per-campaign pacing and quiet hours so automatic posting respects the table's rhythm.

## What Changes

- **`/announce <text>`** posts an announcement from Telegram to the campaign chat bound to the caller's currently resolved campaign. It requires the DM or admin role, rejects empty text, HTML-escapes the body via the existing escaped-reply path, and replies to confirm or to explain the rejection (no bound chat, campaign not enabled, or caller lacks the DM role).
- **`/digest`** lets a linked user subscribe to a weekly summary for their campaign, persisted per `(user, campaign)`. `/digest off` unsubscribes. A scheduler scan sends each due subscription at most once a week, and a user can only subscribe to campaigns they belong to.
- **Per-campaign auto-post cadence**: campaign Telegram settings gain `auto_post_cadence_minutes` (integer 0–1440; `0` = no pacing, the existing behaviour). The scheduler skips automatic recap posting for a campaign until at least that many minutes have elapsed since `auto_post_last_at`. Manual and command-triggered deliveries and retries are never paced.
- **Per-campaign quiet hours**: `quiet_start_hour` and `quiet_end_hour` (optional integers 0–23, both set or both cleared) suppress automatic recap posting while the current UTC hour is inside the window. Quiet hours and cadence are UTC-only and labelled "(UTC)" in the UI.
- **Settings surface**: `GET /api/campaigns/:id/telegram` returns `auto_post_cadence_minutes`, `quiet_start_hour`, `quiet_end_hour` and `auto_post_last_at`; `PUT` accepts the first three (null clears the quiet hours). Endpoint remains DM-only, and the existing `auto_save_media` field keeps working unchanged.

## Capabilities

### New Capabilities

- `telegram-digest`: the `/announce` command for immediate campaign announcements and the per-user weekly digest subscription, its persistence and due-ness scan.
- `telegram-auto-post-scheduling`: per-campaign automatic-post cadence pacing and UTC quiet hours, their settings API fields, and the scheduler gates that apply them.

### Modified Capabilities

- None. Previous behaviour is unchanged: when `auto_post_cadence_minutes` is `0` and quiet hours are unset, automatic recap posting behaves exactly as before, and no existing endpoint signature changes.

## Impact

- Migrations: `db/migrations/registry.go`, new `db/migrations/068_add_telegram_digest_subscriptions.go` and `db/migrations/069_add_telegram_auto_post_schedule.go`.
- Registry/consistency: `db/schema_consistency_test.go` (`rawOnlyTables` entry for the new table), `db/migrations_registry_test.go` (regenerated fingerprint hash).
- Backend: new `telegram/announce.go` and `telegram/digest.go`, `telegram/commands.go` (command registration), `telegram/scheduler.go` (`scanDigests` wiring plus the pacing/quiet gates), `telegram/settings.go` (new fields and update helper), `handlers/routes_telegram.go` (settings API merged with `auto_save_media`).
- Frontend: `ts/lib/telegram.ts`, `ts/campaign/telegram.ts`, plus new `data-testid`s that must also be referenced under `tests/`.
- Tests: Go unit tests in `telegram/` and `handlers/`, vitest unit tests for the TS modules, and `tests/telegram.spec.ts` Playwright coverage; coverage floors stay enforced.
- No breaking API changes: new fields are optional, and cadence `0` plus unset quiet hours preserve today's automatic posting.
