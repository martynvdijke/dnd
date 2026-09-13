## 1. Data model

- [x] 1.1 Ent schemas: `CampaignKnowledge` (campaign_id, title, content, source, status enum, shared, status_history JSON, timestamps) and `CampaignKnowledgeKnownBy` (knowledge_id, character_id, unique pair)
- [x] 1.2 Generate migration; verify `task test` migration tests pass

## 2. Registry & visibility

- [x] 2.1 Add `OwnerCampaignShared` ownership rule to `registry/registry.go` (owner always; members only when shared) with subquery support
- [x] 2.2 Register the `knowledge` entity: Searchable, Transferable, Linkable
- [x] 2.3 Tests: member search/listing excludes unshared entries; owner sees all; existing ownership rules unaffected

## 3. Handlers & routes

- [x] 3.1 CRUD handlers with status validation and status-history append
- [x] 3.2 Known-by add/remove endpoints; bulk reveal action (all PCs known-by + shared=true)
- [x] 3.3 Entity-link wiring via existing linking handlers; 404 guard for direct non-owner access to unshared entries
- [x] 3.4 Handler tests covering every spec scenario (CRUD, lifecycle history, known-by, links, visibility, bulk reveal)

## 4. UI

- [x] 4.1 Campaign knowledge panel: HTMX list partial grouped/filterable by status (data-testids referenced in tests/)
- [x] 4.2 Detail view: mentions-enabled editor, source field, share toggle, status control, entity-link picker, known-by character picker
- [x] 4.3 Member-facing rendering shows only shared entries; DM controls hidden for members
- [x] 4.4 Vitest units for any new client modules (filters, picker glue)

## 5. Verification

- [x] 5.1 `go vet ./... && go build ./...` clean
- [x] 5.2 `task test` green including new handler/registry tests
- [x] 5.3 `npm run typecheck && npm run test:unit` green
- [x] 5.4 e2e: create entry as DM → verify invisible to member → share → visible in member list and search → bulk reveal marks all PCs
