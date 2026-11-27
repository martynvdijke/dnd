# compendium-npc-linking Specification

## Purpose
Link character-associated NPCs from compendium/global NPC pool.

## Requirements

### Requirement: NPC linking from global NPC pool

The NPC linking flow for characters SHALL allow searching across all NPCs (not just the current user's NPCs) when linking an NPC to a character. The link picker SHALL include a search input that queries the global NPC pool via the API.

#### Scenario: Search all NPCs when linking

- **WHEN** the user opens the "Link NPC" modal for a character
- **THEN** the NPC selector SHALL include a search input that finds NPCs across all users (DM-visible), not just the current user's NPCs

#### Scenario: Link NPC to character

- **WHEN** the user selects an NPC from the search results and confirms
- **THEN** a new `CharacterNPC` record SHALL be created linking the character and the selected NPC with the chosen relationship and notes

### Requirement: Link NPC from compendium monster entry

The NPC linking flow SHALL also allow creating a new NPC from a compendium monster entry. The "Link NPC" modal SHALL have a tab or section for "From Compendium Monster" that opens the monster compendium picker. When a monster is selected, a new NPC SHALL be created with the monster's name and basic stats, then linked to the character.

#### Scenario: Create NPC from compendium monster

- **WHEN** the user selects "From Compendium Monster" in the NPC link modal
- **THEN** the monster compendium picker SHALL open, and after selecting a monster, a new NPC SHALL be created and linked to the character

#### Scenario: New NPC pre-filled from monster data

- **WHEN** a new NPC is created from a compendium monster
- **THEN** the NPC's name, race, and class SHALL be pre-filled from the monster entry data
