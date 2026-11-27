# e2e-ai-features Specification

## Purpose
Playwright e2e tests for AI generation modal, endpoint listing, text/image generation flows.

## Requirements

### Requirement: AI generation modal e2e coverage
The Playwright e2e tests SHALL verify the AI generation modal UI and behavior.

#### Scenario: Open AI generation modal
- **WHEN** user clicks an AI generation trigger button (e.g., in one-shot editor, NPC editor)
- **THEN** the AI generation modal SHALL open with mode-appropriate options

#### Scenario: Generate text with AI
- **WHEN** user enters a prompt and clicks "Generate" in the AI modal
- **THEN** generated text SHALL appear in the modal result area (or error if no endpoint configured)

#### Scenario: Insert generated text
- **WHEN** user clicks "Insert" after generating text
- **THEN** the generated text SHALL be inserted into the target field

#### Scenario: Regenerate text
- **WHEN** user clicks "Generate Again" after an initial generation
- **THEN** a new result SHALL be generated

#### Scenario: Close AI modal without inserting
- **WHEN** user clicks close on the AI modal without inserting
- **THEN** the modal SHALL close and no changes SHALL be applied to the target field

### Requirement: AI endpoint listing e2e coverage
The Playwright e2e tests SHALL verify the AI endpoint management UI.

#### Scenario: DM can access AI endpoint list
- **WHEN** DM user triggers an AI generation
- **THEN** the modal SHALL load available AI endpoints (text/image) from the API
