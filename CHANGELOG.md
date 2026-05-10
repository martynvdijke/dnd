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
