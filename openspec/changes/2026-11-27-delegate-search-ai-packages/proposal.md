## Why

TD-007 in `docs/tech-debt.md` tracked four handlers that still carried private copies of logic already extracted into the `search/` and `ai/` packages. The packages were not unused — `handlers/search.go`, `handlers/ai_search.go`, `handlers/ai_draft_lore.go` and `telegram/*` already consume them — so two implementations of the same provider protocol and the same retrieval pipeline had to be maintained side by side:

- `handlers/ai.go` reimplemented the OpenAI-compatible chat-completion transport (`generateText`, `generateChat`) that `ai/client.go` already provides, including its own timeout handling and error sanitisation.
- `handlers/copilot.go` reimplemented campaign RAG retrieval (`retrieveCampaignContext`) identical to `search.RetrieveCampaignContext`.
- `handlers/compendium.go` reimplemented compendium recall (`SearchCompendium`) identical to `search.SearchCompendium`.
- `handlers/settings.go` reimplemented the `ai_enabled` setting read (`aiEnabled`) identical to `ai.IsEnabled`.

Duplication of this kind drifts silently: a fix to FTS query building, visibility filtering, error mapping or timeouts has to be applied twice, and the copies were already diverging (the handler copy of the AI transport ignored the request context, and the handler copy of compendium recall scanned the generic-entry FTS columns in the wrong order).

## What Changes

- **`handlers/ai.go`** — `generateText` and `generateChat` become thin wrappers over `ai.GenerateText` / `ai.GenerateChat`; `aiGenError` becomes a type alias of `ai.AIGenError` so every existing call site (including image generation, vision and import handlers that construct the type) keeps compiling and matching; the local HTTP transport is deleted.
- **`handlers/events.go`** — `SetAppVersion` forwards the build version to `ai.SetAppVersion`, so the delegated transport stamps the same `villum/<version>` User-Agent as before.
- **`handlers/copilot.go`** — `retrieveCampaignContext` delegates to `search.RetrieveCampaignContext` and adapts `[]search.Source` back to the handler's JSON shape (`copilotSource`, which omits `subtitle`). The helper now takes the request context.
- **`handlers/compendium.go`** — `SearchCompendium` (routed at `GET /compendium/search`, `handlers/routes_compendium.go:14`) validates the request, maps query parameters onto `search.CompendiumParams` and delegates; the endpoint contract is unchanged.
- **`handlers/settings.go`** — `aiEnabled` delegates to `ai.IsEnabled`.
- **Markers** — the five `ponytail: TD-007` comments are removed.
- **Registry** — `docs/tech-debt.md` marks TD-007 resolved and links this change.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `backend-modularity`: adds a requirement that the AI provider transport and the campaign/compendium retrieval pipelines have exactly one implementation, in `ai/` and `search/`, with handlers delegating to them.

## Impact

- **Code**: `handlers/ai.go`, `handlers/copilot.go`, `handlers/compendium.go`, `handlers/settings.go`, `handlers/events.go`; focused test additions in `handlers/settings_ai_test.go` and `handlers/copilot_test.go`.
- **APIs**: no contract change. Request/response shapes and status codes are identical. One observable bug fix: schema-based `compendium_entries` hits from `/api/compendium/search` now carry the real `type`/`name` (the handler copy scanned `e.data` into `type` and `cs.type_name` into the JSON blob, so those hits arrived with a JSON string as their type and an empty name).
- **Packages**: `ai/` and `search/` are unchanged; they are now the only implementations.
- **Tests/CI**: `go build ./...`, `go vet ./...`, `go test ./handlers/ ./search/ ./ai/ ./dice/`; coverage floors unchanged (`handlers/` ≥40%).
