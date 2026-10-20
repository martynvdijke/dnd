## MODIFIED Requirements

### Requirement: Player table view
The system SHALL provide a player-facing Table view that shows the current character summary, a dice tray, live initiative, and revealed handouts, and SHALL hide DM-only controls while active. When enabled via Session Mode, the Table view SHALL render by default as a non-blocking docked HUD so the player can simultaneously interact with their character sheet and session quick-actions bar (`#sheetQuickActions`), and SHALL provide a toggle (`data-testid="table-view-dock-toggle"`) to switch between the docked HUD and an expanded full-screen focus layout.

#### Scenario: Player opens table view
- **WHEN** a player enables table view
- **THEN** they see their character, the dice tray, live initiative, and revealed handouts, with no DM editing controls

#### Scenario: Character sheet and quick actions remain interactive in session mode
- **WHEN** a player enables Session Mode while viewing a character sheet
- **THEN** the Table view HUD is visible AND the player can still click character sheet tabs and `#sheetQuickActions` buttons without exiting Session Mode

#### Scenario: Player toggles between docked HUD and expanded focus mode
- **WHEN** a player clicks `table-view-dock-toggle` while the Table view is open
- **THEN** the Table view switches between the compact docked HUD layout and the expanded full-screen layout

#### Scenario: Handout appears live
- **WHEN** the DM reveals a handout while a player is in table view
- **THEN** the handout appears for the player without a manual refresh
