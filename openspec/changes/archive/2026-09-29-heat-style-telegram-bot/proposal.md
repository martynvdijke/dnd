# Change: Heat-style Telegram bot runtime with character commands

## Why

Villum's Telegram integration grew as a purpose-built recap-delivery bot: a hand-rolled Bot API client, settings read ad hoc from `app_settings` and `campaign_telegram_settings` by several packages, and a hard-coded command switch with no menu, no keyboards, and no way to grow. The HEAT project (`~/projects/heat`) shows a maintainable shape for the same problem: a supervised `go-telegram/bot` client that follows settings changes live, one command registry that drives dispatch, `/help`, and the native Telegram command menu, and inline keyboards for guided interaction.

Porting that runtime lets Villum's bot grow from "recap poster" into a table companion — overviews, character sheets and stats, claiming a character, guided character creation — without duplicating transport, settings, and dispatch logic a fourth time.

## What Changes

- **Runtime port**: replace the hand-rolled `telegram/client.go` + `poller.go` with `github.com/go-telegram/bot` (the library HEAT uses). A supervisor re-reads settings on a ticker and starts, stops, or restarts the client when token/mode change, with no server restart.
- **Single settings surface**: every Telegram setting (bot token, webhook secret, mode, grace minutes, update offset, per-campaign chat settings) is read/written through typed accessors in `telegram/config.go`; handlers, scheduler, and routing stop issuing their own SQL. All keys keep their encryption and env overrides.
- **Command registry**: one `command` table drives dispatch, `/help`, and `SetMyCommands`, with categories (`characters`, `campaign`, `notifications`, `bot`), aliases, and per-command navigation keyboards.
- **New commands**: `/overview`, `/characters`, `/stats`, `/sheet`, `/claim`, `/unclaim`, `/create`, `/subscribe`, `/unsubscribe`, `/status`, joining the existing `/help`, `/start <code>`, `/unlink`, `/recap`, `/lastrecap`.
- **Character claiming**: a linked Telegram user can bind one character as their table character (`/claim`, tapped from an inline keyboard); `/create` walks a guided flow (name → race → class → level → campaign) and creates the character through the server's existing creation path, then claims it. `/stats` and `/sheet` default to the claimed character.
- **Transport parity**: polling runs through the library; webhook mode keeps the existing Gin endpoint and secret check, feeding decoded updates into the running client. A native command menu is registered on every client start.
- Recap delivery, linking, auto-post, and DM opt-in behavior is preserved; the existing `telegram-recap-bot` spec's configuration requirement is amended to require live application through the single settings surface.

## Capabilities

### New Capabilities

- `telegram-bot-runtime`: supervised `go-telegram/bot` client (polling + webhook), unified settings access with live reconfiguration, command registry, native command menu, inline keyboards, and notification/subscription commands.
- `telegram-character-commands`: table companion commands — campaign overview, character list, sheet, and stats; claiming an existing character; guided character creation; claim-scoped defaults and access rules.

### Modified Capabilities

- `telegram-recap-bot`: the bot configuration requirement now requires settings to be applied through a single settings surface and live reconfiguration (no restart) for token/mode changes; recap commands become registry commands with unchanged behavior.

## Impact

- **Code**: `telegram/` package is largely rewritten (`client.go`, `poller.go`, `dispatcher.go`, `handlers.go`, `config.go`, `webhook.go` gain/replace logic; new `commands.go`, `keyboard.go`, `settings.go`, `characters.go`, `claims.go`); `handlers/routes_telegram.go` loses direct settings SQL and wires the bot character-creation callback; `handlers/characters_crud.go` extracts a reusable creation path; `handlers/recaps.go` and `router.go`/`app.go` wiring.
- **Dependencies**: adds `github.com/go-telegram/bot` (zero transitive Telegram deps; already vendored in HEAT, same Go toolchain). The hand-rolled client is removed.
- **Data**: new `telegram_character_claims` table (migration 063) with one claim per Telegram user and one claimant per character; existing tables and columns are untouched.
- **APIs**: existing Telegram HTTP routes keep their shapes; `/api/telegram/status` gains the claimed character; admin settings payload is unchanged.
- **Tests**: telegram Go unit tests reworked around the library options (mock server via `TELEGRAM_API_BASE`), new tests for registry, claims, create flow, and settings; `tests/telegram-mock.ts` gains `setMyCommands`/`getUpdates` long-poll behavior; `tests/telegram.spec.ts` gains command-surface coverage.
- **Not in scope**: Telegram media upload/download (tracked by the separate `telegram-rich-media` change, whose WIP is preserved on `feat/telegram-rich-media`), dice rolling, and any change to web UI behavior.
