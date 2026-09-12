## 1. Data model

- [x] 1.1 Migration: add `grid_units` TEXT to `campaign_maps` and `snap_to_grid` INTEGER to `campaign_map_pins` (default-safe, nullable)
- [x] 1.2 Extend map/pin structs and CRUD handlers to carry the new fields; validation tests

## 2. Measurement

- [x] 2.1 Ruler mode toggle in map view (data-testid referenced in tests/): drag path overlay with live distance readout, cleared on exit
- [x] 2.2 Distance math in image coordinates: euclidean segment sum × grid_size → units; pixel fallback when grid_size unusable; unit tests for the math (vitest)
- [x] 2.3 Calibration UI for owner: drag-along-one-edge sets grid_size; units label editor persisted via map update endpoint

## 3. Tokens

- [x] 3.1 Snap-to-grid flag on pin create/edit UI; snap math rounds to nearest cell center (unit-tested)
- [x] 3.2 Owner-only drag with save-on-drop through existing pin update endpoint; no drag affordance for members
- [x] 3.3 Token rendering as labeled colored circles distinct from icon pins

## 4. Verification

- [x] 4.1 `go vet ./... && go build ./...` clean
- [x] 4.2 `task test` green including handler tests for new fields
- [x] 4.3 `npm run typecheck && npm run test:unit` green including snap/distance math units
- [x] 4.4 e2e: calibrate → measure reads correct units → drop token snaps → reload keeps position → member sees token but cannot drag
