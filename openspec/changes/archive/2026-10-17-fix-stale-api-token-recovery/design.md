# Design

## Decisions

- **Recover in the client, not the server.** The server already distinguishes
  "missing token" (401, len 30) from "invalid token" (401, len 29). The client is
  the only place that knows which token it sent, so it owns recovery. No new
  server surface is needed as a result.
- **One retry, mutation-only.** Recovery is limited to non-`GET`/`HEAD`/`OPTIONS`
  requests that carried a token. This bounds the work, avoids loops, and matches
  the middleware's exemption of reads.
- **A registered handler, not a hard dependency.** `api()` stays free of
  bootstrap/localStorage concerns. `setUnauthorizedHandler` (main and admin
  clients) lets each shell wire its own re-provisioning routine; tests can stub
  it.
- **Create, don't rotate, on bootstrap.** Rotating an existing token replaced a
  secret another device already held. Creating a fresh per-device token keeps
  provisioning free of cross-device side effects. The rotate endpoint remains
  available for deliberate manual use.
- **Drop the admin `'admin'` key fallback.** It could load another user's token
  from `localStorage`; the username is always known at bootstrap time.

## Risks

- Token accumulation: each device/storage-clear creates a token. Acceptable —
  tokens are inspectable and revocable in the admin UI, and the alternative
  (rotation) caused cross-device breakage.
- A recovery that still 401s surfaces the original error after one retry, so
  behaviour degrades to the previous state rather than looping.

## Testing

- Unit tests for `ts/lib/api.ts` and `ts/admin/state.ts` cover: retry-and-succeed,
  no retry on GET, no retry without a token, and error surfaced when recovery
  does not help.
