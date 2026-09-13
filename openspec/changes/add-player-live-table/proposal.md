## Why
The DM screen and player phones are disconnected: players roll on their own device and announce the result out loud, the DM copies HP/initiative by hand, and secrets are revealed by reading aloud. Villum already has the primitives — a WebSocket hub, a dice engine, campaign-scoped combat, and shared knowledge — but no live table channel connecting them. This change wires a player's phone roll to the DM screen and a DM reveal to the player phones.

## What Changes
- Extend the WebSocket protocol (envelope `{type,payload}`) with campaign-scoped server→client events: `dice_roll`, `combat_update`, `knowledge_reveal`.
- After `POST /api/roll` for a campaign-owned character, broadcast the result to that campaign's connected members.
- Broadcast combat mutations (create/update/delete entry, next turn) to campaign members.
- Broadcast knowledge entries when `shared` flips true or bulk reveal runs — never DM-only entries.
- Add a read-mostly **Table view** for player devices (reusing `session-mode`) showing the current character, a dice tray, live initiative, and revealed handouts.
- Hide DM-only controls while in Table view.

## Capabilities
### New Capabilities
- `player-live-table`: campaign-scoped live events (dice, combat, handouts) and the player table view.

### Modified Capabilities
<!-- None: existing knowledge visibility, dice and combat APIs keep their current requirements; live events are additive. -->

## Impact
Backend: `handlers/ws.go` (broadcast hooks + event constants), `handlers/dice.go`, `handlers/combat.go`, `handlers/knowledge.go`, `router.go` (WS group unchanged). Frontend: `ts/init.ts` (message switch), `ts/dice.ts`, `ts/combat-tracker.ts`, `ts/session-mode.ts`, `ts/knowledge.ts`. No new dependencies, no DB migration.
