## ADDED Requirements

### Requirement: AI-backed recap generation

The recap generation endpoint SHALL produce a session recap from the campaign's own content. When AI is enabled and a text endpoint is configured, it SHALL use the configured model; otherwise it SHALL return the existing template-based recap. The response SHALL indicate whether AI was used.

#### Scenario: Generated with AI
- **WHEN** the DM requests recap generation while a text endpoint is enabled
- **THEN** the returned recap is model-generated from the campaign's characters, timeline, quests, sessions, and conditions, and is marked as AI-generated

#### Scenario: Template fallback
- **WHEN** the DM requests recap generation while AI is disabled or no text endpoint is configured
- **THEN** the existing template recap is returned and is marked as not AI-generated

#### Scenario: Draft is not auto-persisted
- **WHEN** a recap is generated
- **THEN** no recap row is created until the DM explicitly saves the draft

### Requirement: Recap generation access and errors

Recap generation SHALL require the campaign owner (DM) and SHALL fail with a clear error for provider failures without leaving partial state.

#### Scenario: Non-owner denied
- **WHEN** a non-owner member requests recap generation
- **THEN** the request is denied

#### Scenario: Provider failure
- **WHEN** the configured text endpoint fails or times out
- **THEN** the request returns a clear error and no recap content is persisted

### Requirement: Generate-with-AI editor actions

The journal, notes, and session write-up editors SHALL provide a "Generate with AI" action that fills the editor with AI-generated text using the existing AI generation modal, and SHALL be unavailable when AI is disabled.

#### Scenario: Generate into an editor
- **WHEN** the DM activates "Generate with AI" on an editor with a hint
- **THEN** the generated text is inserted into that editor's target field

#### Scenario: AI disabled hides action
- **WHEN** AI is disabled site-wide
- **THEN** the generate action is disabled or hidden on those editors

### Requirement: Recaps are retrievable by the copilot

Published campaign recaps SHALL be indexed and retrievable by the campaign copilot, respecting campaign membership and visibility, so answers can cite a recap.

#### Scenario: Copilot cites a recap
- **WHEN** a member asks about a past session that has a write-up
- **THEN** the copilot can use that write-up as context and return it as a cited source

#### Scenario: Non-member cannot retrieve
- **WHEN** a user who is not a campaign member queries the copilot
- **THEN** the recap is not retrievable
