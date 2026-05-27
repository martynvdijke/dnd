## ADDED Requirements

### Requirement: Equipped loadout summary panel
The system SHALL display a collapsible "Loadout" panel showing all currently equipped items grouped by slot type (weapon, armor, shield, ring, etc.).

#### Scenario: Loadout panel visible on inventory tab
- **WHEN** the inventory tab is rendered
- **THEN** the system SHALL show a "Loadout" panel at the top, collapsible via a toggle

#### Scenario: Loadout shows equipped items grouped by category
- **WHEN** the loadout panel is expanded
- **THEN** the system SHALL group equipped items by category (Weapon, Armor, Shield, Ring, Wondrous Item, Other) AND show each item's name, and for weapons show damage dice and damage type

#### Scenario: Empty loadout shows placeholder
- **WHEN** no items are equipped
- **THEN** the loadout panel SHALL show "No items equipped" in muted text

#### Scenario: Loadout updates on equip toggle
- **WHEN** an item is equipped or unequipped
- **THEN** the loadout panel SHALL update immediately

#### Scenario: Item category determines slot group
- **WHEN** an item has `is_equipped = true`
- **THEN** the system SHALL use its `category` field to determine which slot group it appears in (weapon items grouped under "Weapons", armor under "Armor", etc.)
