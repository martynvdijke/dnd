## [2.0.3](https://github.com/martynvdijke/dnd/compare/v2.0.2...v2.0.3) (2026-05-28)


### Bug Fixes

* **deps:** update all non-major dependencies ([ab35c98](https://github.com/martynvdijke/dnd/commit/ab35c987084372966685068a3d28526c298c01ac))

## [2.0.2](https://github.com/martynvdijke/dnd/compare/v2.0.1...v2.0.2) (2026-05-28)

## [2.0.1](https://github.com/martynvdijke/dnd/compare/v2.0.0...v2.0.1) (2026-05-28)


### Reverts

* Revert "fix: align ci.yaml with ci.yml - only test chromium in Release workflow" ([e3086d4](https://github.com/martynvdijke/dnd/commit/e3086d4be292bf3a5eaecb17f0701d01f8186aaf))

# [2.0.0](https://github.com/martynvdijke/dnd/compare/v1.16.0...v2.0.0) (2026-05-28)


### Bug Fixes

* add input validation and fix CI failures ([95bf1cd](https://github.com/martynvdijke/dnd/commit/95bf1cd0f4996c1e85171f07e59419cb9afd56a2))
* align ci.yaml with ci.yml - only test chromium in Release workflow ([624c116](https://github.com/martynvdijke/dnd/commit/624c116a00a9cd1bc3c39c64b18f955a4b44013e))
* **ci:** build frontend JS in test-e2e job ([31ad3da](https://github.com/martynvdijke/dnd/commit/31ad3da0967f56ddc9ec3411c4f04a22615ef026))


### Features

* add 13 isolated handler test files with testutil package for shift-left testing ([687492b](https://github.com/martynvdijke/dnd/commit/687492bc2a276e424313f6d23c93cd843b74cd47))
* player UX overhaul with responsive navigation, session mode, and PWA support ([092f388](https://github.com/martynvdijke/dnd/commit/092f3881aa1907acb407ea93f82acbb9eb9ceecc))


### BREAKING CHANGES

* Complete UI restructuring with mobile-first navigation (bottom tab bar + sidebar), session mode for live-play, bottom sheet component, context-aware FAB, PWA support with service worker and manifest, and touch gesture handling. The old top navbar is replaced with a responsive dual-navigation system.

Changes include:
- Bottom tab bar (Characters, Party, Dice, Compendium, More) on mobile
- Collapsible sidebar on desktop
- Session mode with sessionStorage persistence
- Bottom sheet component for mobile overlays
- Context-aware FAB with per-view actions
- PWA support (manifest.json, service worker)
- Touch gestures (pull-to-refresh, swipe-to-dismiss)
- Responsive CSS with mobile-first breakpoints
- Vite multi-entry build configuration
- Comprehensive Go API tests for static files and HTML structure
- Playwright E2E tests for navigation, PWA, and session mode
- Gitignore openspec/ and .opencode/ directories

# [1.16.0](https://github.com/martynvdijke/dnd/compare/v1.15.3...v1.16.0) (2026-05-27)


### Bug Fixes

* strip RPG notation suffix from die label in extractDieLabel ([d48db1f](https://github.com/martynvdijke/dnd/commit/d48db1f8f6605d9b8d7374f4d13f8e95c2914f69))


### Features

* add migrations and tests for campaign completeness features ([265cecf](https://github.com/martynvdijke/dnd/commit/265cecfa40ea7e8f7ad92a7865044c843b1c690c))
* add prek hooks, fix test reliability, containerized E2E ([fbba261](https://github.com/martynvdijke/dnd/commit/fbba26194190172c7e1c8f6bb943bf3448230b5a))
* campaign completeness - dashboard, party inventory, session planner, difficulty calculator, treasure generator ([21aacc0](https://github.com/martynvdijke/dnd/commit/21aacc015229afb5a1612dd97ac75476c1da43a5))
* rpg dice notation engine with 3d polyhedral dice ([f883e56](https://github.com/martynvdijke/dnd/commit/f883e568d668a30348a4bc26229f30d7f8bc8790))

## [1.15.3](https://github.com/martynvdijke/dnd/compare/v1.15.2...v1.15.3) (2026-05-26)


### Bug Fixes

* remove calendar test cases deleted in campaign cleanup ([d4c3bf3](https://github.com/martynvdijke/dnd/commit/d4c3bf3a6e2ef51d393bdc1124a180c36864009b))
* remove orphan shopsNavItem reference that breaks login flow ([52ae0ec](https://github.com/martynvdijke/dnd/commit/52ae0ec9b7efd666451a5cc55b5b6a5f59e4bb59))
* resolve 87 e2e test failures (nav links, API responses, HTMX query, test fixes) ([9440059](https://github.com/martynvdijke/dnd/commit/9440059bdcfd3b3bb1adf44bb20b73fd3f9aa3f5))

## [1.15.2](https://github.com/martynvdijke/dnd/compare/v1.15.1...v1.15.2) (2026-05-25)


### Bug Fixes

* add factionsView to showView, mobile nav handling, checklist timing ([9faf3c3](https://github.com/martynvdijke/dnd/commit/9faf3c3ee797b187e4c33669797aff5204130294))
* clue type constraint (physical_evidence not valid) ([9fd39ee](https://github.com/martynvdijke/dnd/commit/9fd39eee8a8836cf68a12e6e9d5985c2e3bb5590))
* **docker:** include vite.config.ts and use npm run build:ts to bundle app.js ([406755c](https://github.com/martynvdijke/dnd/commit/406755cd1ecbe594eeb47c47506ee5e1bb08f87e))
* improved checklist test timing and htmx initialization ([38a5190](https://github.com/martynvdijke/dnd/commit/38a51900572230be2ecabfa2c6bf4d8f89d2a588))
* invalid timezone UTC+1, use Europe/Amsterdam instead ([274aa2c](https://github.com/martynvdijke/dnd/commit/274aa2c20df3cdd3b9674feeb8eadf6cd69ecac0))
* remove stalePr from renovate.json (no longer valid in Renovate v37) ([ffe6c91](https://github.com/martynvdijke/dnd/commit/ffe6c91dcdff1cf893ec9e00009ea32d729ffa52))
* remove stalePrAge from renovate.json (removed in Renovate v37) ([4387c5b](https://github.com/martynvdijke/dnd/commit/4387c5b70166970785f932ffb0cb41bee8a61e30))
* replace c.HTML() with renderTemplate() in oneshot handlers to fix nil pointer panic ([ac3b751](https://github.com/martynvdijke/dnd/commit/ac3b7516445134eb37d5b1e443022fe3d176132b))
* resolve Playwright test failures (h1 strict mode, version regex, mobile nav) ([4296c96](https://github.com/martynvdijke/dnd/commit/4296c965f99da90c89723b11921fc3ad28f3376c))
* template pipe bugs, pregen double-write, dashboard nil-ptr, checklist target, and test expectations ([6daf197](https://github.com/martynvdijke/dnd/commit/6daf197cae8040e6114877cf3fdcce5c59509954))

## [1.15.1](https://github.com/martynvdijke/dnd/compare/v1.15.0...v1.15.1) (2026-05-24)


### Bug Fixes

* **tests:** correct build tag in playwright config from fts5 to sqlite_fts5 ([2ec075e](https://github.com/martynvdijke/dnd/commit/2ec075e0afa4fcaee018b795d607a560a42c84a8))

# [1.15.0](https://github.com/martynvdijke/dnd/compare/v1.14.0...v1.15.0) (2026-05-24)


### Features

* **oneshot:** add clue/mystery tracker with reveal/hide, dependencies, NPC/location links ([6eab978](https://github.com/martynvdijke/dnd/commit/6eab978bada155628e891b06f85d1dd69839a8f1))
* **oneshot:** add DM screen with quick reference, quick actions, and DM notes ([f05ee9c](https://github.com/martynvdijke/dnd/commit/f05ee9c57856f73eb79a617516e33e73d37bbc6f))
* **oneshot:** add pregenerated characters with quick gen, party balance check ([cf54d7e](https://github.com/martynvdijke/dnd/commit/cf54d7e2c5ae6ba549ebffb3a860a557d3483507))
* **oneshot:** add prep dashboard with checklist, session flow, and all-in-one DM view ([0564711](https://github.com/martynvdijke/dnd/commit/05647110a21b94164fa6f28094f34f1a6fe4dd67))

# [1.14.0](https://github.com/martynvdijke/dnd/compare/v1.13.4...v1.14.0) (2026-05-24)


### Features

* **oneshot:** add one-shot adventure builder with full CRUD, template generation, and tests ([9b620cb](https://github.com/martynvdijke/dnd/commit/9b620cb0556f83b543b90f4d2377bdeba78f98ce))
* **oneshot:** add random generators for hooks, dungeon dressing, taverns, and encounters ([d1012da](https://github.com/martynvdijke/dnd/commit/d1012da0db6e0e1bd9562eae53aa9b41cb35a70a))
* **oneshot:** add session pacing dashboard with timer, scene advance, and alerts ([c5f7230](https://github.com/martynvdijke/dnd/commit/c5f7230ab20036000e9e370b8acb8196259472ac))

## [1.13.4](https://github.com/martynvdijke/dnd/compare/v1.13.3...v1.13.4) (2026-05-24)


### Bug Fixes

* **tech-debt:** api() window assignment bug, duplicate CSS, inline styles ([f6db7b1](https://github.com/martynvdijke/dnd/commit/f6db7b1d029ac110fda0deed2f51cd8fb03540d8))

## [1.13.3](https://github.com/martynvdijke/dnd/compare/v1.13.2...v1.13.3) (2026-05-23)


### Bug Fixes

* **ui:** add aria-hidden to icons, fix FA version to 6.7.2 ([ee04cce](https://github.com/martynvdijke/dnd/commit/ee04cce127d1e9696f854e24d787e261778f79ee))

## [1.13.2](https://github.com/martynvdijke/dnd/compare/v1.13.1...v1.13.2) (2026-05-23)


### Bug Fixes

* **deps:** update dependency marked to v18 ([f8cd83c](https://github.com/martynvdijke/dnd/commit/f8cd83c821a61ed880825f1234d69dfe573f1be3))
* **deps:** update tiptap monorepo to v3 ([f3edffd](https://github.com/martynvdijke/dnd/commit/f3edffd9b4db93ba8e98886320e567bdc893e9f2))

## [1.13.1](https://github.com/martynvdijke/dnd/compare/v1.13.0...v1.13.1) (2026-05-22)


### Bug Fixes

* **deps:** update all non-major dependencies ([#12](https://github.com/martynvdijke/dnd/issues/12)) ([cdce3f9](https://github.com/martynvdijke/dnd/commit/cdce3f941dbd18e3d952e7baf6ef1d50ac91de57))

# [1.13.0](https://github.com/martynvdijke/dnd/compare/v1.12.2...v1.13.0) (2026-05-21)


### Bug Fixes

* playwright tests - expose bootstrap global for HTMX modals, fix journal TipTap editor interaction ([3631653](https://github.com/martynvdijke/dnd/commit/3631653c4ef1aef54d1e4f3f87dcd28f68b0ec1a)), closes [#journalEditor](https://github.com/martynvdijke/dnd/issues/journalEditor) [#journalEntry](https://github.com/martynvdijke/dnd/issues/journalEntry)
* resolve Playwright test failures ([02ae149](https://github.com/martynvdijke/dnd/commit/02ae149f34e402c579841fe5921c40ee6475a39e))
* resolve Playwright test failures - race conditions, build script, and timeouts ([b88ef2c](https://github.com/martynvdijke/dnd/commit/b88ef2c49cdf7f23acb78a4236ab52085be9deb4))


### Features

* modernize build, graph, and journal/wiki UX ([2df3422](https://github.com/martynvdijke/dnd/commit/2df34225572f17c484aed0f4d4aa0f2deb5dc99f))

## [1.12.2](https://github.com/martynvdijke/dnd/compare/v1.12.1...v1.12.2) (2026-05-20)


### Bug Fixes

* ensure Gotify notification always fires on release workflow ([6a1f471](https://github.com/martynvdijke/dnd/commit/6a1f47189a113575a8311175216a9955daad78a2))

## [1.12.1](https://github.com/martynvdijke/dnd/compare/v1.12.0...v1.12.1) (2026-05-20)


### Bug Fixes

* correct CI build tags for FTS5 support ([58b1abe](https://github.com/martynvdijke/dnd/commit/58b1abe71d163eb9f3cfa47316d8ca027cbebbe9))

# [1.12.0](https://github.com/martynvdijke/dnd/compare/v1.11.0...v1.12.0) (2026-05-20)


### Bug Fixes

* add TypeScript build step before go vet in CI ([de31331](https://github.com/martynvdijke/dnd/commit/de31331a5d44e11f152edf9300cd1ed513741a88))
* align Ent table names with migration schemas and fix test ([5f455c1](https://github.com/martynvdijke/dnd/commit/5f455c1ea8d5c2920a546ef204f6c31f4a7168f7))
* align htmx form templates with test expectations and fix modal interactions ([02797e1](https://github.com/martynvdijke/dnd/commit/02797e1b18015290e54a8b7ae04f46a93763d5a9))
* change inventory form button from 'Add' to 'Add Item' to match test expectation ([c134320](https://github.com/martynvdijke/dnd/commit/c1343208c4a784ba33cb03b89a0c42599da07b9c))
* change inventory form button text from 'Add' to 'Add Item' to match test expectation ([c3be71f](https://github.com/martynvdijke/dnd/commit/c3be71f57c53c11e1217e086c176541352d51cb9))
* reduce flaky test failures from modal backdrop and parallel contention ([6b4072b](https://github.com/martynvdijke/dnd/commit/6b4072b5c119e19902938b619f7ab5718d43a966))
* restore test suite - table names, swap mode, and tab rendering ([2bb2a52](https://github.com/martynvdijke/dnd/commit/2bb2a523fc90ec580e913351d99eb1eb8c15dcc8))


### Features

* integrate ent ORM alongside gin and migrate core handlers ([619b3aa](https://github.com/martynvdijke/dnd/commit/619b3aa97495f02f988d1eb2acc2591929a2fb51))

# [1.11.0](https://github.com/martynvdijke/dnd/compare/v1.10.0...v1.11.0) (2026-05-18)


### Features

* Add action/combat log with statistics ([a84c066](https://github.com/martynvdijke/dnd/commit/a84c066e7d6c6fdd5283bcb0ca9378784e34e46d))
* Add all feature routes and comprehensive tests ([39cb9f1](https://github.com/martynvdijke/dnd/commit/39cb9f11567e0ab1ad772254f40c53edb27973e9))
* Add campaign dashboard with aggregated overview data ([392d466](https://github.com/martynvdijke/dnd/commit/392d46601111be65d79862ed9369fc738c197232))
* Add campaign recap generator with auto-generation ([73bf798](https://github.com/martynvdijke/dnd/commit/73bf7988cbf24d7b36c96f9413301b267e47bea8))
* Add downtime activity tracker with day advancement ([19fbb09](https://github.com/martynvdijke/dnd/commit/19fbb09e3a1593181478592858e3a11b8f8058df))
* Add homebrew content manager for custom game content ([2ce09ee](https://github.com/martynvdijke/dnd/commit/2ce09ee63ac8ecdf072254d897bf99141e05be38))
* Add interactive campaign world map with pins and fog of war ([a913fc0](https://github.com/martynvdijke/dnd/commit/a913fc07e1ce7f9f7ce4597f3cb687fc24ee8d79))
* Add level-up planner with build tree visualization ([6da8609](https://github.com/martynvdijke/dnd/commit/6da86092d1ac668fdd6c421057403683038c5986))
* add opentelemetry tracing and prometheus metrics support ([a1004a3](https://github.com/martynvdijke/dnd/commit/a1004a34a648adbb8939373b56779a10e0965e5a))
* Add quick reference panel for D&D rules ([8df36d5](https://github.com/martynvdijke/dnd/commit/8df36d594316ba516821752b8d047c45dac8bfa5))
* Add spell slot & resource tracker with rest recovery ([77c8fac](https://github.com/martynvdijke/dnd/commit/77c8fac6ffb5f433c75fa20b0fd2b954ca3f113b))

# [1.10.0](https://github.com/martynvdijke/dnd/compare/v1.9.0...v1.10.0) (2026-05-15)


### Features

* add sharable character/party links and built-in email sending ([d0aa613](https://github.com/martynvdijke/dnd/commit/d0aa6134c577ce68f4016a37e224d59e11f5f137))

# [1.9.0](https://github.com/martynvdijke/dnd/compare/v1.8.0...v1.9.0) (2026-05-15)


### Features

* add campaign-level graph showing all entities and relationships using vis-network ([41298f1](https://github.com/martynvdijke/dnd/commit/41298f1cc6a9b36219bacf2e8394e13ea2a562c2))

# [1.8.0](https://github.com/martynvdijke/dnd/compare/v1.7.0...v1.8.0) (2026-05-14)


### Bug Fixes

* resolve Playwright test failures on CI, ensure parallel execution safety ([147c78f](https://github.com/martynvdijke/dnd/commit/147c78f3cada64115488316d26010566fe398a05)), closes [#tabBar](https://github.com/martynvdijke/dnd/issues/tabBar)


### Features

* add campaign-level combat tracker with drag-and-drop and inline actions ([6a93ec7](https://github.com/martynvdijke/dnd/commit/6a93ec772275e127aacd05ed07031bc20a5bc40d))
* add crafting system with recipes and progress tracking ([4e3f7e9](https://github.com/martynvdijke/dnd/commit/4e3f7e9e2f568c7650f24d7428074c7349e86a72))
* add global advanced search across all content types ([8954564](https://github.com/martynvdijke/dnd/commit/8954564a8db2d836f1d542cf68263aae6626c73e))
* add shops/trading system and campaign wiki with markdown ([9cda7f0](https://github.com/martynvdijke/dnd/commit/9cda7f04719d4431fb5f733714f795be5dd262e0))

# [1.7.0](https://github.com/martynvdijke/dnd/compare/v1.6.0...v1.7.0) (2026-05-13)


### Bug Fixes

* update release prepareCmd to also update Version in main.go ([4b87eee](https://github.com/martynvdijke/dnd/commit/4b87eeeb64dee8a3a46b1b5b65866a976b440b37))


### Features

* add WebSocket live updates, feat edit, auto-calc proficiency/perception/spell DC ([a0e4dcb](https://github.com/martynvdijke/dnd/commit/a0e4dcb42bc4d0ca54e0d241db8513bb84525a8a))

# [1.6.0](https://github.com/martynvdijke/dnd/compare/v1.5.1...v1.6.0) (2026-05-13)


### Features

* add party naming to campaigns with party_name field ([eb3633d](https://github.com/martynvdijke/dnd/commit/eb3633d62f1f3f12be126fab9149afdbd9e57cf9))

## [1.5.1](https://github.com/martynvdijke/dnd/compare/v1.5.0...v1.5.1) (2026-05-13)

# [1.5.0](https://github.com/martynvdijke/dnd/compare/v1.4.0...v1.5.0) (2026-05-12)


### Features

* add conditions tracker, concentration checks, feats, companions, factions, weather, notes, HP calc, random char gen, character comparison ([5fefc60](https://github.com/martynvdijke/dnd/commit/5fefc608ca0e6d7823824a9bcfaa773bcd00304c))
* add frontend UI for conditions, concentration, feats, companions, factions, weather, notes, HP calc, random char, comparison ([a979de2](https://github.com/martynvdijke/dnd/commit/a979de26ff8c4e86e435633fa08f29fa33814e01))

# [1.4.0](https://github.com/martynvdijke/dnd/compare/v1.3.2...v1.4.0) (2026-05-12)


### Bug Fixes

* correct TestSpellcasting copy-paste bug and add missing fts5 build tag to test script ([51f3a2b](https://github.com/martynvdijke/dnd/commit/51f3a2b4836253aec0606aface84e6a2d808ffeb))
* spellcasting field names, add JSON seed loading, D&D API fallback, and TTRPG system/source support ([5ae5cca](https://github.com/martynvdijke/dnd/commit/5ae5cca126b1a317ab76f6e3847a0ed1b536503f))
* update test selectors for locations UI changes and add loading overlay CSS ([3ec4bc1](https://github.com/martynvdijke/dnd/commit/3ec4bc174023a09fb9eec21be094c52a59410fbf))


### Features

* add campaign members with DM role, email settings UI, footer version, backup improvements, fix advantage/disadvantage ([c58d32f](https://github.com/martynvdijke/dnd/commit/c58d32f85edc944ead93ccf350a82015863805e4))
* add frontend UI for portrait, multi-class, encounter builder, calendar, and timeline ([98db31b](https://github.com/martynvdijke/dnd/commit/98db31b3875f416265e57c4c5291a759fa4a7e26))
* add Leaflet map view for locations with interactive markers and sidebar ([7f903e1](https://github.com/martynvdijke/dnd/commit/7f903e109762483a436c2e6a13d6d321c50048bb))
* add portrait upload, multi-classing, encounter builder, calendar, and timeline features ([f40fd05](https://github.com/martynvdijke/dnd/commit/f40fd05f71c2004dcb33c0e1fc908d038e0840b2))

## [1.3.2](https://github.com/martynvdijke/dnd/compare/v1.3.1...v1.3.2) (2026-05-12)


### Bug Fixes

* default DB_PATH to /db/villum.db when running in Docker ([3b7c36a](https://github.com/martynvdijke/dnd/commit/3b7c36aee8f5a3c19edce5fa5deee56812686d9b))

## [1.3.1](https://github.com/martynvdijke/dnd/compare/v1.3.0...v1.3.1) (2026-05-12)

# [1.3.0](https://github.com/martynvdijke/dnd/compare/v1.2.1...v1.3.0) (2026-05-11)


### Features

* add 3D dice animation, media upload support, email settings, and cleanup tasks ([332f573](https://github.com/martynvdijke/dnd/commit/332f57321493a24e809b167e60f54b43cfc2ce0b))

## [1.2.1](https://github.com/martynvdijke/dnd/compare/v1.2.0...v1.2.1) (2026-05-11)


### Bug Fixes

* strip export {} from compiled JS output for TS6 compatibility ([33f3422](https://github.com/martynvdijke/dnd/commit/33f3422030abb2ea018e299691630424b4efdfab))

# [1.2.0](https://github.com/martynvdijke/dnd/compare/v1.1.0...v1.2.0) (2026-05-11)


### Features

* add email field to user model for user invite flow ([ee20bb5](https://github.com/martynvdijke/dnd/commit/ee20bb59984b4cb86c0cb89fa61781903a828f4e))

# [1.1.0](https://github.com/martynvdijke/dnd/compare/v1.0.2...v1.1.0) (2026-05-11)


### Bug Fixes

* correct ability scores test selector and stabilize sequential dice roll test ([7fa1698](https://github.com/martynvdijke/dnd/commit/7fa169855c95dfa78001adeb67cfc793184ce266))
* prevent loading overlay from blocking modal interactions and stabilize flaky NPC test ([2b843a3](https://github.com/martynvdijke/dnd/commit/2b843a3908fe7f3963fa43051aed9528a72a7b3d))


### Features

* add animated dice with nat 20/1 highlighting, floating action button for quick actions, and sort utility ([9d5ebb6](https://github.com/martynvdijke/dnd/commit/9d5ebb6cc7830f15a79434c5a261b5d8b4c5c03f))
* add dark mode toggle with localStorage persistence across all pages ([15dd741](https://github.com/martynvdijke/dnd/commit/15dd741ac40670975bc8fbd09048fc337b20b377))
* add debounced auto-save for inline edit fields (ability scores, HP, details, etc.) ([7f75d43](https://github.com/martynvdijke/dnd/commit/7f75d433d89f82ecf5e8dfe28efbe3664656f689))
* add keyboard shortcuts (n/d/p/c/1-9/?/Esc/T) and character search bar ([509de5a](https://github.com/martynvdijke/dnd/commit/509de5a725630584bcf3908a568f119d29e34bc4))
* add loading overlay spinner during API calls ([730bdaa](https://github.com/martynvdijke/dnd/commit/730bdaa02cc867098dc40ed124dcc02b05752aa5))
* add tooltips on hover for ability scores, skill calculations, and combat stats ([36a0219](https://github.com/martynvdijke/dnd/commit/36a021970366be6fc35b84165dd10be8c9687458))
* add XP progress bar showing progress to next level ([4bef237](https://github.com/martynvdijke/dnd/commit/4bef237023e67a782312268b53c27392b888b807))
* improve dice roller UI with quick select dice, adv/dis buttons, and visual die faces ([f241ffa](https://github.com/martynvdijke/dnd/commit/f241ffa9ed74825bcb5e7d2e7d704f53e094a416))
* improve empty states with consistent icons and helpful messages ([f19ad8d](https://github.com/martynvdijke/dnd/commit/f19ad8daefbd094e51852777cfba46148bd91db8))
* improve spell list layout with card-style display and full spell details ([0c63bb5](https://github.com/martynvdijke/dnd/commit/0c63bb5047733420e3381485fdc15e6d7cd34d6e))
* make ability scores (str/dex/con/int/wis/cha) editable from stats tab ([9351bc9](https://github.com/martynvdijke/dnd/commit/9351bc92eed95a1aa072b099cb2d9c88e86dded1))

## [1.0.2](https://github.com/martynvdijke/dnd/compare/v1.0.1...v1.0.2) (2026-05-10)


### Bug Fixes

* rename Docker image references from traces to dnd in release workflow ([c97ee19](https://github.com/martynvdijke/dnd/commit/c97ee19a0bac7a2db6d475281d4c262d9454fbcb))

## [1.0.1](https://github.com/martynvdijke/dnd/compare/v1.0.0...v1.0.1) (2026-05-10)

# 1.0.0 (2026-05-10)


### Bug Fixes

* auth checks, level-up hit dice, and add comprehensive tests ([f233639](https://github.com/martynvdijke/dnd/commit/f233639ce181ce2721e99ab62590c1619b7723be))
* e2e test suite - 110→8 failures (160/178 passing) ([a02962f](https://github.com/martynvdijke/dnd/commit/a02962f7c20012dd6b1d9c771445eede6399910f)), closes [#diceSection](https://github.com/martynvdijke/dnd/issues/diceSection)
* fixes ([6d63eab](https://github.com/martynvdijke/dnd/commit/6d63eab4837deaac0c05700a663c0106c96ca33c))
* handle mobile nav collapse, Firefox NS_BINDING_ABORTED, and compendium dupes in e2e tests ([026d478](https://github.com/martynvdijke/dnd/commit/026d478d91522b4124fea464826956e733b8d5e5))
* more fixes ([ca1ffa2](https://github.com/martynvdijke/dnd/commit/ca1ffa2b2d0a9840a7139c020d39552fadb31ffc))
* more fixes v2 ([a10ae8c](https://github.com/martynvdijke/dnd/commit/a10ae8c5cbb0a936605023ee6efcaf64832b6140))
* playwright test config and tests, update stale-branches workflow ([4413178](https://github.com/martynvdijke/dnd/commit/4413178a431ca7c12def3ced5c0682f9642e4f20))
* playwright tests failing in CI due to export {}; in compiled JS and waitForURL mismatch ([3c9afb0](https://github.com/martynvdijke/dnd/commit/3c9afb0dfb61e37fc41134357a9d976fdd968346))
* replace waitForTimeout with condition-based waits in dice/delete e2e tests ([0683426](https://github.com/martynvdijke/dnd/commit/0683426449b1e735194ba900dd62fc5623241155))
* replace waitForTimeout with waitFor in death save / concentration tests for mobile modal backdrop ([44f0ee1](https://github.com/martynvdijke/dnd/commit/44f0ee18901cd0646d08711b9669154b2dd67b5b))
* sync TS source with compiled JS for dice view, history class, and deleteChar ([fffd248](https://github.com/martynvdijke/dnd/commit/fffd2483ad36c5226df1d865e60a7c4caaaffeb2))
* update release config to use correct file paths ([a1587da](https://github.com/martynvdijke/dnd/commit/a1587da1a9862a1f3bd1d9484fc475a88e1f0ef5))
* update test selectors to match HTML structure and fix strict mode violations ([f41fdc5](https://github.com/martynvdijke/dnd/commit/f41fdc5bf9b7cfbba373e808b47184a3b01f6280))
* use domcontentloaded for page navigations in setup tests to avoid CDN timeout ([65c5001](https://github.com/martynvdijke/dnd/commit/65c5001c777d9aa0dc8acb723526a6aa53343880))
* wait for sheet and character card rendering in e2e tests ([6849b3a](https://github.com/martynvdijke/dnd/commit/6849b3a6a86718c9b45d981172b13665bced817a))


### Features

* Bootstrap 5.3 + Font Awesome UI overhaul ([1cb7254](https://github.com/martynvdijke/dnd/commit/1cb7254ecf0fc7618178d3d7ca2d7b1c40142b2a))
* Phase 1.1 - Automated ability score modifiers ([aa6462a](https://github.com/martynvdijke/dnd/commit/aa6462a1b3eaff406bbba00bd7e570cb97dd43c5))
* Phase 1.2 - Skill/save/ability check roller with advantage ([89702e3](https://github.com/martynvdijke/dnd/commit/89702e328169e8a0f58d825cad2571d8736a954b))
* Phase 1.3 - Initiative tracker and combat system ([577d469](https://github.com/martynvdijke/dnd/commit/577d46994cb3a62dd9af3e7f4e10e27a0c5eeda3))
* Phase 1.4 - Hit dice manager with individual spending ([64feef2](https://github.com/martynvdijke/dnd/commit/64feef21ab6d24c23c6a724549072acf40281925))
* Phase 2.1 - Full SRD spell list (254 spells) ([a7eafe4](https://github.com/martynvdijke/dnd/commit/a7eafe439c488efa7b900cb7cda3bad7afe40ab3))
* Phase 2.2-2.3 - Random generators and magic items ([7e07ae0](https://github.com/martynvdijke/dnd/commit/7e07ae052f110ad0f56d22363921593ddf8e98cc))
* Phase 2.2-7 - Generators, health/metrics, UI polish ([6de1979](https://github.com/martynvdijke/dnd/commit/6de19790815f6edf3099022be245f2c7b3340eac))
