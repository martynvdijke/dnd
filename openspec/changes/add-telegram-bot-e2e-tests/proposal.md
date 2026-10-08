# feat: Telegram bot e2e tests for /login and /claim flows

## Why

The Telegram bot login (email magic-link) and claim (inline-button callback) flows have no end-to-end coverage. Regressions (e.g. silent `/status`) are caught late.

## What Changes

- Add `tests/telegram-bot-flows.spec.ts` with Playwright e2e tests:
  - `/login` full email flow via SMTP mock + magic-link redemption in the same browser context.
  - `/claim` inline-button flow via `callback_query` payload `claim:<id>`, re-claim and `/unclaim`.
  - `/status` linked-user regression.
- No production code changes.

## Capabilities

### New Capabilities
- `telegram-bot-e2e`: end-to-end coverage for bot login, claim and status via webhook + mocks.

## Impact

- Tests only; no API or schema changes.
- Depends on `tests/smtp-mock.ts` + `workerData.smtpMock`/`BASE_URL` wiring from the concurrent lane.
