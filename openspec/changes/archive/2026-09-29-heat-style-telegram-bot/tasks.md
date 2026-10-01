# Tasks: Heat-style Telegram bot runtime with character commands

## 1. Foundation

- [x] Add `github.com/go-telegram/bot` to `go.mod`/`go.sum` (same version family as HEAT) and confirm `go build ./...` passes
- [x] Create migration `063_telegram_character_claims.go` (table `telegram_character_claims` with `telegram_user_id` PK, `character_id` UNIQUE FK to `characters`, `created_at`), register it in `db/migrations/registry.go`, and add it to the schema consistency test
- [x] Create `telegram/settings.go`: `Settings` struct + `LoadSettings()`, single KV primitive for stored values, env overrides (`TELEGRAM_BOT_TOKEN`, `TELEGRAM_WEBHOOK_SECRET`, `TELEGRAM_API_BASE`), and `CampaignTelegramSettings` + `GetCampaignTelegramSettings`/`UpsertCampaignTelegramSettings` accessors
- [x] Rewrite `telegram/config.go` accessors to delegate to the single settings surface; keep signatures used by tests and handlers; keep encryption and masking behavior
- [x] Replace direct `campaign_telegram_settings` SQL in `handlers/routes_telegram.go`, `telegram/scheduler.go`, and `telegram/routing.go` with the new accessors; behavior unchanged (`GetCampaignTelegram`/`SetCampaignTelegram` responses identical)
- [x] Add unit tests for `LoadSettings` (stored values, env overrides, encryption) and the campaign settings accessors (upsert, partial update semantics, missing row)

## 2. Runtime Port

- [x] Rewrite `telegram/client.go` on `github.com/go-telegram/bot` (`WithServerURL` from settings, `WithDefaultHandler`, `WithAllowedUpdates(message, callback_query)`, HTML parse mode) and remove the hand-rolled HTTP client, `poller.go`, and `RetryAfterError`
- [x] Implement the supervisor (15s ticker + immediate wake): read settings, create/replace client when `(token, apiBase)` changes, run polling only for polling/auto-resolved-polling, register/delete webhook on transport change, and stop cleanly for `off`/no token
- [x] Implement package-level `SendMessage`, `SendDocument`, `GetChat`, `SetWebhook`, `DeleteWebhook`, `GetBotUsername` over the supervised client with a short-lived fallback client built from stored settings
- [x] Rewrite `webhook.go` to validate the secret exactly as before and feed updates to the running client via `ProcessUpdate`, recording the update offset for the admin payload
- [x] Make admin settings save wake the supervisor so token/mode changes apply immediately (still converging within the ticker in all cases)
- [x] Add supervisor and sender unit tests: start on token, stop on clear, restart on token change, no restart on unrelated change, polling vs webhook client paths, webhook 403 behavior, offset recording

## 3. Command Registry and Keyboards

- [x] Create `telegram/commands.go` with the `command` struct, registry, categories (`characters`, `campaign`, `notifications`, `bot`), `findCommand` alias resolution, `helpText`, and `escapeHTML`
- [x] Move `/help`, `/start`, `/unlink`, `/recap`, `/lastrecap` into the registry with unchanged behavior and messages; implement `/help@BotName` stripping and unknown-command help
- [x] Register the native command menu from the registry on every client start (`SetMyCommands`), excluding hidden commands
- [x] Create `telegram/keyboard.go`: navigation keyboard (`nav:/<command>`), campaign keyboards, claim candidate keyboard, level picker, and the testable `handleCallbackData` router with access enforcement
- [x] Add `/subscribe`, `/unsubscribe`, `/status` commands backed by the identity DM flag and settings, with confirmation replies
- [x] Add registry/keyboard unit tests: help grouping, alias dispatch, unknown command, menu payload, callback routing for each namespace, hostile-text escaping

## 4. Character Commands

- [x] Create `telegram/claims.go`: claim/unclaim store, candidate listing (characters the linked user may edit, excluding those claimed by others), claim re-validation, and clearing claims on unlink
- [x] Implement `/claim` (candidate keyboard + id/name argument) and `/unclaim` with the refusal and unavailable-claim replies from the spec
- [x] Implement `/characters` with the claimed marker, campaign column, empty state, and access restriction
- [x] Implement `/sheet` and `/stats` renderers (identity, abilities with modifiers, HP/AC/initiative/speed/proficiency/passive perception; sheet adds conditions, currency, and feature/spell/inventory counts) with claimed/selected/default resolution and not-found responses
- [x] Implement `/overview` (campaign list with role and character count; campaign detail with members, party classes/levels, latest recap) restricted to membership
- [x] Extract `createCharacterCore(ctx, uid, ch models.Character)` from `CreateCharacter` in `handlers/characters_crud.go` so HTTP and bot share validation, defaults, currency creation, and campaign attach
- [x] Add the `SetCharacterCreator` seam in `telegram` and wire `handlers.BotCharacterCreator` from `router.go`/`app.go`
- [x] Implement the guided `/create` flow (name → race → class → level 1–20 → campaign picker with none), validation-in-place, `/cancel` and other-command abort, auto-claim, and failure reporting
- [x] Add unit tests: claim uniqueness and re-validation, unlink clearing, candidate filtering, listing/sheet/stats/overview rendering and denials, create-flow transitions/validation/abort, fake creator success and failure

## 5. End-to-End and Regression

- [x] Extend `tests/telegram-mock.ts`: answer `setMyCommands` (store the payload), honor `getUpdates` long-poll `timeout` with a real delay and no busy loop, and keep `sendMessage` recording working for HTML messages
- [x] Extend `tests/telegram.spec.ts`: exercise `/help`, `/characters`, `/claim`, `/sheet`, and `/create` through the webhook against the mock, asserting `sentMessages` content, plus a happy-path recap delivery regression
- [x] Verify existing telegram e2e flows still pass (admin token/mode/grace, account link/unlink, campaign chat binding, webhook 403) and no new `data-testid` is introduced without e2e references
- [x] Run `go test ./telegram/... ./handlers/...` with coverage and confirm the Go floors (total ≥20%, `handlers/` ≥40%) and vitest floors are unaffected
- [x] Run `task ci` (typecheck, vitest, build, e2e, coverage gates) until green

## 6. Ship

- [x] Commit with Conventional Commits (e.g. `feat(telegram): port bot runtime to go-telegram/bot with character commands`), push the branch, and open the PR with `gh pr create --fill`
- [x] Wait for checks (`gh pr checks --watch`); on red, read `gh run view --log-failed`, fix, push, and re-wait
- [x] Merge when green (`gh pr merge --squash --delete-branch`) and archive the OpenSpec change
