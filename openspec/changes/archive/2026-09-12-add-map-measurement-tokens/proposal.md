## Why

Campaign maps already render with fog of war and pins, but at the table they can't answer "how far is it?" or show party/monster positions the way a battle map should. DMs fall back to eyeballing or external VTTs. Adding measurement and grid-snapped tokens makes Villum's maps genuinely usable for running combat — without sliding into building a full VTT.

## What Changes

- Add a measure/ruler mode to the map view: click-drag draws a path overlay displaying live distance
- Distance uses each map's existing `grid_size` calibration; add a per-map units label (e.g. "ft") so results read "30 ft"; maps without a usable grid show pixel distance instead of nothing
- Add a scale-calibration affordance so the DM can set/correct `grid_size` from the map UI (click two corners of a known square)
- Extend map pins with optional token behavior: snap-to-grid placement, drag-with-snap repositioning, persisted server-side so positions survive reload and are visible to all campaign viewers
- Explicitly out of scope: line of sight, player-side fog editing, WebSocket live drag sync, initiative/combat-tracker integration

## Capabilities

### New Capabilities

- `map-measurement`: Ruler mode with grid-based distance, units labeling, and scale calibration
- `map-tokens`: Grid-snapped, draggable, persistently positioned map tokens

### Modified Capabilities

- *(none)*

## Impact

- **Frontend**: Map code in `ts/app.ts` (Leaflet) gains ruler overlay, snap math, drag handling, calibration UI; new controls need data-testids referenced in tests/
- **Backend**: `handlers/maps.go` — pin update path reused for position persistence; small additions: `grid_units` column on `campaign_maps`, `snap_to_grid` flag on pins
- **Database**: One migration adding two columns; no new tables
- **Realtime**: None added — save-on-drop + fetch-on-load deliberately avoids WebSocket work
