## [2.17.3](https://github.com/martynvdijke/dnd/compare/v2.17.2...v2.17.3) (2026-06-17)

## [2.17.2](https://github.com/martynvdijke/dnd/compare/v2.17.1...v2.17.2) (2026-06-16)


### Bug Fixes

* remove concurrency block from reusable CI workflow [skip ci] ([6c85b0c](https://github.com/martynvdijke/dnd/commit/6c85b0c08b6724186124e8e76486f2396644578f))
* remove dimension-based 3D dice check and duplicate CI push trigger ([b437928](https://github.com/martynvdijke/dnd/commit/b4379287b20840750cab6bbfd759951f56e34053))
* trigger release with corrected CI workflow (remove concurrency from reusable workflow) ([ea7d665](https://github.com/martynvdijke/dnd/commit/ea7d665d85c9fb2a9e9715690a37c475a15e2a7f))

## [2.17.1](https://github.com/martynvdijke/dnd/compare/v2.17.0...v2.17.1) (2026-06-16)


### Bug Fixes

* CI stabilization — per-worker port cleanup, login timeout hardening, smoke test race fixes ([aed0854](https://github.com/martynvdijke/dnd/commit/aed0854ae94861cb488559727f6a26c1b70eb736)), closes [#charGrid](https://github.com/martynvdijke/dnd/issues/charGrid)

# [2.17.0](https://github.com/martynvdijke/dnd/compare/v2.16.6...v2.17.0) (2026-06-16)


### Bug Fixes

* add .first() selectors and test.slow() for Playwright retry resilience ([78daac4](https://github.com/martynvdijke/dnd/commit/78daac4fcd548ab2433f13a10726683ab1365ef4))
* add sqlite_fts5 build tag to Dockerfile ([017a3e4](https://github.com/martynvdijke/dnd/commit/017a3e4fb00a874becbb9678045f82ae8ecdca38))
* add TypeScript type annotations to test files ([4f72ebf](https://github.com/martynvdijke/dnd/commit/4f72ebf07237758dc08f64fe918a306cb50bdb3c))
* limit coverage measurement to tested packages via -coverpkg ([b326b19](https://github.com/martynvdijke/dnd/commit/b326b1961e1246207f81fa4dc500e5fce7bce216))
* per-worker DB isolation for parallel Playwright tests ([36469f5](https://github.com/martynvdijke/dnd/commit/36469f59d4938c6e646476826156a35b89ccafae))
* release sed pattern for health.go whitespace and add .dockerignore ([3a26217](https://github.com/martynvdijke/dnd/commit/3a2621715010c28e8bef5ddeeb17affe65752963))
* remove explicit timeouts from oneshot waitForFunction/expect — test.slow() handles it ([84faf21](https://github.com/martynvdijke/dnd/commit/84faf21080a3ab3a4785fddbddf7e4d91b3130f8))
* target specific schema tabs/content by type_name to avoid retry duplicates ([dfde5ec](https://github.com/martynvdijke/dnd/commit/dfde5ec8095cd557d5518ab028438e72e2eaa215))
* use single worker in CI to prevent parallel DB corruption; remove fragile schema tab screenshot ([4b81f58](https://github.com/martynvdijke/dnd/commit/4b81f58567356cc8dfb844bad0e41c65eee33442))


### Features

* show imported schema entries in compendium view ([7c6258c](https://github.com/martynvdijke/dnd/commit/7c6258c80167c33cafbfa8df77a053f4206ea773))
* spell compendium browse with HTMX filtering, pagination, and detail expansion ([fe44889](https://github.com/martynvdijke/dnd/commit/fe448892383f605b88b203ff7055259134a03879))

## [2.16.6](https://github.com/martynvdijke/dnd/compare/v2.16.5...v2.16.6) (2026-06-14)


### Bug Fixes

* run go mod tidy to clean up stale dependencies ([a76da9c](https://github.com/martynvdijke/dnd/commit/a76da9c3626e5337f4ebb79b37e5720236ddd4b6))

## [2.16.5](https://github.com/martynvdijke/dnd/compare/v2.16.4...v2.16.5) (2026-06-14)


### Bug Fixes

* extracted modules not imported, state split between app.ts and state.ts ([504df81](https://github.com/martynvdijke/dnd/commit/504df81c00e03739527857d50c8e1465f5cc615c))
* missing currentUser/currentChar imports in party.ts, fix campaign.go indent ([e52f742](https://github.com/martynvdijke/dnd/commit/e52f742299f6c5926dfa08b8ff6ab308208da69c))
* replace hand-rolled dice with @dice-roller/rpg-dice-roller via API ([36ef409](https://github.com/martynvdijke/dnd/commit/36ef409ef8eee3b0735bd1e3b1de85bca919d870))

## [2.16.4](https://github.com/martynvdijke/dnd/compare/v2.16.3...v2.16.4) (2026-06-14)


### Bug Fixes

* expose api() on window so E2E tests can find and wait for SPA initialization ([6a1d83e](https://github.com/martynvdijke/dnd/commit/6a1d83e3a210f8a5eb98398206445cc46758ec23))
* increase Playwright test timeout to 30 min (15 min was too short) ([284edfc](https://github.com/martynvdijke/dnd/commit/284edfc595ea83547694f1b7345bf7ebaae69c63))
* restore window.getCurrentView/setCurrentView for keyboard shortcuts, fix responsive test breakpoint ([727ad1f](https://github.com/martynvdijke/dnd/commit/727ad1f13e54997dee668e3765d54e504b488e72))
* run full chromium test suite in CI with 15-min timeout ([19407ee](https://github.com/martynvdijke/dnd/commit/19407ee18a81546ec53315de64418d660d9fa88f))

## [2.16.3](https://github.com/martynvdijke/dnd/compare/v2.16.2...v2.16.3) (2026-06-14)


### Bug Fixes

* add onsubmit='return false' to setup/login forms preventing native GET submission ([3238531](https://github.com/martynvdijke/dnd/commit/3238531745ff28cced9f306a91f465e4b19ec11b))
* always use 'go run' in Playwright webServer (skip stale pre-built binary) ([2dcb99e](https://github.com/martynvdijke/dnd/commit/2dcb99eefb0d5057c9e396d9f468c7f882a02411))
* always use go run in Playwright webServer (skip hanging ./villum-server binary) ([b67a828](https://github.com/martynvdijke/dnd/commit/b67a82885e22abcf52f100cff74e223a13e3c457))
* e2e setup form race condition + pre-built binary for CI ([f379b7e](https://github.com/martynvdijke/dnd/commit/f379b7edd522dd490993314501cfb2480547604c))
* prevent race condition in setup/login form handler attachment ([1354398](https://github.com/martynvdijke/dnd/commit/135439829459641f8a116b61076ec56de220243a))
* rebuild setup.js/login.js with handler-first fix + add setup/login to build:ts ([531d0eb](https://github.com/martynvdijke/dnd/commit/531d0eb7672dcde757a572a02f8e1805b71862c9))
* restore pre-built binary check in Playwright webServer config ([cf286fe](https://github.com/martynvdijke/dnd/commit/cf286fe3afdd2f82cc107f3e7d1772a949b4ab89))
* run only setup tests in CI to verify release fix, use sqlite_fts5 tag ([978f70f](https://github.com/martynvdijke/dnd/commit/978f70f86898d970d9fda561e327e386021c055c))
* server-side /login → /setup redirect when no admin exists ([6ea81c2](https://github.com/martynvdijke/dnd/commit/6ea81c277e04b60c862bfb4c5579ba2b389e03d2))
* stop tsc from overwriting Vite IIFE output with ESM (root cause of setup/login CI failure) ([993901b](https://github.com/martynvdijke/dnd/commit/993901b474c4986681a5f688c6c191363f7b6879))
* suppress pre-existing tsc errors now that modules import from app.ts ([46ba2d1](https://github.com/martynvdijke/dnd/commit/46ba2d1cef6bc5f3af1b981407ebaed9f45a0b55))
* use consistent sqlite_fts5 tag for CI build, skip mobile-chrome to reduce test time ([6daa257](https://github.com/martynvdijke/dnd/commit/6daa257b505b5384042d09c0884489dc33f5747c))

## [2.16.2](https://github.com/martynvdijke/dnd/compare/v2.16.1...v2.16.2) (2026-06-12)

## [2.16.1](https://github.com/martynvdijke/dnd/compare/v2.16.0...v2.16.1) (2026-06-12)


### Bug Fixes

* **deps:** update all non-major dependencies ([#29](https://github.com/martynvdijke/dnd/issues/29)) ([7496d98](https://github.com/martynvdijke/dnd/commit/7496d98e70c95afe03559704b2fad46ab83b4d94))

# [2.16.0](https://github.com/martynvdijke/dnd/compare/v2.15.1...v2.16.0) (2026-06-11)


### Bug Fixes

* reduce flaky e2e tests by improving login wait reliability ([19991ac](https://github.com/martynvdijke/dnd/commit/19991acb6aa7192143f9fb1872f57e533af11104))


### Features

* add structured logging to external dependency handlers ([72c0427](https://github.com/martynvdijke/dnd/commit/72c042759b17853945a6a7b4cab8ad066ea413bc)), closes [#external-dependency-logging](https://github.com/martynvdijke/dnd/issues/external-dependency-logging)

## [2.15.1](https://github.com/martynvdijke/dnd/compare/v2.15.0...v2.15.1) (2026-06-11)


### Bug Fixes

* **ui:** render compendium pickers as modal body, not nested modal ([862675e](https://github.com/martynvdijke/dnd/commit/862675e014b87def564234b6f9bdd7b3d5c8864a)), closes [#genericModalBody](https://github.com/martynvdijke/dnd/issues/genericModalBody)

# [2.15.0](https://github.com/martynvdijke/dnd/compare/v2.14.3...v2.15.0) (2026-06-11)


### Bug Fixes

* use consistent Gin route param names to avoid panic ([23bdfcb](https://github.com/martynvdijke/dnd/commit/23bdfcb1e668e82b89415192edfe81450545fa15))


### Features

* compendium linking for spells, items, monsters ([f94fbf5](https://github.com/martynvdijke/dnd/commit/f94fbf500cbccc1157e8ec5ccb27dcd36a530aaf))

## [2.14.3](https://github.com/martynvdijke/dnd/compare/v2.14.2...v2.14.3) (2026-06-10)


### Bug Fixes

* update tests to use new stepper selectors after sheet revamp ([d4882c3](https://github.com/martynvdijke/dnd/commit/d4882c349ec3f3177b825c0da07893d7b23abf3f))
* update tests to use unified compendium API and fix admin panel navigation ([e497416](https://github.com/martynvdijke/dnd/commit/e497416390037afb01f7ed691997ac5c22c19d28))

## [2.14.2](https://github.com/martynvdijke/dnd/compare/v2.14.1...v2.14.2) (2026-06-10)


### Bug Fixes

* **deps:** update github.com/dop251/goja digest to 348e6be ([#28](https://github.com/martynvdijke/dnd/issues/28)) ([c6cfa4c](https://github.com/martynvdijke/dnd/commit/c6cfa4ce17f75846f3d3dce42e30f2b94027bf7c))

## [2.14.1](https://github.com/martynvdijke/dnd/compare/v2.14.0...v2.14.1) (2026-06-10)


### Bug Fixes

* flaky 3D dice test and compendium bulk selection on mobile ([e589d93](https://github.com/martynvdijke/dnd/commit/e589d93ffdd228159b0aecddde35977a4b360b88))
* reduce CI workers, fix combat tracker mobile, fix wiki mobile tests ([b32ad80](https://github.com/martynvdijke/dnd/commit/b32ad80b50ef4bf1bb6be011ea8636d1d64bf9f7))

# [2.14.0](https://github.com/martynvdijke/dnd/compare/v2.13.0...v2.14.0) (2026-06-09)


### Bug Fixes

* add actions:read and checks:read perms to ci job for otel in reusable workflow ([8a85b57](https://github.com/martynvdijke/dnd/commit/8a85b57225eec290fb535120bcd247500cb5067e))
* delete child records before character to satisfy FK constraints, add legacy admin compendium routes ([b268d05](https://github.com/martynvdijke/dnd/commit/b268d05dc7a9b1c6ae7a99d16deecbfb146e1e06)), closes [#shopsGrid](https://github.com/martynvdijke/dnd/issues/shopsGrid) [#shopSelect](https://github.com/martynvdijke/dnd/issues/shopSelect)
* make test PNG content unique to avoid dedup hash collision ([ab57ded](https://github.com/martynvdijke/dnd/commit/ab57dedf1e7e3a84e568352ef29b5d31892b8140))
* rename githubToken to otelToken for otel-cicd-action@v4 ([dcfb0c6](https://github.com/martynvdijke/dnd/commit/dcfb0c6f1c1a940b95d7d61518e1cda8bdb85c29))
* replace gin.Logger() with RequestLogger() middleware that logs to AppLog ([31b3b64](https://github.com/martynvdijke/dnd/commit/31b3b642ae2f07ce2a5f963b594c22a5bcb33498))
* use ESM import for zlib instead of require in Playwright test ([c15f136](https://github.com/martynvdijke/dnd/commit/c15f136b5bfd79c5d67d2db0e787be49da2a75e7))
* use githubToken instead of otelToken for otel-cicd-action@v4 ([6737297](https://github.com/martynvdijke/dnd/commit/6737297e2e182de09eaddc3d528cef0940c03582))
* use githubToken instead of otelToken for otel-cicd-action@v4 input ([db6d091](https://github.com/martynvdijke/dnd/commit/db6d0919b6a6c981e7e1af721498bb4de4e8ac33))


### Features

* add otlpAuthorization input for Bearer auth ([f214d3a](https://github.com/martynvdijke/dnd/commit/f214d3a973d0bf125c25d003a7033fcaa90b7427))
* **compendium:** add DM-scoped search, linking handlers, and remove legacy admin routes ([e69d6ab](https://github.com/martynvdijke/dnd/commit/e69d6abf9c82e2dbca99510a6e29ef70405bbcd4))
* full shop system with DM management and trading UI ([b139d98](https://github.com/martynvdijke/dnd/commit/b139d984a75566dc7938cf5c890e4c619f643297))

# [2.13.0](https://github.com/martynvdijke/dnd/compare/v2.12.1...v2.13.0) (2026-06-07)


### Bug Fixes

* update admin compendium E2E test for unified UI ([72dc4dc](https://github.com/martynvdijke/dnd/commit/72dc4dcc2a2d9ab78f1fa5c1a443024a688cf4cb))


### Features

* unified compendium CRUD interface ([4a8311f](https://github.com/martynvdijke/dnd/commit/4a8311fee6fe329f87f44704f6c97a269510982d))

## [2.12.1](https://github.com/martynvdijke/dnd/compare/v2.12.0...v2.12.1) (2026-06-06)


### Bug Fixes

* wire OTel settings tab in admin panel frontend ([3de94b4](https://github.com/martynvdijke/dnd/commit/3de94b445128ea93aeab96aa600fef85cb73ab5b))

# [2.12.0](https://github.com/martynvdijke/dnd/compare/v2.11.0...v2.12.0) (2026-06-06)


### Bug Fixes

* resolve 46 E2E test failures across chromium, firefox, and mobile-chrome ([6337d3a](https://github.com/martynvdijke/dnd/commit/6337d3abbc1f7493ca89067b6aec5eb7088bd0aa))


### Features

* add OTel endpoint admin configuration with DB-backed settings ([4dbe09a](https://github.com/martynvdijke/dnd/commit/4dbe09a3bd522847666e4985e9246a130f1f0759))
* central admin logging with OTel export ([5bd9b09](https://github.com/martynvdijke/dnd/commit/5bd9b09712717c6433f67cd8e0569893fe226080))

# [2.11.0](https://github.com/martynvdijke/dnd/compare/v2.10.0...v2.11.0) (2026-06-06)


### Features

* add self-hosted Umami analytics integration ([66e2e09](https://github.com/martynvdijke/dnd/commit/66e2e09133c0c439b4ab789de05d8a3232c5eb7a))

# [2.10.0](https://github.com/martynvdijke/dnd/compare/v2.9.6...v2.10.0) (2026-06-05)


### Features

* add OpenTelemetry tracing and metrics support ([f40cfef](https://github.com/martynvdijke/dnd/commit/f40cfefbd4a62058c2e5c1f9433ca825adf15d79))

## [2.9.6](https://github.com/martynvdijke/dnd/compare/v2.9.5...v2.9.6) (2026-06-05)


### Bug Fixes

* **deps:** update all non-major dependencies ([#26](https://github.com/martynvdijke/dnd/issues/26)) ([c8cd68e](https://github.com/martynvdijke/dnd/commit/c8cd68ed97f5695252ef13e9f5402ac9ca1dbc9f))

## [2.9.5](https://github.com/martynvdijke/dnd/compare/v2.9.4...v2.9.5) (2026-06-05)


### Bug Fixes

* handle omitempty null/undefined and uniqueName scope in e2e tests ([024499c](https://github.com/martynvdijke/dnd/commit/024499c9e40fbeeb6a2d1bb06baeb57be5aa5106))
* use unique filenames in upload tests to prevent duplicate collision in parallel runs ([2e1e221](https://github.com/martynvdijke/dnd/commit/2e1e221d5a6d623a20fb7dc4df4490d971bac57d))

## [2.9.4](https://github.com/martynvdijke/dnd/compare/v2.9.3...v2.9.4) (2026-06-04)


### Bug Fixes

* remove _fk DSN param, keep explicit PRAGMA foreign_keys=ON ([6083f3d](https://github.com/martynvdijke/dnd/commit/6083f3d15178f9daf5818e3ded4adfa4d26a13e7))
* run safe ALTER TABLE after ent.Schema.Create to avoid column loss ([5891c43](https://github.com/martynvdijke/dnd/commit/5891c4317456f26bc896a95504ac0db5c28d382a))
* update SeedCompendiumSchemas to UPDATE existing schemas (not INSERT OR IGNORE) so new fields apply to existing database schemas ([e421e2b](https://github.com/martynvdijke/dnd/commit/e421e2b683cb84666f51780833c4540e3861c2c9))

## [2.9.3](https://github.com/martynvdijke/dnd/compare/v2.9.2...v2.9.3) (2026-06-04)


### Bug Fixes

* add missing /api/admin/compendium-import route for frontend JSON import ([3db8f46](https://github.com/martynvdijke/dnd/commit/3db8f4625f360894376b66e48415e57075928e93))

## [2.9.2](https://github.com/martynvdijke/dnd/compare/v2.9.1...v2.9.2) (2026-06-04)


### Bug Fixes

* add monsters support to compendium admin management ([c8622f0](https://github.com/martynvdijke/dnd/commit/c8622f0f614b386eb989315c09b942590342c93f))

## [2.9.1](https://github.com/martynvdijke/dnd/compare/v2.9.0...v2.9.1) (2026-06-04)


### Bug Fixes

* show actual error reasons in AI endpoint, fix JSON import schema change hiding preview, improve GitHub fetch error messages ([3a7c562](https://github.com/martynvdijke/dnd/commit/3a7c562af7e700a179c43f8e55d498d90a3ff2fa))

# [2.9.0](https://github.com/martynvdijke/dnd/compare/v2.8.1...v2.9.0) (2026-06-04)


### Bug Fixes

* rename :schema_id to :id in routes to resolve Gin wildcard conflict ([5ebb897](https://github.com/martynvdijke/dnd/commit/5ebb89728fc2283065b2719d5d01a1a73b5785f1))


### Features

* add dynamic compendium schema system with import/export UI ([6ce113a](https://github.com/martynvdijke/dnd/commit/6ce113ad6456a5098ce90514d2eafede8ac2a8a6))

## [2.8.1](https://github.com/martynvdijke/dnd/compare/v2.8.0...v2.8.1) (2026-06-04)

# [2.8.0](https://github.com/martynvdijke/dnd/compare/v2.7.0...v2.8.0) (2026-06-04)


### Features

* add custom monster creator, campaign monster roster, and wire into UI ([e5993df](https://github.com/martynvdijke/dnd/commit/e5993df98b4e488fc4d8e0b14df8dacbf520a676))
* add JSON monster seed (147 SRD monsters) and API import from dnd5eapi.co ([f0be5a0](https://github.com/martynvdijke/dnd/commit/f0be5a092896d2d2eeef1d2e65c1fe5b71a9d0c5))

# [2.7.0](https://github.com/martynvdijke/dnd/compare/v2.6.0...v2.7.0) (2026-06-04)


### Features

* add AI prompt UI with DM endpoint & Docker-aware backups ([6136ade](https://github.com/martynvdijke/dnd/commit/6136ade411454b487b798ef4fdc01b75af0a1227))

# [2.6.0](https://github.com/martynvdijke/dnd/compare/v2.5.2...v2.6.0) (2026-06-03)


### Bug Fixes

* resolve HTMX handler param ID bugs and e2e test CSRF/confirm issues ([f4bff90](https://github.com/martynvdijke/dnd/commit/f4bff90e5ddf85715350362b81094b6f86f8f08b))


### Features

* add scene dialogs, fix act editing bugs ([8831c8d](https://github.com/martynvdijke/dnd/commit/8831c8d38bb04745abed305d806b53218d7ec1d9))

## [2.5.2](https://github.com/martynvdijke/dnd/compare/v2.5.1...v2.5.2) (2026-06-03)


### Bug Fixes

* **deps:** update github.com/dop251/goja digest to 1f200ca ([5c8dfd2](https://github.com/martynvdijke/dnd/commit/5c8dfd2a4f5fd45bd3a32947edd7c6eeb954e26d))

## [2.5.1](https://github.com/martynvdijke/dnd/compare/v2.5.0...v2.5.1) (2026-06-03)


### Bug Fixes

* remove openspec files ([67b6652](https://github.com/martynvdijke/dnd/commit/67b66525ecaf6431dd2627a7b4995151b1c04223))

# [2.5.0](https://github.com/martynvdijke/dnd/compare/v2.4.0...v2.5.0) (2026-06-02)


### Bug Fixes

* add compendium_monster_id to ent EncounterMonster schema ([f764cd4](https://github.com/martynvdijke/dnd/commit/f764cd47a11b25d8be22ccfa950f39da0213b650))
* include seed_monsters.go and ai.go in git, fix :eid/:id route conflict in main.go ([cb44601](https://github.com/martynvdijke/dnd/commit/cb4460193916a1ec004237811d3eedf6101adb8c))
* resolve duplicate /api/encounters/:id/monsters route conflict ([c404a13](https://github.com/martynvdijke/dnd/commit/c404a13099fa01ac3850520527812712e2dd00b3))


### Features

* add NPC and monster management to campaigns and one-shots ([d806872](https://github.com/martynvdijke/dnd/commit/d8068729a52ccc840247ac9d3e5892fb6eeffe5c))

# [2.4.0](https://github.com/martynvdijke/dnd/compare/v2.3.0...v2.4.0) (2026-06-01)


### Bug Fixes

* add CSRF token to Playwright upload API tests ([1fd245b](https://github.com/martynvdijke/dnd/commit/1fd245b7e4d0133c101602e726cfaaf92194a5dc))
* return id in duplicate upload response (was missing, broke upload-links test) ([a803baa](https://github.com/martynvdijke/dnd/commit/a803baadd4cdb9caa3d54806a043eef4142ba56b))
* update Playwright upload tests to not depend on seeded campaign ([996bab7](https://github.com/martynvdijke/dnd/commit/996bab7cb66d0bbe3cfe0d6e31ef1c29a724271b))
* use explicit JSON string for upload-links POST body ([8e681ed](https://github.com/martynvdijke/dnd/commit/8e681edfdbf4b7b92c1a75b0e01f88ae09838ada))
* use page.evaluate with fetch for upload-links tests ([aa9a668](https://github.com/martynvdijke/dnd/commit/aa9a66809db16bccc9f7cfcc2e1d31b53e858d7d))
* use Playwright multipart API for upload tests ([5812c71](https://github.com/martynvdijke/dnd/commit/5812c71bd544e209322e88a19919dd32d5085f64))


### Features

* add file upload system with upload_links, media galleries, NPC portraits, and crop tool ([b32cb8c](https://github.com/martynvdijke/dnd/commit/b32cb8c599410e04017b9abae70698a81377f58e))

# [2.3.0](https://github.com/martynvdijke/dnd/compare/v2.2.0...v2.3.0) (2026-06-01)


### Bug Fixes

* update party.spec.ts to use sidebar nav for desktop ([20fffd3](https://github.com/martynvdijke/dnd/commit/20fffd34db00abc13f7deda2029371411542c277))
* update Playwright tests to use sidebar navigation ([c874d3a](https://github.com/martynvdijke/dnd/commit/c874d3ad3b0c048a2aa44a73769b88e503860472)), closes [#appSidebar](https://github.com/martynvdijke/dnd/issues/appSidebar)


### Features

* add mini-campaign support, NPC story hooks, and campaign overview ([30ed427](https://github.com/martynvdijke/dnd/commit/30ed427f470c438026534f51348404e780fa17a5))

# [2.2.0](https://github.com/martynvdijke/dnd/compare/v2.1.1...v2.2.0) (2026-05-31)


### Bug Fixes

* sort scenes by sort_order in eager load to fix reorder test ([57459d0](https://github.com/martynvdijke/dnd/commit/57459d001789c636a8963f2aebf382569d8384cd))


### Features

* nested sub-acts, scene sort_order, act-level shops/items/encounters ([065ab33](https://github.com/martynvdijke/dnd/commit/065ab33d4b0a04469ab08e91aae79fca13a01287))

## [2.1.1](https://github.com/martynvdijke/dnd/compare/v2.1.0...v2.1.1) (2026-05-31)

# [2.1.0](https://github.com/martynvdijke/dnd/compare/v2.0.5...v2.1.0) (2026-05-30)


### Features

* add act-level planning with ent-backed NPCs, notes, and HTMX details ([50e9a4c](https://github.com/martynvdijke/dnd/commit/50e9a4c622bdebf3f97a04e805466e47d7a591b7))

## [2.0.5](https://github.com/martynvdijke/dnd/compare/v2.0.4...v2.0.5) (2026-05-29)


### Bug Fixes

* give DMs combat nav access and fix sidebar/nav visibility ([6e0180a](https://github.com/martynvdijke/dnd/commit/6e0180ae1a2f176cecfe20d91a5d127576847b1f))

## [2.0.4](https://github.com/martynvdijke/dnd/compare/v2.0.3...v2.0.4) (2026-05-29)


### Bug Fixes

* FAB button tree-shaking and DM role in admin panel ([0c425fc](https://github.com/martynvdijke/dnd/commit/0c425fcd7fab90b10182bc8d5d7c2d9296d455b4))

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
