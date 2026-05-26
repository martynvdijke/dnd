## 1. Role System (normal | dm | admin)

- [ ] 1.1 Add `DMRequired()` middleware function in `middleware/auth.go`
- [ ] 1.2 Update `DMRequired()` to allow dm + admin roles
- [ ] 1.3 Apply `DMRequired()` to all one-shot routes in `main_test.go` router
- [ ] 1.4 Update `ListOneShotAdventures` to allow DM to see all one-shots with `?all=true`
- [ ] 1.5 Update `ListCharacters` to allow DM to see all characters with `?all=true`
- [ ] 1.6 Add route for DM to set user roles (admin-only for role promotion)
- [ ] 1.7 Write tests: DM middleware, normal user denied, admin bypass, DM sees all characters

## 2. Party First-Class Entity

- [ ] 2.1 Create `parties` SQL table migration (id, user_id, name, description, created_at, updated_at)
- [ ] 2.2 Create Party model struct in `models/`
- [ ] 2.3 Create Party CRUD handlers in `handlers/party.go`
- [ ] 2.4 Register party routes in router (GET/POST/PUT/DELETE `/api/parties`)
- [ ] 2.5 Add faction routes under party (`/api/parties/:id/factions`)
- [ ] 2.6 Migrate factions table: add `party_id`, make `campaign_id` optional
- [ ] 2.7 Write migration script to create parties from campaign party_names
- [ ] 2.8 Write tests: party CRUD, faction under party, party with uploads

## 3. One-Shot Items with Uploads

- [ ] 3.1 Create `one_shot_items` SQL table (id, adventure_id, name, description, quantity, weight, value_gp, is_magical, attunement, notes, created_at)
- [ ] 3.2 Create OneShotItem model struct
- [ ] 3.3 Create one-shot item CRUD handlers (`handlers/oneshot_items.go`)
- [ ] 3.4 Register item routes under one-shots (`/api/oneshot-adventures/:id/items`)
- [ ] 3.5 Add upload support for items (extend upload handlers for `owner_type="item"`)
- [ ] 3.6 Write tests: item CRUD, item uploads

## 4. One-Shot Shops

- [ ] 4.1 Add `oneshot_adventure_id` column to shops table
- [ ] 4.2 Create one-shot shop handlers (`/api/oneshot-adventures/:id/shops`)
- [ ] 4.3 Update shop model to support both campaign_id and oneshot_adventure_id
- [ ] 4.4 Write tests: create/update/delete shop in one-shot context

## 5. One-Shot Monsters

- [ ] 5.1 Create `monster_library` SQL table
- [ ] 5.2 Create `one_shot_scene_monsters` SQL table
- [ ] 5.3 Create MonsterLibrary model struct
- [ ] 5.4 Create OneShotSceneMonster model struct
- [ ] 5.5 Create monster library CRUD handlers
- [ ] 5.6 Create scene monster handlers (add from library, add inline, list, remove)
- [ ] 5.7 Register monster routes
- [ ] 5.8 Write tests: library CRUD, add library monster to scene, add inline monster, list/remove

## 6. One-Shot Linked Player Characters

- [ ] 6.1 Create `one_shot_pcs` SQL table (id, adventure_id, character_id, notes, created_at)
- [ ] 6.2 Create one-shot PC link/unlink handlers
- [ ] 6.3 Register one-shot PC routes
- [ ] 6.4 Write tests: link PC, list linked PCs, unlink PC

## 7. NPC Two-Tier System

- [ ] 7.1 Add `npc_type` column to npcs table ("simple"|"full")
- [ ] 7.2 Add full NPC sheet fields (JSON columns for traits, actions, etc.)
- [ ] 7.3 Update NPC model to include npc_type and full sheet fields
- [ ] 7.4 Update NPC create/update handlers to support both tiers
- [ ] 7.5 Add backstory field for simple NPCs
- [ ] 7.6 Write tests: create simple NPC, create full NPC, list shows type, update tier

## 8. NPC↔Item Linking

- [ ] 8.1 Create `one_shot_npc_items` SQL table (id, npc_id, item_id, relationship, notes)
- [ ] 8.2 Create NPC-item link/unlink handlers
- [ ] 8.3 Register NPC-item routes
- [ ] 8.4 Write tests: link NPC to item, list links, unlink

## 9. Campaign Cleanup

- [ ] 9.1 Remove calendar routes from router
- [ ] 9.2 Remove party_name from Campaign (after migration)
- [ ] 9.3 Remove factions route from campaign (factions now under party)
- [ ] 9.4 Update campaign GET response to exclude factions/calendar
- [ ] 9.5 Archive calendar data to backup table
- [ ] 9.6 Write tests: campaign creation without party_name, campaign no longer returns factions

## 10. Inline-Editable Durations

- [ ] 10.1 Add HTMX inline-edit partial for act/scene duration
- [ ] 10.2 Wire hx-trigger="blur" and hx-trigger="keydown[Enter]" on duration fields
- [ ] 10.3 Create PUT endpoints for individual field updates on acts/scenes
- [ ] 10.4 Write tests: inline duration update via API

## 11. Visual Act/Scene Tree with Drag-Reorder

- [ ] 11.1 Update one-shot detail template to render tree structure
- [ ] 11.2 Add SortableJS dependency to frontend
- [ ] 11.3 Create reorder endpoints (POST `/api/oneshot-acts/reorder`, `/api/oneshot-scenes/reorder`)
- [ ] 11.4 Wire drag-and-drop to reorder endpoints
- [ ] 11.5 Write tests: reorder acts, reorder scenes

## 12. Universal File Upload System

- [ ] 12.1 Add `owner_type` and `owner_id` columns to uploads table
- [ ] 12.2 Create polymorphic upload endpoint (POST `/api/upload` with owner_type + owner_id)
- [ ] 12.3 Update GET `/api/uploads` to filter by owner_type + owner_id
- [ ] 12.4 Add validation for owner_type values
- [ ] 12.5 Wire party uploads to use new upload system
- [ ] 12.6 Write tests: upload to party, upload to item, list by owner, delete, invalid owner_type

## 13. Existing Tests Pass

- [ ] 13.1 Run all existing tests and verify they still pass after migrations
- [ ] 13.2 Fix any broken tests from schema changes
