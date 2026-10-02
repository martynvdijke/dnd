## Why

The AI drafting assistant can fail silently on reasoning models. Draft calls
are hardcoded to 4000 max tokens; hidden reasoning consumes that budget, so the
model returns either an empty reply or JSON cut off mid-string. The app stores
the empty/partial reply without an error, the parser falls back to showing raw
JSON as a chat bubble, and no draft box or commit button appears. The endpoint's
own `max_tokens` setting is ignored for drafts, and the 60s request timeout is
too short for reasoning models. Separately, the UI has no way to import a draft
JSON directly: a DM who already has valid JSON (from the assistant or
elsewhere) can only re-paste it into the chat, and there is no way to keep
asking the AI to change parts of a draft after it has been imported.

## What Changes

- Resolve the draft token budget from the endpoint's `max_tokens` setting,
  clamped to a safe range, with a generous default instead of the hardcoded
  4000.
- Give draft requests their own longer timeout so reasoning models have time to
  finish.
- Detect empty replies and `finish_reason: "length"` truncation and return an
  actionable error instead of silently storing a broken assistant turn.
- Expand the assistant's output instructions into a full JSON contract
  ("skill") for one-shot drafts: field rules, allowed values, size limits, and
  guidance for adapting a named published adventure.
- Add a generic draft import: paste a JSON draft (bare object or the full
  `{"status","message","draft"}` assistant reply) for any AI-draftable entity
  type — one-shot, NPC, location, encounter, faction, campaign, quest, item —
  and create the entity with all linked content.
- Let an import replace an existing entity, so applying a revised draft updates
  the created entity instead of duplicating it.
- Add a stateless AI revision endpoint: given an entity type, the current draft
  JSON and an instruction, the assistant applies only the requested changes and
  returns the full revised object.
- Add an import/refine UI: a modal with an entity-type selector, a JSON
  textarea, and an "Ask AI" box. After a successful import the modal stays
  open; every AI revision is applied to the imported entity immediately, so a
  DM can keep asking the AI to edit parts of it.

## Capabilities

### New Capabilities
- `ai-draft-import`: import a pasted JSON draft into a new or existing entity
  of any AI-draftable type, plus the import/refine UI that applies AI revisions
  after import.

### Modified Capabilities
- `ai-structured-drafting`: draft requests get a usable token budget, a longer
  timeout, and explicit failure reporting instead of silent empty/truncated
  replies; the assistant receives a detailed one-shot output contract; and a
  stateless revision mode lets the assistant edit an existing draft JSON.

## Impact

- `handlers/ai.go`: per-call timeout for chat completions; empty-content
  detection.
- `handlers/ai_draft.go`: endpoint-aware token budget, draft timeout, failure
  surfacing, expanded one-shot system prompt, shared one-shot writer, update
  writers for every entity type.
- `handlers/ai_import.go` (new) + `handlers/routes_ai.go`: generic
  `POST /api/ai/import` (create or replace).
- `handlers/oneshot_json_import.go` (new) + `handlers/routes_oneshot.go`:
  `POST /api/oneshot-adventures/import` (one-shot convenience route).
- `handlers/ai_revise.go` (new): `POST /api/ai/revise`.
- `ts/draft-import.ts` (new), `ts/app.ts`, and templates
  (`oneshot_list.html`, `campaign_npcs_section.html`, `locations_list.html`,
  `campaign_encounters_section.html`, `factions_list.html`): import buttons and
  the import/refine modal.
- Tests: `handlers/ai_draft_test.go`, new `handlers/ai_import_test.go`, new
  `handlers/ai_revise_test.go`, new `ts/draft-import.test.ts`,
  `tests/oneshot.spec.ts`.
