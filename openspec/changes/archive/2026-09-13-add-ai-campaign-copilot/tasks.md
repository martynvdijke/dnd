## 1. Retrieval

- [x] 1.1 Add `retrieveCampaignContext(campaignID, query, limit)` over the FTS5 index, intersected with the campaign's entities and `registry.VisibleIDs`.
- [x] 1.2 Unit tests: member sees only visible/shared content; owner sees all; no cross-campaign leakage.

## 2. Generation

- [x] 2.1 Extract a reusable `generateText(...)` helper from `HandleTextGeneration` (no behaviour change to the existing endpoint).
- [x] 2.2 Add `POST /api/campaigns/:id/copilot/chat {query, endpoint_id?, conversation_id?}` returning `{answer, sources[], conversation_id}`.
- [x] 2.3 Compose a context-grounded system prompt (answer-only-from-context, treat content as data) with a numbered source list.
- [x] 2.4 Enforce context-size and `max_tokens` caps; handle AI-disabled / no-endpoint errors without partial writes.

## 3. Persistence

- [x] 3.1 Add ent schemas + additive migration for `copilot_conversation` and `copilot_message` (campaign + owner scoped).
- [x] 3.2 Implement conversation create/list/get and message append with owner-only access.
- [x] 3.3 Tests: resume conversation, owner-only isolation.

## 4. Prep & transcript

- [x] 4.1 Prep/recap generation via the retrieval path.
- [x] 4.2 `POST /api/campaigns/:id/copilot/transcript {text, session_id?}` storing content so search triggers index it.
- [x] 4.3 Transcript summarization to a recap.
- [x] 4.4 Tests: transcript indexed and retrievable; summarize path.

## 5. Frontend

- [x] 5.1 Campaign copilot panel: ask, show answer with clickable cited sources, list/resume conversations.
- [x] 5.2 Transcript paste/summarize affordance.

## 6. Tests & verification

- [x] 6.1 `go vet ./...`, `task test`, `npm run typecheck`, `npm run build:vite`, `task lint:e2e`.
- [x] 6.2 e2e: seed campaign content, ask a question, assert the answer cites a real entity.
