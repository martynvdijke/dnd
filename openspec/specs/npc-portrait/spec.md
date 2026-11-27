# npc-portrait Specification

## Purpose
Portrait image support for NPC entities (display, upload, update).

## Requirements

### Requirement: NPC portrait_url field

NPCs SHALL support a `portrait_url` field for displaying a character portrait image.

#### Scenario: Create NPC with portrait
- **WHEN** user creates an NPC with a `portrait_url` value
- **THEN** the NPC is saved with that portrait_url and displayed in the NPC list and detail views

#### Scenario: Update NPC portrait
- **WHEN** user updates an NPC's `portrait_url`
- **THEN** the portrait updates in all views

#### Scenario: NPC list shows portrait thumbnail
- **WHEN** user views the NPC list
- **THEN** each NPC entry shows a small thumbnail of their portrait (or a default icon if none)

#### Scenario: Upload portrait from NPC form
- **WHEN** user clicks upload on the NPC form
- **THEN** the file is uploaded via POST /api/upload with owner_type=npc and the returned URL is set as portrait_url
