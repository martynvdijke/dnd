# compendium-item-linking Specification

## Purpose
Link character inventory items from compendium equipment entries.

## Requirements

### Requirement: Compendium equipment reference on inventory items

The `InventoryItem` model SHALL have a new nullable field `compendium_equipment_id` referencing `compendium_equipment.id`. When set, the item's name, weight, category, and description SHALL be pre-filled from the compendium entry on creation. The item data SHALL be stored as a copy (not dynamically resolved).

#### Scenario: Add item from compendium equipment

- **WHEN** the user clicks "Link from Compendium" in the inventory tab
- **THEN** a compendium equipment picker modal SHALL open, showing searchable equipment list

#### Scenario: New inventory item copies compendium data

- **WHEN** the user selects a compendium equipment entry and confirms
- **THEN** a new inventory item SHALL be created with `compendium_equipment_id` set, and name, weight, category, description pre-filled from the compendium entry

#### Scenario: Manual item creation unchanged

- **WHEN** the user clicks "Add Item" (not from compendium)
- **THEN** the existing manual item creation flow SHALL work as before, with `compendium_equipment_id` set to null

### Requirement: Compendium link indicator in inventory list

Inventory items that are linked to a compendium entry SHALL display a small compendium icon or badge next to their name in the inventory list. Hovering over the badge SHALL show the compendium source name.

#### Scenario: Linked item badge visible

- **WHEN** rendering an inventory item with `compendium_equipment_id` set
- **THEN** a compendium badge/icon SHALL be shown next to the item name

### Requirement: Unlink compendium reference

Linked inventory items SHALL have an "Unlink" option that removes the `compendium_equipment_id` reference while keeping the item data intact.

#### Scenario: Unlink keeps item data

- **WHEN** the user clicks "Unlink" on a compendium-linked item
- **THEN** the `compendium_equipment_id` SHALL be set to null, but the item SHALL remain with all its data unchanged
