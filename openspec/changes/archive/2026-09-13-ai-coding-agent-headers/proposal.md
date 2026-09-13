## Why

Villum's outbound AI provider calls currently send Go's default `User-Agent` and no session identifier, so they are indistinguishable from generic SDK or proxy traffic. The provider attributes coding-agent traffic purely via request headers — a self-identifying `User-Agent` plus a stable `x-opencode-session` per conversation for routing and prompt-cache affinity. Without them Villum gets misclassified, with worse routing and no cache benefit.

## What Changes

- Send a self-identifying `User-Agent` (`villum/<version>`) on every outbound AI provider request instead of the Go HTTP default or a generic SDK name.
- Send a stable `x-opencode-session` value per conversation on every outbound AI provider request.
- Accept an optional inbound session id (so multi-turn callers get cache affinity) and generate + return one when absent.
- Centralize both headers in one helper so all four AI POST sites (test text/image, text generation, image generation) stay consistent.
- Explicitly NOT changing request pacing, throughput shaping, or traffic mimicry — compliance is headers only.

## Capabilities

### New Capabilities

- `ai-provider-attribution`: Outbound AI provider attribution headers (self-identifying User-Agent + stable per-conversation session id) on all AI POST sites.

### Modified Capabilities

- *(none — `ai-text-generation`, `ai-image-generation`, `ai-endpoint-management` spec dirs are empty; no existing requirements change)*

## Impact

- **Backend**: `handlers/ai.go` only — one header helper + session passthrough/generation; reuses existing `handlers.AppVersion` (injected from `main.Version`). No DB, no migration, no new dependencies.
- **API**: `POST /api/ai/text`, `POST /api/ai/image`, and endpoint-test responses gain an optional `session_id` in/out; outbound provider POSTs gain two headers.
- **Frontend**: none required (clients may optionally store and resend `session_id` for follow-up turns).
