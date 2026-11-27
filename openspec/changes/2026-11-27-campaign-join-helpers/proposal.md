## Why

TD-006 in `docs/tech-debt.md` tracked campaign-scoped character reads that spelled the
`campaign_characters` membership out inline. The 2026-11-27 re-measure found 35 joins across
18 non-test handler files: every one of them re-derived "which characters belong to this
campaign", and several files carried the same query verbatim.

The repetition was not harmless. The predicate is a visibility rule — if it drifts, a character
from another campaign leaks into a dashboard, a recap, an email highlight or the table screen.
Six of the repeated shapes also encoded the same lookups ("the character's campaign", "is this
user a DM of the character's campaign", "every campaign the character belongs to") with
slightly different SQL each time, so a schema change had to be applied by hand in a dozen
places.

## What Changes

- **New `handlers/campaign_character_access.go`** holding the shared shapes:
  - `campaignCharacterJoinSQL` — the one definition of the `characters c` →
    `campaign_characters cm` visibility join (a compile-time constant SQL fragment, appended
    to child-table queries that already join `characters`).
  - `campaignCharactersIn(db, campaignID)` — the party roster (id, name, race, class, level,
    hp, ac, portrait), ordered by name.
  - `isCampaignDMForCharacter(db, characterID, userID)` — the "dm" note-visibility gate.
  - `firstCampaignIDForCharacter(db, characterID)` — the lowest membership, used for
    bookkeeping when a character is in several campaigns.
  - `campaignIDsForCharacter(db, characterID)` — every campaign the character belongs to.
- **32 joins across 15 handlers collapse** onto those helpers: `combat_automation`,
  `dashboard`, `dice`, `dm_screen`, `email`, `hp_authority`, `htmx_character`, `knowledge`,
  `notes`, `oneshot_htmx_pacing`, `recaps`, `share`, `spells_cast`, `table_screen`, `ws`.
- **Three joins stay as they are**, because their semantics are unique: the
  `campaign_id AND character_id` guard in `party.go`, the dynamic IN-list / dynamic-where
  queries in `transfer.go`, and the `INSERT` in `transfer_import.go`.
- **`search/campaign.go` is not converted**: it hand-writes the same join, but `handlers`
  imports `search`, so a helper in `handlers` cannot be used there. It already centralises its
  own reads behind a local `queryIDs` helper and uses context-aware queries.
- **No behaviour change**: identical SQL semantics (same visibility rule, same joins — including
  the `characters` join that excludes orphaned child rows — same ordering, same returned
  fields, same error handling, same `db.DB` non-context calls).
- **Registry**: `docs/tech-debt.md` marks TD-006 resolved with the refreshed evidence.

## Capabilities

### New Capabilities
<!-- none: this is a refactor with no requirement change -->

### Modified Capabilities
<!-- none: every endpoint keeps its contract; the visibility rule is unchanged -->

## Impact

- **Code**: `handlers/campaign_character_access.go` (new), the 15 handlers listed above,
  `handlers/campaign_character_access_test.go` (new).
- **APIs**: no contract change. Dashboard, DM screen, table screen, recap template, highlight
  email, dice and WebSocket fan-out all return/emit what they did before.
- **Tests/CI**: `go build ./...`, `go vet ./...`, `go test ./handlers/... ./search/...`,
  `scripts/ci/check-concatenated-sql.sh`, `scripts/ci/check-detached-context.sh`. The new
  helpers are pinned by `handlers/campaign_character_access_test.go`; no existing test changed.
