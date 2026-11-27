## 1. AI transport delegation

- [x] 1.1 Replace the local `generateText` / `generateChat` HTTP transport in `handlers/ai.go` with thin wrappers over `ai.GenerateText` / `ai.GenerateChat`; delete the duplicated payload, header, timeout, response-cap, truncation and error-mapping code.
- [x] 1.2 Replace `handlers.aiGenError` with `type aiGenError = ai.AIGenError` so existing constructions (`ai_vision.go`, `ai_import.go`, `ai_draft.go`) and assertions (`copilot.go`, `copilot_extract.go`, `ink.go`, `ai_draft.go`, `ai.go`) keep working unchanged.
- [x] 1.3 Forward the build version: `handlers.SetAppVersion` calls `ai.SetAppVersion`, and the wrappers re-sync `handlers.AppVersion` before delegating so the `villum/<version>` User-Agent is unchanged (pinned by `TestAIProviderHeaders_SuppliedSessionAndUA`).
- [x] 1.4 Verification: `go test ./handlers -run 'TestGenerateText|TestAIProviderHeaders|TestAIText|TestAIDraft'` passes; provider payload/headers asserted by the mock-provider tests.

## 2. Search delegations

- [x] 2.1 Replace `handlers/copilot.go` `retrieveCampaignContext` with a delegation to `search.RetrieveCampaignContext`, adapting `[]search.Source` to `[]copilotSource` (subtitle dropped, shape unchanged); give the helper a `context.Context` and pass `c.Request.Context()` at both call sites (`handleCopilotChat`, transcript summarise).
- [x] 2.2 Replace `handlers/compendium.go` `SearchCompendium` body with input validation plus a delegation to `search.SearchCompendium`, mapping `q`/`type`/`class`/`level`/`school`/`category`/`cr`/`monster_type` onto `search.CompendiumParams` and normalising an empty result to `[]`; `GET /compendium/search` stays registered in `handlers/routes_compendium.go`.
- [x] 2.3 Replace `handlers/settings.go` `aiEnabled` with a delegation to `ai.IsEnabled`.
- [x] 2.4 Update the four `retrieveCampaignContext` call sites in `handlers/copilot_test.go` to pass `context.Background()`.
- [x] 2.5 Remove the five `ponytail: TD-007` markers (`ai.go` ×2, `copilot.go`, `compendium.go`, `settings.go`).
- [x] 2.6 Verification: `go test ./handlers -run 'TestRetrieveCampaignContext|TestCompendium|TestSearchCompendium|TestCopilot|TestAIEnabled'` passes, including the 400/`type`-filter contracts of `/api/compendium/search`.

## 3. Focused tests

- [x] 3.1 Add `handlers/settings_ai_test.go` pinning the delegated `aiEnabled` contract: missing setting ⇒ enabled, `"0"` ⇒ disabled, `"1"` ⇒ enabled.
- [x] 3.2 Keep the existing coverage of the delegated paths (mock provider for `generateText`, visibility fixtures for `retrieveCampaignContext`, type-filter fixtures for compendium search) and confirm the `handlers/` coverage floor (≥40%) still holds.

## 4. Registry and documentation

- [x] 4.1 Update `docs/tech-debt.md`: TD-007 → resolved with the corrected symptom and a link to this change; refresh TD-001/TD-002/TD-006 evidence and statuses; resolve TD-002/TD-003/TD-004; add TD-008 for non-context-aware raw SQL.
- [x] 4.2 Record the wrapper-vs-direct decision, the error-type alias, the version forwarding and every behavioural delta in `design.md`.

## 5. Verification gates

- [x] 5.1 `go build ./...` passes.
- [x] 5.2 `go vet ./...` passes.
- [x] 5.3 `go test ./...` passes, at minimum `./handlers/ ./search/ ./ai/ ./dice/`.
- [x] 5.4 `bash scripts/ci/check-detached-context.sh` passes (no new `context.TODO()`/`context.Background()` in handlers; the copilot helper now takes the request context).
- [x] 5.5 `bash scripts/ci/check-concatenated-sql.sh` passes (no SQL assembled from client input).
- [x] 5.6 Conventional commits, scoped to `handlers/`, `docs/tech-debt.md` and this change folder; no push (PR flow owns that).
