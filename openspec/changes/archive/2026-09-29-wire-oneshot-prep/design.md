## Context

The one-shot planner has accumulated a broad feature set (adventures, act/scene tree, clues, per-act NPCs and notes, monsters, items, shops, encounters, pregens, checklist, pacing, DM screen, session-flow print). Two things block actually using it as a prep-and-run workflow:

1. **Orphaned surfaces.** The prep dashboard, DM screen, clue board, pregens, pacing dashboard, and session-flow view have no entry point from `oneshot_detail.html` or `oneshot_list.html`. They cross-link only to each other or are reachable solely by typing undocumented URLs (as the e2e specs do).
2. **Dead wiring and data-loss bugs.** `oneshot_prep_dashboard.html` points at `/htmx/clues` and `/htmx/generators` (neither exists) and calls `showView('clues')` (no such view). `oneshot_dm_screen.html` points at `/oneshot-generators` and uses the `/dice` hash route as an href. `oneshot_pregens.html` calls `showPregenForm`/`editPregen`, `oneshot_act_details.html` calls `showAddActNoteForm`, and `oneshot_scene_dialogs.html` calls `initDialogSort` — all undefined. Saving act notes PUTs a notes-only form to `HtmxUpdateAct`, which unconditionally overwrites title, description, and estimated minutes with empty values. The compendium equipment import inserts into a nonexistent `oneshot_items.scene_id`; the compendium-monster unlink button targets `/api/oneshots/monsters/:id/link` while the route is `/api/oneshot-monsters/:id/link`; the red-herring flag is never read by the HTMX clue handlers. The pacing clock is incremented only in browser JavaScript, so every htmx swap (pause, resume, next scene, reload) resets the displayed elapsed time to a stale database value.

Additionally, `openspec/changes/link-oneshot-acts-encounters` is fully implemented and tested but unarchived with unchecked tasks, so the spec ledger is misleading.

Constraints: keep the change a wiring-and-reliability pass — no new views; follow existing htmx/SortableJS patterns; keep e2e lint rules (no `waitForTimeout`), testid discipline, and coverage floors; keep DM-facing behavior consistent with the existing templates.

## Goals / Non-Goals

**Goals:**

- Every prep/run surface is reachable from the one-shot detail page in one click, and every button on those surfaces works or is removed.
- Act notes and act fields save independently without clobbering.
- Pacing elapsed time is accurate across pause/resume/next-scene/reload.
- The compendium equipment import, compendium-monster unlink, red-herring flag, pregen create/edit, act DM notes, dialog reordering, and adventure search deep-link all work.
- The five API-only one-shot generators are reachable from the existing DM tools modal.
- `link-oneshot-acts-encounters` is archived with its tasks checked.

**Non-Goals:**

- One-shot scheduling/RSVP, public share links or a player-facing one-shot view.
- Scene-level items (schema and UI do not exist), act-level monster UI, clue dependency/NPC/location graph editing UI.
- Pregen-to-player assignment; pregens stay user-scoped.
- Hardening the DM gate/ownership checks on the one-shot HTMX routes (a real gap, but a security change tracked separately, not part of wiring).
- Any new one-shot feature surface beyond exposing what is already implemented.

## Decisions

1. **Prep toolbar on the one-shot detail header.** One row of buttons — Prep Dashboard, DM Screen, Clues, Pregens, Session Flow, Run — each loading the existing fragment into `#oneshotSection`. "Run" calls `StartPacingSession` for the adventure, then loads the prep dashboard (which shows the live session), so a DM never has to find the session id. Alternative: separate nav entries per surface — rejected; it grows navigation for a workflow that always starts from an adventure.

2. **Dedicated act-notes endpoint instead of partial-update magic.** The inline notes textarea will PUT to a new `PUT /htmx/oneshot-acts/:id/notes` handler that updates only `notes`. The full act form keeps `HtmxUpdateAct`. Alternative: make `HtmxUpdateAct` update only the fields present in the POST body — rejected as implicit and easy to regress; an explicit endpoint is trivially testable.

3. **Pregen create/edit as htmx form routes.** New `GET /htmx/pregens/new`, `GET /htmx/pregens/:id/edit`, `POST /htmx/pregens`, `PUT /htmx/pregens/:id` render/submit `oneshot_pregen_form.html` and refresh the pregen list, reusing the existing JSON handler logic for validation where practical. Alternative: build the form in JavaScript around the JSON API — rejected; it duplicates field work in JS and breaks the template pattern used everywhere else.

4. **Implement `showAddActNoteForm` as an inline JSON-backed form** in `ts/app/oneshot.ts` (title + content, POST `/api/oneshot-acts/:id/notes`, then refresh the act-details fragment). Alternative: drop the button — rejected; act DM notes then have no creation path in the UI despite the list rendering them.

5. **Dialog reorder via `initDialogSort`.** Add the missing SortableJS initializer in `ts/app/oneshot.ts` (mirroring `initOneShotTree`), POST-ing the existing `PUT /api/oneshot-scenes/:id/dialogs/reorder`. The scene-dialogs template keeps its guarded call and gains a data attribute for the scene id instead of relying on a global. Alternative: remove the sortable markup — rejected; the backend and route already support it.

6. **Timestamp-based pacing accrual instead of client ticks.** Elapsed becomes authoritative server-side: start/resume sets `started_at = now`; pause/complete/next-scene persists `elapsed_seconds += now - started_at` before clearing the anchor; reads display `elapsed_seconds + (now - started_at)` while running (same for `scene_timings`). The browser's 1-second counter stays for smooth display only. Alternative: heartbeat the existing `POST /api/session-pacing/:id/tick` every 5s from the dashboard — rejected; it still shows stale values after a reload or missed heartbeat, and it writes constantly.

7. **Fix the adventure-scoped pacing endpoint** with a `GetAdventurePacing` handler that resolves the adventure's latest session (404 when none) and shares the session-loading code with `/api/session-pacing/:id`.

8. **Drop `scene_id` from the compendium equipment import** (both the generic-entry and equipment branches), keeping `act_id`. Alternative: add a `scene_id` column — rejected; no UI or handler supports scene-level items and the model field is unused.

9. **One-shot generators in the DM tools modal.** Extend `showDmTools()` with a one-shot section (adventure hook, dungeon dressing, tavern, urban encounter, road encounter) calling the existing `/api/generate/*` endpoints, with a compact result area. Alternative: a new generators page/view — rejected; the DM tools modal already hosts exactly this kind of tool and is one function to extend.

10. **Adventure search deep-links open the detail.** Handle `adventure` in `navigateSearchResult` (and the router's hash handler) by showing the one-shot view and loading `/htmx/oneshot-adventures/:id` into `#oneshotSection`. Alternative: change `handlers/search.go` to emit `#/oneshot` — rejected; it throws away the selected result.

11. **Archive in the same PR, separate commit.** Check off `link-oneshot-acts-encounters` tasks, run `openspec archive`, and include the archived change + merged spec (`openspec/specs/oneshot-act-encounters/`) so the ledger matches reality.

## Risks / Trade-offs

- **Template wiring can regress silently** (these are exactly the bugs that shipped). Mitigation: every item gets at least one Go handler test or Playwright assertion, and `task lint` catches unused testids.
- **Pacing semantics change breaks existing tests that assume tick arithmetic.** Mitigation: update `oneshot_planning_test.go`/pacing tests in the same commit; e2e already starts sessions through the API and renders the dashboard.
- **Archiving a stale change can surface spec merge conflicts.** Mitigation: archive early in the work and let CI validate; the delta spec is small and self-contained.
- **DM tools modal grows beyond a comfortable size.** Mitigation: five buttons + one result area, collapsed by default, reusing the existing `dmToolsResult` container.
- **Removing the `tick` route** (if no callers remain) could break external scripts. Mitigation: keep the route registered but reimplement it as a no-op accrual-refresh if anything still calls it; verify with a repo-wide grep before deletion.
