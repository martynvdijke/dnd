## 1. Schema & Foundation

- [ ] 1.1 Add `exhaustion_level` (int, default 0) field to Character ent schema
- [ ] 1.2 Add `is_identified` (bool, default false) field to InventoryItem ent schema
- [ ] 1.3 Create `PartyItem` ent schema (id, campaign_id, name, quantity, notes, created_at)
- [ ] 1.4 Add `sort_order` field to CharacterResource ent schema for custom resource ordering
- [ ] 1.5 Create `SessionPlan` ent schema (id, campaign_id, title, session_date, status, dm_notes, planned_encounters JSON, npc_ids JSON, player_goals JSON, expected_duration, created_at, updated_at)
- [ ] 1.6 Run `go generate ./ent` to regenerate Ent code
- [ ] 1.7 Add party inventory and session plan API endpoints (CRUD handlers)

## 2. Player: Encumbrance System

- [ ] 2.1 Implement client-side encumbrance calculator in TypeScript (STR×5/×10/×15 thresholds, coin weight at 50/lb)
- [ ] 2.2 Add encumbrance display to inventory tab header with color-coded states
- [ ] 2.3 Add speed penalty and ability check/save disadvantage notes for encumbered/heavily-encumbered states
- [ ] 2.4 Ensure encumbrance recalculates on any inventory or currency change

## 3. Player: Attunement Tracker

- [ ] 3.1 Implement attunement counter (X/3) in TypeScript counting equipped items with attunement=true
- [ ] 3.2 Display attunement counter in inventory tab header with color states (normal/yellow/red)
- [ ] 3.3 Add "Requires Attunement" badge to inventory item rows
- [ ] 3.4 Update attunement count on equip/unequip toggle

## 4. Player: Exhaustion Tracking

- [ ] 4.1 Add exhaustion level display (0-6) to combat or stats tab
- [ ] 4.2 Show mechanical effect text per exhaustion level
- [ ] 4.3 Add up/down controls to adjust exhaustion level via API
- [ ] 4.4 Exhaustion auto-reduces by 1 on long rest (update rest log handler)

## 5. Player: Passive Investigation & Insight

- [ ] 5.1 Calculate passive Investigation (10 + INT mod + proficiency if proficient)
- [ ] 5.2 Calculate passive Insight (10 + WIS mod + proficiency if proficient)
- [ ] 5.3 Display both scores alongside passive Perception on stats tab with modifier breakdown

## 6. Player: Magic Item Identification

- [ ] 6.1 Add identified/unidentified badge to magical inventory items
- [ ] 6.2 Hide description, damage, and properties for unidentified items
- [ ] 6.3 Add toggle button to mark items as identified/unidentified via PATCH API
- [ ] 6.4 Server-side: expose `is_identified` field in inventory item API responses

## 7. Player: Equipped Loadout Panel

- [ ] 7.1 Build collapsible "Loadout" panel component in TypeScript
- [ ] 7.2 Group equipped items by category (weapons, armor, shield, ring, wondrous, other)
- [ ] 7.3 Show weapon damage dice and type for equipped weapons
- [ ] 7.4 Update loadout when items are equipped/unequipped

## 8. Player: Spell Preparation Workflow

- [ ] 8.1 Add "Prepare Spells" button to spells tab header
- [ ] 8.2 Build batch spell preparation modal showing all known spells grouped by level with checkboxes
- [ ] 8.3 Create PUT /api/characters/:id/spells/prepare endpoint accepting array of spell IDs
- [ ] 8.4 Show prepared count vs max preparable in modal
- [ ] 8.5 Refresh spells tab on successful batch save

## 9. Player: Party Inventory & Treasury

- [ ] 9.1 Build party inventory UI in party view (add items, assign to character, notes)
- [ ] 9.2 Build party treasury display (total coins across all characters)
- [ ] 9.3 Implement coin-split functionality (divide coins evenly among party members)
- [ ] 9.4 Implement "Assign to Character" flow that creates InventoryItem and removes from party

## 10. DM: Faction Reputation Visualization

- [ ] 10.1 Implement reputation-to-tier mapping function (Hostile/Unfriendly/Neutral/Friendly/Allied)
- [ ] 10.2 Add horizontal reputation bars with fill color per tier
- [ ] 10.3 Show tier label and exact value on hover tooltip
- [ ] 10.4 Scale bar width proportionally within -50 to +50 range

## 11. DM: Encounter Difficulty Calculator

- [ ] 11.1 Implement DMG XP threshold lookup in TypeScript
- [ ] 11.2 Add party level and size inputs to encounter builder view
- [ ] 11.3 Implement encounter multiplier for multiple monsters (×1.5, ×2, ×2.5, ×3, ×4)
- [ ] 11.4 Add difficulty label (Easy/Medium/Hard/Deadly) with color coding
- [ ] 11.5 Add visual XP budget meter from Easy through Deadly thresholds

## 12. DM: Session Planner

- [ ] 12.1 Build session planner list view (reverse chronological, status badges)
- [ ] 12.2 Build create/edit session plan form (date, title, encounters, NPCs, goals, notes)
- [ ] 12.3 Implement status lifecycle (Planned → Ready → In Progress → Completed)
- [ ] 12.4 Link encounter templates to session plans via selector

## 13. DM: Treasure Generator

- [ ] 13.1 Implement DMG Individual Treasure tables per CR range
- [ ] 13.2 Implement DMG Hoard Treasure tables (coins, gems, art, magic items)
- [ ] 13.3 Build treasure generator UI (CR range input, roll count, generate button)
- [ ] 13.4 Display generated treasure with total GP value
- [ ] 13.5 Add "Add to Party Loot" button to transfer generated treasure

## 14. DM: Campaign Dashboard

- [ ] 14.1 Build dashboard aggregation endpoint (or client-side aggregation from existing endpoints)
- [ ] 14.2 Display upcoming calendar events (next 5)
- [ ] 14.3 Display recent session recaps (last 3)
- [ ] 14.4 Display active combat status
- [ ] 14.5 Display recent dice rolls across party
- [ ] 14.6 Display party member quick-stats cards (name, class, HP bar, status)
- [ ] 14.7 Responsive card grid layout (2-3 cols desktop, stacked mobile)

## 15. Testing

- [ ] 15.1 Write Go unit tests for new ent schema migrations
- [ ] 15.2 Write Go unit tests for session plan and party inventory handlers
- [ ] 15.3 Write E2E tests for encumbrance display and attunement counter
- [ ] 15.4 Write E2E tests for spell preparation workflow
- [ ] 15.5 Write E2E tests for treasure generator and difficulty calculator
- [ ] 15.6 Write E2E tests for campaign dashboard rendering
