## 1. Adapter seam

- [ ] 1.1 Add `handlers/import_ecosystem.go` with `CharacterSource` / `CompendiumSource` interfaces and a source-id registry.
- [ ] 1.2 Add `POST /api/import/external {source,kind,payload|url,dry_run,dedup_action,field_mapping?}` dispatching to adapters and the existing cores.
- [ ] 1.3 SSRF-guard URL fetch (loopback/private rejection + bounded read), reusing the `SaveGeneratedImage` pattern.

## 2. Character adapters

- [ ] 2.1 D&D Beyond adapter → `[]ImportCharacter` (abilities, multiclass levels, HP, AC, inventory).
- [ ] 2.2 Foundry actor adapter → `[]ImportCharacter` including owned weapons/equipment.
- [ ] 2.3 Fixture tests for both adapters (golden payload → expected canonical shape).

## 3. Compendium adapters

- [ ] 3.1 5e.tools adapter → `[]map[string]any` with preset mappings per known category and detect-and-suggest fallback.
- [ ] 3.2 Foundry JSON source-pack adapter; explicit clear error for LevelDB `.db` packs.
- [ ] 3.3 Fixture tests for both adapters.

## 4. Preview, dedup, rollback

- [ ] 4.1 Implement dry-run preview across the new endpoint (row-level create/skip/duplicate).
- [ ] 4.2 Add `source` + created-ids to import logs (additive migration).
- [ ] 4.3 Make rollback delete created rows for character and compendium imports.
- [ ] 4.4 Tests: dry-run writes nothing; each dedup policy; rollback deletes only what it created.

## 5. Frontend

- [ ] 5.1 Import dialog (source picker, payload/url, preview table, dedup policy, commit).
- [ ] 5.2 Show import logs and a working rollback action.

## 6. Tests & verification

- [ ] 6.1 `go vet ./...`, `task test`, `npm run typecheck`, `npm run build:vite`, `task lint:e2e`.
- [ ] 6.2 e2e: import a fixture character with dry-run then commit; roll it back.
