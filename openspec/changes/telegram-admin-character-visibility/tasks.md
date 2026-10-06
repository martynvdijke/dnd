# Tasks

## 1. Implement
- [x] 1.1 Add the administrator bypass to `editableCharacterSQL` (`telegram/claims.go`)
- [x] 1.2 Update all six call sites to pass the user id an extra time
      (`telegram/claims.go`, `telegram/characters.go`)

## 2. Tests
- [x] 2.1 Add `TestAdminSeesAllCharacters` (admin sees + opens another user's character; stranger still denied)
- [x] 2.2 Fix `TestHPRejectNonEditable` intruder role from `admin` to `user`

## 3. Verify
- [x] 3.1 `go build ./...`
- [x] 3.2 `go test ./telegram/...`
- [ ] 3.3 `task ci` / prek pre-push parity suite

## 4. Release
- [ ] 4.1 Push branch, open PR, watch checks
- [ ] 4.2 Merge (squash) and confirm the release workflow
