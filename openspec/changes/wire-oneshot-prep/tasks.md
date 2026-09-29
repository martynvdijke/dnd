## 1. Prep/run toolbar and navigation

- [x] 1.1 Add a Prep Dashboard button to the one-shot detail header loading `/htmx/oneshot-adventures/:id/dashboard` into `#oneshotSection`
- [x] 1.2 Add DM Screen, Clues, Pregens, and Session Flow buttons loading their existing fragments into `#oneshotSection`
- [x] 1.3 Add a Run button that starts or resumes the adventure's pacing session and then loads the prep dashboard
- [x] 1.4 Add a Back-to-adventure control to the prep dashboard, DM screen, clue board, and pregens fragments

## 2. Fix broken links and undefined JavaScript

- [x] 2.1 Point the prep dashboard "Add clues" action at `/htmx/oneshot-adventures/:id/clues` (drop `showView('clues')` and `/htmx/clues`)
- [x] 2.2 Point the prep dashboard Generators action at the DM tools modal (drop `/htmx/generators`)
- [x] 2.3 Fix the one-shot DM screen generator link and change the `/dice` href to the real hash route
- [x] 2.4 Implement `initDialogSort` in `ts/app/oneshot.ts` (SortableJS, PUT to the existing dialog reorder route) and pass the scene id from the template
- [x] 2.5 Implement `showAddActNoteForm` in `ts/app/oneshot.ts` (inline title/content form, POST `/api/oneshot-acts/:id/notes`, refresh act details)
- [x] 2.6 Replace `showPregenForm`/`editPregen` with htmx form routes (`GET /htmx/pregens/new`, `GET /htmx/pregens/:id/edit`, `POST /htmx/pregens`, `PUT /htmx/pregens/:id`) and a `oneshot_pregen_form.html` template

## 3. Editing reliability fixes

- [x] 3.1 Add `PUT /htmx/oneshot-acts/:id/notes` and repoint the inline act notes textarea at it, updating only `notes`
- [x] 3.2 Fix the compendium-monster unlink URL in `oneshot_monsters_section.html` to `/api/oneshot-monsters/:id/link`
- [x] 3.3 Drop `scene_id` from both `oneshot_items` INSERTs in `handlers/oneshot_import.go`
- [x] 3.4 Add the red-herring checkbox to `oneshot_clue_form.html` and parse it in `HtmxCreateClue`/`HtmxUpdateClue`
- [x] 3.5 Handle the `adventure` search result type by loading the adventure detail into the one-shot view (router + `navigateSearchResult`)

## 4. Pacing clock

- [x] 4.1 Accrue elapsed time from timestamps: anchor `started_at` on start/resume, persist the delta on pause/complete/next-scene, display `elapsed + now - started_at` while running (session and scene timings)
- [x] 4.2 Add `GetAdventurePacing` resolving the adventure's latest running/paused session (404 when none) and point `GET /api/oneshot-adventures/:id/pacing` at it
- [x] 4.3 Keep or reimplement `POST /api/session-pacing/:id/tick` only if a caller remains (grep first); otherwise remove it with its route

## 5. Generators in DM tools

- [x] 5.1 Add the five one-shot generators to `showDmTools()` with a result area wired to the existing `/api/generate/*` endpoints

## 6. Tests

- [x] 6.1 Go tests: act notes-only save preserves fields; import creates items at adventure and act level; adventure pacing lookup returns session and 404; red-herring create/update; monster unlink
- [x] 6.2 Playwright: toolbar entry points render their surfaces; pacing elapsed advances and survives pause/resume; pregen create/edit; dialog reorder; red-herring checkbox; search deep-link
- [x] 6.3 Vitest for new `ts/app/oneshot.ts` helpers where applicable; keep coverage floors

## 7. OpenSpec and verification

- [x] 7.1 Check off `link-oneshot-acts-encounters` tasks and archive the change (`openspec archive`), committing the merged spec under `openspec/specs/oneshot-act-encounters/`
- [ ] 7.2 `task ci` green (includes e2e and coverage gates), PR opened and merged via the standard flow
