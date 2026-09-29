## Why

The one-shot planner is feature-complete but its prep and run surfaces are unreachable: `oneshot_detail.html` has no entry point to the prep dashboard, DM screen, clue board, pregen manager, pacing dashboard, or session-flow print view, so a DM must know undocumented URLs to use them. Several of those surfaces link to routes that do not exist or call undefined JavaScript, three editing actions fail outright (compendium equipment import writes a nonexistent `scene_id` column, the compendium-monster unlink URL does not match its registered route, and act-notes autosave wipes the act's title, description, and duration), and the pacing clock never advances because nothing sends timer ticks. The `link-oneshot-acts-encounters` change is fully implemented in code but left unarchived with unchecked tasks, so the spec ledger is stale.

## What Changes

- Add a prep/run toolbar to the one-shot detail page with entry points to the prep dashboard, DM screen, clue board, pregens, session-flow print view, and start/resume pacing.
- Fix broken links and undefined JavaScript: prep dashboard clue and generator buttons, DM screen generator link and dice link, `showAddActNoteForm`, `showPregenForm`/`editPregen`, and `initDialogSort` (real SortableJS reorder calling the existing dialog-reorder endpoint).
- Fix editing reliability: split act notes into a dedicated HTMX notes handler so a notes-only save cannot clobber act fields; correct the compendium-monster unlink URL; drop the nonexistent `scene_id` from the compendium equipment import; persist the red-herring flag in the HTMX clue form; make the pacing clock advance with a heartbeat and fix the adventure-scoped pacing lookup; repair the search deep link for adventure results.
- Expose the five one-shot generators (adventure hook, dungeon dressing, tavern, urban encounter, road encounter) in the existing DM tools modal instead of leaving them API-only.
- Verify and archive the completed `link-oneshot-acts-encounters` change, checking off its tasks.

## Capabilities

### New Capabilities

- `oneshot-prep-run-ui`: a one-shot exposes working entry points to its prep dashboard, DM screen, clue board, pregens, session-flow print view, and pacing; the pacing dashboard shows a live-moving clock and can advance, pause, resume, and complete a session.
- `oneshot-editing-reliability`: one-shot editing actions work without data loss or dead controls — act details and act notes save independently, scene dialogs drag-reorder, pregens can be created/edited, red herrings can be marked, compendium monster unlink and compendium equipment import succeed, and search results open the one-shot detail.

### Modified Capabilities

- `dm-tooling`: the five one-shot generators become reachable from the DM tools UI (previously API-only).

## Impact

- Templates: `handlers/templates/oneshot_detail.html`, `oneshot_prep_dashboard.html`, `oneshot_dm_screen.html`, `oneshot_pregens.html`, `oneshot_act_details.html`, `oneshot_scene_dialogs.html`, `oneshot_monsters_section.html`, `oneshot_clue_form.html`.
- Handlers: `handlers/oneshot_htmx_acts.go`, `handlers/oneshot_import.go`, `handlers/oneshot_pacing.go`, `handlers/oneshot_htmx_clues.go`, `handlers/routes_oneshot.go`, `handlers/search.go`.
- Frontend: `ts/app/oneshot.ts`, `ts/app/dm-tools.ts`, static templates.
- Tests: Go handler tests (act notes, import, unlink, pacing), Playwright specs for the toolbar and pacing, vitest for new helpers.
- OpenSpec: archive `link-oneshot-acts-encounters`.
