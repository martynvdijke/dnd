## Context

The conversational drafting feature shipped in PR #169 (`ai-adventure-generation`)
but was never archived. Production testing against a reasoning model
(`deepseek-v4.1-flash` via an OpenAI-compatible endpoint) showed three failure
modes: the hardcoded 4000-token draft budget is consumed by hidden reasoning,
producing an empty reply; a larger budget still truncates mid-JSON because the
full one-shot draft needs more room; and the 60s text timeout is shorter than
the 79s a 16k-token reasoning draft took. In all three cases the handler stored
the assistant reply without complaint, `parseAIDraftReply` fell back to a raw
chat bubble, and the UI offered no draft box, no commit, and no way to import
the JSON manually.

The endpoint row already has a `max_tokens` column (currently set to an absurd
400000000 by the admin UI) that draft calls ignore. The UI has no import path at
all, and no way to keep working with the AI on a draft once it has been
imported.

## Goals / Non-Goals

**Goals:**
- Draft calls get a token budget large enough for a full one-shot and a timeout
  that fits reasoning models.
- Empty or truncated model replies surface as actionable errors, never as a
  silently stored empty assistant turn.
- The assistant receives a detailed output contract for one-shot drafts so
  replies fit the budget and map cleanly onto the schema.
- DMs can paste a draft JSON (bare or the assistant envelope) for any
  AI-draftable entity type and create or replace that entity with all linked
  content.
- After importing, the DM can keep asking the AI to change specific parts and
  see those changes applied to the created entity.

**Non-Goals:**
- Streaming replies or provider-specific reasoning options.
- Automatic retry with a larger budget (the error tells the DM what to do).
- Persistent AI edit sessions on arbitrary entities (the revision endpoint is
  stateless; the JSON in the modal is the state).
- Validating imported content against compendium/statblock rules.

## Decisions

### Endpoint-aware token budget with a clamp

`aiDraftMaxTokens` becomes a resolver: read the endpoint's `max_tokens`, clamp
it to 4 000–32 000, and fall back to 16 000 when unset. 16k completed a full
Wolves-of-Welton draft in the production repro; the clamp protects against the
admin UI storing nonsensical values (400000000). A hardcoded 16k was rejected
because admins already have a knob and some providers bill per requested token.

### Per-call timeout instead of raising the global text timeout

`generateChat` gains a `timeout time.Duration` parameter. Draft calls pass
`aiDraftTimeout = 180s`; free-text generation keeps the existing 60s. Raising
`aiTextTimeout` globally was rejected because unrelated generation endpoints
should not hang for three minutes.

### Truncation and emptiness are errors, not chat turns

Handlers check `finish_reason == "length"` and an empty trimmed reply after
`generateChat` and respond 502 with a message that names the fix ("raise
max_tokens for this endpoint or ask for a smaller draft"). The assistant turn
is not appended (the turn endpoint rolls back the user turn, matching its
existing network-error path). Returning 200 with an error bubble was rejected:
the frontend already toasts `aiGenError` responses and the session should stay
clean for a retry.

### One output contract, kept next to the schemas

`aiDraftSystemPrompt` keeps its envelope rules and gains a common contract
section plus a detailed one-shot "skill": field-by-field rules, allowed values,
size limits (3–5 acts, 2–4 scenes per act, 4–8 NPCs/locations, 3–6 encounters,
3–8 clues, short descriptions) and an instruction to adapt a named published
adventure faithfully into the schema. The size limits are the truncation
mitigation.

### One writer for AI commit and direct import

`commitAIOneShot` is split so the insert work lives in
`createOneShotFromDraft(ctx, uid, campaignID, raw)`, returning counts. The
commit path calls it with the session draft; the import handler calls it with
the pasted draft. Keeping two writers was rejected: the two paths must produce
identical structures.

### One generic import endpoint, create or replace

`POST /api/ai/import` takes `{entity_type, campaign_id?, parent_id?,
entity_id?, json}`. Without `entity_id` it creates through the existing
`commitAIDraft` dispatch (which already has a writer for every type); with
`entity_id` it replaces the content of that entity through new update writers.
The one-shot route `POST /api/oneshot-adventures/import` remains as a
convenience that pins `entity_type` to one-shot. Replacing rather than
duplicating is what makes post-import AI revisions idempotent: the modal sends
the revised JSON with the same `entity_id` and the entity updates in place.
A single generic endpoint was chosen over one import route per entity type
because the envelope/fence parsing, ownership checks and error shape are
identical.

### Replacement semantics are non-destructive where entities are shared

On one-shot replace: the adventure row updates; acts/scenes and clues are
replaced (they cascade from the adventure); encounter templates linked to the
adventure are unlinked and pruned only when no other adventure or scene
references them; NPCs and locations are matched by name within the owner's
entities and reused, so shared entities survive edits. Renaming an NPC in a
revision creates a new NPC and leaves the old one in the owner's list — deleting
user-visible entities automatically was rejected as too destructive.

### Revision is stateless; the JSON is the state

`POST /api/ai/revise` takes `{entity_type, json, instruction, endpoint_id?}`,
builds a system prompt from the same contract plus a "change only what is
asked, return the complete object" addendum, and returns the standard
`{status, message, draft}` envelope. No draft session row is created: the modal
holds the current JSON client-side and sends it on every turn, which works
identically before and after import and avoids a schema migration. Reusing
`ai_draft_sessions` for edit mode was rejected because commit semantics there
create new entities, not update existing ones.

### Import UI reuses the generic modal

A new `ts/draft-import.ts` renders the import/refine modal in the existing
generic modal: an entity-type selector (one-shot, NPC, location, encounter,
faction, campaign), a JSON textarea, an "Ask AI" instruction box, and an
Import/Apply button. Buttons on the one-shot, NPC, location, encounter and
faction lists open it with the right type. After a successful import the modal
stays open, shows "ask the AI to change parts", and every revision is
auto-applied with a replace call. On failure the error shows in the modal and
the textarea keeps its content.

## Risks / Trade-offs

- [A draft can still exceed 32k with reasoning] → Size limits in the contract
  plus the truncation error; the DM can ask for a smaller draft or raise the
  endpoint limit.
- [Imported JSON may create duplicate NPCs/locations] → Same behavior as the
  existing commit path; imports are explicit user actions. Revisions reuse
  same-named entities.
- [Replacing a one-shot loses manual edits made outside the draft] → The DM
  explicitly asked the AI to apply a revision; the modal shows the JSON that
  will be applied.
- [The parent change is unarchived, so `ai-structured-drafting` is not yet in
  `openspec/specs/`] → This change's delta must be archived after
  `ai-adventure-generation`; the delta headers match that change's capability
  name so archiving composes.
- [The generic modal may already be open behind the AI draft modal] → The
  import entry points live on list toolbars, not inside the draft modal.

## Migration Plan

No database migration. Deploy, then drafts immediately honor the endpoint's
`max_tokens` (clamped) and the 180s timeout, and imports work for every
AI-draftable entity type. Rollback is a revert; no stored state changes shape.

## Open Questions

None.
