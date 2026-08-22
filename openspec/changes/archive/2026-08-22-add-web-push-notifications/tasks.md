## 1. Foundation

- [x] 1.1 Add `github.com/SherClockHolmes/webpush-go` dependency
- [x] 1.2 Migration: create `push_subscriptions` (unique endpoint, user_id, p256dh, auth, expiration_time, created_at) and `push_mutes` (user_id, campaign_id, unique pair)
- [x] 1.3 Push delivery helper: build VAPID config from `app_settings`, send payload `{title, body, url, tag}`, prune on 404/410/expired

## 2. Subscription & preference endpoints

- [x] 2.1 `POST /api/push/subscribe` and `POST /api/push/unsubscribe` (authed, upsert/delete by endpoint)
- [x] 2.2 `GET/PUT /api/campaigns/:id/push-mute` for the per-user mute toggle
- [x] 2.3 Handler tests: subscribe/upsert/unsubscribe, auth rejection, mute round-trip

## 3. Admin configuration

- [x] 3.1 Settings accessors + routes for VAPID keys/subject and both reminder leads following the email-settings pattern; auto-generate keys when absent on save; never log or expose the private key to non-admins. Leads: `push_reminder_lead_minutes` (external feed events, default 60) and `push_session_reminder_lead_days` (local sessions, default 1)
- [x] 3.2 Admin UI section in `static/admin.html` + `ts/admin.ts`: key display/generate, subject, lead time, "send test push" button (with data-testid referenced in tests/)
- [x] 3.3 Admin handler tests: save/read round-trip, 403 for non-admin, test-push endpoint

## 4. Service worker & client flow

- [x] 4.1 `static/sw.js`: `push` listener (parse payload, fallback title) and `notificationclick` (focus or open `url`)
- [x] 4.2 `ts/pwa.ts`: opt-in subscribe control — feature-detect, request permission on click, subscribe with server public key, POST to `/api/push/subscribe`; hide controls when unsupported
- [x] 4.3 Campaign member UI: mute toggle wired to the mute endpoint (data-testid referenced in tests/)
- [x] 4.4 Vitest unit tests for the subscription/mute client modules

## 5. Triggers

- [x] 5.1 Recap published hook: fan-out push to subscribed, unmuted campaign members after recap create/mark-sent (fire-and-forget goroutine)
- [x] 5.2 `StartPushReminderScheduler()` in the backup-scheduler pattern, started from `main.go`: scan upcoming events each minute, send once per event per member at lead time, dedup marker
- [x] 5.3 Tests: fan-out skips muted users and prunes dead endpoints; scheduler sends exactly one reminder per event for both feed events and local sessions; all-day/unparseable feed events skipped

## 6. Verification

- [x] 6.1 `go vet ./... && go build ./...` clean
- [x] 6.2 `task test` green including new tests
- [x] 6.3 `npm run typecheck && npm run test:unit` green
- [x] 6.4 Manual smoke: real browser subscription → test push from admin → recap publish push received; muted user receives nothing
