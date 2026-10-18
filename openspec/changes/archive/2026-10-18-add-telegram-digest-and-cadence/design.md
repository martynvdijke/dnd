## Context

The Telegram bot (`telegram/` package, migration 062) auto-posts recaps from a one-minute ticker in `StartTelegramAutoPostScheduler`. `scanAutoPost` finds unsent recaps whose session end is older than the grace window and calls `DeliverRecap(id, "auto")` for every campaign whose `campaign_telegram_settings` row is enabled with `auto_post_enabled`. `GetCampaignTelegramSettings`/`UpsertCampaignTelegramSettings` already carry `auto_save_media`, and `handlers/routes_telegram.go` exposes the settings to the SPA as `GET`/`PUT /api/campaigns/:id/telegram` (DM-only). Command dispatch is a single `commandRegistry` slice in `telegram/commands.go`, with helpers `resolveMemberCampaign`, `campaignRole`, and `sendEscapedReply` for campaign resolution, role checks, and safe HTML replies.

Constraint: `campaign_telegram_settings` is raw-managed — it is created by plain DDL in `db/migrations/062_telegram.go`, listed in `rawOnlyTables` in `db/schema_consistency_test.go`, and has no ent model. New columns therefore belong in the versioned migration chain, not in `db/migrations/safe_alters.go` (which exists only for ent-owned tables whose columns `ent.Schema.Create` would otherwise drop). `db/migrations_registry_test.go` pins a SHA-256 fingerprint of the registry and must be regenerated whenever migrations change.

## Goals / Non-Goals

**Goals:**

- Let a DM post an announcement to the campaign chat from Telegram with a clear confirmation or rejection.
- Let any member opt into a weekly digest for a campaign, with at-most-once-per-week delivery.
- Let each campaign pace automatic recap posts with a minimum minute interval and suppress them during recurring UTC quiet hours.
- Expose cadence, quiet hours, and the last automatic post time through the existing settings API and UI without changing any existing field or signature.

**Non-Goals:**

- Digests with configurable content, custom week boundaries, or non-weekly frequencies; the schema allows a `frequency` column but only `weekly` is implemented.
- Scheduled/silent sending hours that queue and flush later; quiet hours suppress, they do not defer.
- Per-member time zones, DST handling, or local-time quiet windows — everything is UTC.
- Changing recap delivery, media posting, or the auto-save behaviour already shipped.

## Decisions

1. **Announcements reuse the resolved-campaign and role helpers.** `/announce <text>` resolves the caller's campaign with `resolveMemberCampaign`, confirms DM/admin with `campaignRole`, rejects empty text, and sends through the existing `sendEscapedReply`-style HTML-escaped path to the bound chat. Every rejection (no bound chat, campaign not enabled, not a DM) is a reply to the caller, never a silent drop.
   - Alternative: a new admin-only HTTP endpoint — rejected; the point is to announce from the chat the table is already in.

2. **Digest subscriptions are a new raw-managed table.** Migration 068 creates
   `telegram_digest_subscriptions(user_id INTEGER NOT NULL, campaign_id INTEGER NOT NULL, frequency TEXT NOT NULL DEFAULT 'weekly', last_sent_at TEXT, created_at TEXT NOT NULL, PRIMARY KEY(user_id, campaign_id))`
   plus `CREATE INDEX idx_telegram_digest_subscriptions_campaign ON telegram_digest_subscriptions(campaign_id)`. The primary key enforces one subscription per `(user, campaign)`; `frequency` is stored even though only `weekly` is produced today.
   - Alternative: reuse a generic subscriptions table — rejected; nothing else needs this shape, and a raw table matches the existing Telegram tables.

3. **New columns live in migration 069.** The four `ALTER TABLE campaign_telegram_settings ADD COLUMN` statements (`auto_post_cadence_minutes INTEGER NOT NULL DEFAULT 0`, `quiet_start_hour INTEGER`, `quiet_end_hour INTEGER`, `auto_post_last_at TEXT`) are plain DDL inside versioned migration 069. Because `campaign_telegram_settings` is raw-managed and has no ent model, `ent.Schema.Create` will not rebuild or drop it, so there is no need to route these through `db/migrations/safe_alters.go`.
   - Alternative: put the ALTERs in `safe_alters.go` — rejected; that file is only for ent-owned tables.

4. **Migration numbering is 068 and 069.** `db/migrations/registry.go` is at version 67 on `main`; this change appends 068 (digest table) then 069 (the ALTERs). The new table is added to `rawOnlyTables` in `db/schema_consistency_test.go`, and the hard-coded SHA-256 fingerprint in `db/migrations_registry_test.go` is regenerated.
   - Alternative: one combined migration — rejected; keeping table creation and column additions separate keeps each step reviewable.

5. **Digest due-ness is a seven-day window over `last_sent_at`.** A subscription is due when `last_sent_at` is null or older than seven days; `created_at` records opt-in time but does not gate the first send beyond that. A user can only subscribe to or unsubscribe from campaigns they belong to (the same membership helper the rest of the bot uses). `scanDigests(now)` is wired into the same one-minute ticker as `scanAutoPost`.
   - Alternative: a cron-style weekday gate — rejected; a rolling week is simpler and avoids catch-up surprises.

6. **Send/claim ordering makes a repeat the acceptable failure mode.** Because due-ness is a window rather than a consumed flag, a crash between sending and updating the timestamp would re-send on the next tick. The implementation writes the claim (`last_sent_at`) before or atomically with the send so that the worst case is a missed digest, not a duplicated one; if the ordering is chosen the other way, a repeat send is the accepted failure mode. This is recorded deliberately rather than hidden.
   - Alternative: a per-period dedupe key — rejected as over-engineering for a weekly summary.

7. **Cadence and quiet hours gate only automatic posting.** `scanAutoPost` checks, in order: the campaign is enabled and auto-post is on (unchanged); no post is due before `auto_post_cadence_minutes` have elapsed since `auto_post_last_at`; the current UTC hour is not inside the quiet window. Only the automatic path updates `auto_post_last_at`. `DeliverRecap` invoked from a command or a manual/retry path is never paced and never advances the cadence clock for its own sake, so on-demand delivery stays immediate.
   - Alternative: pace all deliveries — rejected; a DM's explicit request must never be delayed.

8. **Quiet window semantics are explicit.** `quiet_start_hour`/`quiet_end_hour` are UTC hours 0–23, inclusive at the start hour and exclusive at the end hour, wrapping across midnight when `start > end`. Either both are set or both are null; the API rejects a half-set pair with 400. Unset quiet hours never suppress anything.
   - Alternative: minute-of-day windows — rejected; hourly granularity matches the feature request and the UI.

9. **The settings API grows additively.** `GET` adds `auto_post_cadence_minutes`, `quiet_start_hour`, `quiet_end_hour`, and `auto_post_last_at`; `PUT` accepts `auto_post_cadence_minutes`, `quiet_start_hour`, `quiet_end_hour` (null clears both quiet hours) and continues to accept `auto_save_media` with identical behaviour. Values outside 0–1440 (cadence) or 0–23 (hours) are rejected with 400. The endpoint stays DM-only.
   - Alternative: a separate scheduling endpoint — rejected; the settings row, UI section, and DM check are already in one place.

## Risks / Trade-offs

- **Repeat digests on crash.** Due-ness is a window, so a failure between send and timestamp write can re-send on the next tick. Mitigate by ordering the claim so a missed digest is preferred over a duplicate, or by accepting a repeat as the failure mode; either way, record the choice in the code.
- **Quiet hours are UTC-only**, which can land a DM's local evening in the wrong window. Mitigate by labelling the fields "(UTC)" in the UI and documenting it in the requirement; per-user time zones are out of scope.
- **Stale `auto_post_last_at` blocks a first post.** If a campaign sets cadence after a long silence, the last automatic post time still governs; this is intended pacing, but the UI shows the field so a DM can read why a post is waiting.
- **Migration fingerprint churn.** Forgetting to regenerate the registry hash in `db/migrations_registry_test.go` reddens the suite; the task list makes it an explicit step.
- **`/digest` naming collisions.** A member may expect a digest of everything; the reply and the command description must state clearly that it is a weekly campaign summary, and `/digest off` must be unambiguous.
