# oneshot-npc-linking Specification

## Purpose
Link NPCs with story context to one-shots and allow using them in combat encounters.

## Requirements

### Requirement: One-shot can link NPCs with story context

A one-shot adventure SHALL support linking NPCs with metadata: story hook text, role (ally/enemy/neutral), and a combat-ready flag. Linked NPCs SHALL appear in the one-shot detail view.

#### Scenario: Link NPC to one-shot
- **WHEN** user opens the NPC linking interface in a one-shot
- **THEN** they SHALL search and select an NPC, providing story hook text and role

#### Scenario: View linked NPCs in one-shot
- **WHEN** user views a one-shot detail
- **THEN** all linked NPCs SHALL be shown with their name, role badge, and story hook

#### Scenario: Remove NPC from one-shot
- **WHEN** user clicks remove on a linked NPC
- **THEN** the NPC SHALL be unlinked from the one-shot

### Requirement: One-shot NPCs can be used in combat

Linked NPCs with the combat-ready flag SHALL be available to add to the combat tracker when running a combat encounter within the one-shot.

#### Scenario: Add NPC to combat from one-shot
- **WHEN** user opens the combat tracker from a one-shot
- **THEN** combat-ready NPCs linked to that one-shot SHALL appear as selectable combatants

#### Scenario: NPC in combat shows story context
- **WHEN** a linked NPC is added to combat
- **THEN** the NPC's entry in combat SHALL show its story hook text as a tooltip

### Requirement: NPC linking API

The system SHALL provide API endpoints to list, add, and remove NPC links for a one-shot adventure. The link SHALL include: `npc_id`, `oneshot_adventure_id`, `story_hook` (text), `role` (enum: ally/enemy/neutral), `combat_ready` (boolean).

#### Scenario: List linked NPCs for one-shot
- **WHEN** client calls `GET /api/oneshot-adventures/{id}/npcs`
- **THEN** the response SHALL include all linked NPCs with their metadata
