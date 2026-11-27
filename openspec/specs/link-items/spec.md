# link-items Specification

## Purpose
Search and link compendium equipment from character sheet inventory.

## Requirements

### Requirement: Character sheet item linking from compendium
The character sheet inventory section SHALL allow the user to search the compendium for equipment and link it to the character instead of creating an inline item.

#### Scenario: Search for compendium equipment
- **WHEN** user clicks "Add Item from Compendium" button on the character sheet inventory section
- **THEN** a search modal/picker appears with a text input and search capabilities
- **WHEN** user types a search term
- **THEN** the system queries `GET /api/compendium/search?type=equipment&q=<term>` and displays matching equipment as a list with name, category, cost, weight
- **WHEN** user clicks an equipment result
- **THEN** the item is linked to the character (stored as `compendium_equipment_id` reference) with default quantity 1 and appears in the inventory list

#### Scenario: Render linked item in inventory
- **WHEN** a character has a linked compendium equipment item
- **THEN** the item row SHALL display a "📖 Compendium" badge/indicator
- **WHEN** user clicks the linked item row
- **THEN** the item card is fetched from `GET /compendium/card/equipment/:id` and displayed, showing full compendium data including description, cost, weight, properties

#### Scenario: Link with override quantity
- **WHEN** user links a compendium equipment item
- **THEN** the system SHALL prompt for quantity (default 1) before creating the link
- **THEN** the quantity SHALL be stored on the `inventory_items` row alongside the `compendium_equipment_id` reference

### Requirement: Backend API for compendium item linking
The backend SHALL provide endpoints to link/unlink compendium equipment from a character.

#### Scenario: Link item endpoint
- **WHEN** `POST /api/characters/:id/inventory/link` is called with `{"compendium_equipment_id": 42, "quantity": 1}`
- **THEN** a new `inventory_item` row is created with `character_id`, `compendium_equipment_id`, `quantity`, and `name` copied from the compendium entry
- **THEN** the response SHALL include the created inventory item row

#### Scenario: Unlink item endpoint
- **WHEN** `DELETE /api/characters/:id/inventory/:itemId/link` is called
- **THEN** `compendium_equipment_id` is set to NULL on that inventory item row
- **WHEN** the inventory item has no other inline data beyond quantity/name, it SHALL be deleted
