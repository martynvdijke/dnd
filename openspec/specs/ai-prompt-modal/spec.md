# ai-prompt-modal Specification

## Purpose
A reusable modal UI for entering a prompt, selecting an AI endpoint, triggering text/image generation, displaying results, and re-running with the same prompt for varied output.

## Requirements

### Requirement: Open AI prompt modal
The system SHALL provide a reusable Bootstrap modal for AI generation that can be opened from anywhere in the app.

#### Scenario: Open modal via FAB
- **WHEN** a user clicks the "Generate with AI" option in the floating action button menu
- **THEN** the AI prompt modal opens in text mode with no target context or pre-filled prompt

#### Scenario: Open modal via API call
- **WHEN** `window.openAIGenModal(mode, targetId, promptHint, title)` is called
- **THEN** the modal opens with the specified mode (text/image), an optional target field ID, an optional pre-filled prompt hint, and an optional title

### Requirement: Select AI endpoint
The modal SHALL display a dropdown of enabled AI endpoints filtered by the current generation mode (text/image).

#### Scenario: Load endpoints on open
- **WHEN** the modal opens
- **THEN** the system fetches enabled endpoints matching the mode via `GET /api/ai/endpoints?type=<mode>` and populates the dropdown

#### Scenario: No endpoints available
- **WHEN** no enabled endpoints exist for the selected mode
- **THEN** the dropdown shows "No enabled endpoints available" and the Generate button is disabled

#### Scenario: Load failure
- **WHEN** the endpoint fetch fails
- **THEN** the dropdown shows "Failed to load endpoints" and the Generate button is disabled

### Requirement: Generate text with AI
The system SHALL send the prompt to the configured AI text endpoint and display the result.

#### Scenario: Successful text generation
- **WHEN** the user selects a text endpoint, enters a prompt, optionally enters a system prompt, and clicks Generate
- **THEN** the system sends `POST /api/ai/generate/text` with `endpoint_id`, `prompt`, and `system`, displays a loading spinner, and upon success shows the generated text in the result area

#### Scenario: Missing endpoint or prompt
- **WHEN** the user clicks Generate without selecting an endpoint or entering a prompt
- **THEN** a toast error message is shown and no API call is made

### Requirement: Generate image with AI
The system SHALL send the prompt to the configured AI image endpoint and display the generated image.

#### Scenario: Successful image generation
- **WHEN** the user selects an image endpoint, enters a prompt, and clicks Generate
- **THEN** the system sends `POST /api/ai/generate/image` with `endpoint_id` and `prompt`, displays a loading spinner, and upon success shows the image in the result area

#### Scenario: Image generation shows first image
- **WHEN** the response contains an `images` array
- **THEN** the first image URL is displayed in an `<img>` tag within the modal

### Requirement: Generate Again with same prompt
The system SHALL allow the user to re-run the same prompt to produce different results.

#### Scenario: Regenerate from result
- **WHEN** the user clicks "Generate Again" after a successful generation
- **THEN** the same prompt is re-sent to the same endpoint, the result area is replaced with the new output

#### Scenario: Regenerate preserves prompt text
- **WHEN** the user clicks "Generate Again"
- **THEN** the prompt textarea retains the previously entered text

### Requirement: Insert generated text into target
The system SHALL allow the user to insert generated text into the target field.

#### Scenario: Insert text
- **WHEN** the user clicks "Insert" after text generation and a target field ID is set
- **THEN** the generated text is appended to the target field's existing content, a success toast is shown, and the modal closes

#### Scenario: Replace text
- **WHEN** the user clicks "Replace" after text generation and a target field ID is set
- **THEN** the target field's content is replaced with the generated text, a success toast is shown, and the modal closes

### Requirement: Keep last result for regeneration
The system SHALL store the last generated result so it can be reused or regenerated.

#### Scenario: Last result stored
- **WHEN** a generation completes
- **THEN** the result text or image URL is stored in module-level variables for use by Insert, Replace, and Generate Again actions

#### Scenario: Last result cleared on new modal open
- **WHEN** the modal opens
- **THEN** the previous last result is cleared and the result display area is hidden
