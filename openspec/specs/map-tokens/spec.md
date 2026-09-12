# map-tokens Specification

## Purpose
TBD - created by archiving change add-map-measurement-tokens. Update Purpose after archive.
## Requirements
### Requirement: Token pins with snap-to-grid
Pins SHALL support a snap-to-grid flag. When enabled, created or moved tokens SHALL be positioned at the nearest grid-cell center before persisting.

#### Scenario: Drop snaps to cell center
- **WHEN** the owner drops a snap-enabled token anywhere within a grid cell
- **THEN** the stored position is that cell's center

### Requirement: Token dragging with persistence
The campaign owner SHALL be able to drag snap-enabled tokens; the new position SHALL persist server-side on drop and survive reload. Non-owner viewers SHALL see token positions but not drag handles.

#### Scenario: Position survives reload
- **WHEN** the owner drags a token and any viewer reloads the map
- **THEN** the token renders at its dropped position

#### Scenario: Member cannot move tokens
- **WHEN** a non-owner member views the map
- **THEN** tokens render normally but no drag affordance is available

### Requirement: Token rendering
Snap-enabled pins SHALL render as labeled colored circles distinct from standard pin icons, using the pin's existing name/color fields.

#### Scenario: Token is visually distinct
- **WHEN** a map contains both regular pins and snap-enabled tokens
- **THEN** tokens render as filled labeled circles while regular pins keep their icon style
