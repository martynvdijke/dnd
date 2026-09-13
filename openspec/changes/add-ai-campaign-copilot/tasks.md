## 1. Retrieval

- [ ] 1.1 Add `retrieveCampaignContext(campaignID, query, limit)` over the FTS5 index, intersected with the campaign's entities and `registry.VisibleIDs`.
- [ ] 1.2 Unit tests: member sees only visible/shared content; owner sees all; no cross-campaign leakage.

## 2. Generation

- [ ] 2.1 Extract a reusable `generateText(...)` helper from `HandleTextGeneration` (no behaviour change to the existing endpoint).
- [ ] 2.2 Add `POST /api/campaigns/:id/copilot/chat {query, endpoint_id?, conversation_id?}` returning `{answer, sources[], conversation_id}`.
- [ ] 2.3 Compose a context-grounded system prompt (answer-only-from-context, treat content as data) with a numbered source list.
- [ ] 2.4 Enforce context-size and `max_tokens` caps; handle AI-disabled / no-endpoint errors without partial writes.

## 3. Persistence

- [ ] 3.1 Add ent schemas + additive migration for `copilot_conversation` and `copilot_message` (campaign + owner scoped).
- [ ] 3.2 Implement conversation create/list/get and message append with owner-only access.
- [ ] 3.3 Tests: resume conversation, owner-only isolation.

## 4. Prep & transcript

- [ ] 4.1 Prep/recap generation via the retrieval path.
- [ ] 4.2 `POST /api/campaigns/:id/copilot/transcript {text, session_id?}` storing content so search triggers index it.
- [ ] 4.3 Transcript summarization to a recap.
- [ ] 4.4 Tests: transcript indexed and retrievable; summarize path.

## 5. Frontend

- [ ] 5.1 Campaign copilot panel: ask, show answer with clickable cited sources, list/resume conversations.
- [ ] 5.2 Transcript paste/summarize affordance.

## 6. Tests & verification

- [ ] 6.1 `go vet ./...`, `task test`, `npm run typecheck`, `npm run build:vite`, `task lint:e2e`.
- [ ] 6.2 e2e: seed campaign content, ask a question, assert the answer cites a real entity.
