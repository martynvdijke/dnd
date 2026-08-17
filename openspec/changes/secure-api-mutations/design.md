## Context

Villum already has session and admin middleware and three public TRMNL reads. The access token currently mixed read integration with write authorization. The new contract separates public polling from protected application mutations.

## Goals / Non-Goals

### Goals

- Remove token requirements from the three read-only TRMNL endpoints.
- Protect every mutation with user identity plus bearer token and preserve ownership/admin checks.
- Make token creation and revocation available to users.

### Non-Goals

- Removing session, admin, or ownership authorization.
- Changing TRMNL response shapes.
- Committing a real API secret.

## Decisions

- Add a user-owned token model with hash, metadata, expiry, and revocation state.
- Validate bearer tokens after session authentication and before mutation handlers.
- Explicitly classify character/campaign/wiki/upload/companion/link/timeline/condition/transfer writes and default unknown non-GET routes to protection.
- Show secrets only at creation and provide revoke/rotate operations.

## Risks / Trade-offs

Public reads improve TRMNL setup but expose the configured read data by design. Existing API clients must gain a user session and token for writes.

## Migration Plan

1. Add token storage and lifecycle endpoints.
2. Update route middleware and remove auth from only the three read routes.
3. Add authorization and public-read tests.
4. Update the TRMNL configuration to remove its read token and provision write clients.

## Open Questions

- Should token scopes distinguish campaign management from character management?
