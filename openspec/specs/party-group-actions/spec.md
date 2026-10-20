# party-group-actions Specification

## Purpose
TBD - created by archiving change party-radar-group-actions-and-loot. Update Purpose after archive.
## Requirements
### Requirement: Party passive senses radar
The Party View SHALL display Passive Perception, Passive Insight, and Passive Investigation for each character in the party (`10 + ability modifier + proficiency bonus if proficient`), along with party-wide highest passive scores.

#### Scenario: Viewing party passive scores
- **WHEN** a user opens Party View with characters in the party
- **THEN** each character card and the party summary display Passive Perception, Passive Insight, and Passive Investigation

### Requirement: Party group rolls and party rest
The Party View SHALL allow triggering a Group Roll (ability check, skill check, or saving throw) across all party characters simultaneously and SHALL allow triggering a Short Rest or Long Rest for all party characters at once.

#### Scenario: Running a group Stealth check for the party
- **WHEN** the user selects `Stealth` in the Party Group Roll control and clicks Roll
- **THEN** the system rolls a Stealth check for each party character and displays all results together in Party View

#### Scenario: Triggering a party Long Rest
- **WHEN** the user clicks the Party Long Rest button in Party View
- **THEN** each character in the party executes a Long Rest via `/api/characters/:id/rest` and the Party View refreshes their HP

### Requirement: Atomic party loot claiming into character inventory
The system SHALL provide `POST /api/campaigns/:id/party-items/:itemId/claim` accepting `character_id` and optional `quantity` (defaulting to full quantity). The endpoint SHALL require campaign access, create a corresponding inventory item on the target character, and decrement or remove the party item atomically.

#### Scenario: Claiming a party item into a character's inventory
- **WHEN** a campaign member calls `POST /api/campaigns/:id/party-items/:itemId/claim` with a valid `character_id` for a party item with `quantity: 1`
- **THEN** the item is added to that character's `inventory_items` and removed from the campaign's `party_items`

#### Scenario: Claiming partial quantity of a stacked party item
- **WHEN** a campaign member claims `quantity: 2` from a party item that has `quantity: 5`
- **THEN** the character receives an inventory item with `quantity: 2` and the party item remains with `quantity: 3`
