## Why

Campaign recaps are the natural artifact to share publicly — players want to send a session write-up to absent players, or post it in a group chat. Today recaps are locked behind campaign membership: `supportedShareTypes` in `handlers/share.go` excludes `recap`, so `POST /api/share` returns 400 and there is no public route. The share-link infrastructure (tokens, expiry, revocation, public JSON + HTML pages) already exists for notes, journals, maps, and uploads; recaps just need to join it.

## What Changes

- Add `recap` to the share-link allowlist and ownership resolution (`canShareEntity` → campaign membership).
- Add a public JSON case (`GetSharedEntity`) and a server-rendered HTML page (`GetSharedPage` + `share_recap.html`) for shared recaps.
- Add a share button to each recap card in the Sessions view (`ts/campaign/recaps.ts`) that creates the link and shows the copy/email modal.
- Add `recap` to the frontend share type labels.

## Capabilities

### New Capabilities

- `public-recap-sharing`: campaign recaps can be published as revocable, optionally-expiring public links rendered as a standalone noindex page.

### Modified Capabilities

- `public-sharing`: the share-link allowlist and public entity/page dispatch gain the `recap` entity type.

## Impact

- `handlers/share.go` (allowlist, `canShareEntity`, `shareEntityLabel`, `GetSharedEntity`, `GetSharedPage`, `SharedRecapView`)
- `handlers/templates/share_recap.html` (new)
- `ts/campaign/recaps.ts` (share button + `shareRecap`), `ts/share.ts` (label)
- Tests: `handlers/share_recap_test.go` (new), `tests/recaps.spec.ts`
