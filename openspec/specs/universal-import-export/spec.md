# universal-import-export Specification

## Purpose
Per-entity and full-campaign JSON export/import with schema validation, dry-run preview, relationship preservation, and conflict handling.

## Requirements

### Requirement: Per-entity-type JSON export
The system SHALL provide authenticated JSON export for each major entity type: campaigns, one-shot adventures, NPCs, locations, quests, encounters, custom monsters, shops, factions, journal entries, timeline events, and homebrew entries. Each export document SHALL use the versioned envelope `{ "format": "villum-transfer", "version": <n>, "type": "<type>", "entities": [...] }` and each entity SHALL carry a stable `ref` identifier.

#### Scenario: Export NPCs as JSON
- **WHEN** a user exports the NPC list
- **THEN** a JSON file downloads containing a valid `villum-transfer` document with all NPCs the user can view, each with a stable `ref`

#### Scenario: Export preserves relationships by reference
- **WHEN** a user exports an adventure that links an NPC and a location
- **THEN** the exported document contains those links expressed via the entities' `ref` values, not local database IDs

### Requirement: Full campaign bundle export
The system SHALL provide a campaign bundle export that includes the campaign and its owned content (members excepted): NPCs, locations, quests, adventures with acts/scenes/clues, encounters, factions, timeline events, wiki pages, and party inventory, with all cross-references expressed via `ref` values.

#### Scenario: Bundle contains complete campaign content
- **WHEN** a DM exports a campaign bundle
- **THEN** the resulting document contains the campaign and all content owned by that campaign, and every internal cross-reference uses a `ref` present in the document

### Requirement: Import with validation and dry-run preview
The system SHALL provide import endpoints that accept `villum-transfer` documents and run in three phases: schema validation (reject malformed or unsupported-version documents with per-entity errors), a dry-run preview reporting entity counts and conflicts, and — only after explicit user confirmation — a transactional apply that remaps `ref` values to new local IDs and recreates cross-entity links.

#### Scenario: Invalid document rejected with errors
- **WHEN** a user uploads a document with an unsupported `version` or malformed entities
- **THEN** the import is rejected and the response lists per-entity validation errors

#### Scenario: Dry-run preview before apply
- **WHEN** a user uploads a valid bundle
- **THEN** the system returns a preview with counts per entity type and detected conflicts, without writing any data

#### Scenario: Confirmed import applies atomically
- **WHEN** a user confirms an import after preview and any entity fails during apply
- **THEN** the entire import is rolled back and no partial data persists

#### Scenario: Imported bundle recreates links
- **WHEN** a campaign bundle containing an adventure linked to an NPC is imported
- **THEN** the imported adventure is linked to the newly created NPC in the local database

### Requirement: Import UI on entity list views
Each major entity list view SHALL expose Import and Export actions, and the campaign view SHALL expose Export Bundle and Import Bundle actions. The import flow SHALL surface validation errors and the dry-run preview before any data is written.

#### Scenario: Import flow from the UI
- **WHEN** a user chooses Import on the NPCs view, selects a valid file, reviews the preview, and confirms
- **THEN** the new NPCs appear in the list and a success summary is shown

### Requirement: Import audit log
The system SHALL record an import log entry for each applied import, including file name, entity counts, user, and timestamp, viewable in the admin area.

#### Scenario: Import is logged
- **WHEN** an import is applied
- **THEN** an import log entry exists recording the file name, per-type counts, importing user, and timestamp
