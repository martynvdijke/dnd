## 1. Role System

- [x] 1.1 Add DMRequired() middleware to middleware/auth.go
- [x] 1.2 Add DM required group to main.go route setup for one-shot routes
- [x] 1.3 Add GET /characters/all endpoint for DM to see all user characters
- [x] 1.4 Restrict one-shot handlers to DM/admin role check
- [x] 1.5 Update login UI to reflect role-based navigation

## 2. Database Schema Changes

- [x] 2.1 Add owner_type and owner_id columns to uploads table
- [x] 2.2 Create parties table (id, user_id, name, description, created_at)
- [x] 2.3 Add party_id column to factions table
- [x] 2.4 Create oneshot_items table
- [x] 2.5 Add oneshot_adventure_id column to shops table
- [x] 2.6 Create oneshot_monsters table
- [x] 2.7 Create monster_library table
- [x] 2.8 Create npc_item_links table
- [x] 2.9 Create oneshot_player_characters table
- [x] 2.10 Add is_full and expanded stat fields to npcs table
- [x] 2.11 Update Ent schema for Upload to include owner_type/owner_id
- [x] 2.11b Update Ent schema for NPC to include is_full/backstory/ac/speed/skills/saves/features/actions

## 3. Party CRUD (Backend)

- [x] 3.1 Create Party Go model struct in models/restructure.go
- [x] 3.2 Implement CreateParty handler in handlers/party.go
- [x] 3.3 Implement GetParty/ListParties handler in handlers/party.go
- [x] 3.4 Implement UpdateParty handler in handlers/party.go
- [x] 3.5 Implement DeleteParty handler in handlers/party.go
- [x] 3.6 Implement GET /parties/:id/factions in handlers/party.go
- [x] 3.7 Implement POST /parties/:id/factions in handlers/party.go
- [x] 3.8 Implement GET /parties/:id/uploads in handlers/party.go
- [x] 3.9 Register party routes in main.go

## 4. Party UI (Frontend)

- [x] 4.1 Add Party view to app.html navigation (already existed)
- [x] 4.2 Create HTMX template for party list (used JS inline rendering instead)
- [x] 4.3 Create HTMX template for party detail
- [x] 4.4 Create HTMX fragment for party faction list
- [x] 4.5 Add SortableJS for party UI elements
- [x] 4.6 Wire party CRUD in app.ts (showCreateParty, doCreateParty, renameParty, doRenameParty, deleteParty)

## 5. One-Shot Items Backend

- [x] 5.1 Create OneShotItem Go model struct in models/restructure.go
- [x] 5.2 Implement CreateOneShotItem handler in handlers/oneshot_items.go
- [x] 5.3 Implement ListOneShotItems handler in handlers/oneshot_items.go
- [x] 5.4 Implement UpdateOneShotItem handler in handlers/oneshot_items.go
- [x] 5.5 Implement DeleteOneShotItem handler in handlers/oneshot_items.go
- [x] 5.6 Implement GET /oneshot-items/:id/uploads in handlers/oneshot_items.go
- [x] 5.7 Register one-shot item routes in main.go (DM group)

## 6. One-Shot Items Frontend

- [x] 6.1 Add items section to one-shot adventure view in HTMX (lazy-loadeds in oneshot_detail.html)
- [x] 6.2 Create HTMX fragment for item list within one-shot (oneshot_items_section.html)
- [x] 6.3 Create HTMX fragment for item detail/upload modal
- [x] 6.4 Wire item CRUD in app.ts/HTMX

## 7. NPC↔Item Links

- [x] 7.1 Create NPCItemLink Go model struct in models/restructure.go
- [x] 7.2 Implement CreateNPCItemLink handler in handlers/npc_item_links.go
- [x] 7.3 Implement ListItemsForNPC handler in handlers/npc_item_links.go
- [x] 7.4 Implement ListNPCsForItem handler in handlers/oneshot_items.go
- [x] 7.5 Implement DeleteNPCItemLink handler in handlers/npc_item_links.go
- [x] 7.6 Register NPC-item link routes in main.go
- [x] 7.7 Create HTMX fragments for NPC↔Item link UI (shows linked NPCs in item section, needs unlink button)
- [x] 7.8 Wire NPC↔Item linking in frontend

## 8. One-Shot Monsters

- [x] 8.1 Create OneShotMonster Go model struct in models/restructure.go
- [x] 8.2 Create MonsterLibrary Go model struct in models/restructure.go
- [x] 8.3 Implement inline monster CRUD for acts/scenes in handlers/oneshot_monsters.go
- [x] 8.4 Implement monster library CRUD in handlers/oneshot_monsters.go
- [x] 8.5 Implement quick-add from library to act/scene in handlers/oneshot_monsters.go
- [x] 8.6 Implement full-stat monster rendering (frontend)
- [x] 8.7 Register monster routes in main.go
- [x] 8.8 Create HTMX templates for monster cards in tree view (oneshot_monsters_section.html)
- [x] 8.9 Create monster library browser in HTMX

## 9. One-Shot Shops

- [x] 9.1 Move shop CRUD behind DMRequired middleware (already in DM group)
- [x] 9.2 Update shop create handler to accept oneshot_adventure_id in handlers/oneshot_shops.go
- [x] 9.3 Add shops section to one-shot adventure view in HTMX (lazy-loaded in oneshot_detail.html)
- [x] 9.4 Create HTMX fragment for one-shot shop list (oneshot_shops_section.html)
- [x] 9.5 Wire one-shot shop management in frontend

## 10. Linked Player Characters

- [x] 10.1 Create OneShotPC Go model struct in models/restructure.go
- [x] 10.2 Implement LinkCharacterToOneShot handler in handlers/oneshot_pcs.go
- [x] 10.3 Implement ListLinkedCharacters handler in handlers/oneshot_pcs.go
- [x] 10.4 Implement UnlinkCharacterFromOneShot handler in handlers/oneshot_pcs.go
- [x] 10.5 Register linked PC routes in main.go
- [x] 10.6 Create HTMX fragment for linked character browser/selector (oneshot_pcs_section.html)
- [x] 10.7 Wire character linking UI

## 11. NPC Tier System

- [x] 11.1 Update NPC Go model struct with is_full and expanded fields
- [x] 11.2 Add NPC tier toggle in CreateNPC/UpdateNPC handlers
- [x] 11.3 Update NPC Ent schema with new fields (regenerated Ent code)
- [x] 11.4 Update NPC UI to conditionally show full stat block
- [x] 11.5 Create HTMX fragment for full NPC stat block

## 12. Polymorphic File Uploads

- [x] 12.1 Update Upload Go model struct with owner_type/owner_id
- [x] 12.2 Update HandleUpload handler to accept owner_type/owner_id params
- [x] 12.3 Update GetUploads handler to filter by owner
- [x] 12.4 Update Ent Upload schema with owner_type/owner_id fields
- [x] 12.5 Add upload UI controls to item, party, NPC views
- [x] 12.6 Wire upload functionality in frontend (HTMX)

## 13. Campaign Cleanup

- [x] 13.1 Remove calendar routes from main
- [x] 13.3 Remove calendar HTMX routes and handler functions from htmx.go
- [x] 13.2 Remove calendar tab from campaign UI in app.html/templates (removed nav item + view div + TS function)
- [x] 13.4 Remove factions tab from campaign UI (removed nav item, moved to Party)
- [x] 13.6 Update campaign nav to remove calendar, factions (done in app.html+TS)
- [x] 13.5 Remove old standalone shops tab (kept for DM/admin, moved under one-shot)

## 14. Inline Editing & Tree UI

- [x] 14.1 Add PATCH /oneshot-acts/:id/duration endpoint
- [x] 14.2 Add PATCH /oneshot-scenes/:id/duration endpoint
- [x] 14.3 Add act reorder endpoint (PUT /oneshot-adventures/:id/acts/reorder)
- [x] 14.4 Add scene reorder endpoint (PUT /oneshot-acts/:id/scenes/reorder)
- [x] 14.5 Install SortableJS (npm or CDN)
- [x] 14.6 Create HTMX tree view template for acts/scenes
- [x] 14.7 Wire SortableJS drag-reorder in app.ts
- [x] 14.8 Add inline editing to act/scene duration fields in templates
- [x] 14.9 Register inline editing routes in main.go
