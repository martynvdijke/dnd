## 1. Backend Party Loot Claim Endpoint

- [x] 1.1 Implement `HandleClaimPartyItem` (`POST /api/campaigns/:id/party-items/:itemId/claim`) in `handlers/party_inventory.go` and register the route in `handlers/routes_campaigns.go`.
- [x] 1.2 Add Go unit tests in `handlers/party_inventory_test.go` covering full claim, partial quantity claim, and unauthorized access rejection.

## 2. Frontend Party Passives Radar, Group Rolls, Loot Claim & Generator Persistence

- [x] 2.1 Add the Party Passives Radar (Passive Perception, Passive Insight, Passive Investigation), Group Roll bar (`data-testid="party-group-roll-select"`, `data-testid="party-group-roll-btn"`), and Party Rest buttons (`data-testid="party-short-rest-btn"`, `data-testid="party-long-rest-btn"`) to `ts/party.ts` and `static/app.html`.
- [x] 2.2 Add a character selector and Claim button (`data-testid="party-item-claim-btn"`) to `renderPartyInventory()` in `ts/app/campaign-dashboard.ts`.
- [x] 2.3 Add `data-testid="dm-tools-save-btn"` ("Save to Campaign") in `ts/app/dm-tools.ts` to persist generated Taverns (`/api/locations`), Adventure Hooks (`/api/quests`), Names (`/api/npcs`), and Treasure Hoards (`/api/campaigns/:id/party-items`).

## 3. Verification & Quality Gates

- [x] 3.1 Add Vitest unit tests for party passive calculations, group roll execution, party item claim UI, and DM Tools save actions.
- [x] 3.2 Add Playwright E2E tests in `tests/` covering all new `data-testid` selectors (`party-group-roll-select`, `party-group-roll-btn`, `party-short-rest-btn`, `party-long-rest-btn`, `party-item-claim-btn`, `dm-tools-save-btn`).
- [ ] 3.3 Run `openspec validate party-radar-group-actions-and-loot --strict` and `task ci`.
