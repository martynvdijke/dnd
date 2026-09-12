## Context

`campaign_maps` rows carry `image_url`, `width`/`height`, an existing `grid_size` int (pixel edge of one grid cell), `fog_of_war` JSON, and an active flag. Pins (`campaign_map_pins`) already store `x/y`, `icon`, `color`, `name`, `is_hidden`, and entity links; the web tracker pattern is plain CRUD via `handlers/maps.go`. The map view renders in `ts/app.ts` on Leaflet (image overlay, not a geographic CRS), so all distances are pixel-space math — no geo formulas apply.

## Goals / Non-Goals

**Goals:**
- Answer "how far?" in game units with one drag
- Tokens that stay where the DM dropped them, for every viewer, after reload
- Calibration good enough for battle maps: one known square sets the scale

**Non-Goals:**
- Line-of-sight or fog automation
- Live multi-user drag sync (WebSocket) — positions sync on drop, not continuously
- Player-draggable tokens — movement is DM-side this change
- Token images/portraits — colored labeled circles first
- Combat-tracker integration (auto-placing combatants)

## Decisions

### 1. Reuse `grid_size`; add only `grid_units` TEXT
Calibration UI writes the existing column by letting the DM drag along one known grid square edge; units label is free text (default "ft"). When `grid_size <= 0`, measurement falls back to raw pixels.

**Why:** The schema already models cell size; a second column for the label completes it without a calibration matrix. **Alternative:** full affine calibration (origin + rotation) — rejected; scanned maps are axis-aligned often enough and rotation doubles the snap math for marginal gain.

### 2. Euclidean distance, rounded to whole units
Path length = sum of segment lengths in cells × grid_size → units, displayed live while dragging.

**Why:** Simple, predictable, matches the DMG variant most tables accept. **Alternative:** 5e alternating-diagonal (5-10-5) counting — rejected as default because it requires diagonal-path detection per segment pair; noted as a possible future toggle.

### 3. Tokens are pins with `snap_to_grid = true`
No new entity: pin gains the boolean flag. When set, create/drag endpoints round x/y to the nearest cell center client-side before persisting. Rendering uses the existing pin layer with token styling (filled circle, label).

**Why:** Pins already have position/color/links/visibility and their own CRUD; a parallel tokens table would duplicate all of it. **Alternative:** standalone `map_tokens` table — rejected as redundant schema.

### 4. Persistence: save-on-drop through the existing pin update endpoint
Drag ends → PATCH position → server persists → other viewers see it on next map load/poll. No WebSocket traffic.

**Why:** Meets "survives reload, visible to members" at near-zero infra cost; continuous sync is explicitly out of scope.

### 5. Movement permission: owner-only this change
Only the campaign owner sees drag handles and calibration controls; members get read-only rendering including token positions.

**Why:** Matches current pin management permissions; player-side movement opens griefing/consistency questions deferred wholesale.

## Risks / Trade-offs

- **Non-square or rotated source maps** → Calibration measures one edge; rotated grids will mis-snap. Mitigation: document limitation; pixel fallback keeps measurement usable.
- **Zoom-dependent hit targets** → Snap math runs in image coordinates (map CRS), not screen pixels, so zoom level cannot skew results; tested at multiple zooms.
- **Stale positions between viewers** → Accepted until/if realtime sync is prioritized; refresh shows truth.
- **Ruler clutter on small screens** → Measure mode is an explicit toggle; overlay clears on mode exit.

## Migration Plan

1. Migration adds `grid_units` to `campaign_maps` and `snap_to_grid` to `campaign_map_pins` (nullable/default-safe).
2. Ship backend additions, then frontend ruler/calibration/tokens.
3. Rollback: revert commit; added columns are inert.

## Open Questions

- Should hidden pins (`is_hidden`) render as tokens for members when snapped? (Proposal: keep hidden semantics unchanged.)
- Default units string when unset ("ft" vs blank).
