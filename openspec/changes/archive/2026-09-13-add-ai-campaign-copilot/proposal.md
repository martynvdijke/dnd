## Why
Villum already stores an enormous amount of campaign context — wiki, NPCs, quests, sessions, journals, factions, knowledge entries — and already indexes all of it in an FTS5 table for universal search. It also already talks to OpenAI-compatible providers for one-shot text generation. What it cannot do is answer a question *about the campaign itself* ("who still owes us a favour?", "what did we never resolve about the cult?") or turn a wall of session transcript into indexed campaign content. This change adds a retrieval-augmented copilot on top of the search index and the existing AI layer.

## What Changes
- Add campaign-scoped retrieval that reuses the FTS5 `entity_search_index` and the existing permission filtering.
- Add `POST /api/campaigns/:id/copilot/chat`: retrieve relevant campaign entities, answer the question, and return the **cited** source entities.
- Persist conversations and messages (new campaign-scoped entities) so a DM can pick up a thread.
- Add a prep assistant that generates session prep / a recap from the campaign's own content via the same retrieval path.
- Add transcript ingestion: accept a session transcript, store it, let the search triggers index it, and summarize it into a recap.
- Reuse the existing AI provider config, headers, error handling and `ai_enabled` gate; add no provider code.

## Capabilities
### New Capabilities
- `ai-campaign-copilot`: campaign-scoped retrieval, cited Q&A, conversation history, prep generation and transcript ingestion.

### Modified Capabilities
<!-- None: the existing `ai-text-generation` endpoint is reused unchanged; the copilot composes its own system/context prompt and calls the same provider path. -->

## Impact
Backend: new `handlers/copilot.go` + route, retrieval helper over `handlers/search.go` / `db/search_index.go`, reuse of `handlers/ai.go`. New ent schemas `copilot_conversation` / `copilot_message` (migration). Frontend: new copilot panel (`ts/copilot.ts`) in the campaign area. No new dependencies; no streaming in the MVP (existing single-shot generation, 64KB cap).
