# map-measurement Specification

## Purpose
TBD - created by archiving change add-map-measurement-tokens. Update Purpose after archive.
## Requirements
### Requirement: Ruler measurement mode
The map view SHALL provide an explicit measure mode in which click-drag draws a path overlay and displays the live path length. Exiting measure mode SHALL clear the overlay.

#### Scenario: Measure a straight line
- **WHEN** the DM drags from one point to another in measure mode on a calibrated map
- **THEN** the overlay shows the path and its length in the map's units, updating live during the drag

#### Scenario: Multi-segment path
- **WHEN** the drag path includes direction changes
- **THEN** the displayed distance is the sum of segment lengths

#### Scenario: Overlay clears on exit
- **WHEN** measure mode is toggled off
- **THEN** no ruler overlay remains on the map

### Requirement: Grid-based distance with units
Distance SHALL be computed from each map's `grid_size` calibration multiplied by the configured units label. Maps without a usable `grid_size` SHALL display pixel distance instead of failing.

#### Scenario: Calibrated map shows game units
- **WHEN** measuring on a map with grid_size 50 and units "ft"
- **THEN** a 3-cell drag displays approximately "150 ft"

#### Scenario: Uncalibrated map falls back to pixels
- **WHEN** measuring on a map with grid_size unset or zero
- **THEN** the readout shows the pixel distance rather than an error

### Requirement: Scale calibration
The campaign owner SHALL be able to calibrate a map by dragging along one known grid-square edge, which sets `grid_size`, and to set the units label. Calibration persists per map.

#### Scenario: Calibrate from a known square
- **WHEN** the owner completes the calibration drag across one grid square
- **THEN** grid_size is stored for that map and subsequent measurements use it
