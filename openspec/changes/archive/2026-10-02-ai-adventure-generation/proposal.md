## Why

DMs can already generate free text with an AI endpoint, but there is no way to
turn that into usable campaign data, and the one-shot generation button only
fills a base adventure with acts and scenes. DMs want to describe an idea
conversationally, be asked clarifying questions, steer the result, and then
have the finished adventure — NPCs, locations, encounters and clues included —
created for them.

## What Changes

- Add a conversational, resumable AI drafting assistant for the DM-facing
  entity types (one-shot adventure, NPC, location, encounter, faction,
  campaign, quest, item). The assistant asks clarifying questions and offers
  suggestions, and only returns a structured draft once the DM approves.
- Persist each drafting session so the conversation can be reconnected to
  across modal open/close and page reloads, and send full history to the model
  each turn so it has real memory.
- Allow a draft to be committed with one click, creating the entity (and, for
  one-shots, their acts, scenes, NPCs, locations, encounters and clues).
- Add a "Draft with AI" entry point on the one-shot list and a chat modal.
- Fix AI endpoint handling: normalize a base URL that already contains an
  operation path (e.g. `/chat/completions`) so requests are not double-pathed,
  reject non-absolute base URLs, and make the admin "Test" result show the real
  provider response.
- Fix one-shot and campaign-overview views sticking on their loading ornament
  when revisited (htmx 2.x skips re-processing unchanged elements).

## Capabilities

### New Capabilities
- `ai-structured-drafting`: conversational, resumable, steerable AI generation
  that produces structured drafts and creates the corresponding entities.

### Modified Capabilities
- `ai-endpoint-management`: base URLs are normalized and validated on save, and
  request URLs are built through a single helper.

## Impact

- New table `ai_draft_sessions` (migration 064).
- New handlers `handlers/ai_draft.go`; new DM routes under `/api/ai/draft`.
- `handlers/ai.go` gains `generateChat`; `generateText` delegates to it.
- New frontend module `ts/ai-draft.ts`, modal markup in `static/app.html`, and
  a button in `handlers/templates/oneshot_list.html`.
- `ts/app.ts` one-shot/campaign-overview loading path.
- `ts/admin/integrations.ts` and `static/admin.html` AI endpoint form/test.
