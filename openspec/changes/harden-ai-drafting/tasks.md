## 1. Draft reliability (backend)

- [x] 1.1 Add a per-call timeout parameter to `generateChat`; free-text keeps 60s, drafts use a new 180s `aiDraftTimeout`
- [x] 1.2 Replace `aiDraftMaxTokens` with a resolver that reads the endpoint `max_tokens` (clamp 4000–32000, default 16000)
- [x] 1.3 Detect empty replies and `finish_reason == "length"` in `StartAIDraft`/`AIDraftTurn`; return actionable 502 errors without storing a broken assistant turn
- [x] 1.4 Tests: endpoint budget clamp/default, empty reply, truncated reply

## 2. One-shot output contract

- [x] 2.1 Expand `aiDraftSystemPrompt` with common envelope rules and a detailed one-shot field contract (allowed values, size limits, named-adventure adaptation)
- [x] 2.2 Tests: prompt includes the one-shot contract, allowed values and size limits

## 3. Generic import (backend)

- [x] 3.1 Extract `createOneShotFromDraft` from `commitAIOneShot`; return created counts; keep the commit path behavior unchanged
- [x] 3.2 Add `POST /api/oneshot-adventures/import` accepting `{campaign_id?, json}`; strip fences, unwrap the assistant envelope, require a title, enforce campaign membership
- [x] 3.3 Add `POST /api/ai/import` accepting `{entity_type, campaign_id?, parent_id?, entity_id?, json}`; create through the existing commit dispatch, replace through update writers
- [x] 3.4 Update writers for every entity type with ownership checks; one-shot replace updates the adventure and replaces acts/scenes/clues, reuses same-named NPCs/locations, prunes only unreferenced encounter templates
- [x] 3.5 Tests: create for one-shot/NPC/location/encounter/faction/campaign, replace one-shot and a simple type, envelope, fences, malformed JSON, missing name/title, unknown type, scope access, created counts

## 4. AI revision (backend)

- [x] 4.1 Add `POST /api/ai/revise` accepting `{entity_type, json, instruction, endpoint_id?}`; resolve the first enabled text endpoint when unset; return the standard draft envelope; validate empty/truncated replies
- [x] 4.2 Tests: revision applies only the requested change (prompt contract), chatting reply, truncated/empty failure

## 5. Import/refine UI (frontend)

- [x] 5.1 New `ts/draft-import.ts`: entity-type selector, JSON textarea, Import/Apply button, "Ask AI" instruction box, inline errors, auto-apply revisions after import, keep text on failure
- [x] 5.2 Import the module from `ts/app.ts`; add "Import JSON" buttons to `oneshot_list.html`, `campaign_npcs_section.html`, `locations_list.html`, `campaign_encounters_section.html`, `factions_list.html`
- [x] 5.3 Vitest: import success refreshes, revision before import updates the textarea, revision after import applies a replace, failure keeps text
- [x] 5.4 E2E: import a small JSON draft and assert the adventure appears

## 6. Verify and release

- [ ] 6.1 `go test -tags sqlite_fts5 ./...`, `npx vitest run`, `npx tsc --noEmit`, `go vet ./...`
- [ ] 6.2 Full CI parity via prek pre-push (or `task ci`)
- [ ] 6.3 Commit with Conventional Commit messages on `feat/ai-draft-import`
- [ ] 6.4 Push and open the PR; wait for checks; merge when green
- [ ] 6.5 Archive `ai-adventure-generation` first, then this change, so `ai-structured-drafting` exists in `openspec/specs/`
