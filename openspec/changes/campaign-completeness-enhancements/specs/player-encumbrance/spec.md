## ADDED Requirements

### Requirement: Encumbrance display on inventory tab
The system SHALL display the character's current encumbrance state in the inventory tab header.

#### Scenario: Encumbrance shown for unencumbered character
- **WHEN** a character's total carried weight (inventory items + coin weight at 50 coins/lb) is less than STR × 5
- **THEN** the inventory tab SHALL show "Encumbrance: Unencumbered (weight/capacity)" in green

#### Scenario: Encumbrance shown for encumbered character
- **WHEN** a character's total carried weight is between STR × 5 and STR × 10
- **THEN** the inventory tab SHALL show "Encumbrance: Encumbered (weight/capacity)" in yellow, AND the character's speed SHALL be noted as reduced by 10

#### Scenario: Encumbrance shown for heavily encumbered character
- **WHEN** a character's total carried weight is between STR × 10 and STR × 15
- **THEN** the inventory tab SHALL show "Encumbrance: Heavily Encumbered (weight/capacity)" in red, AND the character's speed SHALL be noted as reduced by 20, AND the character SHALL have disadvantage on ability checks, attack rolls, and saving throws using STR/DEX/CON

#### Scenario: Encumbrance shown for push/drag/lift beyond capacity
- **WHEN** a character's total carried weight exceeds STR × 15
- **THEN** the inventory tab SHALL show "Over Capacity! (weight exceeds STR × 15)" in red, AND the character SHALL be unable to move

#### Scenario: Coin weight included in encumbrance calculation
- **WHEN** calculating total carried weight
- **THEN** the system SHALL add (CP + SP + EP + GP + PP) / 50 pounds to the inventory item weight total

#### Scenario: Encumbrance updates on inventory change
- **WHEN** an inventory item is added, removed, quantity changed, or equipped state toggled
- **THEN** the encumbrance display SHALL update immediately
