## Why

Every long campaign accumulates a gap between what the DM knows and what the *party* has actually been told. Today that gap lives in the DM's head or scattered across wiki pages and notes. There is no way to answer "what do the players know about the Ember Guild?", which leads to retconned reveals, forgotten rumors, and clues that never land. A first-class party-knowledge tracker makes information asymmetry manageable instead of memorized.

## What Changes

- Add a campaign-scoped knowledge entity: entries representing rumors, secrets, clues, and misinformation
- Status lifecycle per entry: `rumor → confirmed → revealed`, plus `false` for deliberate misinformation
- Source field (who/where it came from), rich content with existing mention support
- Cross-linking to NPCs, locations, quests, and other entities via the existing `entity_links` mechanism
- "Known by" links recording which PCs have been told each entry
- Visibility split: DM-only drafts vs entries shared with the party; shared entries appear in universal search for members, DM-only ones do not leak
- Campaign sub-view UI (list + detail, filterable by status) consistent with existing campaign panels

## Capabilities

### New Capabilities

- `party-knowledge`: Campaign knowledge entries with status lifecycle, source tracking, PC "known by" links, entity cross-links, DM/party visibility split, and campaign UI

### Modified Capabilities

- *(none)*

## Impact

- **Backend**: New ent schema + migration (`campaign_knowledge`, `campaign_knowledge_known_by`); CRUD handlers; registry onboarding in `registry/registry.go` including a new campaign-shared visibility rule; entity-link support
- **Frontend**: Campaign sub-view panel (HTMX partials per existing campaign patterns); status filters; known-by picker reusing the entity/character picker
- **Database**: Two new tables; no changes to existing tables
- **Search**: New entries appear in FTS5-backed universal search subject to the visibility rule
