## Context

Villum ships a PWA: `static/sw.js` (cache strategy, version-substituted at serve time) registered by `ts/pwa.ts`. Server-side, `app_settings` is the established key/value store for admin-managed config (email SMTP, eink, autosave interval — see `handlers/settings.go`), and `handlers/backup.go` provides the background-scheduler pattern (`StartBackupScheduler()` started from `main.go`). Recap creation lives in `handlers/recaps.go` (`CreateCampaignRecap`, `MarkRecapAsSent`); calendar events live in `db/campaign_events.go`.

Everything user-facing is behind session auth with CSRF middleware; public routes are the exception list in `RegisterPublicRoutes`. Push subscriptions cannot use cookies — the browser's push service delivers payloads to the service worker directly.

## Goals / Non-Goals

**Goals:**
- Reliable delivery of two notification types: recap published, session starting soon
- Zero-config for players beyond granting permission; admin configures keys once
- Clean lifecycle: stale/expired subscriptions are pruned automatically
- Respect user attention: per-campaign mute, no marketing-style noise

**Non-Goals:**
- Real-time gameplay notifications (turn timers, combat pings) — WebSocket already covers in-app realtime
- Email fallback logic or digest batching
- Per-notification-type granularity beyond campaign mute
- iOS Safari < 16.4 / non-installed-PWA workarounds

## Decisions

### 1. `github.com/SherClockHolmes/webpush-go` as the delivery library
**Why:** De-facto standard Go implementation of the Web Push protocol (VAPID + payload encryption); maintained, minimal deps. **Alternative considered:** hand-rolling RFC 8291 encryption — rejected, high-risk crypto code for no benefit.

### 2. VAPID keys + subject in `app_settings`, managed like email settings
Keys auto-generate on first save if absent; admin UI shows public key, allows regeneration and a test push. Stored keys are never returned in full to non-admins and never logged.

**Why:** Matches the existing generic settings mechanism exactly (email/eink/autosave pattern). **Alternative:** env vars — rejected, inconsistent with how all other integrations are configured.

### 3. New `push_subscriptions` table via the existing migration mechanism
Columns: `id, user_id, endpoint (unique), p256dh, auth, expiration_time (nullable), created_at`. One row per browser subscription; a user may have several. Delivery failure with 404/410 (or expired) deletes the row.

**Why:** Raw SQL access via `db/` matches the majority of simple entities; unique endpoint prevents duplicate rows on re-subscribe. **Alternative:** JSON blob on users — rejected, unqueryable for scheduler fan-out.

### 4. Triggers hook into existing flows, not new abstractions
- Recap published → fire-and-forget goroutine after recap create/mark-sent succeeds
- Session reminder (hybrid, two sources) → `StartPushReminderScheduler()` scans each minute:
  - **External feed events** (`google_events_cache` rows with parseable future `start_time` and `all_day = 0`): minute-precision reminder at `now >= start_time - lead_minutes` (admin setting, default 60)
  - **Local calendar sessions** (`campaign_calendar_events` rows with `event_type = 'session'` and future date-only `event_date`): day-granular "Session coming up" reminder when the event is within `lead_days` (admin setting, default 1)
  - Dedup via a `reminder_sent_at` marker on the subscription-independent event level (keyed by source + event id)

Payload JSON: `{title, body, url, tag}` — `tag` collapses duplicates; `url` deep-links to the recap/event.

**Why:** No event-bus machinery for two trigger points. **Alternative:** generic notification-events table + plugins — YAGNI today.

### 5. Service worker handles display; page handles subscribe
`sw.js` gains `push` (show notification from payload, fallback title if parse fails) and `notificationclick` (focus existing window or open `url`). Subscription flow in `ts/pwa.ts`: request permission on explicit user action, `pushManager.subscribe` with the server-provided VAPID public key, POST subscription to the server; unsubscribe mirrors it.

**Why:** Standard split; keeps sw.js dumb and cache-versioning intact.

### 6. Mute storage: `push_mutes(user_id, campaign_id)`
Scheduler and recap fan-out skip muted members. Toggle lives in the campaign member UI.

**Why:** Two columns answer every future "should this user get X for campaign Y" question; reuses member-list UI surface.

## Risks / Trade-offs

- **Push permission prompts annoy users** → Permission is only requested from an explicit opt-in control, never on load.
- **iOS quirks (installed PWA required)** → Feature detection hides controls where unsupported; documented in admin help text.
- **Scheduler fan-out cost** → One minute cadence with an indexed `start_time` scan is trivial at this scale; sends are batched per event.
- **Key loss/regeneration orphans subscriptions** → Regenerating VAPID keys invalidates old subscriptions; endpoints fail delivery once and are pruned by rule 3.
- **New dependency supply-chain** → Pinned via go.mod; small transitive surface.

## Migration Plan

1. Ship migration (new tables only), backend package, routes, admin section.
2. Ship sw.js/pwa.ts changes — version bump forces SW update on next load.
3. Admin generates/saves VAPID keys, sends test push.
4. Rollback: feature-flag off via removing admin keys? No — rollback is code-level: revert commit; orphaned tables are harmless.

## Resolved Questions

- Reminder lead time → **two admin settings**, matching the hybrid sources: `push_reminder_lead_minutes` (external feed events, default 60) and `push_session_reminder_lead_days` (local calendar sessions, default 1).
- Recap notifications fire only on explicit publish/send — drafts never notify.
