# ai-image-generation Specification

## Purpose
Send prompts to configured image-generation AI endpoints and display or save generated images (NPC portraits, location images, item illustrations). Conversational drafting is covered separately by the ai-structured-drafting capability; this spec covers the stateless image endpoints only.

## Requirements

### Requirement: User can select an image AI endpoint

The system SHALL present available enabled image-type AI endpoints for selection when the user triggers image generation.

#### Scenario: List available image endpoints

- **WHEN** a user clicks "Generate Image" in a content area that supports image generation
- **THEN** the system SHALL show a dropdown of all enabled image-type AI endpoints

#### Scenario: No image endpoints configured

- **WHEN** a user clicks "Generate Image" and no image-type AI endpoints are configured or enabled
- **THEN** the system SHALL display a message "No image AI endpoints configured. Ask your admin to configure image generation endpoints."

### Requirement: Image generation endpoint contract

`POST /api/ai/generate/image` SHALL accept `{endpoint_id, prompt, n?, size?}`, authenticate as a DM-or-admin session through the `/api` route group (with CSRF protection), and return generated image data URLs. The count `n` SHALL be clamped to 1–4, and the image size SHALL default to the endpoint's configured size or `1024x1024`. The request SHALL time out after 120 seconds.

#### Scenario: Successful generation

- **WHEN** a DM or admin posts a valid `endpoint_id` and prompt
- **THEN** the endpoint responds 200 with the generated image data for the requested count

#### Scenario: Invalid request body

- **WHEN** the request is missing `endpoint_id` or `prompt`
- **THEN** the endpoint responds 400 with `"endpoint_id and prompt are required"`

#### Scenario: Endpoint not available

- **WHEN** the requested endpoint is not an enabled image endpoint
- **THEN** the endpoint responds 404 `"enabled image endpoint not found"`

### Requirement: Generated image can be saved to the app

`POST /api/ai/save-image` SHALL accept `{url}` (and optional entity association) from a DM-or-admin session, download the image, and store it as an upload record in the media directory for use as NPC portraits, location images, or item illustrations. The URL SHALL be validated as `http(s)`, its host SHALL resolve to a publicly reachable address (loopback, private, link-local and unspecified hosts are rejected), and the response content type SHALL be an image type.

#### Scenario: Save generated image

- **WHEN** a DM or admin posts the URL of a generated image
- **THEN** the system SHALL download the image, store it in the media directory, create an upload record, and associate it with the relevant entity

#### Scenario: Non-public URLs are rejected

- **WHEN** the posted URL resolves to a loopback, private, link-local, or unspecified address
- **THEN** the endpoint responds 400 and no fetch is performed

#### Scenario: Non-image responses are rejected

- **WHEN** the URL returns a non-image content type or the download fails
- **THEN** the endpoint responds 400 (non-image) or 502 (download failure) and no upload is created

### Requirement: Image generation degrades safely

Image generation SHALL be gated by the site-wide `ai_enabled` setting and SHALL surface provider failures as user-friendly messages without exposing raw API details.

#### Scenario: AI features disabled

- **WHEN** the `ai_enabled` app setting is off and any AI generation endpoint is called
- **THEN** the endpoint responds 403 `"AI features are disabled"`

#### Scenario: Provider failure is reported cleanly

- **WHEN** the image endpoint returns an error (content policy violation, rate limit, invalid request) or does not respond within 120 seconds
- **THEN** the system SHALL display a user-friendly error message and SHALL NOT expose raw provider details
