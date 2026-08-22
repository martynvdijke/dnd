## 1. Backend endpoint

- [x] 1.1 Add `GetTRMNLCombatStatus` handler: public read-only (matching shipped TRMNL endpoints), `campaign_id` query param, ordered query over `combat_entries` (`initiative_roll DESC, turn_order ASC`), computed `initiative_total`, condition resolution against the built-in `ConditionTypes` list
- [x] 1.2 Register the route in the public route group alongside the existing `/api/trmnl/` endpoints
- [x] 1.3 Add unit tests: public access returns sorted entries with resolved conditions; empty campaign / missing `campaign_id` → 200 with empty list; unknown condition tokens skipped

## 2. TRMNL plugin

- [x] 2.1 Add the combat polling URL entry and `campaign_id` custom field to `trmnl/src/settings.yml`
- [x] 2.2 Add `combat` display-mode blocks to all four layouts in `trmnl/src/`: active row distinct, idle message on empty payload, `+N more` cap on half/quadrant layouts
- [x] 2.3 Run `trmnlp lint` locally (Taskfile task) and confirm CI lint job passes

## 3. Verification

- [x] 3.1 `go vet ./... && go build ./...` clean
- [x] 3.2 `task test` green including new handler tests
- [x] 3.3 Manual smoke: poll `/api/trmnl/combat?token=...&campaign_id=...` against a seeded campaign and eyeball JSON ordering/conditions
