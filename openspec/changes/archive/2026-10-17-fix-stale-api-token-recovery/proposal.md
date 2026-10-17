# fix: stale API token recovery

## Why

The web app provisions a per-user API token for mutating requests. `GET`s are
exempt from the API-token check, so a token that has been rotated or revoked
elsewhere still lets every read succeed, while every mutation returns
`401 {"error":"invalid API token"}`. The bootstrap trusted any token found in
`localStorage` and never re-validated it, so the browser stayed wedged until the
user cleared storage or reloaded with a fresh bootstrap. This was observed live:
admin AI/endpoint saves failed with 401 while `/api/user/me` and other GETs
returned 200.

A secondary footgun made it worse: when no token was stored, the bootstrap
*rotated* an existing token to obtain a secret, which invalidated the secret
already stored on every other device — wedging them in turn.

## What Changes

- The shared `api()` client (main app and admin shell) detects a `401` on a
  mutating request that carried a token, invokes a registered recovery handler,
  and retries the request **once**.
- The bootstrap registers that handler: it drops the stored secret and
  provisions a fresh token, then `api()` retries transparently.
- Provisioning now always **creates a new per-device token** instead of rotating
  an existing one, so bootstrapping one device can no longer invalidate another.
- The admin bootstrap no longer falls back to the literal `'admin'` storage key,
  which could load another account's token.

## Capabilities

### New Capabilities
- `api-token-recovery`: transparent recovery from an invalidated API token on
  the first mutating request.

## Impact

- Frontend only; no server or schema changes.
- Files: `ts/lib/api.ts`, `ts/init.ts`, `ts/admin/state.ts`, `ts/admin.ts`, plus
  unit tests.
