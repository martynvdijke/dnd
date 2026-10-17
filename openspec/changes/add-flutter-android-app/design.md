## Context

Villum (`dnd`) already exposes rich JSON REST endpoints under `/api` (`characters`, `campaigns`, `combat/attack`, `compendium/search`, `draft`, `generate/*`, `notes`, `npcs`, `quests`) and real-time WebSocket updates on `/ws`.

## Goals / Non-Goals

**Goals:**
- Deliver a tactile dark-parchment / Material 3 Flutter Android app in `mobile/` tailored for one-handed use at a D&D table.
- Support instant HP/spell slot mutations, attack rolls, pinch-to-zoom battlemap token movement, and compendium FTS5 search.
- Keep CI and release integration 100% non-blocking (`continue-on-error: true`).

**Non-Goals:**
- Full 3D VTT rendering on mobile (uses 2D `InteractiveViewer` canvas synced via `/api/campaigns/:id/battlemap/sync`).

## Decisions

1. **Interactive Battlemap Canvas**: Use Flutter's `InteractiveViewer` layered with fog-of-war mask and draggable token widgets syncing coordinates to `/api/campaigns/:id/battlemap/tokens`.
2. **Optimistic Character Sheet Updates**: Apply HP, spell slot, and currency deltas optimistically in Riverpod state while persisting to `/api/characters/:id/*`.
3. **Non-Blocking CI & Release**: SHA-pin all GitHub Actions in `.github/workflows/flutter-android.yaml` and `.github/workflows/release.yaml`, set `continue-on-error: true`, and upload `dnd-<version>-android.apk` to GitHub Releases.

## Risks / Trade-offs

- **[Complex Character Sheet Density on Phone Screens]** -> Split character sheet into 5 bottom-nav/swipe tabs (Core/Saves, Combat/HP, Spells, Inventory/Currency, Features/Notes).
