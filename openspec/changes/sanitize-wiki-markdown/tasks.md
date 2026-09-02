## 1. Dependency and helper setup

- [x] 1.1 Add `dompurify` and `@types/dompurify` to `package.json` (`npm install dompurify && npm install -D @types/dompurify`)
- [x] 1.2 Create `ts/lib/markdown.ts` with `renderMarkdown(md: string): string` — `marked.parse()` + `DOMPurify.sanitize()` with restrictive allowlist (strip `on*`, `javascript:`/`data:` URLs, `<script>/<iframe>/<object>/<form>`)
- [x] 1.3 Add vitest unit tests for `renderMarkdown()` covering `<script>` injection, `<img onerror>`, `javascript:` href, and allowed markdown (links, bold, code, lists, tables) — tests initially fail

## 2. Wiki and markdown rendering

- [x] 2.1 Grep for all `marked.parse` call sites (`rg "marked\.parse" ts/`) and replace each with `renderMarkdown()` import from `ts/lib/markdown.ts`
- [x] 2.2 Update `ts/app.ts:2699-2701` wiki rendering (`el.innerHTML = ... renderContent`) to use `renderMarkdown(page.content)` and verify no direct `marked.parse` remains outside the helper
- [x] 2.3 Verify wiki rendering in dev: create/edit wiki page with markdown (headings, links, images, code) and confirm correct display

## 3. Toast and modal hardening

- [x] 3.1 Harden `ts/lib/dom.ts:toast()` to escape `msg` by default via `esc()`; add `{ html?: boolean }` opt-in (keep boolean overload shim for backwards compat)
- [x] 3.2 Audit all `toast(` call sites (`rg "toast\(" ts/`) — migrate intentional HTML callers to `{ html: true }`, confirm the rest rely on safe default
- [x] 3.3 Audit all `showModal(` call sites for unescaped user content interpolation; fix any that inject raw user strings without `esc()`/`renderMarkdown()`

## 4. Verification

- [x] 4.1 `npm run typecheck` clean
- [x] 4.2 `npm run test:unit` green including new sanitizer tests
- [x] 4.3 `npm run build:vite` succeeds (5 bundles)
- [ ] 4.4 `task test:e2e` green — wiki e2e specs still pass with sanitized rendering
- [ ] 4.5 Run full local gates (`task ci`) and ensure no regressions
