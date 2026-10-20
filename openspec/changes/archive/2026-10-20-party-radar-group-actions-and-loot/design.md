## Context

`ts/party.ts` renders `#partyGrid` from `/api/characters` (filtered to the active campaign when selected) and computes total party HP and average AC. However, it does not compute passive senses, cannot execute group rolls, and cannot rest the party together.

Meanwhile, `handlers/party_inventory.go` exposes `GET/POST /api/campaigns/:id/party-items` and `PUT/DELETE /api/campaigns/:id/party-items/:itemId`, rendered in `ts/app/campaign-dashboard.ts` (`renderPartyInventory`). Moving an item from the party stash to a character requires deleting it from Party Inventory and manually creating an inventory item on the character sheet.

Finally, `ts/app/dm-tools.ts` renders procedural generator outputs into `#dmToolsResult` without offering a way to save generated content to the campaign database.

## Goals / Non-Goals

**Goals:**
- Display Passive Perception, Passive Insight, and Passive Investigation for every party character in Party View.
- Allow 1-click Group Rolls (any ability check, skill check, or saving throw) and Party-wide Short/Long Rest in Party View.
- Provide an atomic `POST /api/campaigns/:id/party-items/:itemId/claim` endpoint and UI control to transfer party loot into a character's inventory.
- Add "Save to Campaign" buttons to DM Tools generators (Tavern → Location, Adventure Hook → Quest, NPC/Name → NPC, Treasure Hoard → Party Inventory).

**Non-Goals:**
- Changing the database schema for `party_items` or `inventory_items`.
- Replacing the dedicated Shops or Crafting views.

## Decisions

1. **Atomic Party Item Claim Transaction (`POST /api/campaigns/:id/party-items/:itemId/claim`)**:
   - *Decision*: Implement `HandleClaimPartyItem` in `handlers/party_inventory.go`. It verifies campaign access (`requireCampaignAccess`) and that `character_id` belongs to the campaign (or caller), reads the `PartyItem`, validates `1 <= claimQty <= item.Quantity`, creates an `InventoryItem` on `character_id` (with `name`, `quantity: claimQty`, `description: item.Notes`, `category: "gear"`), and either decrements `item.Quantity` or deletes the `PartyItem` when `Quantity` reaches `0`.
   - *Rationale*: Guarantees no item loss or duplication if a request fails midway.

2. **Passive Senses Derivation in `ts/party.ts`**:
   - *Decision*: Compute Passive Perception (`10 + WIS mod`), Passive Insight (`10 + WIS mod`), and Passive Investigation (`10 + INT mod`), adding `proficiency_bonus` when the character's `skills` map/array marks `Perception`, `Insight`, or `Investigation` as proficient.
   - *Rationale*: Purely derived from existing `/api/characters` payload with zero extra HTTP round-trips.

3. **DM Tools "Save to Campaign" Actions (`ts/app/dm-tools.ts`)**:
   - *Decision*: Store the last generated payload in module state in `ts/app/dm-tools.ts` and render a contextual `data-testid="dm-tools-save-btn"` button inside `#dmToolsResult` that posts to `/api/locations` (for Tavern), `/api/quests` (for Adventure Hook), `/api/npcs` (for Name/NPC), or `/api/campaigns/:id/party-items` (for Treasure Hoard).
   - *Rationale*: Reuses existing CRUD endpoints with zero new backend routes required.

## Risks / Trade-offs

- **[Risk] Claiming an item when no character is selected** → *Mitigation*: Render a compact character dropdown inside the Party Inventory panel (pre-selecting `activeCharacterId` when available) so the user can always pick which party character receives the item.
