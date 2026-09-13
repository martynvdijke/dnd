## Context

All outbound LLM traffic lives in `handlers/ai.go`: `newAIClient()` builds a bare `*http.Client`, and four POST sites set only `Content-Type` + `Authorization` (endpoint-test text at ~L290, endpoint-test image at ~L321, `HandleTextGeneration` at ~L424, `HandleImageGeneration` at ~L559). Go therefore sends its default `Go-http-client/1.1` UA and no session header. `handlers.AppVersion` already exists (injected from `main.Version`, fallback `0.0.0-dev`), and `crypto/rand` + `encoding/hex` are already imported in `ai.go`. Current generation requests (`textGenRequest`/`imageGenRequest`) have no conversation/session concept — each call is single-shot.

## Goals / Non-Goals

**Goals:**
- Provider sees Villum as a self-identified coding-agent client on every AI POST.
- Follow-up turns of one conversation share one `x-opencode-session` for routing/cache affinity, without any server-side conversation store.
- One helper owns both headers so the four sites cannot drift.

**Non-Goals:**
- No request pacing, throughput shaping, or output-rate mimicry — the provider's ask is headers, and "typical traffic" imitation is explicitly out of scope.
- No spoofing another agent's UA or SDK name.
- No conversation persistence, no DB tables, no session lifecycle management beyond generate-and-echo.
- No changes to non-AI outbound calls (compendium `http.Get`, web push, iCal fetch).

## Decisions

### 1. One helper: `setAIProviderHeaders(req, sessionID)`
Sets `User-Agent: villum/<AppVersion or 0.0.0-dev>` and `x-opencode-session: <sessionID>`; called at all four POST sites.
**Why:** Single choke point — grep every caller shows all AI traffic routes through these four constructions, so one helper fixes all callers at once. **Alternative:** per-site inline `Header.Set` — rejected; four copies drift.

### 2. UA reuses `handlers.AppVersion`
`villum/` + `AppVersion`, falling back to `0.0.0-dev` when empty (tests, local builds).
**Why:** Version plumbing already exists (`main.go` → `SetAppVersion`); no new build vars. **Alternative:** hardcode `villum/1.0` — rejected; rots on every release.

### 3. Stateless session: optional inbound `session_id`, else generate-and-echo
`textGenRequest`/`imageGenRequest` (and the test endpoint) accept optional `session_id`; responses echo the effective value. Absent → `crypto/rand` 16 bytes hex. The outbound header always carries the effective value.
**Why:** Gives multi-turn callers cache affinity (resend the echoed id) with zero storage — the laziest stable-session scheme that works across the existing single-shot API. **Alternative:** deterministic hash of (user, endpoint, prompt) — rejected; breaks follow-up affinity when the prompt changes, which is exactly when caching matters. **Alternative:** server-side session table — rejected; new entity + migration for a value the client can trivially hold.

### 4. Endpoint-test probes use an ephemeral session
Test probes generate a fresh id per probe and do not persist it.
**Why:** A connectivity probe is not a conversation; polluting cache keys with probe ids buys nothing.

## Risks / Trade-offs

- **AppVersion empty in some builds** → fallback constant keeps the UA self-identifying; covered by a unit test with `AppVersion = ""`.
- **Custom OpenAI-compatible endpoints reject unknown headers** → `x-opencode-session` is a plain custom header; unknown-header tolerance is near-universal on such endpoints; a failing provider surfaces through the existing non-2xx path unchanged.
- **Clients ignore the echoed id** → degrades to one-session-per-request (today's effective behavior for caching); correctness unaffected.
- **Session id collision across users sharing an endpoint** → random 128-bit ids make collision negligible; cache affinity is per-id, never a correctness boundary.

## Migration Plan

1. Land helper + header wiring + session passthrough/echo behind handler tests.
2. No migration, no config, no frontend change required; clients adopt `session_id` reuse opportunistically.
3. Rollback: revert commit; headers vanish, behavior identical otherwise.

## Open Questions

- None. Header names/values are dictated by the provider ask; session transport (`session_id` JSON field + echo) is the minimal fit for the current single-shot API.
