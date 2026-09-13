## Context

Universal search is real and cheap: `entity_search_index` is an FTS5 table (`entity_type,entity_id,title,subtitle,body`) kept in sync by triggers for 18 entity types (`db/search_index.go`), queried via `buildFTS5Query` + `bm25` ranking in `HandleSearch` (`handlers/search.go:224`), and permission-filtered through `registry.VisibleIDs` (`search.go:331`). Knowledge entries (`handlers/knowledge.go`) already participate. AI access is `HandleTextGeneration` (`handlers/ai.go:385`): request `{endpoint_id,prompt,system?,max_tokens?,session_id?}` → OpenAI-compatible `POST {base_url}/chat/completions`, with `setAIProviderHeaders` (`ai.go:40`), a 60s client, size-bounded reads and sanitized errors. `models/ai.go` holds endpoints; `aiEnabled` gates generation.

Absent: any chat/conversation storage, any streaming (SSE/`Flusher`), any transcript pipeline, and any campaign-scoped search filter.

## Goals / Non-Goals

**Goals**
- Ask questions in natural language and get answers grounded in the campaign's own entities, with citations.
- Keep every retrieved fact inside what the requesting user may already see.
- Persist conversations so context survives a reload.
- Turn a pasted session transcript into searchable campaign content and a summarizable recap.
- Reuse the AI layer exactly as-is (same config, same provider contract, same gate).

**Non-Goals**
- Streaming token output — the MVP is a synchronous single-shot reply; SSE is a follow-up.
- Agents/tool-calling that mutate campaign data ("add a quest") — read-only copilot for now.
- A new embedding/vector store — FTS5 + bm25 is the retriever.
- Cross-campaign or global questions; the copilot is scoped to one campaign.

## Decisions

### 1. Retrieval reuses the FTS5 index, scoped to the campaign
A `retrieveCampaignContext(campaignID, query, limit)` helper runs the existing FTS query, intersects results with the campaign's entity ids, and applies `registry.VisibleIDs` for the requesting user. Knowledge entries keep the `shared` rule (members never retrieve unshared entries).

**Why:** Search already knows how to rank and how to enforce visibility; a new retriever would be a second, drifting visibility authority. **Alternative:** query each source table directly — rejected (re-implements ranking and risks leaks).

### 2. `POST /api/campaigns/:id/copilot/chat` composes an augmented prompt
Request `{query, endpoint_id?, conversation_id?}`. The handler retrieves top-N entities, builds a system prompt instructing the model to answer *only* from the supplied context, appends a compact numbered source list, and calls the shared text-generation path (extracted into a reusable `generateText` used by both `HandleTextGeneration` and the copilot).

**Why:** The existing endpoint stays byte-for-byte compatible; the copilot owns prompt assembly. **Alternative:** add a `context` field to `textGenRequest` — rejected (changes the existing capability's contract for no gain).

### 3. Citations are mandatory in the response
Response `{answer, sources:[{entity_type,entity_id,title,url}], conversation_id}`. Sources are the retrieved entities actually passed to the model, so the DM can verify each claim.

**Why:** A hallucinated campaign fact is worse than "I don't know"; linking sources makes the answer auditable and matches the app's link-rich UI.

### 4. Conversations are campaign-scoped entities
Two ent schemas: `copilot_conversation{campaign_id,user_id,title,created_at}` and `copilot_message{conversation_id,role,content,created_at}`. Visibility follows the existing campaign membership model; conversations are private to their owner by default (owner user).

**Why:** Reuses the registry/ownership patterns and the migration convention. **Alternative:** no persistence (stateless Q&A) — rejected; follow-up questions need history.

### 5. Transcript ingestion stores content, indexing is automatic
`POST /api/campaigns/:id/copilot/transcript {text, session_id?}` chunks the transcript into a campaign-scoped notes/session entry; the existing search triggers index it, so no manual indexing code. Summarization is a copilot call over the ingested entity.

**Why:** The index is trigger-driven; storing content is the whole job. **Alternative:** bespoke transcript table + custom indexing — rejected (duplicates search plumbing).

### 6. Synchronous MVP, size-bounded
The reply uses the existing non-streaming client. Retrieved context and transcript chunks are capped; over-limit content is truncated with a signal in the response.

## Risks / Trade-offs

- **Prompt injection from stored user content** → the system prompt treats retrieved content as data, never instructions; retrieval is read-only so no tool can act on an injection.
- **Hallucinated facts** → answer-only-from-context instruction plus mandatory citations; "insufficient context" is an allowed answer.
- **Secret leakage to members** → retrieval goes through `registry.VisibleIDs` + the knowledge `shared` rule, covered by tests.
- **Cost / token blowup** → fixed top-N retrieval, hard caps on context size and `max_tokens`.
- **AI disabled or no endpoint** → explicit 503-style error consistent with existing `aiEnabled` handling; no partial writes.
- **Transcripts contain PII** → stored as campaign content under existing campaign visibility; no external send except the DM-initiated summarize call through the configured provider.

## Migration Plan
Additive: two new tables (additive migration) and one route. No change to existing AI or search endpoints. Rollback: remove the route/panel; orphaned tables harmless.

## Open Questions
- Should the copilot be DM-only initially, or available to players over the member-visible slice? (Proposal: DM first, then members.)
- Do we add SSE streaming in the same change or as an immediate follow-up? (Proposal: follow-up.)
- Should transcript ingestion create a `session` or a `note`? (Proposal: `session` when `session_id` is given, else `note`.)
