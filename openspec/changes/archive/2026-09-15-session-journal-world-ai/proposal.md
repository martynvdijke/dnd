## Why

Villum has rich per-entity content, but no campaign-level view that ties places to the story timeline, and no good place to write up what happened in a session. Despite `README.md` advertising AI session recaps, `GenerateCampaignRecap` is template-only and never calls the configured model. The character journal/notes tabs also have a dead duplicate path and no deep-link load. This change closes those gaps.

## What Changes

- Add a top-level **World Overview** view: an interactive map of campaign places plus a directory grouped by type/parent, with place detail showing that place's timeline events.
- Make the existing polymorphic timeline link **usable for places**: a location selector on the timeline create/edit form, persisted through the HTMX path (the JSON API already stores `linked_entity_type`/`linked_entity_id`).
- Add a **Campaign Session Journal**: campaign-scoped, rich-text (TipTap) write-ups built on the existing `campaign_recaps` table, a real list/editor experience (today recaps are only listed on the dashboard), linkable to timeline events and places.
- Make **AI summaries real**: `GenerateCampaignRecap` calls the configured text endpoint when AI is enabled, falling back to the current template when it is not; add "Generate with AI" affordances to the journal/session/notes editors; index recaps so the Copilot can cite them.
- Harden the character **journal & notes** tabs: render the active section on open/deep-link, and delete the orphaned duplicate notes module now served by HTMX.
- Give every remaining icon-less menu an icon (character sheet tabs, compendium tabs including dynamic schema pills, party sub-tabs) using the existing Font Awesome pattern.

## Capabilities

### New Capabilities
- `world-overview`: campaign-scoped map + place directory, and place↔timeline-event linking and display.
- `campaign-session-journal`: rich-text campaign session write-ups (list, editor, dates, linkage to events/places).
- `ai-session-summaries`: AI-backed recap/summary generation with template fallback, editor generate actions, and Copilot retrieval of recaps.
- `character-journal-notes`: reliable character journal and notes tabs (persistence + correct render on open/deep-link).

### Modified Capabilities
<!-- No existing spec-level requirements change. All work is additive; the copilot change is expressed as a new requirement under ai-session-summaries. -->

## Impact

- **Frontend**: new `ts/world.ts` and session-journal module; edits to `ts/navigation.ts`, `ts/types.ts`, `ts/router.ts`, `ts/characters/sheet.ts`, `ts/ai.ts`, `ts/copilot.ts`, `ts/lib/tabs.ts`, `ts/party-subtabs.ts`; removal of orphaned `ts/app/notes.ts`.
- **Backend**: `handlers/recaps.go` (AI path), `handlers/htmx_campaign.go` (timeline linked-entity persistence, session-journal/recap routes), `handlers/copilot.go` + `handlers/search*` (index/retrieve recaps), new world/session templates.
- **Data**: no new columns required — `campaign_recaps` and `campaign_timeline_events.linked_entity_*` already exist; optional index on `(linked_entity_type, linked_entity_id)` for lookup speed.
- **APIs**: additive. Existing `/api/locations`, `/api/party`, `/api/timeline`, `/api/campaigns/:id/recaps` reused; new session-journal endpoints follow recap CRUD.
- **Dependencies**: none added; Leaflet and TipTap already present.
- **Tests**: Go handler tests, vitest unit tests, and Playwright e2e (world/timeline link, session journal save/reload, AI recap fallback).
