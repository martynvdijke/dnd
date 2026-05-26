## ADDED Requirements

### Requirement: Items in one-shot adventures
One-shot adventures SHALL have their own items, independent of character inventory. Items SHALL have columns: `id`, `adventure_id`, `name`, `description`, `quantity`, `weight`, `value_gp`, `is_magical`, `attunement`, `notes`, `created_at`.

#### Scenario: Create item in one-shot
- **WHEN** a DM sends POST to `/api/oneshot-adventures/:id/items` with item data
- **THEN** the item SHALL be created and linked to the adventure

#### Scenario: List one-shot items
- **WHEN** a DM sends GET to `/api/oneshot-adventures/:id/items`
- **THEN** all items for that adventure SHALL be returned

#### Scenario: Update one-shot item
- **WHEN** a DM sends PUT to `/api/oneshot-items/:iid` with updated fields
- **THEN** the item SHALL be updated

#### Scenario: Delete one-shot item
- **WHEN** a DM sends DELETE to `/api/oneshot-items/:iid`
- **THEN** the item SHALL be deleted

### Requirement: File uploads on one-shot items
Items in one-shots SHALL support file uploads (images for item appearance, handouts). Uses `owner_type="item"` pattern.

#### Scenario: Upload file to one-shot item
- **WHEN** a user sends POST to `/api/oneshot-items/:iid/upload` with a file
- **THEN** the file SHALL be saved and linked to the item

#### Scenario: List item uploads
- **WHEN** a user sends GET to `/api/oneshot-items/:iid/uploads`
- **THEN** all uploads for the item SHALL be returned
