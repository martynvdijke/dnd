## Context

`ai/` and `search/` were extracted from the handlers (see archived `2026-09-12-split-go-god-files` and the data-access/context changes), and they are already consumed elsewhere in the tree (`handlers/search.go`, `handlers/ai_search.go`, `handlers/ai_draft_lore.go`, `telegram/*`). Four handler entry points were left behind with private copies:

| Handler | Duplicate | Package implementation |
|---|---|---|
| `handlers/ai.go` `generateText` / `generateChat` | chat-completion HTTP transport | `ai.GenerateText` / `ai.GenerateChat` (`ai/client.go`) |
| `handlers/copilot.go` `retrieveCampaignContext` | campaign RAG retrieval | `search.RetrieveCampaignContext` (`search/campaign.go`) |
| `handlers/compendium.go` `SearchCompendium` | compendium recall | `search.SearchCompendium` (`search/compendium.go`) |
| `handlers/settings.go` `aiEnabled` | `ai_enabled` setting read | `ai.IsEnabled` (`ai/client.go`) |

Stakeholders: backend contributors and reviewers; the handlers coverage gate (≥40%); anyone fixing provider or retrieval behaviour, who currently has to patch two copies. Constraint: no API or JSON-shape change, no behaviour change beyond correcting the compendium generic-entry projection.

## Goals / Non-Goals

**Goals:**
- Exactly one implementation of the AI provider transport and of campaign/compendium retrieval.
- Preserve observable behaviour: same endpoints, same JSON shapes, same status codes, same provider headers and payloads, same result ordering and truncation.
- Remove the five `ponytail: TD-007` markers and close TD-007 in the registry.
- Keep the delegated paths covered by the existing tests, plus a focused test for the `aiEnabled` wrapper.

**Non-Goals:**
- Changing the `ai/` or `search/` packages themselves (their behaviour is the target behaviour).
- Replacing the `aiGenError` type across the handler package (image, vision and import handlers construct it; an alias is enough).
- Converting raw `db.DB` calls to `*Context` — that is TD-008, explicitly out of scope here.
- Reworking the copilot or compendium HTTP contracts, pagination, or frontend callers.

## Decisions

### D1: Thin handler wrappers, not direct call-site replacement

`generateText`, `generateChat`, `retrieveCampaignContext` and `aiEnabled` are kept as wrappers that delegate to `ai`/`search`; their bodies are now one or two statements.

Rationale: the alternative (rewriting ~20 call sites across `copilot.go`, `copilot_extract.go`, `recaps.go`, `ai_draft.go`, `ai_revise.go`, `ink.go`, `settings.go`, `ai.go`) produces a much larger diff, touches files outside this change's scope, and re-implements the same parameter marshalling at each site. Wrappers keep the diff reviewable and make the "one implementation" property greppable: the duplicated bodies are gone, and each wrapper is a single delegation.

The exception is `retrieveCampaignContext`, whose signature gains a `context.Context` parameter (2 production call sites, 4 test call sites) because the package implementation needs the request context for its `*Context` queries. Passing `context.Background()` in a wrapper instead would have re-introduced the detached-context debt TD-002 just closed.

Alternatives rejected: direct call-site replacement (large churn, same result); keeping the local copies and deleting the packages (contradicts the extraction).

### D2: `aiGenError` becomes a type alias of `ai.AIGenError`

`handlers` and `ai` defined structurally identical error types. Callers all over the handler package assert `err.(*aiGenError)` and read `.Status`/`.Msg`, and image/vision/import handlers construct `&aiGenError{...}` directly. Declaring `type aiGenError = ai.AIGenError` makes the delegated functions return values that still satisfy every existing assertion and construction, and deletes the duplicate type declaration plus its `Error()` method. No call-site change is needed.

Alternatives rejected: converting `*ai.AIGenError` to `*aiGenError` in the wrappers (works, but leaves two types and a mapping function); migrating every call site to `ai.AIGenError` (churn in files outside this change).

### D3: Provider User-Agent stays `villum/<version>`

`ai.SetAIProviderHeaders` stamps `villum/<ai.AppVersion>`, defaulting to `0.0.0-dev`; the handler transport stamped `villum/<handlers.AppVersion>`. `handlers.AppVersion` is the canonical build version (set by `main.go` and asserted by `TestAIProviderHeaders_SuppliedSessionAndUA`), so `handlers.SetAppVersion` now forwards to `ai.SetAppVersion`, and the two delegation wrappers re-sync it before delegating (tests and embedders set `handlers.AppVersion` directly). Observable behaviour — the header value — is unchanged.

### D4: Behaviour preservation notes per delegation

**AI transport (`generateText` / `generateChat`).** Same payload (`model`, `messages`, optional `max_tokens`), same URL resolution (`AIRequestURL` ≡ `aiRequestURL`), same `Authorization`, `Content-Type`, `User-Agent` and `x-opencode-session` headers, same 64 KiB response cap, same `truncateResponse`/`sanitizeError` limits (200/100 chars), same status mapping (400/404/500/502), same `no_choices` finish reason, same log lines. Two intentional deltas, both improvements that follow from the package being context-aware:
- the request is built with `http.NewRequestWithContext`, so client cancellation now aborts the provider call;
- the effective client timeout is `min(per-call timeout, remaining ctx deadline)`, so a request with a shorter deadline no longer waits out the full per-call timeout.

Neither changes the response for a live request with no deadline.

**Copilot retrieval (`retrieveCampaignContext`).** Same campaign entity set, same entity-type queries, same FTS query builder (`buildFTS5Query` is already a re-export of `search.BuildFTS5Query`), same bm25 ordering, same visibility rules (`registry.VisibleIDs`), same 8-item default limit, same 8,000-character context truncation and same `[n] type/id title` block format. The only shape adaptation is dropping `search.Source.Subtitle`, which `copilotSource` never exposed.

**Compendium recall (`SearchCompendium`).** Same 400 on an empty `q`, same per-type filters (`type`, `class`, `level`, `school`, `category`, `cr`, `monster_type`) and same legacy table queries. Deltas:
- the FTS pass uses `search.BuildFTS5Query` (quoted prefix terms joined by AND) instead of the raw query string. Semantically the same AND-of-terms, but malformed input can no longer raise an FTS syntax error, and prefix matching is slightly broader.
- `search.CompendiumParams` applies a single limit to both the legacy tables and the generic FTS pass, where the handler used 10 and 25. The handler passes `Limit: 25`, a superset of the old result set: no hit that was previously returned can disappear, and the extra rows are only more of the same ordered results.
- schema-based `compendium_entries` hits are now projected correctly: the old handler scanned `e.data` into `type` and `cs.type_name` into the JSON blob, so those results carried a JSON document as their `type` and usually an empty `name`. The delegated path scans them in the declared column order, so `type` is the schema type name and `name` is parsed from the entry JSON. This is the one intentional observable fix; it affects only the generic-entry portion of the response.

**`aiEnabled`.** Same SQL semantics (`ai_enabled` missing or unreadable ⇒ enabled, `"1"` ⇒ enabled), now with the request context.

## Risks / Trade-offs

- **Compendium limit superset** — responses for queries with many matches may contain more rows (up to 25 per legacy table). Consumers render a list and read `[0]`; additive rows are safe. Mitigated by keeping the generic FTS cap unchanged at 25.
- **Compendium projection fix** — a client that (incorrectly) relied on the old `type`-as-JSON payload would see different values. No frontend code reads that field from generic hits; the fix is the documented intent of the endpoint.
- **Version sync** — two variables (`handlers.AppVersion`, `ai.AppVersion`) still exist; the wrappers re-sync on each delegated call. Follow-up (not this change): make `ai.AppVersion` the single variable and have `handlers` read it.
- **Detached context in wrappers** — none introduced; the copilot helper now *receives* the request context.

## Migration Plan

Single PR, no data migration. Rollback is a revert of the delegation commits; the packages are untouched, so reverting restores the handler copies exactly.

## Open Questions

<!-- none -->
