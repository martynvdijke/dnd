## Context

`campaign_characters(campaign_id, character_id)` is the join table that makes a character a
member of a campaign (created in `db/character_campaign_memberships.go`, unique on the pair).
"Campaign-scoped character read" means any query whose result must be limited to the characters
of one campaign: the dashboard counts and roster, the DM screen party list, the table screen,
the recap template builder, the highlight email, the note visibility gate, the dice/WebSocket
fan-out and several per-character lookups.

TD-006 recorded that those reads spelled the membership out inline. The 2026-11-27 measurement
before this change: 35 `campaign_characters` joins in 18 non-test handler files. Three shapes
accounted for almost all of them:

1. **Roster** — `SELECT c.<fields> FROM characters c JOIN campaign_characters cm ON
   cm.character_id=c.id WHERE cm.campaign_id=? ORDER BY c.name`, repeated with four different
   column lists (dashboard, oneshot overview, DM screen, recap builder) and once with a
   per-character conditions subquery (table screen).
2. **Child rows in campaign** — `FROM <child> x JOIN characters c ON c.id=x.character_id JOIN
   campaign_characters cm ON cm.character_id=c.id WHERE cm.campaign_id=?` plus a per-query
   filter/order/limit, repeated for quests, sessions, journal, conditions, downtime activities
   and dice rolls.
3. **Per-character lookups** — `SELECT campaign_id FROM campaign_characters WHERE
   character_id=? ORDER BY campaign_id LIMIT 1` (first membership), `... WHERE character_id=?`
   (all memberships), and the `campaign_members` DM check.

Stakeholders: backend contributors/reviewers; the `handlers/` coverage gate (≥40%); anyone
changing campaign visibility, who currently must find every copy.

## Goals / Non-Goals

**Goals:**
- One definition of the campaign-visibility predicate for handler reads.
- One query for the campaign roster, shared by every consumer.
- Preserve observable behaviour exactly — same rows, same order, same fields, same error paths.
- Close TD-006 in the registry with refreshed, verifiable evidence.

**Non-Goals:**
- Converting the reads to ent traversal. The repo's accepted architecture is "ent for entity
  CRUD, parameterized raw SQL for projections" (TD-001, `docs/adr-data-access.md`); this change
  deduplicates the raw SQL, it does not migrate the access path.
- Converting `db.DB.Query/QueryRow` to `*Context` variants — that is TD-008.
- Touching the three unique join sites, or `search/campaign.go`.

## Decisions

### D1: A constant SQL fragment for the visibility join, executable helpers for the shapes that return data

Two mechanisms, chosen by what actually repeats:

- **`const campaignCharacterJoinSQL = "JOIN campaign_characters cm ON cm.character_id = c.id"`**
  for child-row queries that keep their own projection, filters, ordering and limit. Every such
  query already joins `characters` aliased as `c`, so the fragment composes by string
  concatenation with a compile-time constant (explicitly allowed by
  `scripts/ci/check-concatenated-sql.sh`) and the resulting SQL is textually identical to what
  the handler had inline.
- **Executable helpers** (`campaignCharactersIn`, `isCampaignDMForCharacter`,
  `firstCampaignIDForCharacter`, `campaignIDsForCharacter`) where the repeated thing is a whole
  query whose rows the caller would otherwise scan by hand. These remove both the SQL and the
  `rows.Next()`/`Scan` boilerplate.

Rationale: a fully generic executable helper ("count child rows for a table with a where
clause") would take table names and predicates as strings — a stringly-typed API with no more
safety than the fragment, and much worse to read. A fragment plus four typed helpers is the
minimum that removes the repetition without inventing a query builder.

Alternatives rejected: ent traversal (out of scope, contradicts the accepted hybrid, and the
roster projection has no ent equivalent that preserves the exact ordering/columns); a single
generic `campaignCharacterQuery(select, where)` builder (over-abstraction, stringly typed).

### D2: The child-table join keeps its `JOIN characters` step

The child-row queries could be shortened to
`WHERE character_id IN (SELECT character_id FROM campaign_characters WHERE campaign_id=?)`,
which needs no alias. That is *not* equivalent: the existing `JOIN characters c ON
c.id = x.character_id` also drops child rows whose owning character no longer exists. The
fragment therefore only replaces the `campaign_characters` half of the join, and every
converted query keeps its `characters` join verbatim.

### D3: `campaignCharactersIn` returns a superset row

The four roster consumers read different column subsets: dashboard/oneshot overview (id, name,
race, class, level, hp_current, hp_max, portrait_url), DM screen (id, name, hp_current, hp_max,
ac), recap builder (id, name, race, class). One struct with all nine fields, selected once,
serves all of them; callers read the fields they need. All the columns are `NOT NULL` with
defaults in `ent/schema/character.go`, so the wider projection cannot produce a scan failure
the narrower ones did not.

The table screen is the exception: it also computes a `GROUP_CONCAT` of active conditions per
character, so it keeps its own `SELECT` and uses only the join fragment.

Behaviour notes:
- Ordering is `ORDER BY c.name`, which the oneshot overview already used as `ORDER BY name`;
  the two forms are the same column.
- Scan errors still leave the row's remaining fields zeroed, and the row is still appended —
  matching the previous inline loops, which ignored `Scan` errors.
- Query errors are returned to the caller, which logs the same per-handler warning as before.
  The only log-only difference is that a `rows.Err()` iteration failure now surfaces through the
  same "query failed" warning instead of a separate "rows iteration failed" line; no HTTP
  response changes.

### D4: `isCampaignDMForCharacter` takes `any` parameters

The two call sites (`HtmxListNotes`, `ListCharacterNotes`) hold the character id as the raw
query-string value and the user id as the gin context's `any`. The inline query bound those
values as-is and SQLite coerced them against the integer columns. Keeping the helper's
parameters as `any` preserves that binding exactly; typing them `int64` would force a
`strconv`/type-assertion conversion at the call sites and change the behaviour for non-numeric
input. The doc comment states the intent, and the helper's test pins the string-id case.

### D5: Sites left alone

| Site | Why |
|---|---|
| `handlers/party.go:228` | `WHERE campaign_id=? AND character_id=?` membership guard for a transfer; a different predicate (pair membership, not "characters of a campaign") and a single use |
| `handlers/transfer.go:416` | dynamic `IN (...)` placeholder list built per request |
| `handlers/transfer.go:478` | dynamic `WHERE` clause over the membership table (export/import dump) |
| `handlers/transfer_import.go:325` | `INSERT OR IGNORE` into the join table, not a read |
| `search/campaign.go:42,52-55` | `handlers` imports `search`, so a handler helper is unavailable there; the file already funnels its reads through one context-aware `queryIDs` helper |

## Risks / Trade-offs

- **String-concatenated SQL** — the fragment is a compile-time constant, but reviewers must not
  extend it with variables. Mitigated by the doc comment on the constant, the constant's own
  test (`TestCampaignCharacterJoinSQL`), and `check-concatenated-sql.sh`, which still fails on
  any client value spliced into SQL.
- **Wider roster projection** — reads two extra `NOT NULL` columns per character; negligible on
  SQLite for party-sized result sets.
- **`any` parameters on the DM gate** — loses compile-time typing at two call sites to keep
  binding behaviour identical; considered a fair trade for a pure refactor.
- **Fragment drift** — if a future query needs a different join (e.g. `campaign_members`
  visibility rather than character membership), it must not reuse this constant. The name says
  "character membership".

## Migration Plan

Single PR, no data migration. Rollback is a revert: every converted site previously carried the
full inline query, and no schema or API surface moved.

## Open Questions

<!-- none -->
