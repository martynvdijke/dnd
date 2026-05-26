## ADDED Requirements

### Requirement: Link NPCs to one-shot items
A DM SHALL be able to link NPCs to items within a one-shot adventure. A `one_shot_npc_items` table SHALL store links: `id`, `npc_id`, `item_id`, `relationship` (e.g., "owner", "carrier", "knows_about"), `notes`.

#### Scenario: Link NPC to item
- **WHEN** a DM sends POST to `/api/oneshot-items/:iid/npcs` with `npc_id` and `relationship`
- **THEN** the NPC SHALL be linked to the item

#### Scenario: List NPCs linked to item
- **WHEN** a DM sends GET to `/api/oneshot-items/:iid/npcs`
- **THEN** all NPCs linked to the item SHALL be returned with their relationship

#### Scenario: Unlink NPC from item
- **WHEN** a DM sends DELETE to `/api/oneshot-npc-items/:id`
- **THEN** the NPC-item link SHALL be removed

#### Scenario: List items linked to NPC
- **WHEN** a user sends GET to `/api/npcs/:nid/items` (within one-shot context)
- **THEN** all items linked to the NPC SHALL be returned
