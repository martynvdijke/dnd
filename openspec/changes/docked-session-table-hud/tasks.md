## 1. Docked Table HUD Layout & Toggle

- [x] 1.1 Replace the blocking `position:fixed;inset:0;z-index:1080` inline layout on `#tableView` in `static/app.html` and `static/styles.css` with a docked HUD layout (`body.session-mode`) and an optional `.table-view-expanded` state.
- [x] 1.2 Add `data-testid="table-view-dock-toggle"` in `static/app.html` and wire `toggleTableViewDock()` in `ts/table-view.ts` to switch between docked HUD and expanded focus mode.
- [x] 1.3 Verify `#sheetQuickActions` (`ts/characters/sheet.ts`) renders visibly and remains clickable when `session-mode` is active.

## 2. Verification & Quality Gates

- [x] 2.1 Add Vitest unit tests for `toggleTableViewDock()` and Session Mode + Character Sheet quick actions coexistence.
- [x] 2.2 Add/update Playwright E2E assertions in `tests/` exercising `data-testid="table-view-dock-toggle"` and clicking `#sheetQuickActions` while Session Mode is enabled.
- [ ] 2.3 Run `openspec validate docked-session-table-hud --strict` and `task ci`.
