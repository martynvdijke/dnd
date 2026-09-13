## Why
Villum can already import its own JSON character export and CSV/JSON compendium entries with field mapping, but every table that arrives from another tool must be re-typed by hand. The most common sources — D&D Beyond, Foundry VTT, and 5e.tools — all publish structured data. Villum already has the canonical shapes and the import cores (`importCharacters`, `importCompendiumEntries`); what is missing is a thin set of source adapters that translate those formats into the shapes the cores accept.

## What Changes
- Add a source-adapter seam with two directions: **character** adapters (D&D Beyond, Foundry VTT actor) and **compendium** adapters (5e.tools, Foundry JSON source packs).
- Map each source into the existing canonical `ImportCharacter` shape and, for compendium, into `CompendiumEntry.data` using the existing detect/mapping pipeline.
- Reuse `importCharacters` and `importCompendiumEntries` unchanged for persistence, validation, dedup and logging.
- Add a single import endpoint that always supports a **dry-run preview** before commit.
- Make character-import and compendium-import rollback actually reverse what was created.
- Guard any URL-based fetch against SSRF, consistent with existing external fetch handling.

## Capabilities
### New Capabilities
- `import-ecosystem`: external character and compendium source adapters with preview, dedup and rollback.

### Modified Capabilities
<!-- None: existing transfer and compendium import requirements are unchanged; this adds sources on top of them. -->

## Impact
Backend: new `handlers/import_ecosystem.go` (adapters + route), reuse of `handlers/characters_import.go`, `handlers/compendium_admin_import.go`, `handlers/compendium_admin_entries.go` (rollback), `handlers/compendium.go` (fetch pattern). Models: no shape change expected; possibly a `source` field on import logs. Frontend: `ts/transfer.ts`-style import dialog. No new dependencies for the required sources (Foundry `.db` LevelDB packs are explicitly out of scope).
