## 1. Shared helpers

- [x] 1.1 Add `handlers/campaign_character_access.go` with the visibility join constant
      `campaignCharacterJoinSQL` and the four typed helpers: `campaignCharactersIn`,
      `isCampaignDMForCharacter`, `firstCampaignIDForCharacter`, `campaignIDsForCharacter`.
- [x] 1.2 Document each helper's contract (ordering, returned fields, error meaning) and the
      invariant the constant relies on (`characters` aliased as `c`, `cm.character_id = c.id`).

## 2. Roster consumers collapse onto `campaignCharactersIn`

- [x] 2.1 `handlers/dashboard.go` — `GetCampaignDashboard` character summaries.
- [x] 2.2 `handlers/oneshot_htmx_pacing.go` — `HtmxCampaignOverview` characters.
- [x] 2.3 `handlers/dm_screen.go` — party list with active conditions (keeps its per-character
      conditions query).
- [x] 2.4 `handlers/recaps.go` — `buildRecapTemplate` "Party Members" section.

## 3. Child-row queries use the join fragment

- [x] 3.1 `handlers/dashboard.go` — active quests, active conditions, recent journal, downtime
      activities, recent dice rolls.
- [x] 3.2 `handlers/oneshot_htmx_pacing.go` — same five counts/queries.
- [x] 3.3 `handlers/table_screen.go` — party characters (keeps its conditions subquery) and
      recent rolls.
- [x] 3.4 `handlers/email.go` — character table, recent sessions, active/available quests.
- [x] 3.5 `handlers/recaps.go` — completed quests, active quests, recent sessions, active
      conditions.
- [x] 3.6 `handlers/share.go` — shared party-member list.
- [x] 3.7 `handlers/knowledge.go` — party character ids for a knowledge entry.

## 4. Per-character lookups collapse onto typed helpers

- [x] 4.1 `handlers/htmx_character.go` and `handlers/notes.go` — DM gate for `dm`-visibility
      notes → `isCampaignDMForCharacter`.
- [x] 4.2 `handlers/combat_automation.go`, `handlers/hp_authority.go` (×2),
      `handlers/spells_cast.go` — first membership → `firstCampaignIDForCharacter`.
- [x] 4.3 `handlers/dice.go`, `handlers/ws.go` — all memberships → `campaignIDsForCharacter`.

## 5. Intentionally unchanged

- [x] 5.1 Leave `handlers/party.go:228` (pair membership guard), `handlers/transfer.go:416,478`
      (dynamic IN-list / dynamic where) and `handlers/transfer_import.go:325` (INSERT) — reasons
      recorded in `design.md` D5.
- [x] 5.2 Leave `search/campaign.go` — `handlers` imports `search`, so the helper is
      unreachable there; the file already uses one context-aware `queryIDs` helper.

## 6. Tests

- [x] 6.1 Add `handlers/campaign_character_access_test.go` covering: roster contents, ordering
      and stats; membership list; first-membership ordering and `sql.ErrNoRows`; the DM gate for
      DM, plain member, no-campaign character and a string character id; the join fragment's
      visibility semantics against a child table.
- [x] 6.2 Existing handler tests pass unchanged.

## 7. Registry and documentation

- [x] 7.1 `docs/tech-debt.md`: TD-006 → `resolved`, with the helper names, the converted file
      count and the intentionally-remaining joins as evidence; linked to this change.

## 8. Verification gates

- [x] 8.1 `go build ./...` passes.
- [x] 8.2 `go vet ./...` passes.
- [x] 8.3 `go test ./handlers/... ./search/...` passes.
- [x] 8.4 `bash scripts/ci/check-concatenated-sql.sh` passes (the fragment is a constant, no
      client input is spliced).
- [x] 8.5 `bash scripts/ci/check-detached-context.sh` passes (no new detached contexts).
- [x] 8.6 Conventional commits scoped to `handlers/`, `docs/tech-debt.md` and this change
      folder; no push.
