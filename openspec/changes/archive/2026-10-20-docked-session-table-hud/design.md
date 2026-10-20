## Context

`ts/session-mode.ts` toggles `body.session-mode` and calls `openTableView()` / `closeTableView()` in `ts/table-view.ts`. Meanwhile, `ts/characters/sheet.ts` (`ensureSheetQuickActions()`) checks `document.body.classList.contains('session-mode')` to render a sticky quick-action bar (`#sheetQuickActions`) inside `#sheetView`. In `static/app.html`, `#tableView` has inline styles `position:fixed;inset:0;z-index:1080;overflow-y:auto;` which covers the entire screen and blocks pointer access to `#sheetView`, `#sheetQuickActions`, and all other views.

## Goals / Non-Goals

**Goals:**
- Allow players and DMs in Session Mode to interact with the Character Sheet, `#sheetQuickActions`, Combat Tracker, and Battlemap while keeping live initiative, shared dice rolls, quick dice input, and handouts visible in `#tableView`.
- Preserve all existing `#tableView` test IDs (`table-view`, `table-view-character`, `table-view-dice`, `table-view-dice-input`, `table-view-roll-btn`, `table-view-rolls`, `table-view-turn`, `table-view-handouts`, `table-view-exit`).
- Provide a toggle (`data-testid="table-view-dock-toggle"`) to switch `#tableView` between Docked HUD mode (default) and Expanded Focus mode.

**Non-Goals:**
- Changing the WebSocket broadcast protocol (`dice_roll`, `combat_update`, `knowledge_reveal`).
- Removing any existing child elements from `#tableView`.

## Decisions

1. **Docked HUD by Default, Expandable on Demand**:
   - *Decision*: Move `#tableView` layout rules from inline `inset: 0; z-index: 1080` to CSS classes controlled by `ts/table-view.ts`. By default, opening `#tableView` in Session Mode docks it as a bottom/right HUD panel (`z-index: 1035`, with bottom padding added to `#mainContent` so nothing is obscured). Clicking `data-testid="table-view-dock-toggle"` toggles `.table-view-expanded` for full-screen display when desired.
   - *Rationale*: Keeps `#tableView` and all its child elements visible in the DOM on `openTableView()` (satisfying existing E2E assertions) while unblocking `#sheetView` and `#sheetQuickActions`.

2. **Quick-Preset Dice Buttons in Table HUD**:
   - *Decision*: Alongside `#tableViewDiceInput` (`1d20`), include compact 1-click preset pills (`d20`, `d12`, `d10`, `d8`, `d6`, `d4`) that populate or roll through the existing `rollTableDice()` path.
   - *Rationale*: Typing dice expressions manually on mobile/tablet during live play is slow; 1-click polyhedral triggers use the exact same `/api/roll` endpoint.

## Risks / Trade-offs

- **[Risk] Existing Playwright tests assert `table-view` visibility immediately after clicking `session-mode-toggle`** → *Mitigation*: Docked HUD mode keeps `#tableView` and all its sub-sections visible (`display: block`) when opened, so existing visibility assertions pass without modification.
- **[Risk] Viewport overlap on small screens** → *Mitigation*: Add bottom margin/padding to `body.session-mode #mainContent` equal to the docked HUD height and allow collapsing the HUD body via `table-view-dock-toggle`.

## Migration Plan

1. Update `#tableView` markup in `static/app.html` and styles in `static/styles.css`.
2. Add dock/expand toggle state in `ts/table-view.ts` and ensure `ensureSheetQuickActions()` in `ts/characters/sheet.ts` is reachable while `#tableView` is open.
3. Add unit and Playwright E2E tests covering simultaneous `#tableView` and `#sheetQuickActions` interaction.
