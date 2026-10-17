# Tasks

## 1. Client recovery
- [x] 1.1 Add `setUnauthorizedHandler` + one-shot retry to `ts/lib/api.ts`
- [x] 1.2 Add the same to `ts/admin/state.ts`

## 2. Bootstrap
- [x] 2.1 `ts/init.ts`: create a new token instead of rotating; register the handler
- [x] 2.2 `ts/admin.ts`: create a new token; register the handler; drop the `'admin'` key fallback

## 3. Tests
- [x] 3.1 `ts/lib/api.test.ts`: recovery, GET exemption, no-token, retry-failure
- [x] 3.2 `ts/admin/state.test.ts`: recovery + error surfacing

## 4. Verify
- [x] 4.1 `npx tsc --noEmit`
- [x] 4.2 `npx vitest run --coverage` within floors
- [x] 4.3 `npm run build:vite`

## 5. Release
- [x] 5.1 Push branch, open PR, watch checks
- [x] 5.2 Merge (squash) and confirm the release workflow
