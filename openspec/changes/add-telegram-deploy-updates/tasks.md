# Tasks

## 1. Implement
- [x] 1.1 Add `image:` to `docker-compose.yml` and document the pull-based update
- [x] 1.2 Expose the running build version via bot `/status`
- [x] 1.3 Create OpenSpec change `add-telegram-deploy-updates`

## 2. Tests
- [x] 2.1 `runStatus` includes `Build:` when a version is set
- [x] 2.2 `runStatus` omits the build line when unset

## 3. Verify
- [x] 3.1 `go test ./telegram/...` passes
- [ ] 3.2 `task ci` / prek pre-push parity suite
- [x] 3.3 `openspec validate add-telegram-deploy-updates --strict` passes

## 4. Release
- [ ] 4.1 Push branch, open PR, watch checks
- [ ] 4.2 Merge (squash) and confirm release workflow
