## 1. Backend share support

- [x] 1.1 Add `recap` to `supportedShareTypes` in `handlers/share.go`
- [x] 1.2 Resolve recap ownership in `canShareEntity` via campaign membership
- [x] 1.3 Add recap label to `shareEntityLabel`
- [x] 1.4 Add `case "recap"` to `GetSharedEntity` (public JSON)
- [x] 1.5 Add `SharedRecapView` + `case "recap"` to `GetSharedPage` (public HTML)

## 2. Public page template

- [x] 2.1 Add `handlers/templates/share_recap.html` (noindex, escaped content)

## 3. Frontend

- [x] 3.1 Add `shareRecap(id)` + share button on recap cards in `ts/campaign/recaps.ts`
- [x] 3.2 Add `recap` label to `ts/share.ts`

## 4. Verification

- [x] 4.1 Go test `handlers/share_recap_test.go` (create, public JSON, public HTML, non-member 403)
- [x] 4.2 Playwright e2e in `tests/recaps.spec.ts` (share button → public JSON)
- [x] 4.3 `go build ./...`, `npm run typecheck`, `task lint:e2e`
