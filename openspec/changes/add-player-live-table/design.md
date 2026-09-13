## Context

`handlers/ws.go` already runs an authenticated hub: `WSMessage{Type,Payload}` (ws.go:29), clients keyed by user id (`Hub.clients map[int64][]*WSClient`), with `BroadcastToUser`, `BroadcastToAdmins`, and `BroadcastToCampaignMembers(campaignID,msg)` (ws.go:72-98) that resolves membership from `campaign_members` + `campaigns.user_id`. `HandleWebSocket` (ws.go:149) authenticates the `session` cookie; `readPump` currently discards inbound frames. The frontend `connectWS()` (`ts/init.ts:29`) reconnects every 5s and switches on `character_update` / `party_update` (init.ts:33).

Rolls already have a complete REST path (`POST /api/roll` → `HandleRoll`, handlers/dice.go:45, persisted to `dice_rolls`), combat has campaign-scoped CRUD and `next-turn` (handlers/combat.go), and knowledge has `shared` + `BulkRevealKnowledge` (knowledge.go:365). `ts/session-mode.ts` is a local presentation toggle only.

## Goals / Non-Goals

**Goals**
- A roll made on a player phone appears on every connected campaign screen in real time.
- Initiative/turn and HP changes made by the DM appear on player screens without polling.
- A handout the DM shares appears on player screens; unshared entries never leave the server.
- A player-facing Table view that is simple and safe (no DM controls).

**Non-Goals**
- Anonymous play over a share token — viewers must be authenticated campaign members. Share-token WS auth is a possible follow-up, not this change.
- Voice/video, map token movement, or turn timers.
- Replacing REST mutations with WebSocket commands; the server stays authoritative over HTTP.

## Decisions

### 1. Rolls are submitted over REST, fanned out over WebSocket
Players keep using `POST /api/roll`; after the handler persists the roll, the backend calls `BroadcastToCampaignMembers`. The roll's campaign is resolved from the character row (a roll with no campaign character is not broadcast).

**Why:** `/api/roll` already carries session auth, CSRF and API-token checks plus roll history. Adding inbound client→server WS commands would duplicate auth/CSRF in the WS path and open a second mutation surface. **Alternative:** WS request/response for rolls — rejected (duplicate auth, harder to log).

### 2. New event types on the existing envelope
Add `dice_roll`, `combat_update`, `knowledge_reveal` to the `{type,payload}` protocol and extend the `ts/init.ts:33` switch. Payloads stay small:
- `dice_roll`: `{user_id, username, character_id, expression, total, text}`
- `combat_update`: `{campaign_id}` — clients refetch `GET /api/combat`, so no stale HP/order is ever pushed
- `knowledge_reveal`: the shared knowledge entry (`{id,campaign_id,title,content,source,status}`)

**Why:** Reusing the envelope means no transport change. Sending a *signal* for combat (not the state) keeps one source of truth and avoids partial/ordered pushes.

### 3. Broadcast hooks at the mutation sites
`HandleRoll` (dice.go), combat create/update/delete/next-turn (combat.go), and `UpdateKnowledge` (only when `shared` transitions to true) + `BulkRevealKnowledge` (knowledge.go) each emit one broadcast. No background scanner.

**Why:** Least code, event always matches the committed state. **Trade-off:** every broadcast must resolve campaign membership (`BroadcastToCampaignMembers` does a query); acceptable at table scale.

### 4. Reveal is the DM's single secret-leak boundary
`knowledge_reveal` is emitted only when an entry is already `shared = true`. Unshared entries are never broadcast, matching the existing `knowledgeVisible` rule.

### 5. Table view reuses session mode, not a new SPA view
`session-mode.ts` already toggles a body class. Table view composes existing widgets (character summary, `ts/dice.ts`, `ts/combat-tracker.ts` read-only, revealed knowledge list) behind that toggle, with DM-only controls hidden. A brand-new SPA view is unnecessary.

## Risks / Trade-offs

- **Secret leaks to players** → broadcast only on the shared transition, only to `BroadcastToCampaignMembers`; dedicated tests for member vs non-member and for `shared=false`.
- **Reconnect misses events** → on WS (re)connect the client refetches the active view once (`combat_update` clients already refetch; table view re-pulls knowledge). No event replay buffer in MVP.
- **Broadcast storms during AoE rolls** → one broadcast per roll; fine at table scale. Coalescing can come later.
- **DM-only knowledge in the table view** → the table view requests the same member-filtered knowledge endpoint, so the split is enforced server-side.

## Migration Plan
Additive only, no schema change. Ship backend broadcasts + frontend listeners behind the new event types; rollback reverts the broadcast calls and the client switch arms. Old clients simply ignore unknown types.

## Open Questions
- Should the DM screen also expose an "auto-roll NPC saves" hook for later combat automation? (Out of scope; see `add-combat-automation`.)
- Do we mint a read-only share token for guest spectators later? (Deferred; non-goal for now.)
