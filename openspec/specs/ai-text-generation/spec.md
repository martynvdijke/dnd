# ai-text-generation Specification

## Purpose
Send prompts to configured text-generation AI endpoints and display or insert the results (one-shot openings, premises, hooks, NPC descriptions, and other narrative content). Conversational drafting and revision are covered separately by the ai-structured-drafting capability; this spec covers the stateless text endpoints only.

## Requirements

### Requirement: User can select a text AI endpoint

The system SHALL present available enabled text-type AI endpoints for selection when the user triggers AI generation.

#### Scenario: List available text endpoints

- **WHEN** a user clicks "Generate with AI" in a content area that supports text generation
- **THEN** the system SHALL show a dropdown of all enabled text-type AI endpoints

#### Scenario: No text endpoints configured

- **WHEN** a user clicks "Generate with AI" and no text-type AI endpoints are configured or enabled
- **THEN** the system SHALL display a message "No AI endpoints configured. Ask your admin to configure AI endpoints."

### Requirement: Text generation endpoint contract

`POST /api/ai/generate/text` SHALL accept `{endpoint_id, prompt, system?, max_tokens?, session_id?}`, authenticate as a DM-or-admin session through the `/api` route group (with CSRF protection), and return `{text, finish, session_id}`. When no system prompt is supplied the default system prompt SHALL be used. The request SHALL time out after 60 seconds.

#### Scenario: Successful generation

- **WHEN** a DM or admin posts a valid `endpoint_id` and prompt
- **THEN** the endpoint responds 200 with the generated text, the finish reason, and a session id the caller can reuse for follow-up turns

#### Scenario: Invalid request body

- **WHEN** the request is missing `endpoint_id` or a non-empty prompt
- **THEN** the endpoint responds 400 with `"endpoint_id and prompt are required"`

#### Scenario: Endpoint not available

- **WHEN** the requested endpoint is not an enabled text endpoint
- **THEN** the endpoint responds 404 `"enabled text endpoint not found"`

### Requirement: Text generation degrades safely

Text generation SHALL be gated by the site-wide `ai_enabled` setting and SHALL surface provider failures as user-friendly messages without exposing raw API details.

#### Scenario: AI features disabled

- **WHEN** the `ai_enabled` app setting is off and any AI generation endpoint is called
- **THEN** the endpoint responds 403 `"AI features are disabled"`

#### Scenario: Provider failure is reported cleanly

- **WHEN** the AI endpoint returns an error (timeout, rate limit, invalid request)
- **THEN** the system SHALL display a user-friendly error message without exposing raw API details

#### Scenario: Generation timeout

- **WHEN** the AI endpoint does not respond within 60 seconds
- **THEN** the system SHALL cancel the request and display a timeout error message

### Requirement: Generated text can be inserted into content

The system SHALL allow users to insert or replace content with the generated text.

#### Scenario: Insert generated text

- **WHEN** a user is editing a one-shot's premise or hook field and generates text via AI
- **THEN** the system SHALL offer "Replace" and "Insert after" actions to apply the generated text to the target field

### Requirement: One-shot opening generation shortcut

The system SHALL provide a dedicated "Generate Opening" action in the one-shot editor that pre-fills a context-aware prompt structure.

#### Scenario: Generate one-shot opening

- **WHEN** a user clicks "Generate Opening" on a one-shot adventure with a title and premise already filled
- **THEN** the system SHALL pre-fill the prompt with "Write a compelling opening hook for a D&D one-shot titled '[title]' with this premise: [premise]" and open the AI generation modal

### Requirement: Vision endpoints transcribe ink-note pages

Endpoints of type `vision` SHALL be OpenAI-compatible chat-style endpoints used by the ink-notes OCR flow to transcribe handwritten session pages into text. The OCR flow SHALL only transcribe pages that have no text yet.

#### Scenario: Ink OCR uses a vision endpoint

- **WHEN** a user triggers OCR on an ink capture whose pages have no text and an enabled vision endpoint is configured
- **THEN** each pending page image is transcribed by the vision endpoint and stored with `ocr_source='vision'`

#### Scenario: No vision endpoint degrades to manual entry

- **WHEN** OCR is triggered and no enabled vision endpoint is configured
- **THEN** the flow responds 409 with a "type the text manually" message and a `manual` flag so the UI falls back to manual transcription without data loss

#### Scenario: Vision failure keeps manual fallback

- **WHEN** the vision provider fails or a page image is unavailable
- **THEN** the error is surfaced with the `manual` flag set and previously transcribed text is preserved
