## 1. AI endpoint fix

- [x] 1.1 Add `normalizeAIBaseURL`, `aiRequestURL`, `validateAIBaseURL` in `handlers/ai.go` and route all provider requests through them
- [x] 1.2 Validate/normalize on create and update; return 400 for invalid base URLs
- [x] 1.3 Show the provider message on a failed admin test (`ts/admin/integrations.ts`) and add base-URL help text (`static/admin.html`)
- [x] 1.4 Tests: `handlers/ai_url_test.go`

## 2. Chat primitive

- [x] 2.1 Extract `generateChat` and make `generateText` delegate to it
- [x] 2.2 Keep all existing AI call sites unchanged

## 3. Draft sessions (backend)

- [x] 3.1 Migration 064 + registry entry + schema-consistency allowlist
- [x] 3.2 `handlers/ai_draft.go`: sessions, system prompt, JSON reply parser
- [x] 3.3 Handlers: start, turn (full history), get, list, discard, commit
- [x] 3.4 Commit writers for oneshot (+acts/scenes/NPCs/locations/encounters/clues), npc, location, encounter, faction, campaign, quest, item
- [x] 3.5 Own only your sessions; campaign access check; force `user_id` from session
- [x] 3.6 Register routes in `handlers/routes_ai.go`
- [x] 3.7 Tests: `handlers/ai_draft_test.go`

## 4. Frontend

- [x] 4.1 `ts/ai-draft.ts` chat modal with localStorage reconnect
- [x] 4.2 Modal markup in `static/app.html`; import in `ts/app.ts`
- [x] 4.3 "Draft with AI" button on the one-shot list

## 5. One-shot loading regression

- [x] 5.1 Replace `hx-get`/`hx-trigger="load"` with `htmx.ajax` in `ts/app.ts` (`loadViewFragment`)
- [x] 5.2 E2E test that revisits One-Shots and asserts the ornament is gone

## 6. Verify and release

- [ ] 6.1 `go test -tags sqlite_fts5 ./...`, `npx vitest run`, `npx tsc --noEmit`, `go vet ./...`
- [ ] 6.2 Full CI parity via prek pre-push (or `task ci`)
- [ ] 6.3 Commit with Conventional Commit messages on `feat/ai-structured-generation`
- [ ] 6.4 Push and open PR; wait for checks; merge when green
