# compendium-linking Specification

## Purpose
TBD - created by archiving change link-compendium-equipment-shops-npcs. Update Purpose after archive.
## Requirements
### Requirement: Player links compendium equipment to character inventory

A player with edit access to a character SHALL be able to link a compendium equipment
entry to that character's inventory. Linking SHALL copy the entry's name, weight,
category, and description into a new inventory row and SHALL store the compendium
reference. Unlinking SHALL remove the reference while preserving the inventory row and its
copied data.

#### Scenario: Player links equipment from the compendium
- **WHEN** a player with edit access opens the character's Inventory and selects "Link from Compendium"
- **THEN** a compendium equipment picker opens, searchable by name
- **WHEN** the player picks an equipment entry
- **THEN** an inventory row is created with the entry's name, weight, category, and description, carrying the compendium link

#### Scenario: Player unlinks equipment, data preserved
- **WHEN** a player unlinks a compendium-linked inventory item
- **THEN** the compendium reference is removed and the inventory row and its copied data remain

### Requirement: DM links compendium equipment to shop items

A DM with access to a campaign or one-shot SHALL be able to link a compendium equipment
entry to a shop item. Creating or updating a linked shop item SHALL copy the entry's name,
category, description, weight, and rarity onto the shop item and SHALL store the
compendium reference. Shop items SHALL display a linked indicator. Unlinking SHALL remove
the reference while preserving the shop item and its copied data. When a player buys a
linked shop item, the resulting inventory row SHALL carry the compendium reference and the
copied name, weight, category, and description.

#### Scenario: DM creates a shop item from the compendium
- **WHEN** a DM opens the shop item form and picks an equipment entry from the compendium picker
- **THEN** the form is populated with the entry's name, category, description, weight, and rarity and the compendium reference is stored on save

#### Scenario: Linked shop item shows a badge and can be unlinked
- **WHEN** a shop item has a compendium reference
- **THEN** it displays a linked indicator, and the DM can unlink it
- **WHEN** the DM unlinks it
- **THEN** the reference is removed and the item's copied data remains editable

#### Scenario: Buying a linked shop item propagates the link
- **WHEN** a player buys a shop item that has a compendium reference
- **THEN** the buyer's inventory gains a row with the compendium reference and the copied name, weight, category, and description

### Requirement: NPC item links reference compendium equipment

A DM SHALL be able to link a compendium equipment entry directly to an NPC, in addition to
linking one-shot items. A link SHALL reference exactly one source: either a one-shot item
or a compendium equipment entry. Listing an NPC's linked items SHALL resolve and display
the item name from whichever source the link references.

#### Scenario: DM links compendium equipment to an NPC
- **WHEN** a DM chooses "link item to NPC" and picks a compendium equipment entry
- **THEN** an NPC item link is created referencing the compendium equipment entry

#### Scenario: NPC item list resolves names from both sources
- **WHEN** an NPC has links to both one-shot items and compendium equipment entries
- **THEN** the NPC's item list displays the correct item name for each link

#### Scenario: Creating a link without a source is rejected
- **WHEN** a link is created with neither a one-shot item id nor a compendium equipment id (or both)
- **THEN** the request is rejected with a client error

### Requirement: Character links compendium race, class, and background

A player with edit access to a character SHALL be able to link a compendium race, class,
and background entry to that character. Linking SHALL copy the entry's name into the
character's race, class, or background text field and SHALL store the compendium
reference. Linked fields SHALL display a linked indicator. Unlinking SHALL remove the
reference while preserving the text.

#### Scenario: Player links a compendium race to a character
- **WHEN** a player opens the character details and picks a race from the compendium picker
- **THEN** the race field is set to the entry's name and the compendium reference is stored

#### Scenario: Player links a compendium class to a character
- **WHEN** a player picks a class from the compendium picker on the character details form
- **THEN** the class field is set to the entry's name and the compendium reference is stored

#### Scenario: Player links a compendium background to a character
- **WHEN** a player picks a background from the compendium picker on the character details form
- **THEN** the background field is set to the entry's name and the compendium reference is stored

#### Scenario: Player unlinks a linked identity field, text preserved
- **WHEN** a player unlinks a character's race, class, or background
- **THEN** the compendium reference is removed and the text in the field remains
