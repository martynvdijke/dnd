# ai-provider-attribution Specification

## Purpose
TBD - created by archiving change ai-coding-agent-headers. Update Purpose after archive.
## Requirements
### Requirement: Self-identifying User-Agent on AI provider calls
Every outbound AI provider POST SHALL set `User-Agent` to `villum/<version>` (using the injected app version), never the Go default transport string nor a generic SDK/HTTP-library name.

#### Scenario: Text generation carries Villum UA
- **WHEN** `POST /api/ai/text` proxies a chat-completions request to the provider
- **THEN** the outbound request carries `User-Agent: villum/<version>`

#### Scenario: No generic SDK name
- **WHEN** any AI provider POST is sent (test or generation, text or image)
- **THEN** the `User-Agent` does not claim to be another SDK or bare `Go-http-client`

### Requirement: Stable per-conversation session header
Every outbound AI provider POST SHALL carry `x-opencode-session` with a stable value per conversation so the provider can optimize routing and prompt caching. The value SHALL remain identical across follow-up turns of the same conversation and SHALL differ between unrelated conversations.

#### Scenario: Client-supplied session is forwarded
- **WHEN** the inbound request supplies a `session_id`
- **THEN** the outbound provider request sets `x-opencode-session` to that exact value

#### Scenario: Missing session is generated and returned
- **WHEN** the inbound request omits `session_id`
- **THEN** the server generates a random session id, uses it for `x-opencode-session`, and returns it in the response for reuse on follow-up turns

### Requirement: Full coverage of AI POST sites
All four outbound AI POST sites SHALL send both headers: endpoint-test text, endpoint-test image, text generation, and image generation.

#### Scenario: Endpoint test sends headers
- **WHEN** an endpoint test fires a text or image probe
- **THEN** the probe carries both `User-Agent` and `x-opencode-session`

#### Scenario: Generation sends headers
- **WHEN** a text or image generation request is proxied
- **THEN** the provider request carries both `User-Agent` and `x-opencode-session`

### Requirement: No traffic-shape mimicry
The system SHALL NOT throttle, pace, pad, or otherwise shape request timing or throughput to imitate "typical coding agent traffic", and SHALL NOT spoof another agent's identity. Attribution is headers only.

#### Scenario: No behavior shaping
- **WHEN** AI requests are sent at any rate the user triggers
- **THEN** no artificial delay, batching, or throughput cap is applied for attribution purposes
