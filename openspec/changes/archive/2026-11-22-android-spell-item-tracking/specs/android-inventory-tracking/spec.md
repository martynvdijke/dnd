# android-inventory-tracking Specification

## Purpose

Inventory and currency tracking in the Android app renders and edits real server state; editing can never silently destroy existing data.

## ADDED Requirements

### Requirement: Inventory and currency render from server data

The app SHALL parse the `inventory` and `currency` sub-resources returned by `GET /api/characters/:id` and SHALL render the item list (name, quantity, equipped state) and the coin amounts from the server response.

#### Scenario: Sheet shows carried items and coins

- WHEN a character carrying items and gold is opened
- THEN the INVENTORY tab lists the server's items with quantity and equipped state and shows the server's coin amounts

### Requirement: Inventory item management

The app SHALL support adding, editing, equipping/unequipping, and deleting inventory items through the existing character inventory endpoints, replacing the placeholder action.

#### Scenario: Add an item from the sheet

- WHEN the player adds an item with a quantity from the INVENTORY tab
- THEN the item is persisted via the inventory endpoint and appears in the list

#### Scenario: Delete an item from the sheet

- WHEN the player deletes an item from the INVENTORY tab
- THEN the item is removed via the inventory endpoint and the list reflects the server state

### Requirement: Currency edit preserves server values

The currency edit dialog SHALL pre-fill all coin fields from the server values, and a save SHALL send the complete currency map so that coins the player did not touch retain their server values.

#### Scenario: Editing one coin does not zero the others

- WHEN the player edits one coin type and saves while other coins hold non-zero server values
- THEN the app persists the edited coin and the untouched coins keep their server values
