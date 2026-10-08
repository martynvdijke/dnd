# Tasks

## 1. Implement
- [x] 1.1 Create `tests/telegram-bot-flows.spec.ts` (login/claim/status flows)
- [x] 1.2 Create OpenSpec change `add-telegram-bot-e2e-tests` (proposal + spec delta)

## 2. Tests
- [x] 2.1 `/login` email + magic-link + `/status` linked flow
- [x] 2.2 `/claim` inline-button `claim:<id>` callback flow + re-claim + `/unclaim`
- [x] 2.3 `/status` linked-user regression

## 3. Verify
- [x] 3.1 `npm run typecheck` passes
- [x] 3.2 `task ci` / prek pre-push parity suite
- [x] 3.3 `openspec validate add-telegram-bot-e2e-tests --strict` passes

## 4. Release
- [ ] 4.1 Push branch, open PR, watch checks
- [ ] 4.2 Merge (squash) and confirm release workflow
