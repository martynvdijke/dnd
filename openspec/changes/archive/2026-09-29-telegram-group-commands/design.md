# Design: Telegram group commands and friendlier onboarding

## Context

The bot's commands are private-chat-first: character commands resolve characters through the linked account, campaign commands (`/recap`, `/overview`) resolve campaigns through campaign membership. A Telegram group chat can already be bound to a campaign for recap delivery (`campaign_telegram_settings.chat_id` with `is_enabled=1`), but in that chat the bot knows nothing about the campaign's party items, location visits, open quests, or statistics. Onboarding is a bare command list: `/start` without a code shows help, the unlinked reply is three terse steps, and adding the bot to a group produces no response at all.

The web UI already models all the data needed: `party_items` (campaign-scoped), `quests` per character (statuses `available`, `active`, `complete`, `failed`, `abandoned`), `character_locations` joining characters to `locations` with a relationship (`current`, `hometown`, `visited`, `headquarters`, `quest`, `other`), and campaign analytics (characters, sessions, quests, NPCs, locations, average level) rendered client-side in `ts/party-subtabs.ts`.

## Goals / Non-Goals

**Goals:**

- Read-only group commands that work in a campaign-bound chat without each sender linking a Villum account: `/items`, `/quests`, `/visits`, and campaign statistics.
- A single campaign-context resolver that prefers the bound chat and falls back to the linked user's campaigns.
- Friendlier onboarding: welcome on `/start`, clearer linking instructions, a post-link nudge toward `/claim` and `/create`, and a welcome when the bot is added to a group.
- No schema changes; all queries are projections over existing tables.

**Non-Goals:**

- Mutating commands from Telegram (no creating/editing quests, items, or locations through the bot).
- Dice rolling or live combat.
- Web UI changes.

## Decisions

### D1. Campaign context resolution order

`resolveCampaignContext(c)`:

1. **Bound chat**: `SELECT s.campaign_id, c.name FROM campaign_telegram_settings s JOIN campaigns c ON c.id = s.campaign_id WHERE s.chat_id = ? AND s.is_enabled = 1 LIMIT 1`. A hit wins for any chat type.
2. **Private chat fallback**: if no bound campaign and the sender is linked, reuse the existing `/recap` campaign resolution (`memberCampaignIDs` → single campaign auto-selects, multiple show the picker).
3. **Neither**: a group gets binding instructions ("ask your DM to connect this group from Campaign → Telegram"); a private chat gets the linking instructions.

Alternatives: requiring an explicit `/campaign <id>` argument everywhere (worse UX in the common bound-group case); storing the group's campaign on the sender's identity (wrong scope — the group is the context).

### D2. Access rules for group commands

A bound, enabled campaign chat is the DM's explicit act of sharing campaign data with that group, so the read-only campaign commands are allowed for any member of that group, linked or not. Private-chat command results remain restricted to campaigns the linked user belongs to. No command ever exposes data for a campaign the chat is not bound to.

### D3. `/stats` is context-aware

In a bound campaign chat, `/stats` renders campaign statistics; in a private chat it keeps rendering the claimed character's stats. A separate `/campaignstats` was rejected as undiscoverable; overloading `/stats` matches the user's mental model ("show me the numbers for where I am"). The command's help text notes both modes.

### D4. Renderers and queries

- `/items`: `SELECT name, quantity, notes FROM party_items WHERE campaign_id = ? ORDER BY name`; renders `name × quantity` with notes on a second line, escaped, capped at 40 entries with an "and N more" footer.
- `/quests`: party characters joined to open quests (`status IN ('available','active')`), grouped by character name, with objectives/rewards only when present; capped per character at 5 with a count footer.
- `/visits`: `character_locations` joined to `locations` and `characters`, restricted to party characters, ordered by `character_locations.id DESC`, limit 20; each line shows character, location, type badge, and relationship when not `visited`.
- Campaign statistics mirror the web analytics: characters (campaign roster), sessions (sessions of party characters), quests total + open/completed, NPCs and locations owned by campaign members, average party level to one decimal.

All renderers escape user text via `escapeHTML` and rely on the existing chunking.

### D5. Onboarding

- `/start` without a code: a short welcome ("I'm the Villum bot…") with the three linking steps, "once linked you can…" bullets, and the navigation keyboard.
- `linkingHelpText`: friendlier wording, keeps the three steps, and points at Settings → Telegram as the code source.
- `/start <code>` success: "Linked! … Try /claim to pick your character or /create to make one; /help lists everything." with the navigation keyboard.
- Group add: subscribe to `my_chat_member`; when `NewChatMember.Type` is member/administrator/creator and the previous state was left/kicked/empty (or absent) and the chat is a group/supergroup, send a one-time welcome explaining that a DM can connect the group from Campaign → Telegram and that `/items`, `/quests`, `/visits`, and `/stats` light up once connected. No welcome on ordinary member updates.

### D6. Wiring and files

- New `telegram/group.go`: context resolution, queries, renderers, statistics.
- `telegram/commands.go`: register `/items`, `/quests`, `/visits`; make `/stats` context-aware; help text updates.
- `telegram/bot.go`: handle `Update.MyChatMember`.
- `telegram/client.go`: add `AllowedUpdateMyChatMember` to `WithAllowedUpdates`.
- `telegram/handlers.go`: onboarding texts.

### D7. Testing

Unit tests cover bound-chat resolution (hit, disabled row, missing row), private fallback, renderers with seeded data, empty states, stats math, and the group-add welcome (added vs. updated member). E2E extends `tests/telegram.spec.ts`: create a campaign, bind its chat, seed a party item, a quest, and a location visit through the HTTP API, then post `/items`, `/quests`, `/visits`, `/stats` webhook updates from the group chat and assert the recorded replies.

## Risks / Trade-offs

- [Group members see campaign data without linking] → Read-only, requires a DM to bind the chat, mirrors recap broadcast exposure; no secrets or member PII beyond character names.
- [my_chat_member noise] → Welcome only on transitions into member/administrator/creator from left/kicked/empty; tested.
- [Campaign statistics semantics differ from the web UI] → Mirror the UI exactly (members-owned NPCs/locations, average level one decimal) and document in the spec.
- [Long group messages] → Caps plus the existing 4096-chunk splitter; tested.
- [Unbound group commands confuse users] → Every unbound reply includes the binding path, and the group-add welcome pre-explains it.

## Migration Plan

No data migration. Deploy the binary; the bot starts accepting `my_chat_member` updates. Rollback is a binary revert; no state changes.

## Open Questions

- None blocking. If groups later want mutating commands (e.g. adding party items from chat), that would be a separate change with per-user authorization.
