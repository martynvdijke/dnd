# npc-campaign-linking Specification

## Purpose
Associate NPCs with campaigns and reuse them across adventures within a campaign.

## Requirements

### Requirement: DM can link an existing NPC to a campaign
The system SHALL allow a DM to associate an existing NPC with a campaign, specifying the NPC's role within that campaign.

#### Scenario: Link NPC to campaign
- **WHEN** the DM selects an NPC and a campaign from the campaign NPC management UI and submits the link
- **THEN** the NPC is associated with the campaign in the `campaign_npcs` table with the specified role

#### Scenario: Duplicate link prevention
- **WHEN** the DM attempts to link the same NPC to the same campaign a second time
- **THEN** the system SHALL return an error indicating the NPC is already linked

### Requirement: DM can create an NPC inline from campaign context
The system SHALL allow creating a new NPC directly from the campaign NPC management section without navigating away.

#### Scenario: Create NPC inline
- **WHEN** the DM fills in the inline NPC creation form (name, race, class, description) within the campaign NPC section
- **THEN** the system creates a new NPC record and immediately links it to the campaign with the specified role

### Requirement: DM can view all NPCs linked to a campaign
The system SHALL display a list of all NPCs associated with a campaign, showing their name, role, and any notes.

#### Scenario: View campaign NPCs
- **WHEN** the DM opens the campaign NPC section
- **THEN** the system SHALL display all NPCs linked to that campaign with name, race, class, role, and notes

### Requirement: DM can remove an NPC from a campaign
The system SHALL allow unlinking an NPC from a campaign without deleting the NPC itself.

#### Scenario: Unlink NPC from campaign
- **WHEN** the DM clicks the unlink button next to an NPC in the campaign NPC list
- **THEN** the link between the NPC and campaign is removed, but the NPC record persists

### Requirement: DM can update NPC role within a campaign
The system SHALL allow updating the role and notes of an NPC linked to a campaign.

#### Scenario: Update campaign NPC metadata
- **WHEN** the DM edits the role or notes for a campaign-linked NPC and submits
- **THEN** the system updates the `campaign_npcs` record with the new values
