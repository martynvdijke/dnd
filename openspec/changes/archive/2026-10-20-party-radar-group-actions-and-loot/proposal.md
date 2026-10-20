## Why

During exploration, social encounters, and post-combat downtime:
1. **No DM Party Passives Radar or Group Actions (`ts/party.ts`)**: Party View only displays HP, AC, and a basic character card. DMs constantly need to check **Passive Perception, Passive Insight, and Passive Investigation** across the party without opening each character sheet, and need 1-click **Group Skill Check / Saving Throw** (e.g., "Everyone roll Stealth") and **Party Rest** (Short/Long rest for the whole party at once).
2. **Party Inventory Items Cannot Be Claimed (`handlers/party_inventory.go`, `ts/app/campaign-dashboard.ts`)**: `party_items` is a flat list (`name, quantity, notes`) with a Delete button, forcing players to manually delete the party item and manually re-type it into their character sheet's inventory.
3. **Ephemeral DM Tools Generators (`ts/app/dm-tools.ts`)**: The DM Tools modal generates NPCs, Taverns, Adventure Hooks, and Treasure Hoards, but renders them as static HTML in `#dmToolsResult` with no **"Save to Campaign"** button to persist them as an NPC, Location, Quest, or Party Inventory item.

## What Changes

- Enhance `ts/party.ts` (`#partyView`) with:
  - A **Party Passives Radar** banner showing each party member's Passive Perception, Passive Insight, and Passive Investigation (`10 + ability mod + skill proficiency`).
  - A **Group Roll** bar (`data-testid="party-group-roll-select"`, `data-testid="party-group-roll-btn"`) that rolls a chosen skill check, ability check, or saving throw for every party member and displays the side-by-side results.
  - A **Party Rest** control (`data-testid="party-short-rest-btn"`, `data-testid="party-long-rest-btn"`) that triggers Short or Long Rest across all party characters.
- Add `POST /api/campaigns/:id/party-items/:itemId/claim` in `handlers/party_inventory.go` (and wire it in `ts/app/campaign-dashboard.ts`) so a campaign member can atomically claim a quantity of a Party Inventory item directly into a character's `inventory_items`.
- Add **Save to Campaign** action buttons (`data-testid="dm-tools-save-btn"`) to `ts/app/dm-tools.ts` so generated NPCs/Names can be saved as campaign NPCs (`/api/npcs`), Taverns as Locations (`/api/locations`), Adventure Hooks as Quests (`/api/quests`), and Treasure Hoards as Party Inventory items (`/api/campaigns/:id/party-items`).

## Capabilities

### New Capabilities

- `party-group-actions`: Party Passive Senses Radar (Passive Perception, Insight, Investigation), 1-click Group Skill/Save Rolls, Party Short/Long Rest, and atomic Party Loot claiming into a character's inventory.

### Modified Capabilities

- `dm-tooling`: Add 1-click persistence ("Save to Campaign") for generated NPCs, taverns, adventure hooks, and treasure hoards in the DM Tools modal.

## Impact

- **Backend (`handlers/party_inventory.go`, `handlers/routes_campaigns.go`)**:
  - New `POST /api/campaigns/:id/party-items/:itemId/claim` endpoint that transfers item quantity from `party_items` into `inventory_items` for the target character within a transaction.
- **Frontend (`ts/party.ts`, `ts/app/campaign-dashboard.ts`, `ts/app/dm-tools.ts`, `static/app.html`)**:
  - Passive radar, Group Roll controls, and Party Rest buttons in `ts/party.ts`.
  - Character selector + Claim button on each row in `renderPartyInventory()` (`ts/app/campaign-dashboard.ts`).
  - "Save to Campaign" buttons on generator results in `ts/app/dm-tools.ts`.
- **Tests (`handlers/party_inventory_test.go`, `ts/**/*.test.ts`, `tests/`)**:
  - Handler tests for atomic party item claiming and authorization checks.
  - Vitest unit tests and Playwright E2E coverage for all new `data-testid` attributes.
