# ai-draft-import Specification

## Purpose
TBD - created by archiving change harden-ai-drafting. Update Purpose after archive.
## Requirements
### Requirement: Import a draft JSON for any AI-draftable entity

The system SHALL provide an endpoint that creates or replaces an entity from a
pasted JSON draft for every AI-draftable type: one-shot, NPC, location,
encounter, faction, campaign, quest and item. The input SHALL accept either a
bare draft object or the assistant's full reply envelope
(`{"status","message","draft"}`), and SHALL tolerate markdown code fences
around the JSON. Without `entity_id` the entity SHALL be created; with
`entity_id` the existing entity SHALL be replaced by the draft's content. A
one-shot replacement SHALL include the draft's acts, scenes, NPCs, locations,
encounters and clues. The owner SHALL come from the caller, and campaign-,
character- or adventure-scoped types SHALL require the caller to have access to
that scope. A draft without a non-empty name/title SHALL be rejected.

#### Scenario: Import a bare draft object
- **WHEN** a DM posts a valid one-shot draft with a title
- **THEN** the adventure and its acts, scenes, NPCs, locations, encounters and clues are created

#### Scenario: Import a non-one-shot draft
- **WHEN** a DM posts a valid NPC (or location, encounter, faction, campaign) draft with a name
- **THEN** that entity is created and linked to the supplied campaign scope

#### Scenario: Replace an existing entity
- **WHEN** a DM posts a revised draft with the `entity_id` of an existing entity they own
- **THEN** the entity's content is replaced by the draft and no duplicate is created

#### Scenario: Import the assistant envelope
- **WHEN** a DM posts the full `{"status","message","draft"}` reply
- **THEN** the nested draft is imported and the envelope fields are ignored

#### Scenario: Tolerate markdown fences
- **WHEN** the pasted JSON is wrapped in ```json fences
- **THEN** the fences are stripped and the draft is imported

#### Scenario: Reject malformed JSON
- **WHEN** the pasted text is not valid JSON
- **THEN** the system responds with a clear validation error and creates nothing

#### Scenario: Reject a draft without a name
- **WHEN** the pasted draft has no non-empty name or title
- **THEN** the system rejects the import with a clear error

#### Scenario: Unknown entity type
- **WHEN** the request names an entity type the assistant cannot draft
- **THEN** the system rejects the request with a clear error

#### Scenario: Scope access is required
- **WHEN** a DM imports into a campaign, character or adventure they do not have access to
- **THEN** the system rejects the request and creates nothing

### Requirement: Import and refine UI

The UI SHALL offer an import action that opens a modal containing an entity-type
selector, a JSON textarea, an Import button and an "Ask AI" instruction box.
On a successful import the modal SHALL stay open, a toast SHALL name the created
entity, and the list SHALL refresh. After import, an AI revision SHALL be
applied to the imported entity immediately (replacing its content) and the
textarea SHALL show the revised JSON. Before import, an AI revision SHALL
update the textarea without creating anything. On failure the error SHALL be
shown in the modal and the pasted text SHALL be preserved.

#### Scenario: Successful import from the UI
- **WHEN** a DM pastes a valid draft and submits the import
- **THEN** the entity is created, a toast names it, and the list refreshes

#### Scenario: AI revision before import
- **WHEN** a DM pastes JSON and asks the AI to change a part
- **THEN** the textarea is replaced by the revised draft and nothing is created

#### Scenario: AI revision after import applies immediately
- **WHEN** the DM asks the AI to change a part after a successful import
- **THEN** the revised draft replaces the imported entity's content

#### Scenario: Failed import keeps the text
- **WHEN** the import request fails
- **THEN** the error is displayed in the modal and the textarea still contains the pasted JSON
