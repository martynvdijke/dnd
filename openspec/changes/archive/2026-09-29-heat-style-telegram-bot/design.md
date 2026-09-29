# Design: Heat-style Telegram bot runtime with character commands

## Context

Villum's Telegram feature is one package (`telegram/`) that owns a hand-rolled Bot API client, a polling loop, a Gin webhook handler, a recap scheduler, identity linking, and a five-command dispatcher. Settings are split between `app_settings` key-values (owned by `telegram/config.go`) and `campaign_telegram_settings` rows queried directly by `handlers/routes_telegram.go`, `telegram/scheduler.go`, and `telegram/routing.go`. The HEAT project's `telegram/` package solves the same lifecycle problems with `github.com/go-telegram/bot`: a supervisor that starts/stops/restarts the client as settings change, a command registry that drives dispatch, `/help`, and `SetMyCommands`, and inline keyboards for guided flows.

The user asked for Villum's bot to work the same way as HEAT's, to double-check that all settings use one access method, and to add overview/stats/characters/sheets commands plus claiming or creating a character through the bot. This change ports the runtime and adds the character-facing command surface.

Constraints:

- Webhook transport with the existing Gin endpoint and secret check must keep working (Villum exposes `POST /api/telegram/webhook`; updates are posted by Telegram and validated against `app_settings` secret).
- Existing HTTP routes and admin payload shapes must not change.
- `telegram/` sits below `handlers/` in the import graph (`handlers` imports `telegram`), so character-creation logic must flow into the bot without creating a cycle.
- Coverage floors apply: Go total ≥20%, `handlers/` ≥40%, `middleware/` ≥40%; e2e runs against `./villum-server` with a Playwright Telegram Bot API mock (`tests/telegram-mock.ts`).

## Goals / Non-Goals

**Goals:**

- One supervised runtime built on `github.com/go-telegram/bot` with polling and webhook transports.
- One settings surface: every Telegram setting read/written through typed accessors; live reconfiguration without a server restart.
- One command registry driving dispatch, `/help`, and the native command menu; keyboards for selection flows.
- Character-facing commands: `/overview`, `/characters`, `/sheet`, `/stats`, `/claim`, `/unclaim`, `/create`, plus `/subscribe`, `/unsubscribe`, `/status`.
- Claim a character or create one through the bot, using the server's existing validation and side effects.

**Non-Goals:**

- Telegram media upload/download (owned by the `telegram-rich-media` change, WIP preserved on `feat/telegram-rich-media`).
- Dice rolling, live combat tracking, or any web UI change.
- A general `characters` service package; this design uses a narrow callback seam instead (see Decisions).

## Decisions

### D1. Adopt `github.com/go-telegram/bot` as the only transport

Use the same library and aliases (`tgbot`, `tgmodels`) as HEAT. Its `WithServerURL` option points the client at `TELEGRAM_API_BASE`, preserving the current test/mock setup. Alternatives: keep the hand-rolled client (rejected: user explicitly asked for HEAT parity; it is the source of the split settings and dispatch logic), or `telegram-bot-api/v5` (capable but HEAT already validates the chosen library and its option style suits the supervisor).

### D2. Supervised client, mode controls polling rather than client existence

A `Bot` struct owns one supervisor goroutine and at most one `tgbot.Bot` client:

- Every 15 seconds (plus an immediate wake channel poked by the admin settings save) the supervisor loads `Settings` and compares `(token, apiBase)` with the running client.
- No token, or mode `off` → stop the client (cancel context, nil it).
- Token present → ensure a client exists, recreated only when token/apiBase changed. `Settings.Mode` decides what runs under that client:
  - `polling`: run `client.Start(ctx)` in a goroutine.
  - `webhook`: do not poll; updates arrive via `ProcessUpdate` from the Gin handler.
  - `auto` resolves to `EffectiveMode()` (webhook when `BASE_URL` is HTTPS, else polling).
- On transport change, register or delete the webhook exactly once (`EnsureWebhookState`).

Alternatives: rebuild the client on every settings tick (wasteful, and loses the in-flight long poll for unrelated edits); tie client existence to mode (webhook mode would then need a separate sender path for handlers).

### D3. Package-level API compatibility for senders

`telegram.SendMessage`, `SendDocument`, `GetChat`, `SetWebhook`, `DeleteWebhook`, `GetBotUsername` keep their signatures. They use the supervised client when it exists and otherwise build a short-lived client from the stored token and API base (the `SendTest` pattern from HEAT). This keeps `handlers/routes_telegram.go` and `telegram/routing.go` unchanged at their call sites.

### D4. Webhook handler feeds the running client

`WebhookHandler` keeps the constant-time secret check and 403 behavior, decodes `tgmodels.Update`, records the update offset for the admin payload, and calls `ProcessUpdate` on the supervised client. If no client is running (token cleared mid-flight), it still answers `{"ok":true}` and drops the update. `ProcessUpdate` is public in the library, so no private API or separate HTTP server is needed.

### D5. Command registry modeled on HEAT

A `command` struct (`name`, `usage`, `desc`, `category`, `aliases`, `showNav`, `run`) in a single registry slice drives:

- `findCommand` dispatch after the `/command@BotName` split,
- `/help` rendered per category,
- `SetMyCommands` on every client start (registry minus hidden commands).

Categories: `characters`, `campaign`, `notifications`, `bot`. Existing recap and linking commands move into the registry with unchanged behavior.

### D6. Inline keyboards and callback routing

Callback data namespaces:

- `nav:/<command>` — run a registry command and attach the navigation keyboard (HEAT's pattern).
- `claim:<character_id>` — claim the tapped candidate.
- `sheet:<character_id>` — render that character's sheet.
- `campaign:<id>` — select a campaign for `/overview` or the `/create` flow.
- `create:<step>:<value>` — inline answers inside the create flow (level and campaign pickers).

A navigation keyboard is attached to top-level views (`/overview`, `/characters`, `/sheet`, `/stats`) with buttons for Characters, Sheet, Stats, and Recap. Callback handling lives in `handleCallbackData` so it is testable without a live client.

### D7. Claim persistence: dedicated table keyed by Telegram user

New table `telegram_character_claims`:

```sql
CREATE TABLE IF NOT EXISTS telegram_character_claims (
  telegram_user_id INTEGER PRIMARY KEY,
  character_id     INTEGER NOT NULL UNIQUE REFERENCES characters(id) ON DELETE CASCADE,
  created_at       TEXT NOT NULL DEFAULT (datetime('now'))
);
```

One claim per Telegram user, one claimant per character. Claims are cleared when an identity unlinks and re-validated on every read: if the linked user can no longer edit the character (ownership rule of `canEditCharacter`: own the character or be campaign owner/DM of a campaign the character is attached to), the bot answers that the claim is unavailable and suggests `/unclaim`. A separate table beats a `telegram_identities.character_id` column because the uniqueness side (one claimant per character) and lifecycle (cleared on unlink) are clearer, and identities stay untouched.

### D8. Character creation flows through a callback seam

`telegram` defines:

```go
type CreateCharacterInput struct { Name, Race, Class string; Level int; CampaignID int64 }
type CharacterCreator func(ctx context.Context, userID int64, in CreateCharacterInput) (CreatedCharacter, error)
func SetCharacterCreator(fn CharacterCreator)
```

`router.go` wires `telegram.SetCharacterCreator(handlers.BotCharacterCreator)`. `handlers/characters_crud.go` extracts the ent create + defaults + currency + campaign attach block out of the HTTP handler into `createCharacterCore(ctx, uid, ch models.Character)`, so the HTTP endpoint and the bot share one creation path. Alternatives: `telegram` importing `handlers` (import cycle), duplicating ent creation in the bot (drifts from validation/defaults), or a new service package (bigger refactor; this seam is the migration point if that package ever lands).

### D9. Create flow is in-memory and cancellable

`/create` collects name → race → class → level → campaign through a per-chat flow struct in a mutex-guarded map (HEAT's `pendingQuotes` pattern). Any other command, `/cancel`, or a timeout drops the flow. Flows do not survive a server restart; users re-run `/create`. Persisting half-finished characters was rejected as it would create drafts the web UI does not understand.

### D10. Stats vs sheet split

- `/stats` is the glanceable combat block: ability scores with modifiers, HP, AC, initiative, speed, proficiency bonus, passive perception.
- `/sheet` is the fuller read: identity (race/class/subclass/level), stats block, conditions/exhaustion/inspiration, currency, and counts of features/spells/inventory items, plus a pointer to the web sheet.

Both accept an optional name or id and default to the claimed character. Messages render with HTML parse mode and escape user content (HEAT's `escapeHTML`), chunked at 4096.

### D11. Single settings surface

`telegram/settings.go` introduces `Settings` + `LoadSettings()` and campaign-level `CampaignTelegramSettings` + `GetCampaignTelegramSettings` / `UpsertCampaignTelegramSettings`. All existing accessors delegate to one KV primitive; `handlers/routes_telegram.go`, `telegram/scheduler.go`, and `telegram/routing.go` lose their direct SQL. Env overrides (`TELEGRAM_BOT_TOKEN`, `TELEGRAM_WEBHOOK_SECRET`, `TELEGRAM_API_BASE`) and encryption at rest are centralized and unchanged.

### D12. Testing strategy

- Unit tests in `telegram/`: mock Bot API server via `TELEGRAM_API_BASE`, library `WithServerURL`; test registry/help/menu, callback routing, claim store, create-flow transitions, settings load, sender fallback, supervisor start/stop with a short interval, and webhook authenticity.
- Handler tests for `createCharacterCore` extraction and the campaign settings accessors.
- E2E: extend `tests/telegram-mock.ts` to answer `setMyCommands` and long-poll `getUpdates` with a small delay (the library polls immediately after an empty response); extend `tests/telegram.spec.ts` to exercise `/help`, `/characters`, `/claim`, `/sheet` through the webhook and assert mock `sentMessages`.

## Risks / Trade-offs

- [Library polling hammers the e2e mock] → Mock honors `timeout` and delays empty `getUpdates` responses; assert the mock stays responsive.
- [Webhook update arrives while no client exists] → Answer 200 and drop (same observable behavior as today when the bot is unconfigured); document in tests.
- [Claims pointing at characters that stop being editable] → Re-validate on read; render "unavailable", offer `/unclaim`.
- [In-memory create flow lost on restart or multi-instance deploy] → Accepted: short flow, `/cancel` and retry; documented in D9.
- [Ownership predicate duplicated as SQL in the bot] → Keep the SQL mirror small, cover it with tests, and reference `isDMOfCharacter` in a comment.
- [Migration number collision with `feat/telegram-rich-media`'s 063] → Coordinate at merge time; the later merge renumbers.
- [HTML parse mode breaks on user content containing `<`/`&`] → Escape every interpolated value; test with hostile names.

## Migration Plan

1. Add migration `063_telegram_character_claims.go` (table above, registered in `db/migrations/registry.go`).
2. Land the runtime port and commands behind the same routes; no client-facing API changes.
3. Deploy: replace the binary; the supervisor converges on the stored settings within 15 seconds; no restart of the process is needed to change token/mode.
4. Rollback: revert the binary; the claims table is additive and harmless to leave in place.

## Open Questions

- None blocking. Whether `/stats` should later grow a party-wide mode can be revisited after the per-character version ships.
