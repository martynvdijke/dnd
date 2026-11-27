# ai-endpoint-management Specification

## Purpose
TBD - created by archiving change ai-adventure-generation. Update Purpose after archive.
## Requirements
### Requirement: AI endpoint request URL construction

The system SHALL treat the configured base URL as a base and build operation
URLs (`/chat/completions`, `/images/generations`) from it. If a saved base URL
already includes a trailing operation path, the system SHALL strip it before
storing and before requesting, so the operation path is never duplicated.

#### Scenario: Base URL already contains the operation path
- **WHEN** a DM saves an endpoint whose base URL ends with `/chat/completions`
- **THEN** the stored base URL has that suffix removed and a test/generation request targets `<base>/chat/completions`

#### Scenario: Invalid base URL
- **WHEN** a DM saves an endpoint whose base URL is not an absolute http(s) URL
- **THEN** the system rejects the request and no endpoint is stored

### Requirement: AI endpoint test feedback

The admin AI endpoint test SHALL show the provider's actual response on
failure, so a failure can be distinguished from an unknown error.

#### Scenario: Test returns a provider error
- **WHEN** the endpoint test fails
- **THEN** the UI shows the provider's message (e.g. the HTTP status and body)

### Requirement: Admin can view all AI endpoints

The system SHALL display a list of configured AI endpoints in the admin panel "AI Endpoints" tab.

#### Scenario: View endpoint list

- **WHEN** an admin navigates to the admin panel and clicks the "AI Endpoints" tab
- **THEN** the system SHALL display a table of all configured endpoints showing name, type (text/image), model, base URL, and enabled status

#### Scenario: Empty endpoint list

- **WHEN** no AI endpoints have been configured
- **THEN** the system SHALL display a message "No AI endpoints configured" with a button to add one

### Requirement: Admin can create an AI endpoint

The system SHALL allow admins to add a new AI endpoint configuration with the following fields:
- Name (required, unique, string)
- Type (required, enum: "text" | "image")
- Base URL (required, string, the API endpoint URL)
- API Key (required, string, stored encrypted at rest)
- Model (required, string, the model identifier)
- Temperature (optional, float, 0.0–2.0, default: 0.8, for text endpoints only)
- Max Tokens (optional, integer, 1–16384, default: 1000, for text endpoints only)
- Image Size (optional, string, default: "1024x1024", for image endpoints only)
- Enabled (boolean, default: true)

#### Scenario: Successful endpoint creation

- **WHEN** an admin fills in all required fields and clicks "Save"
- **THEN** the system SHALL create the endpoint, encrypt the API key, store it in the database, and display it in the endpoint list

#### Scenario: Missing required fields

- **WHEN** an admin submits the form without a required field (name, base URL, API key, or model)
- **THEN** the system SHALL display an error message indicating the missing required field

#### Scenario: Duplicate endpoint name

- **WHEN** an admin creates an endpoint with a name that already exists
- **THEN** the system SHALL return an error and not create the duplicate

### Requirement: Admin can edit an AI endpoint

The system SHALL allow admins to modify an existing AI endpoint's configuration.

#### Scenario: Edit endpoint fields

- **WHEN** an admin clicks "Edit" on an endpoint and modifies its fields
- **THEN** the system SHALL update the endpoint configuration and re-encrypt the API key if a new one was provided

#### Scenario: Edit without changing API key

- **WHEN** an admin edits an endpoint but leaves the API key field blank
- **THEN** the system SHALL preserve the existing encrypted API key

### Requirement: Admin can delete an AI endpoint

The system SHALL allow admins to remove an AI endpoint configuration.

#### Scenario: Delete endpoint

- **WHEN** an admin clicks "Delete" on an endpoint and confirms
- **THEN** the system SHALL remove the endpoint from the database, and it SHALL no longer appear in the endpoint list or be available for generation

### Requirement: Admin can test an AI endpoint connection

The system SHALL allow admins to test an AI endpoint by sending a minimal request to verify connectivity and credentials.

#### Scenario: Successful test

- **WHEN** an admin clicks "Test" on an endpoint
- **THEN** the system SHALL send a minimal request to the endpoint and display a success message if the endpoint responds correctly

#### Scenario: Failed test

- **WHEN** an admin clicks "Test" on a misconfigured or unreachable endpoint
- **THEN** the system SHALL display an error message with the failure reason (connection refused, authentication failed, etc.)
