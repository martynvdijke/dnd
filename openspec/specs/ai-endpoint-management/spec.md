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
