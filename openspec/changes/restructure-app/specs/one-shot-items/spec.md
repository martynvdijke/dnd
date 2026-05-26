## ADDED Requirements

### Requirement: Items within one-shots
The system SHALL support items that belong to a one-shot adventure. Items have: name, description, category, quantity, weight, price_gp, is_magical, attunement, notes.

#### Scenario: Create item in one-shot
- **WHEN** a DM makes a POST to `/oneshot-adventures/:id/items` with item details
- **THEN** the system creates the item and returns 201

#### Scenario: List items for one-shot
- **WHEN** a DM makes a GET to `/oneshot-adventures/:id/items`
- **THEN** the system returns items belonging to that one-shot

#### Scenario: Update item
- **WHEN** a DM makes a PUT to `/oneshot-items/:id` with updated fields
- **THEN** the system updates the item and returns 200

#### Scenario: Delete item
- **WHEN** a DM makes a DELETE to `/oneshot-items/:id`
- **THEN** the system deletes the item and returns 200

### Requirement: Items support file uploads
Items SHALL support file uploads via polymorphic owner_type="item" / owner_id=<item_id>.

#### Scenario: Upload file to item
- **WHEN** a DM uploads a file with owner_type="item" and owner_id=<item_id>
- **THEN** the system stores the upload and links it to the item

#### Scenario: List uploads for item
- **WHEN** a DM makes a GET to `/oneshot-items/:id/uploads`
- **THEN** the system returns uploads for that item

### Requirement: NPC to Item links
The system SHALL support linking NPCs to items within a one-shot context. Links have a relationship_type field ("owns", "knows", "wields", etc).

#### Scenario: Link NPC to item
- **WHEN** a DM makes a POST to `/oneshot-adventures/:id/npc-item-links` with npc_id, item_id, and relationship_type
- **THEN** the system creates the link and returns 201

#### Scenario: List NPCs for item
- **WHEN** a DM makes a GET to `/oneshot-items/:id/npcs`
- **THEN** the system returns NPCs linked to that item

#### Scenario: List items for NPC
- **WHEN** a DM makes a GET to `/oneshot-adventures/:id/npcs/:nid/items`
- **THEN** the system returns items linked to that NPC

#### Scenario: Remove NPC-item link
- **WHEN** a DM makes a DELETE to `/npc-item-links/:id`
- **THEN** the system removes the link and returns 200
