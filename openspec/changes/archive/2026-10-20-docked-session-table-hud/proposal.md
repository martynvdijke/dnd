## Why

Enabling Session Mode (`toggleSessionMode()` in `ts/session-mode.ts`) currently applies `body.session-mode`—which renders `#sheetQuickActions` on the Character Sheet (`ts/characters/sheet.ts`)—while simultaneously opening `#tableView` (`ts/table-view.ts`) as a full-screen overlay (`position: fixed; inset: 0; z-index: 1080`). Because `#tableView` covers the entire viewport, players in Session Mode cannot interact with their character sheet tabs, skills, spells, inventory, or the `#sheetQuickActions` bar rendered underneath, and are left with only a single-line summary and a raw `1d20` text box.

## What Changes

- Convert `#tableView` from a blocking full-screen overlay (`inset: 0; z-index: 1080`) into a non-blocking **Docked Live Table HUD** when Session Mode is active, leaving `#mainContent` (Character Sheet, Combat Tracker, Battlemap, Party View) and `#sheetQuickActions` fully interactive.
- Add a dock mode toggle (`data-testid="table-view-dock-toggle"`) on `#tableView` so users can switch between the default **Docked HUD** (compact drawer/split bar showing active turn, quick dice, recent rolls, and revealed handouts) and **Expanded Table Focus** (full-viewport display for secondary monitors or TV table views).
- Ensure `#sheetQuickActions` (Damage, Heal, Condition quick buttons on the Character Sheet) is visible and clickable above/alongside the docked HUD whenever Session Mode is active.

## Capabilities

### New Capabilities

<!-- None; refines the existing `player-live-table` capability. -->

### Modified Capabilities

- `player-live-table`: Update the `Player table view` requirement so Session Mode / Table View renders as a dockable HUD that preserves interaction with the active view and `#sheetQuickActions`, with an explicit toggle to expand or collapse the HUD.

## Impact

- **Frontend (`static/app.html`, `static/styles.css`, `ts/table-view.ts`, `ts/session-mode.ts`, `ts/characters/sheet.ts`)**:
  - Replace full-screen `inset: 0` styling on `#tableView` with docked HUD styling (`body.session-mode`) and an optional `.table-view-expanded` state.
  - Add `data-testid="table-view-dock-toggle"` to toggle between docked and expanded states.
- **Tests (`tests/`, `ts/*.test.ts`)**:
  - Update/add E2E tests verifying that in Session Mode `#tableView` is visible AND character sheet tabs / `#sheetQuickActions` can be clicked without exiting Session Mode.
