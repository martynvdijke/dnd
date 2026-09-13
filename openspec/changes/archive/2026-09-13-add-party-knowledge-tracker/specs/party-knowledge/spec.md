## ADDED Requirements

### Requirement: Knowledge entry CRUD
The system SHALL allow campaign owners (DMs) to create, edit, and delete knowledge entries scoped to their campaign. Each entry SHALL have a title, content, source, status (`rumor`, `confirmed`, `revealed`, `false`), and a shared flag. Invalid status values SHALL be rejected.

#### Scenario: DM creates a rumor
- **WHEN** the DM submits a new entry with title, content, and status `rumor`
- **THEN** the entry is stored and appears in the campaign knowledge list

#### Scenario: Invalid status rejected
- **WHEN** an entry is submitted with a status outside the allowed enum
- **THEN** the request fails with a validation error

### Requirement: Status lifecycle with history
The DM SHALL be able to change an entry's status at any time. Every status change SHALL be recorded with a timestamp in the entry's history.

#### Scenario: Rumor becomes revealed
- **WHEN** the DM changes an entry from `rumor` to `revealed`
- **THEN** the entry shows the new status and the history records the transition

### Requirement: Known-by tracking
The DM SHALL be able to link characters to an entry recording that the PC has been told it, and remove such links.

#### Scenario: Mark a PC as told
- **WHEN** the DM adds a character to an entry's known-by list
- **THEN** the character appears on the entry and the entry is discoverable from the character's perspective

### Requirement: Entity cross-linking
Knowledge entries SHALL participate in the existing `entity_links` mechanism so NPCs, locations, quests, and other linkable entities can be attached, with backlinks rendered on the linked entities.

#### Scenario: Link an NPC to a secret
- **WHEN** the DM attaches an NPC to a knowledge entry
- **THEN** the entry lists the NPC and the NPC's page shows a backlink

### Requirement: Visibility split
Campaign members who are not the owner SHALL see only entries with `shared = true`; the owner SHALL see all entries. Universal search results SHALL respect this split for all users. Unshared entries SHALL NOT be retrievable by non-owner members through any endpoint.

#### Scenario: Member sees only shared entries
- **WHEN** a non-owner member opens the campaign knowledge view or searches
- **THEN** only shared entries appear

#### Scenario: Owner sees everything
- **WHEN** the campaign owner opens the knowledge view
- **THEN** all entries appear regardless of shared flag

#### Scenario: Direct access to unshared entry is denied
- **WHEN** a non-owner member requests an unshared entry directly
- **THEN** the system responds 404

### Requirement: Campaign knowledge UI
The campaign area SHALL provide a knowledge panel listing entries grouped or filterable by status, with a detail view featuring a mentions-enabled editor, known-by picker, entity links, share toggle, and a bulk "reveal to party" action that marks all party characters known-by and sets `shared = true`.

#### Scenario: Filter by status
- **WHEN** the DM filters the list by `false`
- **THEN** only misinformation entries are shown

#### Scenario: Reveal to party in one action
- **WHEN** the DM uses the bulk reveal action on an entry
- **THEN** all campaign PCs are added to known-by and the entry becomes shared
