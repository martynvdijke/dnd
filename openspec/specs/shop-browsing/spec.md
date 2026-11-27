# shop-browsing Specification

## Purpose
Player-facing shop viewer with item listing, category filtering, buy/sell transactions with markup-visible pricing.

## Requirements

### Requirement: Player can browse shops in their campaign
A player SHALL be able to view all shops in their active campaign. The shop list SHALL show name, description, location name (if set), and sell/buy markup percentages.

#### Scenario: Player views campaign shops
- **WHEN** a player fetches `GET /api/shops?campaign_id=X`
- **THEN** the system returns all shops in that campaign with full details

#### Scenario: Player sees empty state
- **WHEN** a player fetches shops in a campaign with no shops
- **THEN** the system returns an empty array

### Requirement: Player can browse a shop's inventory
A player SHALL be able to view all items available in a specific shop with their prices (including markup applied), quantities, and descriptions.

#### Scenario: Player views shop items
- **WHEN** a player fetches `GET /api/shops/:id/items`
- **THEN** the system returns all items in that shop with prices displayed as the base price with markup

#### Scenario: Item shows markup-adjusted price
- **WHEN** a shop has `markup_percent: 150` and an item has `price_gp: 10`
- **THEN** the displayed price SHALL be 15 GP (10 × 1.5) with a note showing the base + markup

### Requirement: Player can filter shop items by category
A player SHALL be able to filter a shop's inventory by item category (e.g., weapon, armor, potion, gear, magic-item).

#### Scenario: Player filters by category
- **WHEN** a player selects a category filter in the shop UI
- **THEN** only items matching that category are displayed

#### Scenario: No items match filter
- **WHEN** a player applies a category filter with no matching items
- **THEN** the UI shows "No items in this category"

### Requirement: Player can buy items from a shop
A player SHALL be able to purchase items from a shop, deducting gold from their character and adding the item to their inventory.

#### Scenario: Player buys an item with sufficient gold
- **WHEN** a player POSTs to `/api/shops/:id/buy` with `item_id`, `character_id`, and `quantity`
- **THEN** the character's gold is reduced by the markup-adjusted price, the item is added to their inventory, stock is decremented, and a transaction is recorded

#### Scenario: Player buys an item with insufficient gold
- **WHEN** a player tries to buy an item they cannot afford
- **THEN** the system returns an error "not enough gold" and no transaction occurs

#### Scenario: Player buys an out-of-stock item
- **WHEN** a player tries to buy an item with `quantity_available: 0`
- **THEN** the system returns an error "item out of stock"

#### Scenario: Player buys from unlimited stock
- **WHEN** a player buys an item with `quantity_available: -1` (unlimited)
- **THEN** the purchase succeeds and stock remains unchanged

### Requirement: Player can see buy confirmation
After a successful purchase, the player SHALL see a confirmation with the item name, quantity, and total price paid.

#### Scenario: Buy confirmation displayed
- **WHEN** a purchase succeeds
- **THEN** the UI shows a toast/alert: "Bought 2x Longsword for 30 GP"

### Requirement: Player can sell items to a shop
A player SHALL be able to sell items from their inventory to a shop, receiving gold at the shop's buy markup rate.

#### Scenario: Player sells an item
- **WHEN** a player POSTs to `/api/shops/:id/sell` with `character_id`, `item_name`, and `quantity`
- **THEN** the character gains gold at the shop's buy markup rate, the item is removed from inventory, and a transaction is recorded

#### Scenario: Player sells an item they don't have
- **WHEN** a player tries to sell an item not in their inventory
- **THEN** the system returns an error

### Requirement: Player can view shop transaction history
A player SHALL be able to view recent buy/sell transactions for their character.

#### Scenario: Player views transactions
- **WHEN** a player fetches `GET /api/shop-transactions?character_id=X`
- **THEN** the system returns the last 20 transactions for that character with item name, quantity, price, type (buy/sell), and timestamp

### Requirement: Shop view shows location on map
If a shop has an associated location with coordinates, the shop detail view SHALL show a link or reference to the location.

#### Scenario: Shop with coordinates shows location
- **WHEN** a player views a shop with a location that has latitude/longitude
- **THEN** the location name is displayed and clickable to view on the map

#### Scenario: Shop without coordinates
- **WHEN** a player views a shop without location or coordinates
- **THEN** only the location name text is shown (if any)
