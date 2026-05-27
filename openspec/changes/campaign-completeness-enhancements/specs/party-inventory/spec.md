## ADDED Requirements

### Requirement: Party inventory and treasury view
The system SHALL provide a party-level inventory view for sharing loot, tracking party funds, and splitting coins among party members.

#### Scenario: Party view shows inventory tab
- **WHEN** the party view is rendered
- **THEN** the system SHALL show a "Party Inventory" section listing items shared to the party

#### Scenario: Add item to party inventory
- **WHEN** a user with DM or party member role clicks "Add to Party Loot"
- **THEN** the system SHALL show a form (item name, quantity, notes) AND on submit add the item to the party inventory

#### Scenario: Assign party item to character
- **WHEN** a DM or party member clicks "Assign" on a party loot item
- **THEN** the system SHALL show a character picker of party members, AND on confirm transfer the item to that character's personal inventory (creating an InventoryItem) and remove it from party inventory

#### Scenario: Party treasury shows total coins
- **WHEN** the party view is rendered
- **THEN** the system SHALL show a "Party Treasury" section with total coins across all party members' currency

#### Scenario: Split coins evenly among party
- **WHEN** a user clicks "Split Coins" in the party treasury
- **THEN** the system SHALL divide the selected coin type equally among all party members AND update each character's currency record

#### Scenario: Party inventory notes
- **WHEN** an item is added to party inventory
- **THEN** the system SHALL allow optional notes (e.g., "found in dragon hoard", "quest reward from Lord Harkin")
