# Change: Telegram group commands and friendlier onboarding

## Why

The Telegram bot now has a proper runtime and character commands, but it is still private-chat-first: onboarding is a bare command list and an unlinked user gets a terse "link first" line, while the richest shared data a table actually wants in a group chat — party inventory, who has visited which places, what quests are still open, campaign statistics — is invisible to the bot even though the campaign chat is already bound for recaps.

This change makes the bot a table companion for bound campaign chats and gives new users a guided first run.

## What Changes

- **Group commands**: `/items` (campaign party inventory), `/quests` (open quests of party characters), `/visits` (party location visits), and campaign statistics shown by `/stats` in a bound campaign chat (character stats stay in private chats).
- **Bound-chat campaign resolution**: commands resolve the campaign from `campaign_telegram_settings` (bound, enabled chat); private chats fall back to the linked user's campaigns with the existing single-campaign/picker behavior; unbound groups get binding instructions.
- **Nicer onboarding**: a friendly `/start` welcome with a navigation keyboard, clearer linking steps for unlinked users, a post-link success message that points at `/claim` and `/create`, and a welcome when the bot is added to a group explaining how to bind the campaign.
- The bot subscribes to `my_chat_member` updates so group additions are noticed.

## Capabilities

### New Capabilities

- `telegram-group-commands`: campaign-scoped group commands (party items, open quests, location visits, campaign statistics) resolved from the bound chat, with empty states and access rules.
- `telegram-bot-onboarding`: first-run welcome, linking guidance, post-link next steps, and the group-added welcome.

## Impact

- **Code**: new `telegram/group.go`; additions in `telegram/commands.go` (items/quests/visits, context-aware stats), `telegram/bot.go` (my_chat_member handling), `telegram/handlers.go` (onboarding texts), `telegram/client.go` (allowed updates).
- **Data**: read-only queries against `party_items`, `quests`, `character_locations`, `locations`, `sessions`, `campaign_characters`, `campaign_members`, `npcs`; no schema changes.
- **Tests**: unit tests for campaign resolution and renderers; e2e group-command coverage through the webhook.
- **Out of scope**: mutating commands (adding items/quests from chat), dice rolling, and any web UI change.
