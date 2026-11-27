# ai-inline-generation Specification

## Purpose
Context-aware "Generate with AI" buttons placed next to editable content fields (oneshots, NPCs, locations, items, journal entries) that pre-fill the prompt modal with contextual instructions for the current field.

## Requirements

### Requirement: Inline "Generate with AI" buttons
The system SHALL provide small trigger buttons next to content fields that open the AI modal pre-configured for that field.

#### Scenario: Text field trigger
- **WHEN** a textarea or input field has the attribute `data-ai-generate="text"` and a sibling or adjacent `.ai-generate-btn` element exists
- **THEN** clicking the button calls `openAIGenModal('text', fieldId, contextHint)` where `fieldId` is the target field and `contextHint` is derived from the field's label or placeholder

#### Scenario: Image field trigger
- **WHEN** a field has `data-ai-generate="image"` with an adjacent `.ai-generate-btn`
- **THEN** clicking the button calls `openAIGenModal('image', fieldId, promptHint)` for image generation

### Requirement: Context-aware prompt hints
The system SHALL pre-fill the prompt textarea with contextual hints derived from the target field.

#### Scenario: Prompt hint from data attribute
- **WHEN** a trigger button has a `data-ai-hint` attribute
- **THEN** the modal's prompt textarea is pre-filled with that hint text

#### Scenario: Prompt hint from label
- **WHEN** no `data-ai-hint` is set but the target field has a visible label
- **THEN** the prompt is pre-filled with a template like "Generate a [label text] for a D&D campaign"

### Requirement: Show/hide system prompt field
The system SHALL show the system prompt field only in text mode.

#### Scenario: System prompt visible for text
- **WHEN** the modal opens in text mode
- **THEN** the system prompt textarea is visible with a default or empty value

#### Scenario: System prompt hidden for image
- **WHEN** the modal opens in image mode
- **THEN** the system prompt textarea is hidden

### Requirement: Insert/Replace buttons visibility
The system SHALL show Insert and Replace buttons only for text mode (not image mode).

#### Scenario: Insert/Replace visible for text
- **WHEN** the modal shows a text result
- **THEN** Insert and Replace buttons are visible

#### Scenario: Insert/Replace hidden for image
- **WHEN** the modal shows an image result
- **THEN** Insert and Replace buttons are hidden; only "Generate Again" is shown

### Requirement: Backend endpoint for DM-accessible endpoint listing
The system SHALL provide a backend endpoint for DMs (non-admin) to list enabled AI endpoints by type.

#### Scenario: List enabled text endpoints
- **WHEN** a DM user calls `GET /api/ai/endpoints?type=text`
- **THEN** the system returns an array of enabled text endpoints with `id`, `name`, `model`, and `type` (no API keys or base URLs)

#### Scenario: List enabled image endpoints
- **WHEN** a DM user calls `GET /api/ai/endpoints?type=image`
- **THEN** the system returns an array of enabled image endpoints with `id`, `name`, `model`, and `type`

#### Scenario: Auth required
- **WHEN** an unauthenticated user calls `GET /api/ai/endpoints`
- **THEN** the system returns 401 Unauthorized
