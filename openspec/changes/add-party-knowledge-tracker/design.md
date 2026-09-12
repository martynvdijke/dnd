## Context

Villum's entity model is centralized in `registry/registry.go`: each type declares table, ownership rule (OwnerUser / OwnerCharacter / OwnerCampaign / OwnerUserOrCampaign / OwnerAdventure / OwnerGlobal), and participation in universal search, transfer, and cross-linking. Campaign-scoped content (wiki pages, timeline events, factions) uses ent schemas under `ent/schema/` with campaign membership visibility resolved by the registry's `campaignMemberSubquery`. Cross-entity links flow through `entity_links`; mentions/backlinks already parse rich text.

The one thing no existing entity models is *information state*: whether a fact is a rumor, verified truth, revealed to players, or a deliberate lie.

## Goals / Non-Goals

**Goals:**
- Answer "what does the party know about X?" in one view
- Track lifecycle from rumor to reveal without losing history of what changed
- Integrate with search/links/mentions so knowledge participates in the existing graph
- Keep DM secrets invisible to players everywhere — including search

**Non-Goals:**
- Per-player granular secrets ("only Thorin knows") beyond the known-by list — the shared flag is entry-level
- Automatic inference (e.g., auto-reveal when a clue is found) — status changes are manual
- Session-recap integration or AI summarization — future work
- Player-authored knowledge entries — players read; only DM-side roles write

## Decisions

### 1. New ent entities `campaign_knowledge` + `campaign_knowledge_known_by`
Fields: `campaign_id`, `title`, `content`, `status` (`rumor|confirmed|revealed|false`), `source`, `shared` (bool), timestamps. Known-by is a join table to characters.

**Why:** Matches how every campaign-scoped feature is modeled (ent schema + migration). Join table rather than CSV keeps "who knows what" queryable per character. **Alternative:** reuse wiki pages with a type field — rejected; wiki has different semantics (no lifecycle, member-editable) and overloading it muddies both.

### 2. Visibility via a new `OwnerCampaignShared` registry rule
Registry gains an ownership variant: visible to campaign owner always; visible to members only when `shared = true`. The knowledge entity registers with it, `Searchable: true`.

**Why:** The registry is the single visibility authority — extending it keeps search/transfer/linking correct without per-handler filtering. DM-only entries then cannot leak through universal search. **Alternative:** register as OwnerCampaign and exclude from search — rejected; shared clues deserve to be findable like any other entity, and handler-level filtering invites drift.

### 3. Status is a closed enum stored as TEXT
Four values, validated on write. Transitions are unconstrained (DMs jump around in practice), but every change appends to a lightweight `status_history` JSON column for the audit trail.

**Why:** Enum-in-TEXT matches existing conventions (no CHECK-migration churn); history column answers "when did we learn this?" without a third table.

### 4. Links ride the existing `entity_links` mechanism
Registering `Linkable: true` makes NPCs/locations/quests/etc. attachable using the standard picker; backlinks render automatically.

**Why:** Zero new linking code; knowledge entries immediately appear in the relationship graph.

### 5. UI as a campaign sub-view panel
List grouped/filterable by status, detail pane with mentions-enabled editor, known-by character picker, share toggle. HTMX partials following the factions/wiki panel pattern; DM sees all statuses, members see only `shared` entries.

**Why:** Consistent navigation; reuses mention parsing and pickers wholesale.

## Risks / Trade-offs

- **Visibility bug leaks a secret** → New ownership rule gets dedicated tests: member search/listing/detail must 404/hide unshared entries; covered before UI work lands.
- **Registry change touches hot paths** → The new rule is additive; existing constants untouched; full test suite gates.
- **Status enum too rigid** → Closed enum + free-text source/content covers real play; revisit only if a fifth state proves necessary.
- **Known-by maintenance burden** → Bulk "reveal to party" action marks all PCs known-by and flips shared in one click.

## Migration Plan

1. Land ent schemas + migration (two new tables, additive).
2. Registry rule + entity registration behind complete handler tests.
3. Handlers, routes, HTMX partials, pickers.
4. Rollback: revert commit; orphaned tables harmless.

## Open Questions

- Should `revealed` entries auto-flip `shared` to true? (Proposal: yes, as a convenience, overridable.)
- Icon choice for registry/UI chips (`fa-eye` vs `fa-lightbulb`).
