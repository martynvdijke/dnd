## ADDED Requirements

### Requirement: Attunement slot counter in inventory
The system SHALL display an attunement slot counter (X / 3) in the inventory tab header, counting items with `attunement = true` AND `is_equipped = true`.

#### Scenario: Attunement counter shows available slots
- **WHEN** the inventory tab is rendered
- **THEN** the system SHALL display "Attunement: X / 3" where X is the count of equipped items with attunement enabled

#### Scenario: Attunement warning when at or over limit
- **WHEN** a character has 3 attuned items
- **THEN** the attunement counter SHALL be shown in yellow with text "Full"
- **WHEN** a character has more than 3 attuned items equipped
- **THEN** the counter SHALL be shown in red with text "Over limit! Only 3 can be attuned"

#### Scenario: Attunement indicator on item rows
- **WHEN** an inventory item has `attunement = true`
- **THEN** the item row SHALL show a small "Requires Attunement" badge or icon
- **WHEN** that item is also equipped
- **THEN** the badge SHALL indicate it counts toward the attunement limit

#### Scenario: Attunement count updates on equip toggle
- **WHEN** a user equips or unequips an attunement-requiring item
- **THEN** the attunement counter SHALL update immediately
