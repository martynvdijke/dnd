// @ts-nocheck — split from monolith
import * as bootstrap from 'bootstrap';
import { expose } from '../lib/expose';
import { esc, attrEscape, capitalize, showModal, hideModal, toast, openCompendiumPicker } from '../lib/dom';
import { api } from '../lib/api';
import { showView } from '../navigation';
import { compendiumSearchModal } from '../compendium-search';

// ─── One-Shot Tree UI (SortableJS Drag-Reorder) ───

expose('initOneShotTree', function (adventureId: number) {
  const actTree = document.getElementById('actTree');
  if (!actTree) return;

  // Sortable acts
  const actsEl = actTree.querySelector('.sortable-acts') || actTree;
  if (actsEl && !(actsEl as any)._sortableInitialized) {
    (actsEl as any)._sortableInitialized = true;
    new (window as any).Sortable(actsEl, {
      handle: '.sortable-handle',
      animation: 150,
      draggable: '.sortable-act',
      onEnd: async function () {
        const order = Array.from(actsEl.querySelectorAll('.sortable-act')).map(el => parseInt(el.getAttribute('data-id') || '0'));
        try {
          await api('PUT', `/api/oneshot-adventures/${adventureId}/acts/reorder`, { order });
        } catch (e: any) { toast(e.message, true); }
      }
    });
  }

  // Sortable scenes within each act
  actTree.querySelectorAll('.sortable-scenes').forEach((scenesEl: any) => {
    if (scenesEl._sortableInitialized) return;
    scenesEl._sortableInitialized = true;
    const actId = parseInt(scenesEl.getAttribute('data-act-id') || '0');
    new (window as any).Sortable(scenesEl, {
      handle: '.sortable-handle',
      animation: 150,
      draggable: '.sortable-scene',
      onEnd: async function () {
        const order = Array.from(scenesEl.querySelectorAll('.sortable-scene')).map((el: any) => parseInt(el.getAttribute('data-id') || '0'));
        try {
          await api('PUT', `/api/oneshot-acts/${actId}/scenes/reorder`, { order });
        } catch (e: any) { toast(e.message, true); }
      }
    });
  });
});

// ─── One-Shot Items ───

expose('showOneShotItemForm', function (adventureId: number) {
  showModal('Add Item', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="itemName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Category</label><input class="form-control" id="itemCategory" placeholder="weapon, armor, potion..."></div>
      <div class="col-3"><label class="form-label">Qty</label><input class="form-control" id="itemQty" type="number" value="1"></div>
      <div class="col-3"><label class="form-label">Weight</label><input class="form-control" id="itemWeight" placeholder="lbs"></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Price (GP)</label><input class="form-control" id="itemPrice" type="number" value="0"></div>
      <div class="col-6"><label class="form-label d-flex gap-3">
        <span><input type="checkbox" id="itemMagical"> Magical</span>
        <span><input type="checkbox" id="itemAttune"> Attunement</span>
      </label></div>
    </div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="itemDesc" rows="2"></textarea>
    <button class="btn btn-sm btn-outline-primary mt-2 ai-generate-btn" type="button" data-ai-mode="text" data-ai-target="itemDesc" data-ai-hint="Write a flavorful item description for a D&D item with this name and category" data-ai-title="Generate Item Description"><i class="fa-solid fa-wand-magic-sparkles me-1"></i>Generate with AI</button></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="itemNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveOneShotItem(${adventureId})">Create</button>
  `);
});

expose('saveOneShotItem', async function (adventureId: number) {
  const name = (document.getElementById('itemName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  await api('POST', `/api/oneshot-adventures/${adventureId}/items`, {
    name,
    category: (document.getElementById('itemCategory') as HTMLInputElement).value,
    quantity: parseInt((document.getElementById('itemQty') as HTMLInputElement).value) || 1,
    weight: (document.getElementById('itemWeight') as HTMLInputElement).value,
    price_gp: parseFloat((document.getElementById('itemPrice') as HTMLInputElement).value) || 0,
    is_magical: (document.getElementById('itemMagical') as HTMLInputElement).checked,
    attunement: (document.getElementById('itemAttune') as HTMLInputElement).checked,
    description: (document.getElementById('itemDesc') as HTMLTextAreaElement).value,
    notes: (document.getElementById('itemNotes') as HTMLTextAreaElement).value,
  });
  hideModal();
  toast('Item created');
  // Refresh items section via HTMX
  const itemsCard = document.querySelector('[hx-get*="/items"]');
  if (itemsCard) htmx.trigger(itemsCard, 'load');
});

// ─── One-Shot Items: Compendium Equipment Picker ───

expose('showCompendiumEquipmentPickerForOneShot', function (adventureId: number) {
  showModal('Equipment Compendium', `<div id="compendiumEquipmentPickerContent" hx-get="/htmx/compendium/equipment/picker-oneshot/${adventureId}" hx-trigger="load" hx-swap="innerHTML"><div class="text-center py-3"><i class="fa-solid fa-spinner fa-spin me-1"></i>Loading...</div></div>`);
});

expose('importCompendiumEquipmentToOneShot', async function (equipmentId: number, adventureId: number, quantity: number, source?: string) {
  try {
    const body: Record<string, any> = source === 'entry'
      ? { compendium_entry_id: equipmentId }
      : { compendium_equipment_id: equipmentId };
    body.adventure_id = adventureId;
    body.quantity = quantity || 1;
    await api('POST', `/api/oneshot-adventures/${adventureId}/import/compendium-equipment`, body);
    toast('Equipment added to one-shot');
    bootstrap.Modal.getOrCreateInstance(document.getElementById('genericModal')).hide();
    const itemsCard = document.querySelector('[hx-get*="/items"]');
    if (itemsCard) htmx.trigger(itemsCard, 'load');
  } catch (e: any) { toast(e.message, true); }
});

expose('editOneShotItem', async function (itemId: number) {
  // Get item data by listing from the adventure context
  showModal('Edit Item', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="editItemName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Category</label><input class="form-control" id="editItemCategory"></div>
      <div class="col-3"><label class="form-label">Qty</label><input class="form-control" id="editItemQty" type="number" value="1"></div>
      <div class="col-3"><label class="form-label">Weight</label><input class="form-control" id="editItemWeight"></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Price (GP)</label><input class="form-control" id="editItemPrice" type="number" value="0"></div>
      <div class="col-6"><label class="form-label d-flex gap-3">
        <span><input type="checkbox" id="editItemMagical"> Magical</span>
        <span><input type="checkbox" id="editItemAttune"> Attunement</span>
      </label></div>
    </div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="editItemDesc" rows="2"></textarea>
    <button class="btn btn-sm btn-outline-primary mt-2 ai-generate-btn" type="button" data-ai-mode="text" data-ai-target="editItemDesc" data-ai-hint="Write a flavorful item description for a D&D item with this name and category" data-ai-title="Generate Item Description"><i class="fa-solid fa-wand-magic-sparkles me-1"></i>Generate with AI</button></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="editItemNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="updateOneShotItem(${itemId})">Save</button>
  `);
});

expose('updateOneShotItem', async function (itemId: number) {
  const name = (document.getElementById('editItemName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  await api('PUT', `/api/oneshot-items/${itemId}`, {
    name,
    category: (document.getElementById('editItemCategory') as HTMLInputElement).value,
    quantity: parseInt((document.getElementById('editItemQty') as HTMLInputElement).value) || 1,
    weight: (document.getElementById('editItemWeight') as HTMLInputElement).value,
    price_gp: parseFloat((document.getElementById('editItemPrice') as HTMLInputElement).value) || 0,
    is_magical: (document.getElementById('editItemMagical') as HTMLInputElement).checked,
    attunement: (document.getElementById('editItemAttune') as HTMLInputElement).checked,
    description: (document.getElementById('editItemDesc') as HTMLTextAreaElement).value,
    notes: (document.getElementById('editItemNotes') as HTMLTextAreaElement).value,
  });
  hideModal();
  toast('Item updated');
  const itemsCard = document.querySelector('[hx-get*="/items"]');
  if (itemsCard) htmx.trigger(itemsCard, 'load');
});

expose('deleteOneShotItem', async function (itemId: number) {
  if (!confirm('Delete this item?')) return;
  await api('DELETE', `/api/oneshot-items/${itemId}`);
  toast('Item deleted');
  const itemsCard = document.querySelector('[hx-get*="/items"]');
  if (itemsCard) htmx.trigger(itemsCard, 'load');
});

// ─── One-Shot Shops ───

expose('showOneShotShopForm', function (adventureId: number) {
  showModal('Add Shop', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="shopName"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="shopDesc" rows="2"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Sell Markup %</label><input class="form-control" id="shopMarkup" type="number" value="100"></div>
      <div class="col-6"><label class="form-label">Buy Markup %</label><input class="form-control" id="shopBuyMarkup" type="number" value="50"></div>
    </div>
    <button class="btn btn-primary w-100" onclick="createOneShotShop(${adventureId})">Create</button>
  `);
});

expose('createOneShotShop', async function (adventureId: number) {
  const name = (document.getElementById('shopName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  await api('POST', `/api/oneshot-adventures/${adventureId}/shops`, {
    name,
    description: (document.getElementById('shopDesc') as HTMLTextAreaElement).value,
    markup_percent: parseFloat((document.getElementById('shopMarkup') as HTMLInputElement).value) || 100,
    markup_buy_percent: parseFloat((document.getElementById('shopBuyMarkup') as HTMLInputElement).value) || 50,
  });
  hideModal();
  toast('Shop created');
  const shopsCard = document.querySelector('[hx-get*="/shops"]');
  if (shopsCard) htmx.trigger(shopsCard, 'load');
});

expose('deleteOneShotShop', async function (shopId: number) {
  if (!confirm('Delete this shop?')) return;
  await api('DELETE', `/api/oneshot-adventures/0/shops/${shopId}`);
  toast('Shop deleted');
  const shopsCard = document.querySelector('[hx-get*="/shops"]');
  if (shopsCard) htmx.trigger(shopsCard, 'load');
});

// ─── One-Shot Monsters ───

expose('deleteOneShotMonster', async function (monsterId: number) {
  if (!confirm('Delete this monster?')) return;
  await api('DELETE', `/api/oneshot-monsters/${monsterId}`);
  toast('Monster deleted');
  const monstersCard = document.querySelector('[hx-get*="/monsters"]');
  if (monstersCard) htmx.trigger(monstersCard, 'load');
});

expose('showCompendiumMonsterPickerForOneShot', function (adventureId: number) {
  showMonsterPicker('oneshot', adventureId, 'compendium');
});

// Unified monster picker modal (monster-management change): opens the shared
// /htmx/monster-picker/<context>/<id> modal with Compendium / My Library tabs
// (+ Campaign Roster for campaign contexts).
expose('showMonsterPicker', function (context: string, contextId: number, tab?: string) {
  const tabQ = tab ? '?tab=' + tab : '';
  showModal('Add Monster', `<div hx-get="/htmx/monster-picker/${context}/${contextId}${tabQ}" hx-trigger="load" hx-swap="innerHTML"><div class="text-center py-3"><i class="fa-solid fa-spinner fa-spin me-1"></i>Loading...</div></div>`);
});

// One-shot monsters: compendium search first (schema-based), library as custom fallback.
expose('showOneShotMonsterSearch', async function (adventureId: number) {
  const entry = await compendiumSearchModal({
    title: 'Add Monster from Compendium',
    schemaType: 'monster',
    context: 'Search the compendium for a monster to add to this one-shot.',
  });
  if (!entry) {
    // "Create Custom" (or dismissed) → monster library (has a New Monster flow)
    showMonsterLibrary(adventureId);
    return;
  }
  try {
    await api('POST', `/api/oneshot-adventures/${adventureId}/import/compendium-entry`, {
      compendium_entry_id: entry.id,
      adventure_id: adventureId,
    });
    toast(`Added ${entry.name} to one-shot`);
    const monstersCard = document.querySelector('[hx-get*="/monsters"]');
    if (monstersCard) htmx.trigger(monstersCard, 'load');
  } catch (e: any) { toast(e.message, true); }
});

expose('importCompendiumMonsterToOneShot', async function (monsterId: number, adventureId: number) {
  try {
    await api('POST', `/api/oneshot-adventures/${adventureId}/import/compendium`, {
      compendium_monster_id: monsterId,
      adventure_id: adventureId,
    });
    toast('Monster imported to one-shot');
    bootstrap.Modal.getOrCreateInstance(document.getElementById('genericModal')).hide();
    const monstersCard = document.querySelector('[hx-get*="/monsters"]');
    if (monstersCard) htmx.trigger(monstersCard, 'load');
  } catch (e: any) { toast(e.message, true); }
});

// ─── One-Shot NPC Linking ───

expose('showOneShotNPCLinkForm', async function (adventureId: number) {
  let npcs: any[];
  try {
    const res = await fetch('/api/npcs');
    npcs = await res.json();
  } catch {
    toast('Failed to load NPCs', true);
    return;
  }
  const list = npcs.map(n => `<div class="form-check"><input class="form-check-input npc-select" type="radio" name="npc_id" value="${n.id}" id="npc_${n.id}"><label class="form-check-label" for="npc_${n.id}"><strong>${esc(n.name)}</strong> ${n.race ? '(' + esc(n.race) + ')' : ''} ${n.class ? '- ' + esc(n.class) : ''}</label></div>`).join('');
  showModal('Link NPC', `
    <div class="mb-3">
      <input class="form-control form-control-sm" id="npcSearchInput" placeholder="Search NPCs..." oninput="filterNPCList(this.value)">
    </div>
    <div id="npcSelectList" style="max-height:200px;overflow-y:auto" class="mb-3 border rounded p-2">${list}</div>
    <div class="mb-2"><label class="form-label">Role</label>
      <select class="form-select" id="npcLinkRole">
        <option value="ally">Ally</option>
        <option value="neutral">Neutral</option>
        <option value="enemy">Enemy</option>
      </select>
    </div>
    <div class="mb-2"><label class="form-label">Story Hook</label><textarea class="form-control" id="npcLinkHook" rows="2" placeholder="Brief story hook or note for this NPC..."></textarea></div>
    <div class="mb-3"><label class="form-label"><input type="checkbox" id="npcLinkCombat"> Combat-ready</label></div>
    <button class="btn btn-primary w-100" onclick="linkOneShotNPC(${adventureId})">Link NPC</button>
  `);
});

expose('filterNPCList', function (query: string) {
  const list = document.getElementById('npcSelectList');
  if (!list) return;
  const q = query.toLowerCase();
  list.querySelectorAll('.form-check').forEach(el => {
    const label = el.querySelector('.form-check-label')?.textContent?.toLowerCase() || '';
    el.classList.toggle('d-none', !label.includes(q));
  });
});

expose('linkOneShotNPC', async function (adventureId: number) {
  const selected = document.querySelector<HTMLInputElement>('.npc-select:checked');
  if (!selected) { toast('Select an NPC', true); return; }
  const npcId = parseInt(selected.value);
  const role = (document.getElementById('npcLinkRole') as HTMLSelectElement).value;
  const storyHook = (document.getElementById('npcLinkHook') as HTMLTextAreaElement).value;
  const combatReady = (document.getElementById('npcLinkCombat') as HTMLInputElement).checked;
  await api('POST', `/api/oneshot-adventures/${adventureId}/npcs`, { npc_id: npcId, role, story_hook: storyHook, combat_ready: combatReady });
  hideModal();
  toast('NPC linked');
  const npcCard = document.querySelector('[hx-get*="/npcs"]');
  if (npcCard) htmx.trigger(npcCard, 'load');
});

expose('deleteOneShotNPC', async function (adventureId: number, npcId: number) {
  if (!confirm('Remove this NPC from the adventure?')) return;
  await api('DELETE', `/api/oneshot-adventures/${adventureId}/npcs/${npcId}`);
  toast('NPC removed');
  const npcCard = document.querySelector('[hx-get*="/npcs"]');
  if (npcCard) htmx.trigger(npcCard, 'load');
});

// ─── Combat-Ready NPCs → Combat Tracker ───

expose('addNPCToCombat', async function (npcId: number, npcName: string) {
  try {
    await api('POST', '/api/combat', {
      name: npcName,
      type: 'npc',
      ac: 10,
      hp_max: 20,
      hp_current: 20,
      initiative_mod: 0,
      source_npc_id: npcId,
    });
    showView('combatTracker');
    (window as any).showCombatTracker();
    toast(`${npcName} added to combat tracker`);
  } catch (e: any) { toast(e.message, true); }
});

expose('addOneShotCombatNPCs', async function (adventureId: number) {
  try {
    const res = await api('GET', `/api/oneshot-adventures/${adventureId}/npcs`);
    const combatNPCs = res.filter((n: any) => n.combat_ready);
    if (!combatNPCs.length) { toast('No combat-ready NPCs in this one-shot'); return; }
    for (const npc of combatNPCs) {
      await api('POST', '/api/combat', {
        name: npc.npc_name,
        type: 'npc',
        ac: 10,
        hp_max: 20,
        hp_current: 20,
        initiative_mod: 0,
        source_npc_id: npc.npc_id,
      });
    }
    showView('combatTracker');
    (window as any).showCombatTracker();
    toast(`${combatNPCs.length} combat-ready NPC(s) added to tracker`);
  } catch (e: any) { toast(e.message, true); }
});

// Monster Library
expose('showMonsterLibrary', function (adventureId: number) {
  showModal('Monster Library', `
    <div class="mb-3 d-flex gap-2">
      <button class="btn btn-outline-primary btn-sm" onclick="showAddLibraryMonster(${adventureId})"><i class="fa-solid fa-plus me-1"></i>New</button>
      <input class="form-control form-control-sm" id="libSearch" placeholder="Search library..." oninput="filterLibraryMonsters()">
    </div>
    <div id="libraryList" class="list-group list-group-flush" style="max-height:50vh;overflow-y:auto">
      <div class="text-muted small py-2">Loading library...</div>
    </div>
  `);
  loadMonsterLibrary(adventureId);
});

async function loadMonsterLibrary(adventureId: number) {
  const list = document.getElementById('libraryList');
  if (!list) return;
  try {
    const monsters = await api('GET', '/api/monster-library');
    if (!monsters.length) {
      list.innerHTML = '<div class="text-muted small fst-italic py-2">No monsters in library yet.</div>';
      return;
    }
    expose('_libraryMonsters', monsters);
    renderLibraryMonsters(adventureId, monsters);
  } catch {
    list.innerHTML = '<div class="text-danger small py-2">Failed to load library.</div>';
  }
}

function renderLibraryMonsters(adventureId: number, monsters: any[]) {
  const list = document.getElementById('libraryList');
  if (!list) return;
  list.innerHTML = monsters.map((m: any) => `
    <div class="list-group-item py-2 px-0 d-flex justify-content-between align-items-start library-monster-item" data-search="${esc(m.name).toLowerCase()}">
      <div>
        <strong>${esc(m.name)}</strong>
        <span class="badge bg-danger ms-1">CR ${esc(m.cr)}</span>
        <span class="text-muted small ms-2">AC ${m.ac} · HP ${m.hp}</span>
      </div>
      <div class="d-flex gap-1">
        <button class="btn btn-sm btn-outline-primary" onclick="quickAddLibraryMonster(${adventureId}, ${m.id})" title="Quick Add"><i class="fa-solid fa-plus"></i></button>
        <button class="btn btn-sm btn-outline-danger" onclick="deleteLibraryMonster(${m.id}, ${adventureId})"><i class="fa-solid fa-trash"></i></button>
      </div>
    </div>
  `).join('');
}

expose('filterLibraryMonsters', function () {
  const q = ((document.getElementById('libSearch') as HTMLInputElement).value || '').toLowerCase();
  const monsters = (window as any)._libraryMonsters || [];
  if (!q) { renderLibraryMonsters(0, monsters); return; }
  const filtered = monsters.filter((m: any) => m.name.toLowerCase().includes(q));
  const list = document.getElementById('libraryList');
  if (!list) return;
  list.innerHTML = filtered.map((m: any) => `
    <div class="list-group-item py-2 px-0 d-flex justify-content-between align-items-start library-monster-item">
      <div>
        <strong>${esc(m.name)}</strong>
        <span class="badge bg-danger ms-1">CR ${esc(m.cr)}</span>
        <span class="text-muted small ms-2">AC ${m.ac} · HP ${m.hp}</span>
      </div>
      <div class="d-flex gap-1">
        <button class="btn btn-sm btn-outline-primary" onclick="quickAddLibraryMonster(0, ${m.id})" title="Quick Add"><i class="fa-solid fa-plus"></i></button>
        <button class="btn btn-sm btn-outline-danger" onclick="deleteLibraryMonster(${m.id}, 0)"><i class="fa-solid fa-trash"></i></button>
      </div>
    </div>
  `).join('') || '<div class="text-muted small fst-italic py-2">No matches.</div>';
});

expose('showAddLibraryMonster', function (adventureId: number) {
  showModal('Add to Library', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="libMonsterName"></div>
    <div class="row g-3 mb-3">
      <div class="col-3"><label class="form-label">AC</label><input class="form-control" id="libMonsterAC" type="number" value="10"></div>
      <div class="col-3"><label class="form-label">HP</label><input class="form-control" id="libMonsterHP" type="number" value="10"></div>
      <div class="col-3"><label class="form-label">CR</label><input class="form-control" id="libMonsterCR" placeholder="1/2"></div>
      <div class="col-3"><label class="form-label">Source</label><input class="form-control" id="libMonsterSource" value="custom"></div>
    </div>
    <div class="row g-3 mb-3">
      ${['str','dex','con','int','wis','cha'].map(s => `<div class="col-2"><label class="form-label">${s.toUpperCase()}</label><input class="form-control" id="libMonster${s.toUpperCase()}" type="number" value="10"></div>`).join('')}
    </div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="libMonsterDesc" rows="2"></textarea></div>
    <div class="mb-3"><label class="form-label">Special Abilities</label><textarea class="form-control" id="libMonsterAbilities" rows="3"></textarea></div>
    <div class="mb-3"><label class="form-label">Actions</label><textarea class="form-control" id="libMonsterActions" rows="3"></textarea></div>
    <button class="btn btn-primary w-100" onclick="createLibraryMonster(${adventureId})">Create</button>
  `);
});

expose('createLibraryMonster', async function (adventureId: number) {
  const name = (document.getElementById('libMonsterName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  await api('POST', '/api/monster-library', {
    name,
    ac: parseInt((document.getElementById('libMonsterAC') as HTMLInputElement).value) || 10,
    hp: parseInt((document.getElementById('libMonsterHP') as HTMLInputElement).value) || 10,
    cr: (document.getElementById('libMonsterCR') as HTMLInputElement).value || '0',
    source: (document.getElementById('libMonsterSource') as HTMLInputElement).value || 'custom',
    str: parseInt((document.getElementById('libMonsterSTR') as HTMLInputElement).value) || 10,
    dex: parseInt((document.getElementById('libMonsterDEX') as HTMLInputElement).value) || 10,
    con: parseInt((document.getElementById('libMonsterCON') as HTMLInputElement).value) || 10,
    int_: parseInt((document.getElementById('libMonsterINT') as HTMLInputElement).value) || 10,
    wis: parseInt((document.getElementById('libMonsterWIS') as HTMLInputElement).value) || 10,
    cha: parseInt((document.getElementById('libMonsterCHA') as HTMLInputElement).value) || 10,
    special_abilities: (document.getElementById('libMonsterAbilities') as HTMLTextAreaElement).value,
    actions: (document.getElementById('libMonsterActions') as HTMLTextAreaElement).value,
    description: (document.getElementById('libMonsterDesc') as HTMLTextAreaElement).value,
    is_full: 1,
  });
  hideModal();
  toast('Monster added to library');
  if (adventureId) (window as any).showMonsterLibrary(adventureId);
});

expose('quickAddLibraryMonster', async function (adventureId: number, libraryId: number) {
  try {
    const libMonsters = (window as any)._libraryMonsters || [];
    const m = libMonsters.find((x: any) => x.id === libraryId);
    if (!m) { toast('Monster not found', true); return; }
    await api('POST', `/api/oneshot-acts/0/monsters`, {
      name: m.name, adventure_id: adventureId,
      ac: m.ac, hp: m.hp, cr: m.cr, source: m.source,
      str: m.str, dex: m.dex, con: m.con, int_: m.int, wis: m.wis, cha: m.cha,
      special_abilities: m.special_abilities, actions: m.actions,
      is_full: m.is_full ? 1 : 0, library_id: libraryId,
    });
    toast('Monster added');
    hideModal();
    const monstersCard = document.querySelector('[hx-get*="/monsters"]');
    if (monstersCard) htmx.trigger(monstersCard, 'load');
  } catch (e: any) { toast(e.message, true); }
});

expose('deleteLibraryMonster', async function (libraryId: number, adventureId: number) {
  if (!confirm('Delete from library?')) return;
  await api('DELETE', `/api/monster-library/${libraryId}`);
  toast('Library entry deleted');
  if (adventureId) loadMonsterLibrary(adventureId);
});

// Act-level monster display
expose('showActMonsters', async function (actId: number) {
  try {
    const monsters = await api('GET', `/api/oneshot-acts/${actId}/monsters`);
    const advId = (window as any).adventureIdForAct ? (window as any).adventureIdForAct(actId) : 0;
    showModal('Act Monsters', `
      <div class="mb-2 d-flex gap-1">
        <button class="btn btn-sm btn-outline-warning" onclick="showCompendiumMonsterPickerForOneShot(${advId})"><i class="fa-solid fa-book-open me-1"></i>From Compendium</button>
      </div>
      ${monsters.length ? monsters.map((m: any) => `
        <div class="inv-item">
          <div>${m.compendium_monster_id
            ? `<a href="javascript:void(0)" onclick="htmx.ajax('GET','/compendium/card/monster/${m.compendium_monster_id}',{target:'#cardContainer',swap:'beforeend'})" class="text-decoration-none"><strong>${attrEscape(m.name)}</strong></a>`
            : `<strong>${esc(m.name)}</strong>`}
          <span class="badge bg-danger">CR ${esc(m.cr)}</span> <span class="text-muted small">AC ${m.ac} · HP ${m.hp}</span></div>
          <button class="btn btn-sm btn-outline-danger" onclick="deleteOneShotMonster(${m.id})"><i class="fa-solid fa-trash"></i></button>
        </div>
      `).join('') : '<div class="text-muted small fst-italic">No monsters in this act.</div>'}
    `);
  } catch (e: any) { toast(e.message, true); }
});

// Scene-level monster display
expose('showSceneMonsters', async function (sceneId: number) {
  try {
    const monsters = await api('GET', `/api/oneshot-scenes/${sceneId}/monsters`);
    const advId = (window as any).adventureIdForScene ? (window as any).adventureIdForScene(sceneId) : 0;
    showModal('Scene Monsters', `
      <div class="mb-2 d-flex gap-1">
        <button class="btn btn-sm btn-outline-warning" onclick="showCompendiumMonsterPickerForOneShot(${advId})"><i class="fa-solid fa-book-open me-1"></i>From Compendium</button>
      </div>
      ${monsters.length ? monsters.map((m: any) => `
        <div class="inv-item">
          <div>${m.compendium_monster_id
            ? `<a href="javascript:void(0)" onclick="htmx.ajax('GET','/compendium/card/monster/${m.compendium_monster_id}',{target:'#cardContainer',swap:'beforeend'})" class="text-decoration-none"><strong>${attrEscape(m.name)}</strong></a>`
            : `<strong>${esc(m.name)}</strong>`}
          <span class="badge bg-danger">CR ${esc(m.cr || '0')}</span> <span class="text-muted small">AC ${m.ac} · HP ${m.hp}</span></div>
          <button class="btn btn-sm btn-outline-danger" onclick="deleteOneShotMonster(${m.id})"><i class="fa-solid fa-trash"></i></button>
        </div>
      `).join('') : '<div class="text-muted small fst-italic">No monsters in this scene.</div>'}
    `);
  } catch (e: any) { toast(e.message, true); }
});

// Helper to find adventure ID for a scene (from DOM)
expose('adventureIdForScene', function (sceneId: number): number {
  const tree = document.getElementById('actTree');
  if (!tree) return 0;
  // Adventure ID is stored on the oneshotSection or nearby
  const detail = tree.closest('[hx-get*="/oneshot-adventures/"]');
  if (detail) {
    const m = detail.getAttribute('hx-get')?.match(/oneshot-adventures\/(\d+)/);
    if (m) return parseInt(m[1]);
  }
  // Fallback: look for any element with hx-get containing adventure ID
  const anyEl = document.querySelector('[hx-get*="oneshot-adventures/"][hx-get*="/monsters"]');
  if (anyEl) {
    const m = anyEl.getAttribute('hx-get')?.match(/oneshot-adventures\/(\d+)/);
    if (m) return parseInt(m[1]);
  }
  return 0;
});

// Helper to find adventure ID for an act (same DOM lookup as scene)
expose('adventureIdForAct', function (actId: number): number {
  return (window as any).adventureIdForScene(actId);
});

// ─── One-Shot Linked Player Characters ───

expose('showLinkPCForm', function (adventureId: number) {
  showModal('Link Character', `
    <p class="text-muted small mb-3">Search for a character to link to this one-shot.</p>
    <div class="mb-3"><input class="form-control" id="pcSearchInput" placeholder="Search characters..." oninput="searchCharactersForLink(${adventureId})"></div>
    <div id="pcLinkResults" class="mb-3" style="max-height:300px;overflow-y:auto"></div>
  `);
});

expose('searchCharactersForLink', async function (adventureId: number) {
  const q = (document.getElementById('pcSearchInput') as HTMLInputElement).value.trim();
  const resultsEl = document.getElementById('pcLinkResults');
  if (!resultsEl) return;
  if (q.length < 1) { resultsEl.innerHTML = ''; return; }
  try {
    const chars = await api('GET', `/api/characters?q=${encodeURIComponent(q)}`);
    resultsEl.innerHTML = chars.length ? chars.map((c: any) => `
      <div class="list-group-item py-2 px-0 d-flex justify-content-between align-items-center">
        <div>
          <strong>${esc(c.name)}</strong>
          <span class="text-muted small ms-2">${esc(c.race)} ${esc(c.class)} · Lvl ${c.level}</span>
        </div>
        <button class="btn btn-sm btn-outline-primary" onclick="linkPCToOneShot(${adventureId}, ${c.id})">Link</button>
      </div>
    `).join('') : '<div class="text-muted small">No characters found.</div>';
  } catch { resultsEl.innerHTML = '<div class="text-danger small">Search failed.</div>'; }
});

expose('linkPCToOneShot', async function (adventureId: number, charId: number) {
  await api('POST', `/api/oneshot-adventures/${adventureId}/characters`, { character_id: charId });
  hideModal();
  toast('Character linked');
  const pcsCard = document.querySelector('[hx-get*="/pcs"]');
  if (pcsCard) htmx.trigger(pcsCard, 'load');
});

expose('unlinkPCFromOneShot', async function (adventureId: number, charId: number) {
  if (!confirm('Unlink this character?')) return;
  await api('DELETE', `/api/oneshot-adventures/${adventureId}/characters/${charId}`);
  toast('Character unlinked');
  const pcsCard = document.querySelector('[hx-get*="/pcs"]');
  if (pcsCard) htmx.trigger(pcsCard, 'load');
});

// ─── NPC↔Item Links ───

expose('showLinkNPCToItem', function (adventureId: number, itemId: number) {
  showModal('Link NPC to Item', `
    <p class="text-muted small mb-3">Find an NPC in this adventure to link:</p>
    <div class="mb-3"><input class="form-control" id="npcLinkSearch" placeholder="Search NPCs..." oninput="searchNPCsForLink(${adventureId}, ${itemId})"></div>
    <div id="npcLinkResults" class="mb-3" style="max-height:300px;overflow-y:auto"></div>
  `);
  (window as any).searchNPCsForLink(adventureId, itemId);
});

expose('searchNPCsForLink', async function (adventureId: number, itemId: number) {
  const q = (document.getElementById('npcLinkSearch') as HTMLInputElement)?.value?.trim() || '';
  const resultsEl = document.getElementById('npcLinkResults');
  if (!resultsEl) return;
  try {
    const npcs = await api('GET', `/api/oneshot-adventures/${adventureId}/npcs${q ? '?q=' + encodeURIComponent(q) : ''}`);
    resultsEl.innerHTML = npcs.length ? npcs.map((n: any) => `
      <div class="list-group-item py-2 px-0 d-flex justify-content-between align-items-center">
        <div><strong>${esc(n.npc_name || n.name)}</strong></div>
        <button class="btn btn-sm btn-outline-primary" onclick="linkNPCToItem(${adventureId}, ${n.npc_id || n.id}, ${itemId})">Link</button>
      </div>
    `).join('') : '<div class="text-muted small">No NPCs found in this adventure.</div>';
  } catch { resultsEl.innerHTML = '<div class="text-danger small">Search failed.</div>'; }
});

expose('linkNPCToItem', async function (adventureId: number, npcId: number, itemId: number) {
  await api('POST', `/api/oneshot-adventures/${adventureId}/npc-item-links`, { npc_id: npcId, item_id: itemId });
  hideModal();
  toast('NPC linked to item');
  const itemsCard = document.querySelector('[hx-get*="/items"]');
  if (itemsCard) htmx.trigger(itemsCard, 'load');
});

expose('unlinkNPCFromItem', async function (linkId: number) {
  if (!confirm('Remove link?')) return;
  await api('DELETE', `/api/npc-item-links/${linkId}`);
  toast('Link removed');
  const itemsCard = document.querySelector('[hx-get*="/items"]');
  if (itemsCard) htmx.trigger(itemsCard, 'load');
});

// ─── Compendium Equipment → NPC Links ───

expose('showLinkCompendiumItemToNPC', function (adventureId: number) {
  openCompendiumPicker({
    title: 'Link Compendium Item to NPC',
    placeholder: 'Search equipment...',
    search: (q) => api('GET', `/api/compendium/equipment?q=${encodeURIComponent(q)}`),
    render: (e: any) => `<div><span class="fw-bold">${esc(e.name)}</span>
      ${e.category ? `<span class="text-muted small ms-1">${esc(e.category)}</span>` : ''}
      ${e.item_rarity ? `<span class="text-muted small"> · ${esc(e.item_rarity)}</span>` : ''}
      ${e.weight ? `<span class="text-muted small"> · ${e.weight} lbs</span>` : ''}</div>`,
    onPick: (e: any) => showLinkCompendiumItemToNPCPick(adventureId, e),
  });
});

expose('showLinkCompendiumItemToNPCPick', async function (adventureId: number, item: any) {
  showModal(`Link "${esc(item.name)}" to NPC`, `
    <p class="text-muted small mb-3">Find an NPC in this adventure to link:</p>
    <div class="mb-3"><input class="form-control" id="npcCompLinkSearch" placeholder="Search NPCs..." oninput="searchNPCsForCompendiumLink(${adventureId}, ${item.id})"></div>
    <div id="npcCompLinkResults" class="mb-3" style="max-height:300px;overflow-y:auto"></div>
  `);
  (window as any).searchNPCsForCompendiumLink(adventureId, item.id);
});

expose('searchNPCsForCompendiumLink', async function (adventureId: number, compId: number) {
  const q = (document.getElementById('npcCompLinkSearch') as HTMLInputElement)?.value?.trim() || '';
  const resultsEl = document.getElementById('npcCompLinkResults');
  if (!resultsEl) return;
  try {
    const npcs = await api('GET', `/api/oneshot-adventures/${adventureId}/npcs${q ? '?q=' + encodeURIComponent(q) : ''}`);
    resultsEl.innerHTML = npcs.length ? npcs.map((n: any) => `
      <div class="list-group-item py-2 px-0 d-flex justify-content-between align-items-center">
        <div><strong>${esc(n.npc_name || n.name)}</strong></div>
        <button class="btn btn-sm btn-outline-primary" onclick="linkCompendiumItemToNPC(${adventureId}, ${n.npc_id || n.id}, ${compId})">Link</button>
      </div>
    `).join('') : '<div class="text-muted small">No NPCs found in this adventure.</div>';
  } catch { resultsEl.innerHTML = '<div class="text-danger small">Search failed.</div>'; }
});

expose('linkCompendiumItemToNPC', async function (adventureId: number, npcId: number, compId: number) {
  try {
    await api('POST', `/api/oneshot-adventures/${adventureId}/npc-item-links`, { npc_id: npcId, compendium_equipment_id: compId });
    hideModal();
    toast('NPC linked to compendium item');
  } catch (e: any) {
    toast(e.message, true);
  }
});
