## Why

Villum's TRMNL plugin currently needs an access token even for read-only stats. Read polling should be public, while campaign and application mutations must require both a signed-in user and an API token.

## What Changes

- Keep `/api/trmnl/character-stats`, `/api/trmnl/campaign-stats`, and `/api/trmnl/characters` public reads.
- Require an authenticated user plus bearer API token for character, campaign, wiki, upload, companion, link, timeline, condition, transfer, and all other writes.
- Add owner-scoped one-time-visible token creation and lifecycle management.

## Capabilities

### New Capabilities

- `api-authentication`

### Modified Capabilities

- `public-api-reads`

## Impact

Villum middleware, API route groups, token persistence, migrations, UI/API token management, tests, and the TRMNL plugin configuration are affected. Existing session/admin authorization remains authoritative.
