## Why

Villum can only reach players between sessions via email recaps. Players who don't live in their inbox miss recaps and forget session times. The app is already installed as a PWA on players' phones — Web Push closes the loop by letting the server pull people back in: "recap published", "session starts in an hour".

## What Changes

- Add Web Push support end-to-end: VAPID key configuration (admin-managed), browser subscription lifecycle, and server-side delivery
- Store push subscriptions per user (endpoint, p256dh key, auth secret, expiration) in a new table
- Notify campaign members when a recap is published/sent
- Send a session reminder push a configurable time before calendar events (default 60 minutes), dispatched by a background scheduler following the backup-scheduler pattern
- Add a `push` event handler + notification display/click behavior to the existing service worker
- Add per-user, per-campaign mute preference so players can silence a campaign without leaving it
- Add an admin settings section: generate/view VAPID keys, subject, and a "send test push" action
- Degrade silently on browsers/devices without push support

## Capabilities

### New Capabilities

- `web-push-notifications`: Subscription lifecycle, VAPID configuration, delivery pipeline, and the recap-published and session-reminder triggers
- `push-preferences`: Per-user mute controls for campaign notifications

### Modified Capabilities

- *(none)*

## Impact

- **Backend**: New push package (subscription CRUD + delivery), hooks in recap creation flow, reminder scheduler started from `main.go` alongside `StartBackupScheduler`, new routes (subscribe/unsubscribe/preferences/test-push); **new Go dependency** (`github.com/SherClockHolmes/webpush-go`)
- **Database**: One migration adding `push_subscriptions` (+ mute storage); no changes to existing tables
- **Frontend**: `static/sw.js` gains `push`/`notificationclick` handlers; `ts/pwa.ts` gains subscription flow and permission UI; small preference toggle in campaign member area
- **Admin**: New settings section in `static/admin.html` + `ts/admin.ts` following the existing email-settings pattern
- **Risks**: iOS requires installed-PWA context for push; documented, not worked around
