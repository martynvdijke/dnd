# public-recap-sharing Specification

## Purpose
TBD - created by archiving change share-public-recaps. Update Purpose after archive.
## Requirements
### Requirement: Campaign recaps can be shared publicly

A campaign member SHALL be able to create a revocable public share link for a campaign recap. The link SHALL resolve to the recap without authentication, both as JSON and as a standalone HTML page.

#### Scenario: Member creates a recap share link

- **WHEN** a campaign member POSTs `/api/share` with `entity_type=recap` and a recap id belonging to their campaign
- **THEN** the server responds 201 with a token and URL

#### Scenario: Non-member cannot share a recap

- **WHEN** a user who is not a member of the recap's campaign POSTs `/api/share` with `entity_type=recap`
- **THEN** the server responds 403

#### Scenario: Public JSON access

- **WHEN** anyone GETs `/api/share/:token` for a valid recap token
- **THEN** the server responds 200 with the recap title, content, session dates, word count, and campaign name

#### Scenario: Public HTML page

- **WHEN** anyone GETs `/share/:token` for a valid recap token
- **THEN** the server responds 200 with a noindex HTML page containing the recap title and content

#### Scenario: Expired or revoked link

- **WHEN** a recap share link has expired or been deleted
- **THEN** the public routes respond 410 (expired) or 404 (unknown/revoked)
