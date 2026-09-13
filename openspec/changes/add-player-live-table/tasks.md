## 1. WebSocket protocol

- [ ] 1.1 Add event-type constants (`dice_roll`, `combat_update`, `knowledge_reveal`) and small payload builder helpers in `handlers/ws.go`.
- [ ] 1.2 Extend the `ts/init.ts:33` `ws.onmessage` switch with handlers for the three new types (refetch or render as appropriate).
- [ ] 1.3 Confirm `BroadcastToCampaignMembers` is used for every new event and never `BroadcastToAdmins`.

## 2. Backend broadcast hooks

- [ ] 2.1 In `handlers/dice.go` `HandleRoll`, resolve the rolling character's campaign and broadcast `dice_roll` after the roll is persisted; skip when no campaign.
- [ ] 2.2 In `handlers/combat.go`, broadcast `combat_update` from create, update, delete, and `next-turn`.
- [ ] 2.3 In `handlers/knowledge.go`, broadcast `knowledge_reveal` when `UpdateKnowledge` sets `shared` true and from `BulkRevealKnowledge`; never for unshared entries.

## 3. Player table view

- [ ] 3.1 Compose a Table view from existing widgets (character summary, dice tray, read-only initiative, revealed handouts) toggled via `ts/session-mode.ts`.
- [ ] 3.2 Hide DM-only controls in table view.
- [ ] 3.3 On WS (re)connect while in table view, refetch initiative and revealed knowledge once.

## 4. Tests & verification

- [ ] 4.1 Go tests: roll broadcasts to members, not non-members; no broadcast without a campaign.
- [ ] 4.2 Go tests: combat mutations broadcast; knowledge broadcast only on shared transition.
- [ ] 4.3 e2e: two users in a campaign; player rolls → DM screen shows the roll; DM reveals a handout → player view shows it; non-member receives nothing.
- [ ] 4.4 `go vet ./...`, `task test`, `npm run typecheck`, `npm run build:vite`, `task lint:e2e`.
