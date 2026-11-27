# armor-class Specification

## Purpose
TBD - created by archiving change add-combat-depth. Update Purpose after archive.
## Requirements
### Requirement: Armor Class derivation
The system SHALL derive a character's Armor Class from equipped body armor, any equipped shield, and the Dexterity modifier when body armor is equipped. Light armor adds the full Dexterity modifier, medium armor adds the Dexterity modifier capped at +2, and heavy armor ignores the Dexterity modifier. Any equipped item with an armor bonus that is not body armor (a shield) adds its bonus to the total, and the unarmored base is 10. When no body armor is equipped the stored manual AC SHALL be used unchanged.

#### Scenario: Light armor adds full Dexterity
- **WHEN** a character with Dexterity 16 equips light armor with base AC 11
- **THEN** the derived AC is 14

#### Scenario: Medium armor caps Dexterity
- **WHEN** a character with Dexterity 18 equips medium armor with base AC 14
- **THEN** the derived AC is 16

#### Scenario: Heavy armor ignores Dexterity
- **WHEN** a character with Dexterity 8 equips heavy armor with base AC 16
- **THEN** the derived AC is 16

#### Scenario: Shield adds its bonus
- **WHEN** a character with derived AC 14 equips a shield granting +2 AC
- **THEN** the derived AC is 16

#### Scenario: Unarmored keeps the manual value
- **WHEN** a character with no equipped body armor has a manual AC of 15
- **THEN** the reported AC is 15

#### Scenario: Equipping armor updates the stored AC
- **WHEN** a character equips body armor
- **THEN** the character's stored AC is recomputed and persisted
