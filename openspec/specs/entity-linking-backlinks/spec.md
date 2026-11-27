# entity-linking-backlinks Specification

## Purpose
Unified cross-entity references, @-mentions in rich text, rendered link chips, and backlink panels on entity views.

## Requirements

### Requirement: Universal entity links
The system SHALL provide a unified link store allowing any entity type to be linked to any other entity type, with links created via API and removable by users with edit permission on the source entity. Links SHALL record source type/ID, target type/ID, creation context (`manual`, `mention`, or `import`), and creation time.

#### Scenario: Link a location to a quest
- **WHEN** a user with edit permission creates a link from a quest to a location
- **THEN** the link is stored with context `manual` and appears in the quest's linked-entities list

#### Scenario: Remove a link
- **WHEN** a user removes an existing link from an entity
- **THEN** the link no longer appears in either entity's link lists

#### Scenario: Deleting an entity cleans up its links
- **WHEN** an entity is deleted
- **THEN** all links where it is source or target are removed

### Requirement: Backlinks panel on entity views
Every entity detail view SHALL display a "Linked from" backlink panel listing all entities that link to the displayed entity, grouped by source entity type, with each entry navigating to the source entity's deep link.

#### Scenario: Backlink appears on target entity
- **WHEN** an adventure links to an NPC and a user views that NPC
- **THEN** the NPC's backlink panel lists the adventure, and activating it opens the adventure

### Requirement: @-mentions in rich text
Rich-text editors (wiki pages, notes, journal entries, DM notes) SHALL support @-mentions: typing `@` opens an autocomplete fed by universal search; selecting a result inserts an inline chip displaying the entity name. On save, the system SHALL reconcile mention-derived links (context `mention`) with the document's current mention set, adding and removing links to match.

#### Scenario: Insert a mention
- **WHEN** a user types `@`, selects "Emberhold Tavern" from the autocomplete, and saves the page
- **THEN** the document renders an inline chip for the location and a `mention` link exists from the document to that location

#### Scenario: Removing a mention removes the link
- **WHEN** a user deletes a previously saved mention chip and saves the document
- **THEN** the corresponding `mention` link is removed

#### Scenario: Mention chip navigates to entity
- **WHEN** a reader clicks a rendered mention chip
- **THEN** the application navigates to the mentioned entity's deep-link URL

### Requirement: Entity picker for manual linking
Entity detail views SHALL provide an "Add link" action opening a searchable entity picker (backed by the universal search endpoint) that allows selecting any linkable entity as a link target.

#### Scenario: Manual link via picker
- **WHEN** a user opens "Add link" on a quest, searches for an NPC, and selects it
- **THEN** a `manual` link from the quest to the NPC is created and displayed
