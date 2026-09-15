## 1. Menu icons (landed)

- [x] 1.1 Add per-section icons to the character sheet tabs (`ts/lib/tabs.ts`, `ts/characters/sheet.ts`)
- [x] 1.2 Add icons to the compendium tabs, including dynamic schema pills (`static/app.html`, `ts/compendium.ts`)
- [x] 1.3 Add icons to the party sub-tabs (`ts/party-subtabs.ts`)

## 2. Journal & notes hardening

- [x] 2.1 Render the active sheet section on open and on hash deep-link; guard against double-load (`ts/characters/sheet.ts`)
- [x] 2.2 Delete the orphaned duplicate notes module `ts/app/notes.ts` and its import; update any references
- [x] 2.3 Add a unit test for deep-link/initial-section rendering and an e2e create→reload persistence check for journal and notes

## 3. World Overview

- [x] 3.1 Add `world` to `ViewState`, the navigation view list, and the hash router (`ts/types.ts`, `ts/navigation.ts`, `ts/router.ts`)
- [x] 3.2 Add World nav entries to the top nav, sidebar, and mobile "more" menu (`static/app.html`, `ts/navigation.ts`)
- [x] 3.3 Add the `#worldView` container and layout (`static/app.html`)
- [x] 3.4 Implement `ts/world.ts`: fetch party members, per-member locations, and campaign timeline; build the place directory grouped by `type` and `parent_id`
- [x] 3.5 Render the Leaflet map with one marker per located place and keep map/directory selection in sync
- [x] 3.6 Place detail: list linked timeline events and link each event back to the place
- [x] 3.7 Add a place `<select>` to `handlers/templates/timeline_form.html`, populated from `/api/locations`
- [x] 3.8 Persist `linked_entity_type`/`linked_entity_id` in `HtmxCreateTimeline`/`HtmxUpdateTimeline` (`handlers/htmx_campaign.go`)
- [x] 3.9 Add a migration indexing `campaign_timeline_events(linked_entity_type, linked_entity_id)`
- [x] 3.10 e2e: link an event to a place, see it on the place, and jump from the event to the world view

## 4. Campaign Session Journal

- [x] 4.1 Add a `sessions` view + navigation entry and a list renderer for campaign write-ups
- [x] 4.2 Build the write-up editor on the existing TipTap setup, wired to recap create/update/delete; compute `word_count` and set `is_edited` on manual save
- [x] 4.3 Support date or date range on a write-up
- [x] 4.4 Link write-ups to places and timeline events (register `recap` as linkable in `entity_links`, or add a join table if not supported) with backlinks
- [x] 4.5 Enforce access control: DM edits/deletes, members read published write-ups, non-members denied
- [x] 4.6 e2e: create/edit/reload a write-up and verify formatting, word count, and links

## 5. AI summaries

- [x] 5.1 Add the AI path to `GenerateCampaignRecap` using the first enabled text endpoint, with the template as fallback and an `ai_used` indicator (`handlers/recaps.go`)
- [x] 5.2 Return clear provider errors without persisting partial state; keep generation DM-only
- [x] 5.3 Add `.ai-generate-btn` actions to the journal, notes, and session write-up editors (`data-ai-mode/target/hint`; no new JS)
- [x] 5.4 Index campaign recaps into `entity_search_index` and add `recap` to `retrieveCampaignContext` so the Copilot can cite them (`handlers/copilot.go`, search/migration layer)
- [x] 5.5 Go tests: AI-generated recap, template fallback when AI disabled/no endpoint, non-owner denied, provider failure

## 6. Verification and delivery

- [x] 6.1 `npm run typecheck`
- [x] 6.2 `npm run test:unit`
- [x] 6.3 `task test:e2e`
- [x] 6.4 Run the full local gates (`task ci` or the pre-push hook) and fix any failures
- [x] 6.5 Open the PR per `AGENTS.md`, watch checks, and merge only when green
