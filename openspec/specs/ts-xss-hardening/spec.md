# ts-xss-hardening Specification

## Purpose
TBD - created by archiving change harden-ts-xss-residual. Update Purpose after archive.
## Requirements
### Requirement: JS-context escaping

The system SHALL NOT construct JS string literals inside `innerHTML` using HTML-only escaping; all `onclick` or other JS-context interpolations of user-controlled data SHALL use a JS-context escape (`jsEscape`/`attrEscape`) or a `data-*` attribute + `addEventListener` pattern.

#### Scenario: No onclick JS interpolation with HTML esc only

- **WHEN** searching the codebase with `rg -n "onclick.*esc\(|onclick.*escapeHtml" ts/`
- **THEN** no matches are found

#### Scenario: Wiki edit button passes user data safely

- **WHEN** rendering the wiki page edit button previously at `ts/app.ts:2710` (`showEditWikiPage(${id},'${esc(title)}',...)`) with a title containing `'`, `"`, `\`, `<`, or `</script>`
- **THEN** the title is transported via `data-*` attributes or `jsEscape` and the click handler receives the exact original title without executing injected JS

#### Scenario: Party member suggestion click populates input safely

- **WHEN** rendering the party member suggestion at `ts/party.ts:431` (`onclick="...value='${esc(u.username)}'"`) with a username containing `'` or `";alert(1)//`
- **THEN** the username is set via `data-*` + event listener (or `jsEscape`) and no JS injection executes

### Requirement: Esc dedup

The system SHALL expose a single canonical `esc` (HTML-escaping) export in `ts/lib/dom.ts` and SHALL NOT retain duplicate `esc`/`escAttr`/`escapeHtml` definitions in `ts/mentions.ts` or `ts/admin.ts`.

#### Scenario: Single esc definition

- **WHEN** searching with `rg -n "function esc\(|function escAttr|function escapeHtml|const esc =" ts/`
- **THEN** only `ts/lib/dom.ts` contains a definition; `ts/mentions.ts:42` and `ts/admin.ts:313`/`ts/admin.ts:2107` no longer define their own

#### Scenario: Imports resolve after dedup

- **WHEN** running `npm run typecheck`
- **THEN** the project typechecks with `ts/mentions.ts` and `ts/admin.ts` importing `esc`/`attrEscape` from `ts/lib/dom.ts` and no local escape helpers

### Requirement: Console guard

The system SHALL NOT emit `console.log`/`console.warn`/`console.error` in production bundles; such calls SHALL be gated by `import.meta.env.DEV` or removed, and an ESLint `no-console` rule SHALL prevent re-introduction (with explicit `// eslint-disable-next-line no-console` allowlist where needed).

#### Scenario: No ungated console in source

- **WHEN** searching with `rg -n "console\.(log|warn|error)" ts/`
- **THEN** every match is either guarded by `import.meta.env.DEV` or has an `eslint-disable` comment documenting the allowlist

#### Scenario: No console strings in prod bundles

- **WHEN** running `npm run build:vite` in production mode and searching `rg -n "console\.(log|warn|error)" static/js/`
- **THEN** no matches are found in emitted `static/js/*.js` and `static/js/chunks/*`

### Requirement: Vite shared chunks

The build SHALL NOT duplicate shared `ts/lib/*` (`dom.ts`, `api.ts`, `expose.ts`, `bootstrap`) and `dompurify` code across the 5 frontend bundles (`app`, `admin`, `pwa`, `setup`, `login`); it SHALL use a multi-entry build with a shared chunk (`manualChunks` or `rollupOptions.input` map) so shared lib code is emitted once.

#### Scenario: Shared chunk emitted

- **WHEN** running `npm run build:vite` and listing `static/js/`
- **THEN** a shared chunk (e.g. `static/js/chunks/*` or `static/js/shared-*.js`) exists and contains `ts/lib/dom.ts`/`api.ts`/`expose.ts` code, and per-entry bundles no longer each inline a full copy

#### Scenario: All entries still load

- **WHEN** serving the app and visiting pages that load `app.js`, `admin.js`, `pwa.js`, `setup.js`, and `login.js` (including any required shared-chunk `<script>` tags)
- **THEN** each page loads without console errors and shared helpers (e.g. `esc`, `api`, `toast`) resolve correctly
