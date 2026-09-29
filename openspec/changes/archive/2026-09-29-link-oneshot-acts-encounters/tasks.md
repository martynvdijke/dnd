## 1. Backend JSON

- [x] 1.1 `LinkOneShotEncounter` accepts optional `act_id` in the JSON body and inserts it (nullable); keep adventure-level when absent.
- [x] 1.2 Include `act_id` in `GetOneShotEncounters` and `loadAdventureEncounters` SELECTs.
- [x] 1.3 Add act-scoped handlers `ListActEncounters`, `LinkActEncounter`, `UnlinkActEncounter` and register `GET/POST/DELETE /api/oneshot-acts/:id/encounters` in `routes_oneshot.go`.
- [x] 1.4 Populate encounter names on each act's `Encounters` in `loadAdventureDetail` so the tree can show counts/labels.

## 2. HTMX act encounter picker

- [x] 2.1 Add `HtmxActEncounters`, `HtmxLinkActEncounter`, `HtmxUnlinkActEncounter` handlers.
- [x] 2.2 Add template `handlers/templates/oneshot_act_encounters.html` (linked list + unlink, candidate `<select>` + link button, empty state).
- [x] 2.3 Register the three `/htmx/oneshot-acts/:id/encounters[...]` routes in `handlers/htmx.go`.
- [x] 2.4 Add an "Encounters" button/badge to the act card in `oneshot_detail.html` opening the fragment in the generic modal.

## 3. Scene encounter field

- [x] 3.1 Pass candidate encounters to `HtmxSceneForm`/`HtmxEditSceneForm` and add the `<select name="encounter_id">` to `oneshot_scene_form.html`.
- [x] 3.2 Parse and persist `encounter_id` in `HtmxCreateScene`/`HtmxUpdateScene`.

## 4. Verification

- [x] 4.1 Go test: link encounter to act, list it, unlink it; adventure-level link still works.
- [x] 4.2 Playwright: from an act, open the Encounters modal, link an encounter, see it listed, unlink it.
- [x] 4.3 `task ci` green, PR merged.
