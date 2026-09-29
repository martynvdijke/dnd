## Why

Recap delivery is limited to in-app and Web Push (installed PWA). Many D&D groups already coordinate in Telegram — including players who never install the PWA or miss Web Push while away from their browser. A recap is exactly the content a group wants to read together in the chat where scheduling already happens.

Telegram fits this better than a WhatsApp bot: Telegram bots can post into group chats/channels *and* receive commands, so one integration covers both "post the recap when the session ends" and "ask the bot for the recap".

## What Changes

- Add a Telegram bot integration: store a bot token (encrypted at rest), post campaign recaps to a linked Telegram chat, and let linked users query recaps from chat
- Link a Villum account to a Telegram identity with a one-time code generated in the web UI and redeemed via `/start <code>`
- Per-campaign Telegram target (group/channel) configured by the campaign DM
- **Manual delivery**: publishing/sending a recap (`SendRecap` / `MarkRecapAsSent`) also posts it to the campaign's Telegram chat
- **Automatic delivery**: a background scheduler posts a recap once its `session_end_date` has passed, when auto-post is enabled for the campaign
- **Inbound queries**: `/recap`, `/lastrecap`, `/help`, `/unlink`; authorization enforced with existing campaign membership
- Delivery log with idempotency, so a recap is never posted twice to the same target
- Chunking to Telegram's 4096-character message limit
- Admin settings surface (bot token, transport mode) and a per-campaign Telegram section

## Capabilities

### New Capabilities

- `telegram-recap-bot`: Bot configuration, account linking, outbound recap delivery (manual + automatic), inbound query commands, and delivery logging/idempotency

### Modified Capabilities

- *(none)* — Web Push and in-app recap behavior are unchanged; Telegram is additive and composes alongside the existing `NotifyCampaignRecapPublished` fan-out

## Impact

- **Backend**: New `telegram/` package (Bot API client, update dispatcher, account linking, delivery routing, auto-post scheduler) plus handlers/routes; hook into the `handlers/recaps.go` publish flow next to `NotifyCampaignRecapPublished`; scheduler registered from `app.go` `registerSchedulers`. No new Go dependency — `net/http` only.
- **Database**: One migration adding `telegram_identities`, `telegram_link_codes`, `telegram_deliveries`, `campaign_telegram_settings`; no changes to existing tables
- **Frontend**: Admin section for bot token + transport mode; per-campaign Telegram target field; "Link Telegram" + DM opt-in in the account/settings area
- **Public surface**: One new unauthenticated endpoint `POST /api/telegram/webhook` (only active in webhook mode), verified via `X-Telegram-Bot-Api-Secret-Token`, registered in `RegisterPublicRoutes` alongside `/api/login` and `/invite/:token`
- **Risks**: Telegram 4096-character message cap; multi-replica polling duplication; freeform recap text vs Telegram parse modes; bot token is a new outbound secret
