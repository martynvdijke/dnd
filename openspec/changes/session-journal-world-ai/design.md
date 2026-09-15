## Context

The pieces already exist but are unconnected:

- **Places**: `locations` is user-scoped (`ent/schema/location.go`), with a self-referential `parent_id` and a `type` string. Characters attach to places via `character_locations`. The character sheet already renders them with Leaflet (`ts/characters/locations.ts`), and the party view aggregates them (`ts/party-subtabs.ts:renderPartyLocations`).
- **Timeline**: `campaign_timeline_events` already carries polymorphic `linked_entity_type` + `linked_entity_id` (`ent/schema/campaigntimelineevent.go:14`) accepted by the JSON API (`handlers/timeline.go`), but the HTMX form/handlers ignore them, so the link is unreachable in the UI.
- **Recaps**: `campaign_recaps` (`ent/schema/campaignrecap.go`) stores title/content/session dates/word_count/is_edited/is_sent and is exposed by `handlers/recaps.go`. `GenerateCampaignRecap` is a pure text template — no LLM call.
- **AI**: a working OpenAI-compatible client (`handlers/ai.go:generateText`), encrypted endpoints, admin config, a generic `.ai-generate-btn` modal (`ts/ai.ts`), and Copilot RAG over `entity_search_index` (`handlers/copilot.go`). AI can be disabled site-wide via `app_settings.ai_enabled`.
- **Journal/notes**: journal works via REST+TipTap; notes works via HTMX; `ts/app/notes.ts` is a dead duplicate (zero callers).

## Goals / Non-Goals

**Goals:**
- A discoverable World Overview: map + place directory, with each place showing its timeline events.
- Make the existing place↔timeline link settable and visible.
- A real campaign Session Journal built on `campaign_recaps` with a rich-text editor.
- Recaps that use the configured model when available, with a template fallback.
- Journal/notes tabs that always load their content, on click and on deep-link.
- Consistent icons on every menu.

**Non-Goals:**
- New world/region/continent hierarchy tables (reuse `parent_id` + `type`).
- Real-time collaboration, streaming AI, or multi-user editing.
- Re-scoping `locations` to campaigns (keep user-scoped + character links).
- A new rich-text library (reuse the existing TipTap setup).

## Decisions

1. **World Overview is a top-level view** (`world`) alongside Timeline/Party, added to the top nav, sidebar, `showMoreNav()`, `ViewState`, and the hash router. *Alternative:* a Party sub-tab — rejected because a world map + cross-entity timeline is world-level, not party-roster-level.

2. **World data via existing endpoints, client-side join.** The view fetches `/api/party`, per-member `/api/characters/:id/locations`, and `/api/timeline?campaign_id=`, then groups places by `type`/`parent_id` and matches events by `linked_entity_id`. *Alternative:* a new `/api/world` endpoint — deferred; the party view already uses this N-request fan-out and campaigns are small. Mark it as a known ceiling: if a campaign has many members, add a server aggregate endpoint.

3. **Link places to events with the existing polymorphic columns**, `linked_entity_type='location'` + `linked_entity_id`, not a new FK. Add an index on the pair for lookup. The HTMX timeline create/update handlers are extended to persist these (the JSON API already does), and `timeline_form.html` gains a location `<select>` populated from `/api/locations`. *Alternative:* a dedicated `location_id` column — rejected as redundant.

4. **Session Journal uses `campaign_recaps` + existing TipTap editor.** Recaps already have title/content/session dates/word_count/is_edited/is_sent and CRUD routes, so no new content table. A new `sessions` view hosts the list + editor. Write-ups link to places/events through the existing `entity_links` mechanism (recaps registered as a linkable entity type); if that is impractical, fall back to an `entity_links`-style join table. Manual edits set `is_edited`.

5. **AI recap with graceful fallback.** `POST /api/campaigns/:id/recaps/generate` assembles the current context sections as an LLM prompt and returns the model output when AI is enabled and a text endpoint exists; otherwise it returns the existing template output. It remains an unsaved draft the DM reviews before `CreateCampaignRecap` persists it. *Alternative:* hard-fail without AI — rejected; the template is a useful offline path and preserves current tests.

6. **Reuse `.ai-generate-btn`** for editor-level summaries (journal/notes/session). It is already delegated by `ts/ai.ts:97`, so new buttons need only `data-ai-mode="text"`, `data-ai-target`, and `data-ai-hint` — no new JS.

7. **Copilot indexes recaps** by adding `recap` to `retrieveCampaignContext`'s campaign set and ensuring recap rows reach `entity_search_index`, so generated write-ups are citable. Expressed as an additive requirement under `ai-session-summaries`.

8. **Journal/notes hardening**: `renderSheet()`/`switchTab` renders the active section on initial open and on hash deep-link (calling the same render/HTMX path, guarded against double-load). Delete `ts/app/notes.ts` and its import; update any test that referenced `renderNotes`.

## Risks / Trade-offs

- **N-request fan-out in World Overview** → acceptable for small campaigns; documented ceiling with a server aggregate endpoint as the upgrade path.
- **`entity_links` may not accept `recap`** → verify registry supports the type first; fall back to a dedicated join table if not.
- **AI recap latency/cost (60s timeout, no streaming)** → template fallback, clear errors via existing `aiGenError`, button spinner/disable; no partial writes since drafts are unsaved.
- **Removing `ts/app/notes.ts` breaks references** → grep `renderNotes`/notes module imports and update tests before deletion.
- **Deep-link rendering may double-load HTMX tabs** → guard with a loaded-section marker.
- **Copilot recap indexing needs FTS triggers** → verify `campaign_recaps` triggers exist; add if missing, else recaps won't be searchable and the requirement fails.

## Migration Plan

Additive only: no data migration. Optional migration adds an index on `campaign_timeline_events(linked_entity_type, linked_entity_id)` and `entity_search_index` triggers for recaps. Deploy is a normal build; rollback is reverting the PR (index/trigger drops are optional and harmless to leave).

## Open Questions

- World place scope: party-linked places only, or also the DM's unlinked user-owned locations? (Leaning party-linked to avoid showing unused prep locations.)
- Should AI recap generation persist immediately as a draft recap row, or always return an unsaved draft for review? (Leaning unsaved draft.)
- Should the session journal support multiple entries per session, or one recap per date range? (Leaning one entry per session, matching the existing schema.)
