## ADDED Requirements

### Requirement: TRMNL reads are public

The service SHALL allow unauthenticated GET access to character stats, campaign stats, and character roster endpoints used by the TRMNL plugin.

#### Scenario: TRMNL polling

- WHEN TRMNL requests the three read endpoints without credentials
- THEN Villum returns the corresponding read payloads

### Requirement: Application writes require user and token

Character, campaign, wiki, upload, companion, link, timeline, condition, transfer, and every create/update/delete route MUST require an authenticated user and a valid bearer API token, while retaining admin/ownership checks.

#### Scenario: Token-only campaign mutation

- WHEN a token is presented without a user session to a campaign mutation
- THEN the service rejects the request and changes nothing

### Requirement: Token management is secure

Users SHALL create, list metadata for, revoke, and rotate their own tokens. A token secret MUST be shown once, stored hashed, and never logged or returned again.

#### Scenario: Cross-user token

- WHEN a user uses another user's token on a write
- THEN the request is rejected without revealing token ownership
