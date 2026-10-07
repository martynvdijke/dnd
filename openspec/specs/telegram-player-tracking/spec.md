# telegram-player-tracking Specification

## Purpose
TBD - created by archiving change telegram-player-tracking. Update Purpose after archive.
## Requirements
### Requirement: Inventory listing

The bot SHALL list the claimed character's inventory on `/inventory` (with `/inv` and `/bag` aliases), showing each item's name, category, quantity, equipped state and attunement state, and SHALL attach inline controls for the allowable actions. The listing SHALL be scoped to a single character the linked user may edit and SHALL NOT reveal another character's items.

#### Scenario: Inventory is listed with controls

- **WHEN** a linked user with a claimed character sends `/inventory`
- **THEN** the reply lists the items with category, quantity, equipped and attunement state and attaches inline controls

#### Scenario: Foreign character is denied

- **WHEN** a linked user requests inventory for a character they may not edit
- **THEN** the bot replies with the same "not found" response as for a nonexistent character and reveals no item data

### Requirement: Inventory management

The bot SHALL let the user equip or unequip an item, change its quantity by one in either direction with a floor of zero, remove it after a confirmation step, and add a new item with `/additem <name> [qty]` defaulting to quantity one. Each mutation SHALL act on the claimed character only and SHALL re-validate edit rights at the time of the action.

#### Scenario: Equip toggles state

- **WHEN** the user taps equip on an unequipped item
- **THEN** the item becomes equipped and the listing reflects it

#### Scenario: Quantity decrements to a floor of zero

- **WHEN** the user taps decrement on an item with quantity one
- **THEN** the quantity stays at one and the bot does not remove the item

#### Scenario: Quantity increments

- **WHEN** the user taps increment on an item
- **THEN** the item's quantity increases by one

#### Scenario: Removal requires confirmation

- **WHEN** the user taps remove on an item
- **THEN** the bot asks for confirmation and removes the item only if the user confirms

#### Scenario: Add item creates the row

- **WHEN** the user sends `/additem Shield +1 2`
- **THEN** an item named "Shield +1" with quantity 2 is added to the claimed character

#### Scenario: Add item without a name is refused

- **WHEN** the user sends `/additem` with no name
- **THEN** the bot explains the usage and adds nothing

#### Scenario: Mutation on a foreign character is refused

- **WHEN** a callback references a character the linked user may not edit
- **THEN** the mutation is refused and no row changes

### Requirement: Spellbook and preparation

The bot SHALL list the claimed character's spells by level on `/spells` (alias `/spellbook`) with their prepared state, and SHALL offer prepare/unprepare controls for spells that can be prepared. Cantrips and always-prepared spells SHALL be shown without preparation controls. `/prepare <spell>` and `/unprepare <spell>` SHALL change the prepared flag for the named spell on the claimed character.

#### Scenario: Spellbook is listed by level

- **WHEN** a linked user with a claimed character sends `/spells`
- **THEN** the reply lists the spells grouped by level with their prepared state

#### Scenario: Cantrips are not toggleable

- **WHEN** the spellbook includes a cantrip
- **THEN** it is shown without a prepare control

#### Scenario: Always-prepared spells are not toggleable

- **WHEN** the spellbook includes an always-prepared spell
- **THEN** it is shown without an unprepare control

#### Scenario: Prepare by button

- **WHEN** the user taps prepare on an unprepared spell
- **THEN** the spell's prepared flag becomes true and the listing reflects it

#### Scenario: Prepare by argument

- **WHEN** the user sends `/prepare <spell>`
- **THEN** the named spell on the claimed character becomes prepared

#### Scenario: Unknown spell is refused

- **WHEN** the user sends `/prepare <spell>` for a spell the claimed character does not have
- **THEN** the bot reports that the spell was not found and changes nothing

### Requirement: Conditions and features

The bot SHALL list the claimed character's active conditions on `/conditions` with remove controls and SHALL add a condition with `/condition <name>`. It SHALL list the claimed character's features read-only on `/features`. Empty or over-long condition names SHALL be rejected.

#### Scenario: Conditions are listed with remove controls

- **WHEN** a linked user with conditions sends `/conditions`
- **THEN** the reply lists the conditions and attaches remove controls

#### Scenario: Add condition

- **WHEN** the user sends `/condition Prone`
- **THEN** a condition named "Prone" is added to the claimed character

#### Scenario: Empty condition is refused

- **WHEN** the user sends `/condition` with no name
- **THEN** the bot explains the usage and adds nothing

#### Scenario: Remove condition

- **WHEN** the user taps remove on a condition
- **THEN** the condition is removed from the claimed character

#### Scenario: Features are read-only

- **WHEN** the user sends `/features`
- **THEN** the reply lists the claimed character's features without mutation controls

### Requirement: Currency management

The bot SHALL show the claimed character's currency on `/money` and SHALL adjust a single denomination with `/money +N|-N <pp|gp|ep|sp|cp>`, clamping the result at zero and never producing a negative amount.

#### Scenario: Currency is shown

- **WHEN** the user sends `/money`
- **THEN** the reply shows the platinum, gold, electrum, silver and copper amounts

#### Scenario: Positive adjustment adds

- **WHEN** the user sends `/money +10 gp`
- **THEN** the gold amount increases by 10

#### Scenario: Overspend is clamped at zero

- **WHEN** the user sends `/money -100 gp` while holding 30 gold
- **THEN** the gold amount becomes zero and is not negative

#### Scenario: Invalid denomination is refused

- **WHEN** the user sends `/money +10 xx`
- **THEN** the bot explains the usage and changes nothing

### Requirement: Tracking authorization

Every tracking listing and mutation SHALL apply only to the claimed character and SHALL verify that the linked user may still edit that character before reading or changing data. Operations SHALL use parameterized statements and SHALL never interpolate user input.

#### Scenario: Unclaimed user is guided

- **WHEN** a linked user without a claim sends a tracking command
- **THEN** the bot explains how to claim a character and changes nothing

#### Scenario: Lost edit rights block the action

- **WHEN** the linked user has lost edit rights to the claimed character
- **THEN** tracking replies report the claim as unavailable and change nothing
