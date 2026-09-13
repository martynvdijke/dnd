## 1. Header helper

- [x] 1.1 Add `setAIProviderHeaders(req *http.Request, sessionID string)` in `handlers/ai.go` setting `User-Agent: villum/<AppVersion or 0.0.0-dev>` and `x-opencode-session`
- [x] 1.2 Add session-id helper (resolve inbound `session_id` or generate 16 random bytes hex) reusing existing `crypto/rand` + `encoding/hex` imports
- [x] 1.3 Wire the helper into all four POST sites: endpoint-test text, endpoint-test image, `HandleTextGeneration`, `HandleImageGeneration`

## 2. Session passthrough

- [x] 2.1 Accept optional `session_id` in `textGenRequest` / `imageGenRequest` (and test endpoint) and echo the effective value in each JSON response
- [x] 2.2 Endpoint-test probes use an ephemeral generated id per probe

## 3. Tests

- [x] 3.1 Handler test: outbound provider request carries `villum/` UA and the supplied `x-opencode-session` (httptest provider capturing headers)
- [x] 3.2 Handler test: omitted `session_id` is generated, used outbound, and echoed in the response; resending it yields the same outbound value
- [x] 3.3 Unit test: empty `AppVersion` falls back to `0.0.0-dev` in the UA; UA never equals `Go-http-client/*` nor claims another SDK

## 4. Verification

- [x] 4.1 `go vet ./... && go build ./...` clean
- [x] 4.2 `task test` green including new AI header tests (keeps `handlers/` coverage floor)
- [x] 4.3 `openspec validate ai-coding-agent-headers` clean
