## Context

Character import already has a reusable core: `importCharacters(ctx,userID,[]ImportCharacter)` (`handlers/characters_import.go:228`) creates a character plus currency, proficiencies, features, spellcasting, spells and inventory in one transaction; `ImportJSON` (`:187`) is the batch entry point. The canonical shape is `ImportCharacter` (name required; sensible defaults for level/HP/hit dice), populated from `models.character.go`.

Compendium import has `importCompendiumEntries` (`handlers/compendium_admin_import.go:57`): dot-notation field mapping (`getNestedValue`), required-field checks, dedup by name, `skip|overwrite|create-new|force` actions, dry-run, and import-log recording. `DetectImportFields` / `DetectImportSchema` (`compendium_admin_schemas.go:502,631`) already discover fields and suggest mappings. `ImportCompendiumBatchJSON` (`:492`) accepts `entries + field_mapping`. External fetch exists in `handlers/compendium.go:547` (`FetchFromDnDApi`, http.Get + JSON decode) and `ImportFromAPI` (`:649`). `transfer.go` / `transfer_import.go` define the cross-instance envelope and FK remapping (`importOrder:450`, `fkColumns:466`).

Gaps: no D&D Beyond / Foundry / 5e.tools parsers; transfer dry-run is effectively a no-op; compendium rollback only flips `status=rolled_back` and does not delete created entries.

## Goals / Non-Goals

**Goals**
- Import a D&D Beyond or Foundry character into Villum without retyping.
- Import a 5e.tools or Foundry JSON compendium pack with a mapping preview and dedup.
- Always preview before committing; always leave a rollback that works.
- Zero new dependencies for the supported sources.

**Non-Goals**
- Foundry **LevelDB** `.db` packs — they need a LevelDB reader; only Foundry **JSON source** packs are supported.
- Live sync / re-import-on-change; import is a one-shot copy.
- D&D Beyond OAuth or scraping behind their login; import accepts an exported/pasted payload or a public URL.
- Cross-system (non-5e) stat conversion.

## Decisions

### 1. An adapter registry keyed by source id
`handlers/import_ecosystem.go` defines:
```go
type CharacterSource interface { Parse(raw []byte) ([]models.ImportCharacter, error) }
type CompendiumSource interface { Parse(raw []byte) ([]map[string]any, error) }
```
Adapters register by id (`dndbeyond`, `foundry-actor`, `5etools`, `foundry-pack`). Detection is by explicit id plus a cheap shape probe, never by guessing.

**Why:** Keeps each format's quirks in one place and lets the endpoint stay a thin dispatcher. **Alternative:** one giant switch — rejected (unreviewable, hard to test).

### 2. Everything funnels through the existing import cores
Character adapters return `[]ImportCharacter` and call `importCharacters`. Compendium adapters return `[]map[string]any` and call `importCompendiumEntries` with a mapping (preset for known 5e.tools categories, otherwise `DetectImportSchema` suggestions).

**Why:** Validation, dedup, transactions and logging are already solved; duplicating them per source is the bug farm we are avoiding.

### 3. Dry-run is mandatory in the API shape
`POST /api/import/external` takes `{source, payload|url, kind, dry_run, dedup_action, field_mapping?}` and returns a preview: would-create / would-skip / duplicates, row by row. Commit is the same call with `dry_run:false`.

**Why:** External payloads are messy; a preview is the difference between a useful importer and a data-loss incident. This also closes the existing transfer dry-run gap for the new path.

### 4. SSRF-guarded URL fetch
The URL path reuses the loopback/private-IP rejection pattern from `SaveGeneratedImage` (`handlers/ai.go:673`) and caps response size (`io.LimitReader`).

### 5. Rollback reverses creations
Record created entity ids per import log; rollback deletes those rows (character import) / those compendium entries, then marks the log rolled back. Requires storing ids alongside the existing import log.

**Why:** A rollback that does not undo anything is worse than none.

## Risks / Trade-offs

- **Format drift (DDB/Foundry versions change)** → adapters are isolated and covered by fixture tests; failures surface as parse errors, never partial writes.
- **Foundry LevelDB packs requested** → documented non-goal with a clear error message.
- **Importing junk into a real campaign** → dry-run default in the UI; dedup policy required; rollback deletes created rows.
- **Rollback deleting user-edited rows** → rollback only targets ids created by that import; edited-then-rolled-back is accepted (log states it).
- **5e.tools shapes vary by category** → preset mappings for known categories, detect-and-suggest fallback otherwise.

## Migration Plan
Additive: new handler, new route, optional `source`/`created_ids` columns on import logs (additive migration). No change to existing import endpoints. Rollback: remove the route and handler; existing imports unaffected.

## Open Questions
- Should Foundry actor import also pull its owned items into inventory? (Proposal: yes, map owned weapons/equipment.)
- Do we need a public "paste your DDB JSON" example in the UI? (Proposal: yes, one sample payload in the dialog.)
