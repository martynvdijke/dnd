## Context

DMs already have a text-generation endpoint and a per-field "generate" button,
but nothing turns a conversation into structured campaign data. The existing
Copilot persists conversations yet sends only the latest query to the model, so
it has no memory. We want a conversational assistant that asks questions,
accepts steering, and only then emits a structured draft that can be committed
to real entities.

## Goals / Non-Goals

- Goals: multi-turn memory, resumable sessions, approval-gated structured
  drafts, one-click commit for the eight DM entity types, a safe commit path
  that always forces `user_id` from the session.
- Non-Goals: streaming token-by-token replies, editing an existing entity from
  a draft, a full chat history browser UI, non-DM/player access.

## Decisions

### Reuse a single chat primitive
`generateText` becomes a thin wrapper over a new `generateChat(ctx, endpointID,
messages []map[string]string, maxTokens *int, sessionID string)`. Multi-turn
callers pass the whole history; single-shot callers keep the old behavior. This
avoids a second provider-request implementation.

### Persist sessions in a raw table, not ent
`ai_draft_sessions` is raw-managed via `db/migrations/064_ai_drafts.go`
(messages and draft stored as JSON text). This mirrors the existing
`copilot_*` approach and avoids an ent migration for a short-lived, JSON-shaped
record. The `user_id` foreign key is intentionally omitted because unit tests
use `user_id = 1` without a seeded user row; handlers still force the id from
the session.

### Approval gate is modelled in the reply
The model must return `{"status":"chatting"|"ready","message":...,"draft":...}`.
`parseAIDraftReply` strips code fences, tolerates a JSON-string draft and
salvages embedded objects, and falls back to `("chatting", rawtext, nil)` so a
non-JSON reply becomes a chat message rather than an error. `commit` refuses
unless status is `ready` and a draft exists.

### Commit writers mirror existing create handlers
Each writer reproduces the insert shape used by the equivalent REST handler
(raw SQL for encounters/factions/one-shot items/party items; ent for NPCs,
locations, campaigns) and additionally enforces ownership that the legacy
handlers sometimes omit: `user_id` from the session, `isCampaignMember` for
campaign-scoped rows, and `canEditCharacterID` / parent checks for quests and
items. URLs come from the existing `entityURL` helper.

### Endpoint base URL normalization
Users pasted the full operation URL (`.../v1/chat/completions`) into a "Base
URL" field, producing `/chat/completions/chat/completions` and a 404. Rather
than documenting it only, we normalize on save and on request
(`normalizeAIBaseURL`/`aiRequestURL`) and reject non-absolute URLs, so old and
new rows both work.

## Risks / Trade-offs

- Storing provider history verbatim grows the `messages` blob; drafts are short
  and sessions are user-discardable, so this is acceptable for now.
- The model may not always honour the JSON contract; the parser's prose
  fallback keeps the UI usable, at the cost of not offering a draft that turn.
- Raw-table sessions need a schema-consistency allowlist entry (added) or the
  consistency test fails.

## Migration Plan

- Forward-only migration 64 creates `ai_draft_sessions`; there is no backfill
  and nothing depends on the table existing for other features.
- Rollback: drop the table (sessions are ephemeral drafts).

## Testing

- Unit (Go): `handlers/ai_url_test.go` for normalization; `handlers/ai_draft_test.go`
  for the parser, start/turn/commit, history propagation, not-ready rejection,
  unsupported entity type, list/discard.
- Unit (TS/vitest): `tsc --noEmit` plus existing suite.
- E2E (Playwright): `tests/oneshot.spec.ts` regression that leaves the One-Shots
  view and returns, asserting the loading ornament is gone.
