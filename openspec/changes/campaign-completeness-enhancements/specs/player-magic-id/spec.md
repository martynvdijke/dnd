## ADDED Requirements

### Requirement: Magic item identification state
The system SHALL allow tracking whether a magic item has been identified, using an `is_identified` boolean field on InventoryItem.

#### Scenario: Identified status shown in inventory
- **WHEN** an inventory item has `is_magical = true`
- **THEN** the system SHALL display an icon or badge indicating whether it's identified (identified: visible name and description) or unidentified (name shown as "Unidentified Item" with description hidden or replaced by " unidentified")
- **WHEN** an item is not magical
- **THEN** no identification badge is shown

#### Scenario: Toggle identified state
- **WHEN** viewing a magical inventory item
- **THEN** the system SHALL provide a button or toggle to mark it as identified or unidentified
- **WHEN** toggled
- **THEN** the system SHALL send a PATCH to update `is_identified` on the item AND update the display

#### Scenario: Unidentified item edit
- **WHEN** an item is marked as unidentified
- **THEN** the system SHALL hide the item's description, damage_dice, damage_type, and weapon_properties from the item row (DM notes remain visible to DMs)
