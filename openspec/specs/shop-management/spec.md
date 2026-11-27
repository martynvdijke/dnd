# shop-management Specification

## Purpose
Campaign-level shop CRUD with location binding, inventory management via compendium lookup, and markup pricing for DMs.

## Requirements

### Requirement: DM can create shops within their campaign
A Dungeon Master SHALL be able to create a new shop associated with a campaign they belong to. The shop SHALL have a name, optional description, optional location (from the locations table), sell markup percentage, and buy markup percentage.

#### Scenario: DM creates a shop with all fields
- **WHEN** a DM sends a POST request to `/api/campaigns/:id/shops` with valid `name`, `description`, `location_id`, `markup_percent`, and `markup_buy_percent`
- **THEN** the system creates a new shop record and returns it with HTTP 201

#### Scenario: DM creates a minimal shop
- **WHEN** a DM sends a POST request with only `name` (required)
- **THEN** the system creates a shop with default values for optional fields

#### Scenario: Non-DM cannot create shops
- **WHEN** a player who is not a DM in the campaign sends a POST to the shop creation endpoint
- **THEN** the system returns HTTP 403 Forbidden

### Requirement: DM can update shop details
A Dungeon Master SHALL be able to update a shop's name, description, location, and markup percentages.

#### Scenario: DM updates shop markup
- **WHEN** a DM sends a PUT request to `/api/shops/:id` with updated `markup_percent`
- **THEN** the system updates the shop record and returns HTTP 200

#### Scenario: DM updates shop location
- **WHEN** a DM sends a PUT request to `/api/shops/:id` with a valid `location_id`
- **THEN** the shop is associated with that location

#### Scenario: DM clears shop location
- **WHEN** a DM sends a PUT request to `/api/shops/:id` with `location_id: null`
- **THEN** the shop's location association is removed

### Requirement: DM can delete a shop
A Dungeon Master SHALL be able to delete a shop and all its associated items and transactions.

#### Scenario: DM deletes a shop
- **WHEN** a DM sends a DELETE request to `/api/shops/:id`
- **THEN** the shop is deleted along with its items and transactions (cascade)

#### Scenario: Non-DM cannot delete shops
- **WHEN** a player who is not a campaign DM sends a DELETE request
- **THEN** the system returns HTTP 403

### Requirement: DM can add items to a shop's inventory
A Dungeon Master SHALL be able to add items to a shop's inventory, optionally sourced from the compendium equipment list.

#### Scenario: DM adds an item with compendium data
- **WHEN** a DM selects an item from the compendium equipment list and POSTs to `/api/shops/:id/items` with `compendium_item_id` and adjusted `price_gp`
- **THEN** the item is created with pre-filled fields from the compendium entry and the DM's price override

#### Scenario: DM adds a custom item
- **WHEN** a DM POSTs to `/api/shops/:id/items` with all item fields manually specified
- **THEN** the custom item is created in the shop's inventory

### Requirement: DM can update shop items
A Dungeon Master SHALL be able to update a shop item's name, category, price, quantity, description, magical status, attunement, and notes.

#### Scenario: DM updates item price and quantity
- **WHEN** a DM sends a PUT request to `/api/shop-items/:id` with updated `price_gp` and `quantity_available`
- **THEN** the system updates the item record and returns HTTP 200

### Requirement: DM can delete shop items
A Dungeon Master SHALL be able to remove an item from a shop's inventory.

#### Scenario: DM deletes a shop item
- **WHEN** a DM sends a DELETE request to `/api/shop-items/:id`
- **THEN** the item is removed from the shop's inventory

### Requirement: Admin retains global shop management
An admin user SHALL be able to create, update, and delete any shop regardless of campaign association.

#### Scenario: Admin creates a global shop
- **WHEN** an admin sends a POST to `/api/admin/shops` with shop data
- **THEN** the shop is created without campaign scope

#### Scenario: Admin deletes a shop from any campaign
- **WHEN** an admin sends a DELETE to `/api/admin/shops/:id`
- **THEN** the shop is deleted regardless of campaign

### Requirement: Shop location is shown in listings
When shops are listed, associated location data SHALL be included if a location is set.

#### Scenario: Shop list includes location data
- **WHEN** a user fetches `GET /api/shops`
- **THEN** each shop includes `location` object with `id`, `name`, `type` if a location is associated
