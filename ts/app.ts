// @ts-nocheck — legacy monolith being split into modules, pre-existing type errors
import * as d3 from 'd3';
import Chart from 'chart.js/auto';
import { marked } from 'marked';
import L from 'leaflet';
import * as bootstrap from 'bootstrap';
expose('bootstrap', bootstrap);
import { showView, setCurrentView, getCurrentView } from './navigation';
import { toggleFabMenu, updateFabForView } from './fab';
import { initBridge } from './lib/bridge';
import { esc, capitalize, showModal, hideModal, toast } from './lib/dom';
import { FilePicker } from './file-picker';
import { initTheme } from './lib/theme';
import { api, setCsrfToken, getCsrfToken } from './lib/api';
import { initShortcuts, showShortcutsHelp, getSections } from './lib/shortcuts';
import { renderSheet, updateField } from './characters/sheet';
import { renderStats } from './characters/stats';
import { renderCombat } from './characters/combat';
import { renderInventory } from './characters/inventory';
import { renderSpells } from './characters/spells';
import './characters/locations';
import './characters/npcs';
import './characters/sessions';
import './characters/quests';
import './characters/journal';
import { animateHpChange } from './lib/animations';
import './compendium';
import './combat-tracker';
import './encounter';
import './party';
import './timeline';
import './factions';
import './character-sheet';
import { expose } from './lib/expose';
import { currentUser, currentChar, currentTab, allLocations, allNPCs, setCurrentChar, setCurrentTab, setAllLocations, setAllNPCs } from './lib/state';

// Expose API helper globally for E2E tests (window.api check)
expose('api', api);

declare const htmx: any;

// ─── FAB ───
// (moved to ts/fab.ts)

// ─── Sort ───

function sortList(key: string, order: 'asc' | 'desc' = 'asc') {
  const container = document.getElementById(key + 'List');
  if (!container) return;
  const items = Array.from(container.querySelectorAll('.inv-item'));
  const sorted = items.sort((a, b) => {
    const va = a.getAttribute('data-sort') || a.textContent?.trim() || '';
    const vb = b.getAttribute('data-sort') || b.textContent?.trim() || '';
    const na = parseFloat(va), nb = parseFloat(vb);
    if (!isNaN(na) && !isNaN(nb)) return order === 'asc' ? na - nb : nb - na;
    return order === 'asc' ? va.localeCompare(vb) : vb.localeCompare(va);
  });
  sorted.forEach(item => container.appendChild(item));
}
expose('sortList', sortList);

// Keyboard Shortcuts handled via import from ./lib/shortcuts

// Global Search → extracted to ts/search.ts
import { showSearchOverlay, hideSearchOverlay, doSearch, initSearch } from './search';

expose('navigateSearchResult', function (type: string, id: number, name: string) {
  if (type === 'characters') {
    openChar(id);
  } else if (type === 'campaigns') {
    showView('characters');
    toast('Campaign: ' + name);
  } else if (['spells','equipment','races','classes','feats','backgrounds'].includes(type)) {
    (window as any).showCompendium();
  } else if (type === 'monsters') {
    (window as any).showCompendium();
    setTimeout(() => (window as any).loadCompendiumTab('monsters'), 100);
  } else if (type === 'npcs') {
    showView('characters');
    toast('NPC: ' + name);
  } else {
    showView('characters');
    toast(name);
  }
});

// Loading, API, Theme, Modal, Toast → imported from ./lib/{dom,api,theme}

// ─── WebSocket + Init moved to ts/init.ts ───

// ─── Character List → extracted to ts/characters/list.ts ───
import { loadCharacters, filterCharacters } from './characters/list';

async function openChar(id: number) {
  try {
    setCurrentChar(await api('GET', `/api/characters/${id}`));
    expose('currentChar', currentChar);
    setCurrentTab('stats');
    showView('sheet');
    renderSheet();
  } catch (e: any) {
    toast(e.message, true);
  }
}
expose('openChar', openChar);

// ─── Character Sheet ───

// ─── Roll / Combat Actions ───

expose('applyDamage', async function () {
  if (!currentChar) return;
  const dmg = parseInt((document.getElementById('dmgInput') as HTMLInputElement)?.value || '0');
  if (!dmg) return;
  const oldHp = currentChar.hp_current;
  const newHp = Math.max(0, currentChar.hp_current - dmg);
  await updateField('hp_current', newHp);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  if (currentChar.concentrating_on) {
    try {
      const conc = await api('POST', `/api/characters/${currentChar.id}/check-concentration`, { damage: dmg });
      if (conc.needs_check) {
        toast(`Concentration check: DC ${conc.dc} (${conc.damage} damage to ${conc.spell_name})`);
        showModal('Concentration Check', `
          <p>You are concentrating on <strong>${esc(conc.spell_name)}</strong>.</p>
          <p>Damage taken: <strong>${conc.damage}</strong></p>
          <p class="fw-bold fs-5">CON Save DC ${conc.dc}</p>
          <div class="d-flex gap-2">
            <button class="btn btn-success flex-grow-1" onclick="doConcentrationSave(${conc.dc})"><i class="fa-solid fa-dice me-1"></i>Roll Save</button>
            <button class="btn btn-danger flex-grow-1" onclick="loseConcentration()"><i class="fa-solid fa-xmark me-1"></i>Lose Spell</button>
          </div>
        `);
      }
    } catch {}
  }
  renderSheet();
  // Animate HP change after re-render
  const bar = document.getElementById('charHpBarFill');
  const hpText = document.getElementById('charHpText');
  if (bar && hpText) {
    bar.style.width = Math.max(0, Math.min(100, (oldHp / currentChar.hp_max) * 100)) + '%';
    animateHpChange(hpText, bar, oldHp, currentChar.hp_current, currentChar.hp_max);
  }
});

// ─── Combat ───

// ─── Exhaustion ───

expose('adjustExhaustion', async function (delta: number) {
  if (!currentChar) return;
  const newLevel = Math.max(0, Math.min(6, (currentChar.exhaustion_level || 0) + delta));
  await api('PATCH', `/api/characters/${currentChar.id}/exhaustion`, { exhaustion_level: newLevel });
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderStats();
});

// ─── Identify Toggle ───

expose('toggleIdentify', async function (id: number) {
  const item = currentChar.inventory.find((i:any) => i.id === id);
  if (!item) return;
  const newVal = item.is_identified === false ? true : false;
  await api('PUT', `/api/inventory/${id}`, { ...item, is_identified: newVal });
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast(newVal ? 'Item identified' : 'Item marked unidentified');
});

expose('addInventory', function () {
  showModal('Add Item', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="invName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Quantity</label><input class="form-control" id="invQty" type="number" value="1"></div>
      <div class="col-6"><label class="form-label">Weight (lbs)</label><input class="form-control" id="invWeight" type="number" value="0" step="0.1"></div>
    </div>
    <div class="mb-3"><label class="form-label">Category</label>
      <select class="form-select" id="invCat">
        <option value="gear">Gear</option><option value="weapon">Weapon</option><option value="armor">Armor</option>
        <option value="potion">Potion</option><option value="scroll">Scroll</option><option value="tool">Tool</option>
        <option value="wondrous">Wondrous Item</option><option value="other">Other</option>
      </select></div>
    <button class="btn btn-primary w-100" onclick="saveInventory(this)"><i class="fa-solid fa-plus me-1"></i>Add</button>
  `);
});

expose('editInventory', function (id:number,name:string,qty:number,cat:string,weight:number,equipped:boolean) {
  showModal('Edit Item', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="invName" value="${esc(name)}"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Quantity</label><input class="form-control" id="invQty" type="number" value="${qty}"></div>
      <div class="col-6"><label class="form-label">Weight (lbs)</label><input class="form-control" id="invWeight" type="number" value="${weight}" step="0.1"></div>
    </div>
    <div class="mb-3"><label class="form-label">Category</label>
      <select class="form-select" id="invCat">${['gear','weapon','armor','potion','scroll','tool','wondrous','other'].map(c=>`<option value="${c}"${c===cat?' selected':''}>${capitalize(c)}</option>`).join('')}</select></div>
    <div class="mb-3"><div class="form-check"><input type="checkbox" class="form-check-input" id="invEquip"${equipped?' checked':''}><label class="form-check-label">Equipped</label></div></div>
    <button class="btn btn-primary w-100" onclick="saveEditInventory(${id},this)">Save</button>
  `);
});

expose('saveInventory', async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/inventory`, {
    name: (document.getElementById('invName') as HTMLInputElement).value,
    quantity: +(document.getElementById('invQty') as HTMLInputElement).value || 1,
    weight: +(document.getElementById('invWeight') as HTMLInputElement).value || 0,
    category: (document.getElementById('invCat') as HTMLSelectElement).value,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast('Item added');
});

expose('saveEditInventory', async function (id:number, btn:HTMLElement) {
  await api('PUT', `/api/inventory/${id}`, {
    name: (document.getElementById('invName') as HTMLInputElement).value,
    quantity: +(document.getElementById('invQty') as HTMLInputElement).value || 1,
    weight: +(document.getElementById('invWeight') as HTMLInputElement).value || 0,
    category: (document.getElementById('invCat') as HTMLSelectElement).value,
    equipped: (document.getElementById('invEquip') as HTMLInputElement).checked,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast('Item updated');
});

expose('deleteInventory', async function (id:number) {
  if (!confirm('Remove this item?')) return;
  await api('DELETE', `/api/inventory/${id}`);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast('Item removed');
});

expose('toggleEquip', async function (id:number) {
  const item = currentChar.inventory.find((i:any) => i.id === id);
  if (!item) return;
  item.equipped = !item.equipped;
  await api('PUT', `/api/inventory/${id}`, item);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderInventory();
  toast(item.equipped ? 'Equipped' : 'Unequipped');
});

expose('enableSpellcasting', async function () {
  currentChar.spellcasting = {
    ability: 'int', save_dc: 10, attack_bonus: 0,
    slots_1_max: 2, slots_1_used: 0,
  };
  await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, currentChar.spellcasting);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
});

expose('addSpell', function () {
  showModal('Add Spell', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="spellName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="spellLevel" type="number" value="0" min="0" max="9"></div>
      <div class="col-6"><label class="form-label">School</label>
        <select class="form-select" id="spellSchool">${['Abjuration','Conjuration','Divination','Enchantment','Evocation','Illusion','Necromancy','Transmutation'].map(s=>`<option value="${s}">${s}</option>`).join('')}</select></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Casting Time</label><input class="form-control" id="spellCast" value="1 action"></div>
      <div class="col-6"><label class="form-label">Range</label><input class="form-control" id="spellRange" value="Self"></div>
    </div>
    <div class="mb-3"><label class="form-label">Components</label><input class="form-control" id="spellComp" value="V,S"></div>
    <div class="mb-3"><label class="form-label">Duration</label><input class="form-control" id="spellDur" value="Instantaneous"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="spellDesc" rows="3"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveSpell(this)">Add Spell</button>
  `);
});

expose('editSpell', function (id:number,name:string,level:number,school:string,prepared:boolean,comp:string,range:string,cast:string,dur:string,desc:string) {
  showModal('Edit Spell', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="spellName" value="${esc(name)}"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="spellLevel" type="number" value="${level}" min="0" max="9"></div>
      <div class="col-6"><label class="form-label">School</label>
        <select class="form-select" id="spellSchool">${['Abjuration','Conjuration','Divination','Enchantment','Evocation','Illusion','Necromancy','Transmutation'].map(s=>`<option value="${s}"${s===school?' selected':''}>${s}</option>`).join('')}</select></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Casting Time</label><input class="form-control" id="spellCast" value="${esc(cast)}"></div>
      <div class="col-6"><label class="form-label">Range</label><input class="form-control" id="spellRange" value="${esc(range)}"></div>
    </div>
    <div class="mb-3"><label class="form-label">Components</label><input class="form-control" id="spellComp" value="${esc(comp)}"></div>
    <div class="mb-3"><label class="form-label">Duration</label><input class="form-control" id="spellDur" value="${esc(dur)}"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="spellDesc" rows="3">${esc(desc)}</textarea></div>
    <div class="mb-3"><div class="form-check"><input type="checkbox" class="form-check-input" id="spellPrep"${prepared?' checked':''}><label class="form-check-label">Prepared</label></div></div>
    <button class="btn btn-primary w-100" onclick="saveEditSpell(${id},this)">Save Spell</button>
  `);
});

expose('saveSpell', async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/spells`, {
    name: (document.getElementById('spellName') as HTMLInputElement).value,
    level: +(document.getElementById('spellLevel') as HTMLInputElement).value || 0,
    school: (document.getElementById('spellSchool') as HTMLSelectElement).value,
    casting_time: (document.getElementById('spellCast') as HTMLInputElement).value,
    range: (document.getElementById('spellRange') as HTMLInputElement).value,
    components: (document.getElementById('spellComp') as HTMLInputElement).value,
    duration: (document.getElementById('spellDur') as HTMLInputElement).value,
    description: (document.getElementById('spellDesc') as HTMLTextAreaElement).value,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
  toast('Spell added');
});

expose('saveEditSpell', async function (id:number, btn:HTMLElement) {
  await api('PUT', `/api/spells/${id}`, {
    name: (document.getElementById('spellName') as HTMLInputElement).value,
    level: +(document.getElementById('spellLevel') as HTMLInputElement).value || 0,
    school: (document.getElementById('spellSchool') as HTMLSelectElement).value,
    casting_time: (document.getElementById('spellCast') as HTMLInputElement).value,
    range: (document.getElementById('spellRange') as HTMLInputElement).value,
    components: (document.getElementById('spellComp') as HTMLInputElement).value,
    duration: (document.getElementById('spellDur') as HTMLInputElement).value,
    description: (document.getElementById('spellDesc') as HTMLTextAreaElement).value,
    prepared: (document.getElementById('spellPrep') as HTMLInputElement).checked,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
  toast('Spell updated');
});

expose('deleteSpell', async function (id:number) {
  if (!confirm('Remove this spell?')) return;
  await api('DELETE', `/api/spells/${id}`);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
  toast('Spell removed');
});

expose('saveSpellPrep', async function () {
  const spellIds: number[] = [];
  (currentChar.spells || []).forEach((s:any) => {
    const cb = document.getElementById(`prep-${s.id}`) as HTMLInputElement;
    if (cb && cb.checked) spellIds.push(s.id);
  });
  await api('PUT', `/api/characters/${currentChar.id}/spells/prepare`, { spell_ids: spellIds });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSpells();
  toast('Spell preparation saved');
});

// ─── Features ───

function renderFeatures() {
  const feats = currentChar.features || [];
  document.getElementById('featuresSection')!.innerHTML = `
    <div class="d-flex justify-content-between align-items-center">
      <h5>Features & Proficiencies</h5>
      <button class="btn btn-primary btn-sm" onclick="addFeature()"><i class="fa-solid fa-plus me-1"></i>Add Feature</button>
    </div>
    <div class="mt-2">
      ${feats.map((f:any) => `
        <div class="card mb-2">
          <div class="card-body py-2 px-3">
            <div class="d-flex justify-content-between align-items-center">
              <div><span class="fw-bold">${esc(f.name)}</span>
                <span class="badge badge-blood ms-1">Lv ${f.level_gained}</span>
                ${f.source ? `<span class="badge badge-gold ms-1">${esc(f.source)}</span>` : ''}
                <p class="mb-0 mt-1 small text-muted">${esc(f.description)}</p></div>
              <button class="btn btn-sm btn-outline-danger" onclick="deleteFeature(${f.id})"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>
        </div>`).join('') || '<div class="empty-state"><i class="fa-solid fa-star fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Features Yet</p><p class="small text-muted">Track class, race, and feat features here.</p></div>'}
    </div>`;
}

expose('addFeature', function () {
  showModal('Add Feature', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="featName"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="featDesc" rows="3"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Source</label><input class="form-control" id="featSource" placeholder="Class, Race, etc."></div>
      <div class="col-6"><label class="form-label">Level Gained</label><input class="form-control" id="featLevel" type="number" value="1"></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveFeature(this)">Add Feature</button>
  `);
});

expose('saveFeature', async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/features`, {
    name: (document.getElementById('featName') as HTMLInputElement).value,
    description: (document.getElementById('featDesc') as HTMLTextAreaElement).value,
    source: (document.getElementById('featSource') as HTMLInputElement).value,
    level_gained: +(document.getElementById('featLevel') as HTMLInputElement).value || 1,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderFeatures();
  toast('Feature added');
});

expose('deleteFeature', async function (id:number) {
  await api('DELETE', `/api/features/${id}`);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderFeatures();
  toast('Feature removed');
});

// ─── Proficiencies ───

expose('addProf', function () {
  showModal('Add Proficiency', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="profName"></div>
    <div class="mb-3"><label class="form-label">Type</label>
      <select class="form-select" id="profType">
        <option value="skill">Skill</option><option value="save">Saving Throw</option><option value="tool">Tool</option>
        <option value="weapon">Weapon</option><option value="armor">Armor</option><option value="language">Language</option>
        <option value="other">Other</option>
      </select></div>
    <button class="btn btn-primary w-100" onclick="saveProf(this)">Add Proficiency</button>
  `);
});

expose('saveProf', async function (btn:HTMLElement) {
  await api('POST', '/api/proficiencies', {
    character_id: currentChar.id,
    type: (document.getElementById('profType') as HTMLSelectElement).value,
    name: (document.getElementById('profName') as HTMLInputElement).value,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSheet();
  toast('Proficiency added');
});

expose('deleteProf', async function (id:number) {
  await api('DELETE', `/api/proficiencies/${id}`);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSheet();
  toast('Proficiency removed');
});

// ─── Details ───

function renderDetails() {
  const c = currentChar;
  const el = document.getElementById('detailsSection')!;
  el.innerHTML = `
    <div class="row g-3">
      <div class="col-md-12 mb-2">
        <label class="form-label">Portrait</label>
        <div class="d-flex align-items-center gap-2">
          ${c.portrait_url ? `<img src="${esc(c.portrait_url)}" class="character-portrait-lg me-2" alt="">` : ''}
          <input type="file" class="form-control form-control-sm" id="portraitUpload" accept="image/*">
          <button class="btn btn-primary btn-sm" onclick="uploadPortrait()"><i class="fa-solid fa-upload me-1"></i>Upload</button>
          <button class="btn btn-outline-info btn-sm" onclick="browsePortrait()"><i class="fa-solid fa-image me-1"></i>Browse</button>
          ${c.portrait_url ? `<button class="btn btn-outline-danger btn-sm" onclick="clearPortrait()"><i class="fa-solid fa-xmark"></i></button>` : ''}
        </div>
      </div>
    </div>
    <div class="row g-3">
      <div class="col-md-4"><label class="form-label">Race</label><input class="form-control form-control-sm" value="${esc(c.race)}" oninput="autoSaveField('race',this)"></div>
      <div class="col-md-4"><label class="form-label">Class</label><input class="form-control form-control-sm" value="${esc(c.class)}" oninput="autoSaveField('class',this)"></div>
      <div class="col-md-4"><label class="form-label">Subclass</label><input class="form-control form-control-sm" value="${esc(c.subclass)}" oninput="autoSaveField('subclass',this)"></div>
    </div>
    <div class="row g-3 mt-1">
      <div class="col-md-4"><label class="form-label">Level</label><input class="form-control form-control-sm" type="number" value="${c.level}" oninput="autoSaveField('level',this)"></div>
      <div class="col-md-4"><label class="form-label">Background</label><input class="form-control form-control-sm" value="${esc(c.background)}" oninput="autoSaveField('background',this)"></div>
      <div class="col-md-4"><label class="form-label">Alignment</label><input class="form-control form-control-sm" value="${esc(c.alignment)}" oninput="autoSaveField('alignment',this)"></div>
    </div>
    <div class="mt-2 form-check">
      <input type="checkbox" class="form-check-input" id="hpAutoCalcCb" ${c.hp_auto_calc ? 'checked' : ''} onchange="autoSaveField('hp_auto_calc',this.checked)">
      <label class="form-check-label small" for="hpAutoCalcCb">Auto-calculate HP from classes</label>
      <button class="btn btn-sm btn-outline-gold ms-2" onclick="calcHP()"><i class="fa-solid fa-calculator me-1"></i>Recalculate HP</button>
    </div>
    <h5 class="mt-3">Multi-Class</h5>
    <div id="multiClassArea">
      ${(c.classes && c.classes.length ? c.classes.map((cc: any) => `
        <div class="inv-item">
          <span class="fw-bold">${esc(cc.class)}</span>
          ${cc.subclass ? `<span class="text-muted small">(${esc(cc.subclass)})</span>` : ''}
          <span class="badge badge-blood ms-1">Lv ${cc.level}</span>
          <span class="badge badge-muted ms-1">${esc(cc.hit_dice)}</span>
          <div class="d-flex gap-1">
            <button class="btn btn-sm btn-outline-primary" onclick="editClass(${cc.id})"><i class="fa-solid fa-pen"></i></button>
            <button class="btn btn-sm btn-outline-danger" onclick="deleteClass(${cc.id})"><i class="fa-solid fa-trash"></i></button>
          </div>
        </div>`).join('')
        : '<div class="text-muted small">Single class. Add multi-class entries below.</div>')}
    </div>
    <button class="btn btn-sm btn-outline-primary mt-1" onclick="addClass()"><i class="fa-solid fa-plus me-1"></i>Add Class</button>
    <hr class="my-3">
    ${['personality_traits','ideals','bonds','flaws','appearance'].map(f => `
      <div class="mb-3"><label class="form-label">${capitalize(f.replace(/_/g,' '))}</label>
      <textarea class="form-control form-control-sm" rows="2" oninput="autoSaveField('${f}',this)">${esc((c as any)[f])}</textarea></div>
    `).join('')}
    <div class="mb-3"><label class="form-label">Backstory</label>
    <textarea class="form-control form-control-sm" rows="4" oninput="autoSaveField('backstory',this)">${esc(c.backstory)}</textarea></div>
    <h5 class="mt-4">Currency <small class="text-muted fw-normal">(tap +/− to adjust)</small></h5>
    <div class="row g-3">
      ${['cp','sp','ep','gp','pp'].map(coin => `
        <div class="col-4 col-md-2"><label class="form-label small">${coin.toUpperCase()}</label>
        <div class="currency-stepper">
          <button class="stepper-btn" onclick="coinStepper('${coin}', -1)" aria-label="Decrease ${coin.toUpperCase()}">−</button>
          <span class="stepper-value" id="coin${coin}">${c.currency?.[coin]||0}</span>
          <button class="stepper-btn" onclick="coinStepper('${coin}', 1)" aria-label="Increase ${coin.toUpperCase()}">+</button>
        </div></div>
      `).join('')}
    </div>
    <div class="mt-3">
      <button class="btn btn-outline-primary btn-sm" onclick="shareCharacter()"><i class="fa-solid fa-share-nodes me-1"></i>Share Character</button>
    </div>`;
}
expose('renderDetails', renderDetails);







// ─── D3 Force Graph ───

function createForceGraph(
  container: HTMLElement,
  data: { nodes: any[], edges: any[] },
  groups: Record<string, { shape: string, color: string }>,
  options?: { linkDistance?: number, chargeStrength?: number }
) {
  const width = container.clientWidth || 800;
  const height = container.clientHeight || 600;

  container.innerHTML = '';

  const svg = d3.select(container)
    .append('svg')
    .attr('width', width)
    .attr('height', height)
    .style('background', 'var(--parchment-light)')
    .style('cursor', 'grab')
    .style('border-radius', '4px')
    .style('display', 'block');

  const strokeColor = '#2c1810';
  const edgeColor = '#8b7355';

  svg.append('defs').append('marker')
    .attr('id', 'arrowhead')
    .attr('viewBox', '0 -5 10 10')
    .attr('refX', 20)
    .attr('refY', 0)
    .attr('markerWidth', 6)
    .attr('markerHeight', 6)
    .attr('orient', 'auto')
    .append('path')
    .attr('d', 'M0,-5L10,0L0,5')
    .attr('fill', edgeColor);

  const g = svg.append('g');

  const zoom = d3.zoom<SVGSVGElement, unknown>()
    .scaleExtent([0.1, 4])
    .on('zoom', (event) => g.attr('transform', event.transform));
  svg.call(zoom);

  const link = g.append('g')
    .selectAll<SVGLineElement, any>('line')
    .data(data.edges)
    .join('line')
    .attr('stroke', edgeColor)
    .attr('stroke-width', (d: any) => d.width || 1)
    .attr('stroke-dasharray', (d: any) => d.dashes ? '6,3' : null)
    .attr('marker-end', 'url(#arrowhead)');

  const linkLabel = g.append('g')
    .selectAll<SVGTextElement, any>('text')
    .data(data.edges.filter((d: any) => d.label))
    .join('text')
    .text((d: any) => d.label)
    .attr('font-size', 10)
    .attr('font-family', 'Vollkorn')
    .attr('fill', '#5c3a2a')
    .attr('text-anchor', 'middle')
    .attr('dy', '-4');

  const node = g.append('g')
    .selectAll<SVGGElement, any>('g')
    .data(data.nodes)
    .join('g')
    .style('cursor', 'pointer');

  node.each(function (d: any) {
    const el = d3.select(this);
    const size = d.size || 15;
    const grp = groups[d.group] || { shape: 'dot', color: '#8b0000' };
    const color = d.color || grp.color;

    const shapeEl = (() => {
      switch (grp.shape) {
        case 'ellipse':
          return el.append('ellipse').attr('rx', size).attr('ry', size * 0.7);
        case 'square':
          return el.append('rect').attr('x', -size).attr('y', -size)
            .attr('width', size * 2).attr('height', size * 2).attr('rx', 3);
        case 'diamond': {
          const pts = `0,-${size} ${size},0 0,${size} -${size},0`;
          return el.append('polygon').attr('points', pts);
        }
        case 'star': {
          const pts: string[] = [];
          for (let i = 0; i < 10; i++) {
            const r = i % 2 === 0 ? size : size * 0.4;
            const a = (i * Math.PI) / 5 - Math.PI / 2;
            pts.push(`${(r * Math.cos(a)).toFixed(1)},${(r * Math.sin(a)).toFixed(1)}`);
          }
          return el.append('polygon').attr('points', pts.join(' '));
        }
        case 'hexagon': {
          const pts: string[] = [];
          for (let i = 0; i < 6; i++) {
            const a = (i * Math.PI * 2) / 6 - Math.PI / 2;
            pts.push(`${(size * Math.cos(a)).toFixed(1)},${(size * Math.sin(a)).toFixed(1)}`);
          }
          return el.append('polygon').attr('points', pts.join(' '));
        }
        case 'triangle':
          return el.append('polygon')
            .attr('points', `0,-${size} ${(size * 0.866).toFixed(1)},${(size * 0.5).toFixed(1)} -${(size * 0.866).toFixed(1)},${(size * 0.5).toFixed(1)}`);
        default:
          return el.append('circle').attr('r', size * 0.5);
      }
    })();

    shapeEl
      .attr('fill', color)
      .attr('stroke', strokeColor)
      .attr('stroke-width', 2);

    const labelSize = d.size > 20 ? 14 : 11;
    const dy = grp.shape === 'dot' ? size * 0.5 + 14 : size + 10;

    el.append('text')
      .text(d.label)
      .attr('dy', dy)
      .attr('text-anchor', 'middle')
      .attr('fill', strokeColor)
      .attr('font-family', 'Playfair Display')
      .attr('font-size', labelSize);

    el.on('mouseenter', () => shapeEl.attr('stroke', '#b8963e').attr('stroke-width', 3))
      .on('mouseleave', () => shapeEl.attr('stroke', strokeColor).attr('stroke-width', 2));
  });

  const drag = d3.drag<SVGGElement, any>()
    .on('start', (event, d) => {
      if (!event.active) sim.alphaTarget(0.3).restart();
      d.fx = d.x;
      d.fy = d.y;
    })
    .on('drag', (event, d) => { d.fx = event.x; d.fy = event.y; })
    .on('end', (event, d) => {
      if (!event.active) sim.alphaTarget(0);
      d.fx = null;
      d.fy = null;
    });

  node.call(drag as any);

  const sim = d3.forceSimulation(data.nodes)
    .force('link', d3.forceLink(data.edges.map((e: any) => ({ ...e, source: e.from, target: e.to })))
      .id((d: any) => d.id)
      .distance(options?.linkDistance || 200))
    .force('charge', d3.forceManyBody().strength(options?.chargeStrength || -300))
    .force('center', d3.forceCenter(width / 2, height / 2))
    .force('collision', d3.forceCollide().radius((d: any) => d.size + 20))
    .on('tick', () => {
      link
        .attr('x1', (d: any) => d.source.x)
        .attr('y1', (d: any) => d.source.y)
        .attr('x2', (d: any) => d.target.x)
        .attr('y2', (d: any) => d.target.y);
      linkLabel
        .attr('x', (d: any) => (d.source.x + d.target.x) / 2)
        .attr('y', (d: any) => (d.source.y + d.target.y) / 2);
      node.attr('transform', (d: any) => `translate(${d.x},${d.y})`);
    });

  const ro = new ResizeObserver(() => {
    const w = container.clientWidth;
    const h = container.clientHeight;
    svg.attr('width', w).attr('height', h);
    sim.force('center', d3.forceCenter(w / 2, h / 2)).alpha(0.3).restart();
  });
  ro.observe(container);

  return sim;
}

// ─── Graph ───

async function renderGraph() {
  const el = document.getElementById('graphSection')!;
  el.innerHTML = `<div class="ornament mb-3">✧ Drawing your web of fate ✧</div>
    <div id="graphContainer" style="width:100%;height:600px;border:1px solid var(--border);border-radius:4px;background:var(--parchment-light)"></div>`;
  try {
    const data = await api('GET', `/api/characters/${currentChar.id}/graph`);
    const container = document.getElementById('graphContainer')!;
    createForceGraph(container, data, {
      character: { shape: 'ellipse', color: '#8b0000' },
      location: { shape: 'square', color: '#b8963e' },
      npc: { shape: 'diamond', color: '#2d6a2d' },
      quest: { shape: 'star', color: '#8b4513' },
      session: { shape: 'dot', color: '#5c3a2a' },
    }, { linkDistance: 200, chargeStrength: -300 });
  } catch (e:any) {
    el.innerHTML += `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load graph: ${esc(e.message)}</p></div>`;
  }
}
expose('renderGraph', renderGraph);

// ─── Analytics ───

async function renderAnalytics() {
  const el = document.getElementById('analyticsSection')!;
  el.innerHTML = '<div class="ornament mb-3">✧ Loading analytics... ✧</div>';
  try {
    const stats = await api('GET', `/api/characters/${currentChar.id}/stats`);
    el.innerHTML = `
      <h5>Campaign Overview</h5>
      <div class="row g-3 mb-3">
        <div class="col-6 col-md-3"><div class="combat-stat"><div class="stat-label">Sessions</div><div class="stat-value">${stats.session_count}</div></div></div>
        <div class="col-6 col-md-3"><div class="combat-stat"><div class="stat-label">Level</div><div class="stat-value">${stats.level}</div></div></div>
        <div class="col-6 col-md-3"><div class="combat-stat text-success"><div class="stat-label">Total XP</div><div class="stat-value">${stats.total_xp_earned}</div></div></div>
        <div class="col-6 col-md-3"><div class="combat-stat" style="color:var(--gold)"><div class="stat-label">Gold Earned</div><div class="stat-value">${stats.total_gold_earned}</div></div></div>
      </div>
      <div class="row g-3 mb-3">
        <div class="col-md-6">
          <div class="card">
            <div class="card-body">
              <h6>Quests (${stats.quests.total})</h6>
              <div class="d-flex gap-1 flex-wrap">
                ${stats.quests.active > 0 ? `<span class="badge badge-blood">${stats.quests.active} Active</span>` : ''}
                ${stats.quests.complete > 0 ? `<span class="badge bg-success">${stats.quests.complete} Complete</span>` : ''}
                ${stats.quests.failed > 0 ? `<span class="badge bg-secondary">${stats.quests.failed} Failed</span>` : ''}
                ${stats.quests.available > 0 ? `<span class="badge badge-gold">${stats.quests.available} Available</span>` : ''}
              </div>
            </div>
          </div>
        </div>
        <div class="col-md-6">
          <div class="card">
            <div class="card-body">
              <h6>Rests</h6>
              <div class="d-flex gap-1 flex-wrap">
                <span class="badge badge-gold">${stats.rests.short} Short</span>
                <span class="badge badge-blood">${stats.rests.long} Long</span>
                ${stats.rests.total_healed > 0 ? `<span class="badge bg-success">${stats.rests.total_healed} HP Healed</span>` : ''}
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="row g-3 mb-3">
        <div class="col-md-6">
          <div class="card">
            <div class="card-body">
              <h6>World</h6>
              <p class="mb-1 small text-muted">${stats.locations_count} Locations explored</p>
              <p class="mb-1 small text-muted">${stats.npc_interactions} NPC interactions</p>
              <p class="mb-1 small text-muted">${stats.journal_count} Journal entries</p>
              <p class="mb-0 small text-muted">${stats.dice_rolls.total_rolls} Dice rolls (avg ${stats.dice_rolls.average.toFixed(1)})</p>
            </div>
          </div>
        </div>
        <div class="col-md-6">
          <div class="card">
            <div class="card-body">
              <h6>Notable NPCs</h6>
              ${stats.top_npcs && stats.top_npcs.length > 0
                ? stats.top_npcs.map((n:any) => `<p class="mb-1 small text-muted">&loz; ${esc(n)}</p>`).join('')
                : '<p class="mb-0 small text-muted fst-italic">No NPC interactions yet</p>'}
            </div>
          </div>
        </div>
      </div>
      <div id="questChartContainer" style="height:200px;max-width:400px;margin:0 auto"></div>`;
    if ((typeof Chart !== 'undefined') && stats.quests.total > 0) {
      const ctx = document.createElement('canvas');
      document.getElementById('questChartContainer')!.appendChild(ctx);
      new Chart(ctx, {
        type: 'doughnut',
        data: {
          labels: ['Active', 'Complete', 'Failed', 'Available', 'Abandoned'],
          datasets: [{
            data: [stats.quests.active, stats.quests.complete, stats.quests.failed, stats.quests.available, stats.quests.abandoned],
            backgroundColor: ['#8b0000', '#2d6a2d', '#666', '#b8963e', '#ccc'],
            borderWidth: 0,
          }]
        },
        options: {
          responsive: true, maintainAspectRatio: false,
          plugins: { legend: { position: 'bottom', labels: { font: { family: 'Vollkorn' } } } }
        }
      });
    }
  } catch (e:any) {
    el.innerHTML = `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load analytics: ${esc(e.message)}</p></div>`;
  }
}
expose('renderAnalytics', renderAnalytics);

// ─── New Character ───

expose('newChar', function () {
  showModal('New Character', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="newName" placeholder="Character name"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Race</label><input class="form-control" id="newRace" list="raceSuggestions"><datalist id="raceSuggestions"></datalist></div>
      <div class="col-6"><label class="form-label">Class</label><input class="form-control" id="newClass" list="classSuggestions"><datalist id="classSuggestions"></datalist></div>
    </div>
    <button class="btn btn-primary w-100" onclick="createChar()"><i class="fa-solid fa-plus me-1"></i>Create</button>
    <div class="text-center mt-2"><button class="btn btn-sm btn-outline-gold" onclick="generateRandomChar()"><i class="fa-solid fa-dice me-1"></i>Random Character</button></div>
  `);
  fetch('/api/compendium/races', { credentials: 'include' }).then(r => r.json()).then((races:any[]) => {
    document.getElementById('raceSuggestions')!.innerHTML = races.map((r:any) => `<option value="${esc(r.name)}">`).join('');
  }).catch(() => {});
  fetch('/api/compendium/classes', { credentials: 'include' }).then(r => r.json()).then((cls:any[]) => {
    document.getElementById('classSuggestions')!.innerHTML = cls.map((c:any) => `<option value="${esc(c.name)}">`).join('');
  }).catch(() => {});
});

expose('createChar', async function () {
  const name = (document.getElementById('newName') as HTMLInputElement).value || 'Unnamed';
  const race = (document.getElementById('newRace') as HTMLInputElement).value;
  const cls = (document.getElementById('newClass') as HTMLInputElement).value;
  try {
    const char = await api('POST', '/api/characters', { name, race, class: cls });
    hideModal();
    if (char.id) await openChar(char.id);
    loadCharacters();
  } catch (e:any) {
    toast(e.message, true);
  }
});

// ─── Import / Export ───

expose('showImport', function () {
  showModal('Import Character', `
    <p class="text-muted fst-italic small mb-3">Paste JSON or upload a file</p>
    <div class="mb-3"><label class="form-label">JSON</label><textarea class="form-control" id="importJson" rows="6" style="font-family:monospace;font-size:0.8rem"></textarea></div>
    <div class="mb-3"><label class="form-label">File</label><input class="form-control" type="file" id="importFile" accept=".json"></div>
    <button class="btn btn-primary w-100" onclick="doImport()"><i class="fa-solid fa-file-import me-1"></i>Import</button>
  `);
});

expose('doImport', async function () {
  const jsonEl = document.getElementById('importJson') as HTMLTextAreaElement;
  const fileEl = document.getElementById('importFile') as HTMLInputElement;
  try {
    let result;
    if (fileEl.files && fileEl.files[0]) {
      const form = new FormData();
      form.append('file', fileEl.files[0]);
      const res = await fetch('/api/characters/import', { method: 'POST', headers: { 'X-CSRF-Token': getCsrfToken() }, credentials: 'include', body: form });
      result = await res.json();
    } else if (jsonEl.value.trim()) {
      result = await api('POST', '/api/characters/import', JSON.parse(jsonEl.value));
    } else {
      toast('Provide JSON or a file', true);
      return;
    }
    toast(`Imported ${Array.isArray(result) ? result.length : 1} character(s)`);
    hideModal();
    loadCharacters();
  } catch (e:any) {
    toast('Import failed: ' + e.message, true);
  }
});

expose('exportChar', async function () {
  if (!currentChar) return;
  try {
    const data = await api('GET', `/api/characters/${currentChar.id}/export`);
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const a = document.createElement('a');
    const url = URL.createObjectURL(blob);
    a.href = url;
    a.download = currentChar.name.replace(/[^a-zA-Z0-9]/g, '_') + '.json';
    a.click();
    URL.revokeObjectURL(url);
  } catch (e:any) {
    toast(e.message, true);
  }
});

// ─── Print ───

expose('printChar', async function () {
  if (!currentChar) return;
  try {
    const res = await fetch(`/api/characters/${currentChar.id}/print`, {
      headers: { 'X-CSRF-Token': getCsrfToken() }, credentials: 'include',
    });
    const text = await res.text();
    const win = window.open('', '_blank');
    if (win) {
      win.document.write(`<pre style="font-family:monospace;font-size:12px;line-height:1.4">${esc(text)}</pre>`);
      win.document.close();
      win.print();
    }
  } catch (e:any) {
    toast(e.message, true);
  }
});

// ─── Transfer UI (import/export via villum-transfer format) ───
import './transfer';

// ─── Party & Campaign → extracted to ts/party.ts ───
import './party';

// ─── Delete Character ───

expose('deleteChar', async function () {
  if (!currentChar) return;
  if (!confirm('Delete this character?')) return;
  try {
    await api('DELETE', `/api/characters/${currentChar.id}`);
    setCurrentChar(null);
    showView('characters');
    await loadCharacters();
    toast('Character deleted');
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Logout ───

expose('logout', async function () {
  await api('POST', '/api/logout');
  window.location.href = '/login';
});

// ─── Portrait Upload ───

expose('uploadPortrait', async function () {
  const input = document.getElementById('portraitUpload') as HTMLInputElement;
  if (!input.files || !input.files[0]) { toast('Select an image', true); return; }
  const form = new FormData();
  form.append('image', input.files[0]);
  try {
    const res = await fetch('/api/upload', {
      method: 'POST', headers: { 'X-CSRF-Token': getCsrfToken() }, credentials: 'include', body: form,
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Upload failed');
    await updateField('portrait_url', data.url);
    setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
    renderSheet();
    toast('Portrait uploaded');
  } catch (e: any) { toast(e.message, true); }
});

expose('browsePortrait', async function () {
  try {
    const url = await FilePicker.pick();
    await updateField('portrait_url', url);
    setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
    renderSheet();
    toast('Portrait set');
  } catch (e: any) { toast(e.message, true); }
});

expose('clearPortrait', async function () {
  await updateField('portrait_url', '');
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSheet();
  toast('Portrait removed');
});

// ─── Race Colors Management ───

expose('showRaceColors', async function () {
  try {
    const data = await api('GET', '/api/race-colors');
    const colors = data.colors || {};
    showModal('Race Colors', `
      <p class="small text-muted">Set colors for character races. These appear as colored badges on the campaign overview and character lists.</p>
      <div id="raceColorsList">
        ${Object.entries(colors).map(([race, color]) => `
          <div class="row g-2 mb-2 align-items-center">
            <div class="col-4"><label class="form-label mb-0 small">${esc(race)}</label></div>
            <div class="col-2"><input type="color" class="form-control form-control-color" value="${esc(color as string)}" data-race="${esc(race)}"></div>
            <div class="col-6"><input class="form-control form-control-sm" value="${esc(color as string)}" data-race="${esc(race)}" oninput="this.previousElementSibling.value=this.value"><span class="badge ms-1" style="background:${esc(color as string)};color:#fff">Preview</span></div>
          </div>
        `).join('')}
      </div>
      <div class="mt-3">
        <div class="row g-2 align-items-center">
          <div class="col-4"><input class="form-control form-control-sm" id="newRaceColorName" placeholder="New race"></div>
          <div class="col-2"><input type="color" class="form-control form-control-color" id="newRaceColorPicker" value="#6c757d"></div>
          <div class="col-3"><input class="form-control form-control-sm" id="newRaceColorValue" value="#6c757d" oninput="document.getElementById('newRaceColorPicker').value=this.value"></div>
          <div class="col-3"><button class="btn btn-sm btn-outline-primary w-100" onclick="addRaceColor()"><i class="fa-solid fa-plus me-1"></i>Add</button></div>
        </div>
      </div>
      <button class="btn btn-primary w-100 mt-3" onclick="saveRaceColors()"><i class="fa-solid fa-save me-1"></i>Save Changes</button>
    `);
  } catch (e: any) { toast(e.message, true); }
});

expose('addRaceColor', function () {
  const name = (document.getElementById('newRaceColorName') as HTMLInputElement).value.trim();
  const color = (document.getElementById('newRaceColorPicker') as HTMLInputElement).value;
  if (!name) { toast('Enter a race name', true); return; }
  const list = document.getElementById('raceColorsList')!;
  list.insertAdjacentHTML('beforeend', `
    <div class="row g-2 mb-2 align-items-center">
      <div class="col-4"><label class="form-label mb-0 small">${esc(name)}</label></div>
      <div class="col-2"><input type="color" class="form-control form-control-color" value="${color}" data-race="${esc(name)}"></div>
      <div class="col-6"><input class="form-control form-control-sm" value="${color}" data-race="${esc(name)}" oninput="this.previousElementSibling.value=this.value"></div>
    </div>
  `);
  (document.getElementById('newRaceColorName') as HTMLInputElement).value = '';
});

expose('saveRaceColors', async function () {
  const colors: Record<string, string> = {};
  document.querySelectorAll('#raceColorsList input[type="color"]').forEach((el) => {
    const input = el as HTMLInputElement;
    const race = input.getAttribute('data-race');
    if (race) colors[race] = input.value;
  });
  await api('PUT', '/api/race-colors', { colors });
  hideModal();
  toast('Race colors saved');
});

// ─── Multi-Class ───

expose('addClass', function () {
  showModal('Add Class', `
    <div class="mb-3"><label class="form-label">Class</label><input class="form-control" id="mcClass"></div>
    <div class="mb-3"><label class="form-label">Subclass</label><input class="form-control" id="mcSubclass"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="mcLevel" type="number" value="1" min="1"></div>
      <div class="col-6"><label class="form-label">Hit Dice</label>
        <select class="form-select" id="mcHD"><option value="d6">d6</option><option value="d8">d8</option><option value="d10" selected>d10</option><option value="d12">d12</option></select></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveClass()"><i class="fa-solid fa-plus me-1"></i>Add</button>
  `);
});

expose('saveClass', async function () {
  await api('POST', `/api/characters/${currentChar.id}/classes`, {
    class: (document.getElementById('mcClass') as HTMLInputElement).value,
    subclass: (document.getElementById('mcSubclass') as HTMLInputElement).value,
    level: +(document.getElementById('mcLevel') as HTMLInputElement).value || 1,
    hit_dice: (document.getElementById('mcHD') as HTMLSelectElement).value,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSheet();
  toast('Class added');
});

expose('editClass', function (id: number) {
  const cc = currentChar.classes.find((c: any) => c.id === id);
  if (!cc) return;
  showModal('Edit Class', `
    <div class="mb-3"><label class="form-label">Class</label><input class="form-control" id="mcClass" value="${esc(cc.class)}"></div>
    <div class="mb-3"><label class="form-label">Subclass</label><input class="form-control" id="mcSubclass" value="${esc(cc.subclass)}"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="mcLevel" type="number" value="${cc.level}" min="1"></div>
      <div class="col-6"><label class="form-label">Hit Dice</label>
        <select class="form-select" id="mcHD">${['d6','d8','d10','d12'].map(d => `<option value="${d}"${d===cc.hit_dice?' selected':''}>${d}</option>`).join('')}</select></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveEditClass(${id})"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
});

expose('saveEditClass', async function (id: number) {
  await api('PUT', `/api/classes/${id}`, {
    class: (document.getElementById('mcClass') as HTMLInputElement).value,
    subclass: (document.getElementById('mcSubclass') as HTMLInputElement).value,
    level: +(document.getElementById('mcLevel') as HTMLInputElement).value || 1,
    hit_dice: (document.getElementById('mcHD') as HTMLSelectElement).value,
  });
  hideModal();
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSheet();
  toast('Class updated');
});

expose('deleteClass', async function (id: number) {
  if (!confirm('Remove this class?')) return;
  await api('DELETE', `/api/classes/${id}`);
  setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
  renderSheet();
  toast('Class removed');
});

// ─── Encounter Builder → extracted to ts/encounter.ts ───
import './encounter';

// ─── Calendar ───

// ─── Timeline → extracted to ts/timeline.ts ───
import './timeline';

// ─── Conditions / Ailments ───

expose('showAddCondition', function () {
  showModal('Add Condition', `
    <div class="mb-3"><label class="form-label">Condition</label>
      <select class="form-select" id="condType">
        <option value="">Custom...</option>
        <option value="blinded">Blinded</option><option value="charmed">Charmed</option>
        <option value="deafened">Deafened</option><option value="exhaustion">Exhaustion</option>
        <option value="frightened">Frightened</option><option value="grappled">Grappled</option>
        <option value="incapacitated">Incapacitated</option><option value="invisible">Invisible</option>
        <option value="paralyzed">Paralyzed</option><option value="petrified">Petrified</option>
        <option value="poisoned">Poisoned</option><option value="prone">Prone</option>
        <option value="restrained">Restrained</option><option value="stunned">Stunned</option>
        <option value="unconscious">Unconscious</option><option value="concentration">Concentration</option>
      </select></div>
    <div class="mb-3" id="condCustomNameDiv"><label class="form-label">Custom Name</label><input class="form-control" id="condName" placeholder="e.g. Cursed"></div>
    <div class="row g-3 mb-3">
      <div class="col-4"><label class="form-label">Duration</label><input class="form-control" id="condDuration" type="number" value="1" min="0"></div>
      <div class="col-4"><label class="form-label">Unit</label>
        <select class="form-select" id="condDurationType">
          <option value="round">Rounds</option><option value="minute">Minutes</option>
          <option value="hour">Hours</option><option value="day">Days</option>
          <option value="permanent">Permanent</option>
        </select></div>
      <div class="col-4"><label class="form-label">Source</label><input class="form-control" id="condSource" placeholder="Spell/effect"></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Save Ends?</label><input class="form-control" id="condSave" placeholder="e.g. con"></div>
      <div class="col-6"><label class="form-label">Save DC</label><input class="form-control" id="condDC" type="number" value="0"></div>
    </div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="condDesc" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveCondition()"><i class="fa-solid fa-plus me-1"></i>Add Condition</button>
  `);
  const sel = document.getElementById('condType') as HTMLSelectElement;
  sel.addEventListener('change', () => {
    const customDiv = document.getElementById('condCustomNameDiv')!;
    customDiv.style.display = sel.value ? 'none' : 'block';
  });
});

expose('saveCondition', async function () {
  const sel = document.getElementById('condType') as HTMLSelectElement;
  const name = sel.value || (document.getElementById('condName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  await api('POST', '/api/conditions', {
    character_id: currentChar.id, name, type: sel.value || 'other',
    duration: +(document.getElementById('condDuration') as HTMLInputElement).value || 1,
    duration_type: (document.getElementById('condDurationType') as HTMLSelectElement).value,
    source: (document.getElementById('condSource') as HTMLInputElement).value,
    saving_throw: (document.getElementById('condSave') as HTMLInputElement).value,
    save_dc: +(document.getElementById('condDC') as HTMLInputElement).value || 0,
    description: (document.getElementById('condDesc') as HTMLTextAreaElement).value,
  });
  hideModal();
  renderCombat();
  toast('Condition added');
});

expose('tickConditions', async function () {
  if (!currentChar) return;
  const result = await api('POST', '/api/conditions/tick', {
    character_id: currentChar.id, count: 1, duration_type: 'round',
  });
  renderCombat();
  if (result.expired > 0) toast(`${result.expired} condition(s) expired`);
  else toast('Rounds advanced');
});

expose('deleteCondition', async function (id: number) {
  await api('DELETE', `/api/conditions/${id}`);
  renderCombat();
});

// ─── Feats ───

async function renderFeats() {
  const el = document.getElementById('featsSection')!;
  if (!currentChar) return;
  try {
    const feats = await api('GET', `/api/feats?character_id=${currentChar.id}`);
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center">
        <h5>Feats</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddFeat()"><i class="fa-solid fa-plus me-1"></i>Add Feat</button>
      </div>
      <div class="mt-2">
        ${feats.length ? feats.map((f: any) => `
          <div class="card mb-2">
            <div class="card-body py-2 px-3">
              <div class="d-flex justify-content-between align-items-start">
                <div>
                  <span class="fw-bold">${esc(f.name)}</span>
                  <span class="badge badge-blood ms-1">Lv ${f.level_gained}</span>
                  ${f.source ? `<span class="badge badge-gold ms-1">${esc(f.source)}</span>` : ''}
                  ${f.prerequisites ? `<span class="badge badge-muted ms-1">${esc(f.prerequisites)}</span>` : ''}
                  <p class="mb-0 mt-1 small text-muted">${esc(f.description)}</p>
                </div>
                <div class="d-flex gap-1">
                  <button class="btn btn-sm btn-outline-primary" onclick="showEditFeat(${f.id},'${esc(f.name)}','${esc(f.description)}','${esc(f.prerequisites)}','${esc(f.source)}',${f.level_gained})"><i class="fa-solid fa-pen"></i></button>
                  <button class="btn btn-sm btn-outline-danger" onclick="deleteFeat(${f.id})"><i class="fa-solid fa-trash"></i></button>
                </div>
              </div>
            </div>
          </div>`).join('')
          : '<div class="empty-state"><i class="fa-solid fa-star fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Feats</p><p class="small text-muted">Track your character feats here (distinct from class/race features).</p></div>'}
      </div>`;
  } catch { el.innerHTML = '<div class="empty-state"><p class="small text-muted">Could not load feats.</p></div>'; }
}

expose('showAddFeat', function () {
  editingFeatId = null;
  showModal('Add Feat', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="featName" list="featSuggestions">
      <datalist id="featSuggestions">
        ${['Alert','Athlete','Actor','Charger','Crossbow Expert','Defensive Duelist','Dual Wielder','Dungeon Delver','Durable','Elemental Adept','Grappler','Great Weapon Master','Healer','Heavily Armored','Heavy Armor Master','Inspiring Leader','Keen Mind','Lightly Armored','Linguist','Lucky','Mage Slayer','Magic Initiate','Martial Adept','Medium Armor Master','Mobile','Moderately Armored','Mounted Combatant','Observant','Polearm Master','Resilient','Ritual Caster','Sentinel','Sharpshooter','Shield Master','Skilled','Skulker','Spell Sniper','Tavern Brawler','Tough','War Caster','Weapon Master'].map(n => `<option value="${n}">`).join('')}
      </datalist></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="featDesc" rows="3"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Prerequisites</label><input class="form-control" id="featPrereq" placeholder="e.g. Str 13+"></div>
      <div class="col-6"><label class="form-label">Source</label><input class="form-control" id="featSource" placeholder="PHB, Tasha's, etc."></div>
    </div>
    <div class="mb-3"><label class="form-label">Level Gained</label><input class="form-control" id="featLevel" type="number" value="1"></div>
    <button class="btn btn-primary w-100" onclick="saveFeat()"><i class="fa-solid fa-plus me-1"></i>Add Feat</button>
  `);
});

let editingFeatId: number | null = null;

expose('showEditFeat', function (id: number, name: string, description: string, prerequisites: string, source: string, level: number) {
  editingFeatId = id;
  showModal('Edit Feat', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="featName" value="${esc(name)}" list="featSuggestions">
      <datalist id="featSuggestions">
        ${['Alert','Athlete','Actor','Charger','Crossbow Expert','Defensive Duelist','Dual Wielder','Dungeon Delver','Durable','Elemental Adept','Grappler','Great Weapon Master','Healer','Heavily Armored','Heavy Armor Master','Inspiring Leader','Keen Mind','Lightly Armored','Linguist','Lucky','Mage Slayer','Magic Initiate','Martial Adept','Medium Armor Master','Mobile','Moderately Armored','Mounted Combatant','Observant','Polearm Master','Resilient','Ritual Caster','Sentinel','Sharpshooter','Shield Master','Skilled','Skulker','Spell Sniper','Tavern Brawler','Tough','War Caster','Weapon Master'].map(n => `<option value="${n}">`).join('')}
      </datalist></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="featDesc" rows="3">${esc(description)}</textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Prerequisites</label><input class="form-control" id="featPrereq" value="${esc(prerequisites)}" placeholder="e.g. Str 13+"></div>
      <div class="col-6"><label class="form-label">Source</label><input class="form-control" id="featSource" value="${esc(source)}" placeholder="PHB, Tasha's, etc."></div>
    </div>
    <div class="mb-3"><label class="form-label">Level Gained</label><input class="form-control" id="featLevel" type="number" value="${level}"></div>
    <button class="btn btn-primary w-100" onclick="saveFeat()"><i class="fa-solid fa-floppy-disk me-1"></i>Save Feat</button>
  `);
});

expose('saveFeat', async function () {
  const name = (document.getElementById('featName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  const data = {
    name,
    description: (document.getElementById('featDesc') as HTMLTextAreaElement).value,
    prerequisites: (document.getElementById('featPrereq') as HTMLInputElement).value,
    source: (document.getElementById('featSource') as HTMLInputElement).value,
    level_gained: +(document.getElementById('featLevel') as HTMLInputElement).value || 1,
  };
  if (editingFeatId) {
    await api('PUT', `/api/feats/${editingFeatId}`, data);
    editingFeatId = null;
    toast('Feat updated');
  } else {
    await api('POST', '/api/feats', { ...data, character_id: currentChar.id });
    toast('Feat added');
  }
  hideModal();
  renderFeats();
});

expose('deleteFeat', async function (id: number) {
  if (!confirm('Remove this feat?')) return;
  await api('DELETE', `/api/feats/${id}`);
  renderFeats();
  toast('Feat removed');
});

// ─── Companions ───

async function renderCompanions() {
  const el = document.getElementById('companionsSection')!;
  if (!currentChar) return;
  try {
    const comps = await api('GET', `/api/companions?character_id=${currentChar.id}`);
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center">
        <h5>Companions & Mounts</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddCompanion()"><i class="fa-solid fa-plus me-1"></i>Add Companion</button>
      </div>
      <div class="row g-3 mt-2">
        ${comps.length ? comps.map((comp: any) => {
          const hpPct = comp.hp_max > 0 ? Math.round((comp.hp_current / comp.hp_max) * 100) : 0;
          const abilMod = (s: number) => { const m = Math.floor((s - 10) / 2); return m >= 0 ? '+' + m : '' + m; };
          return `<div class="col-md-6">
            <div class="card">
              <div class="card-body py-2 px-3">
                <div class="d-flex justify-content-between align-items-start">
                  <div>
                    <span class="fw-bold">${esc(comp.name)}</span>
                    <span class="badge badge-gold ms-1">${esc(comp.type)}</span>
                    ${comp.race_color ? `<span class="badge ms-1" style="background:${comp.race_color};color:#fff">${esc(comp.race)}</span>` : `<span class="badge badge-muted ms-1">${esc(comp.race)}</span>`}
                    ${!comp.is_alive ? '<span class="badge bg-danger ms-1">Deceased</span>' : ''}
                  </div>
                  <div class="d-flex gap-1">
                    <button class="btn btn-sm btn-outline-primary" onclick="editCompanion(${comp.id})"><i class="fa-solid fa-pen"></i></button>
                    <button class="btn btn-sm btn-outline-danger" onclick="deleteCompanion(${comp.id})"><i class="fa-solid fa-trash"></i></button>
                  </div>
                </div>
                <div class="hp-bar mt-2" style="height:8px">
                  <div class="hp-bar-fill" style="width:${hpPct}%;height:100%"></div>
                </div>
                <div class="small text-muted mt-1">HP: ${comp.hp_current}/${comp.hp_max} · AC: ${comp.ac} · Spd: ${comp.speed}</div>
                <div class="small text-muted">STR ${comp.str}(${abilMod(comp.str)}) DEX ${comp.dex}(${abilMod(comp.dex)}) CON ${comp.con}(${abilMod(comp.con)}) INT ${comp.int}(${abilMod(comp.int)}) WIS ${comp.wis}(${abilMod(comp.wis)}) CHA ${comp.cha}(${abilMod(comp.cha)})</div>
                ${comp.abilities ? `<div class="small text-muted mt-1"><i class="fa-solid fa-star me-1"></i>${esc(comp.abilities)}</div>` : ''}
                ${comp.notes ? `<div class="small text-muted">${esc(comp.notes)}</div>` : ''}
              </div>
            </div>
          </div>`;
        }).join('')
        : '<div class="empty-state"><i class="fa-solid fa-dog fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Companions</p><p class="small text-muted">Track familiars, mounts, animal companions, and summoned creatures.</p></div>'}
      </div>`;
  } catch { el.innerHTML = '<div class="empty-state"><p class="small text-muted">Could not load companions.</p></div>'; }
}

expose('showAddCompanion', function () {
  showModal('Add Companion', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="compName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Type</label>
        <select class="form-select" id="compType">
          <option value="familiar">Familiar</option><option value="mount">Mount</option>
          <option value="companion">Companion</option><option value="summoned">Summoned</option>
          <option value="pet">Pet</option>
        </select></div>
      <div class="col-6"><label class="form-label">Race</label><input class="form-control" id="compRace" placeholder="Owl, Warhorse, etc."></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-4"><label class="form-label">HP Max</label><input class="form-control" id="compHP" type="number" value="10"></div>
      <div class="col-4"><label class="form-label">AC</label><input class="form-control" id="compAC" type="number" value="10"></div>
      <div class="col-4"><label class="form-label">Speed</label><input class="form-control" id="compSpeed" type="number" value="30"></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-4"><label class="form-label">STR</label><input class="form-control" id="compStr" type="number" value="10"></div>
      <div class="col-4"><label class="form-label">DEX</label><input class="form-control" id="compDex" type="number" value="10"></div>
      <div class="col-4"><label class="form-label">CON</label><input class="form-control" id="compCon" type="number" value="10"></div>
      <div class="col-4"><label class="form-label">INT</label><input class="form-control" id="compInt" type="number" value="10"></div>
      <div class="col-4"><label class="form-label">WIS</label><input class="form-control" id="compWis" type="number" value="10"></div>
      <div class="col-4"><label class="form-label">CHA</label><input class="form-control" id="compCha" type="number" value="10"></div>
    </div>
    <div class="mb-3"><label class="form-label">Abilities</label><textarea class="form-control" id="compAbilities" rows="2" placeholder="Flyby, Darkvision 60ft, etc."></textarea></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="compNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveCompanion()"><i class="fa-solid fa-plus me-1"></i>Add</button>
  `);
});

expose('saveCompanion', async function () {
  const name = (document.getElementById('compName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  await api('POST', '/api/companions', {
    character_id: currentChar.id, name,
    type: (document.getElementById('compType') as HTMLSelectElement).value,
    race: (document.getElementById('compRace') as HTMLInputElement).value,
    hp_max: +(document.getElementById('compHP') as HTMLInputElement).value || 10,
    hp_current: +(document.getElementById('compHP') as HTMLInputElement).value || 10,
    ac: +(document.getElementById('compAC') as HTMLInputElement).value || 10,
    str: +(document.getElementById('compStr') as HTMLInputElement).value || 10,
    dex: +(document.getElementById('compDex') as HTMLInputElement).value || 10,
    con: +(document.getElementById('compCon') as HTMLInputElement).value || 10,
    int: +(document.getElementById('compInt') as HTMLInputElement).value || 10,
    wis: +(document.getElementById('compWis') as HTMLInputElement).value || 10,
    cha: +(document.getElementById('compCha') as HTMLInputElement).value || 10,
    speed: +(document.getElementById('compSpeed') as HTMLInputElement).value || 30,
    abilities: (document.getElementById('compAbilities') as HTMLTextAreaElement).value,
    notes: (document.getElementById('compNotes') as HTMLTextAreaElement).value,
    is_alive: true,
  });
  hideModal();
  renderCompanions();
  toast('Companion added');
});

expose('editCompanion', async function (id: number) {
  const comps = await api('GET', `/api/companions?character_id=${currentChar.id}`);
  const comp = comps.find((c: any) => c.id === id);
  if (!comp) return;
  showModal('Edit Companion', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="compName" value="${esc(comp.name)}"></div>
    <div class="mb-3">
      <label class="form-label">Portrait</label>
      <div class="d-flex align-items-center gap-2">
        ${comp.portrait_url ? `<img src="${esc(comp.portrait_url)}" class="character-portrait-lg me-2" alt="">` : ''}
        <input type="file" class="form-control form-control-sm" id="compPortraitUpload" accept="image/*">
        <button class="btn btn-primary btn-sm" onclick="uploadCompPortrait(${comp.id})"><i class="fa-solid fa-upload me-1"></i>Upload</button>
        <button class="btn btn-outline-info btn-sm" onclick="browseCompPortrait(${comp.id})"><i class="fa-solid fa-image me-1"></i>Browse</button>
        ${comp.portrait_url ? `<button class="btn btn-outline-danger btn-sm" onclick="clearCompPortrait(${comp.id})"><i class="fa-solid fa-xmark"></i></button>` : ''}
      </div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Type</label>
        <select class="form-select" id="compType">${['familiar','mount','companion','summoned','pet'].map(t => `<option value="${t}"${t===comp.type?' selected':''}>${capitalize(t)}</option>`).join('')}</select></div>
      <div class="col-6"><label class="form-label">Race</label><input class="form-control" id="compRace" value="${esc(comp.race)}"></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-4"><label class="form-label">HP Max</label><input class="form-control" id="compHP" type="number" value="${comp.hp_max}"></div>
      <div class="col-4"><label class="form-label">HP Current</label><input class="form-control" id="compHPCur" type="number" value="${comp.hp_current}"></div>
      <div class="col-4"><label class="form-label">AC</label><input class="form-control" id="compAC" type="number" value="${comp.ac}"></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-2"><label class="form-label">STR</label><input class="form-control" id="compStr" type="number" value="${comp.str}"></div>
      <div class="col-2"><label class="form-label">DEX</label><input class="form-control" id="compDex" type="number" value="${comp.dex}"></div>
      <div class="col-2"><label class="form-label">CON</label><input class="form-control" id="compCon" type="number" value="${comp.con}"></div>
      <div class="col-2"><label class="form-label">INT</label><input class="form-control" id="compInt" type="number" value="${comp.int}"></div>
      <div class="col-2"><label class="form-label">WIS</label><input class="form-control" id="compWis" type="number" value="${comp.wis}"></div>
      <div class="col-2"><label class="form-label">CHA</label><input class="form-control" id="compCha" type="number" value="${comp.cha}"></div>
    </div>
    <div class="mb-3"><label class="form-label">Abilities</label><textarea class="form-control" id="compAbilities" rows="2">${esc(comp.abilities)}</textarea></div>
    <button class="btn btn-primary w-100" onclick="saveEditCompanion(${id})"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
});

// ─── Companion Portrait ───

let compPortraitUrl: string = '';

expose('uploadCompPortrait', async function (id: number) {
  const input = document.getElementById('compPortraitUpload') as HTMLInputElement;
  if (!input.files || !input.files[0]) { toast('Select an image', true); return; }
  const form = new FormData();
  form.append('image', input.files[0]);
  try {
    const res = await fetch('/api/upload', {
      method: 'POST', headers: { 'X-CSRF-Token': getCsrfToken() }, credentials: 'include', body: form,
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Upload failed');
    compPortraitUrl = data.url;
    hideModal();
    toast('Portrait uploaded — re-open to confirm');
  } catch (e: any) { toast(e.message, true); }
});

expose('browseCompPortrait', async function (id: number) {
  try {
    const url = await FilePicker.pick();
    compPortraitUrl = url;
    hideModal();
    toast('Portrait selected — re-open to confirm');
  } catch (e: any) { toast(e.message, true); }
});

expose('clearCompPortrait', async function (id: number) {
  compPortraitUrl = '';
  hideModal();
  toast('Portrait cleared');
});

expose('saveEditCompanion', async function (id: number) {
  const portraitUrl = compPortraitUrl || (document.getElementById('compPortraitUrl') as HTMLInputElement)?.value || '';
  await api('PUT', `/api/companions/${id}`, {
    name: (document.getElementById('compName') as HTMLInputElement).value,
    type: (document.getElementById('compType') as HTMLSelectElement).value,
    race: (document.getElementById('compRace') as HTMLInputElement).value,
    hp_max: +(document.getElementById('compHP') as HTMLInputElement).value || 10,
    hp_current: +(document.getElementById('compHPCur') as HTMLInputElement).value || 10,
    ac: +(document.getElementById('compAC') as HTMLInputElement).value || 10,
    str: +(document.getElementById('compStr') as HTMLInputElement).value || 10,
    dex: +(document.getElementById('compDex') as HTMLInputElement).value || 10,
    con: +(document.getElementById('compCon') as HTMLInputElement).value || 10,
    int: +(document.getElementById('compInt') as HTMLInputElement).value || 10,
    wis: +(document.getElementById('compWis') as HTMLInputElement).value || 10,
    cha: +(document.getElementById('compCha') as HTMLInputElement).value || 10,
    speed: +(document.getElementById('compSpeed') as HTMLInputElement).value || 30,
    portrait_url: portraitUrl,
    abilities: (document.getElementById('compAbilities') as HTMLTextAreaElement).value,
    notes: (document.getElementById('compNotes') as HTMLTextAreaElement).value,
    is_alive: true,
  });
  compPortraitUrl = '';
  hideModal();
  renderCompanions();
  toast('Companion updated');
});

expose('deleteCompanion', async function (id: number) {
  if (!confirm('Remove this companion?')) return;
  await api('DELETE', `/api/companions/${id}`);
  renderCompanions();
  toast('Companion removed');
});

// ─── Notes ───

async function renderNotes() {
  const el = document.getElementById('notesSection')!;
  if (!currentChar) return;
  try {
    const notes = await api('GET', `/api/notes?character_id=${currentChar.id}`);
    const groups: Record<string, any[]> = { general: [], backstory: [], quest: [], lore: [], dm: [], other: [] };
    notes.forEach((n: any) => { if (groups[n.category]) groups[n.category].push(n); else groups.other.push(n); });
    let html = `
      <div class="d-flex justify-content-between align-items-center">
        <h5>Notes</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddNote()"><i class="fa-solid fa-plus me-1"></i>New Note</button>
      </div>`;
    for (const [cat, items] of Object.entries(groups)) {
      if (!items.length) continue;
      html += `<h6 class="mt-3 text-muted">${capitalize(cat)}</h6>`;
      for (const n of items) {
        const visIcon = n.visibility === 'dm' ? '<i class="fa-solid fa-eye-slash ms-1 text-muted" title="DM only"></i>' : '';
        html += `<div class="card mb-2">
          <div class="card-body py-2 px-3">
            <div class="d-flex justify-content-between align-items-start">
              <div><span class="fw-bold">${esc(n.title)}</span> ${visIcon}
                <span class="badge badge-muted ms-1">${esc(n.visibility)}</span></div>
              <div class="d-flex gap-1">
                <button class="btn btn-sm btn-outline-primary" onclick="editNote(${n.id})"><i class="fa-solid fa-pen"></i></button>
                <button class="btn btn-sm btn-outline-danger" onclick="deleteNote(${n.id})"><i class="fa-solid fa-trash"></i></button>
              </div>
            </div>
            <div class="mt-1 small text-muted" style="white-space:pre-wrap">${esc(n.content).substring(0, 300)}</div>
          </div>
        </div>`;
      }
    }
    if (!notes.length) html += '<div class="empty-state"><i class="fa-solid fa-note-sticky fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">No notes yet. Keep track of campaign information, backstory details, and DM secrets.</p></div>';
    el.innerHTML = html;
  } catch { el.innerHTML = '<div class="empty-state"><p class="small text-muted">Could not load notes.</p></div>'; }
}

expose('showAddNote', function () {
  showModal('New Note', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="noteTitle" placeholder="Note title"></div>
    <div class="mb-3"><label class="form-label">Content</label><textarea class="form-control" id="noteContent" rows="6"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Visibility</label>
        <select class="form-select" id="noteVis">
          <option value="player">Player Only</option><option value="both">Player & DM</option>
          <option value="dm">DM Only</option>
        </select></div>
      <div class="col-6"><label class="form-label">Category</label>
        <select class="form-select" id="noteCat">
          <option value="general">General</option><option value="backstory">Backstory</option>
          <option value="quest">Quest</option><option value="lore">Lore</option>
          <option value="dm">DM</option><option value="other">Other</option>
        </select></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveNote()"><i class="fa-solid fa-plus me-1"></i>Create Note</button>
  `);
});

expose('saveNote', async function () {
  await api('POST', '/api/notes', {
    character_id: currentChar.id,
    title: (document.getElementById('noteTitle') as HTMLInputElement).value,
    content: (document.getElementById('noteContent') as HTMLTextAreaElement).value,
    visibility: (document.getElementById('noteVis') as HTMLSelectElement).value,
    category: (document.getElementById('noteCat') as HTMLSelectElement).value,
  });
  hideModal();
  renderNotes();
  toast('Note created');
});

expose('editNote', async function (id: number) {
  const notes = await api('GET', `/api/notes?character_id=${currentChar.id}`);
  const n = notes.find((x: any) => x.id === id);
  if (!n) return;
  showModal('Edit Note', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="noteTitle" value="${esc(n.title)}"></div>
    <div class="mb-3"><label class="form-label">Content</label><textarea class="form-control" id="noteContent" rows="6">${esc(n.content)}</textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Visibility</label>
        <select class="form-select" id="noteVis">${['player','both','dm'].map(v => `<option value="${v}"${v===n.visibility?' selected':''}>${capitalize(v)}</option>`).join('')}</select></div>
      <div class="col-6"><label class="form-label">Category</label>
        <select class="form-select" id="noteCat">${['general','backstory','quest','lore','dm','other'].map(c => `<option value="${c}"${c===n.category?' selected':''}>${capitalize(c)}</option>`).join('')}</select></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveEditNote(${id})"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
});

expose('saveEditNote', async function (id: number) {
  await api('PUT', `/api/notes/${id}`, {
    title: (document.getElementById('noteTitle') as HTMLInputElement).value,
    content: (document.getElementById('noteContent') as HTMLTextAreaElement).value,
    visibility: (document.getElementById('noteVis') as HTMLSelectElement).value,
    category: (document.getElementById('noteCat') as HTMLSelectElement).value,
  });
  hideModal();
  renderNotes();
  toast('Note updated');
});

expose('deleteNote', async function (id: number) {
  if (!confirm('Delete this note?')) return;
  await api('DELETE', `/api/notes/${id}`);
  renderNotes();
  toast('Note deleted');
});

// ─── Factions → extracted to ts/factions.ts ───
import './factions';

// ─── Shops State ───

let selectedShopShops: any[] = [];
let allCampaigns: any[] = [];
let compendiumEquipment: any[] = [];

async function ensureCompendiumEquipment() {
  if (compendiumEquipment.length === 0) {
    try { compendiumEquipment = await api('GET', '/api/compendium/equipment'); } catch {}
  }
}

// ─── Main Shops View ───

expose('showShops', async function () {
  showView('shops');
  const el = document.getElementById('shopsContent')!;
  try {
    const [shops, campaigns] = await Promise.all([
      api('GET', '/api/shops'),
      api('GET', '/api/campaigns').catch(() => []),
    ]);
    allCampaigns = campaigns;
    selectedShopShops = shops;

    const isDM = currentUser?.role === 'dm' || currentUser?.role === 'admin';

    let html = `<div class="d-flex justify-content-between align-items-center flex-wrap gap-2 mb-3">
      <h1 class="h4 mb-0"><i class="fa-solid fa-store me-2"></i>Shops & Trading</h1>
      <div class="d-flex gap-2">
        ${isDM ? `<button class="btn btn-gold btn-sm" onclick="showNewShop()"><i class="fa-solid fa-plus me-1"></i>New Shop</button>` : ''}
        <button class="btn btn-outline-primary btn-sm" onclick="showTransactions()"><i class="fa-solid fa-receipt me-1"></i>Transactions</button>
      </div>
    </div>`;

    html += `<div class="mb-3"><input class="form-control" id="shopSearch" placeholder="Search shops by name, location..." oninput="filterShops()"></div>`;
    html += `<div class="row g-3" id="shopsGrid">`;

    if (!shops.length) {
      html += `<div class="col-12"><div class="empty-state"><i class="fa-solid fa-store fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Shops Found</p><p class="small text-muted">${isDM ? 'Create your first shop using the button above.' : 'Ask your DM to add shops to the world!'}</p></div></div>`;
    } else {
      for (const s of shops) {
        html += renderShopCard(s, isDM);
      }
    }

    html += `</div>`;

    el.innerHTML = html;
  } catch (e: any) {
    el.innerHTML = `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load shops: ${esc(e.message)}</p></div>`;
  }
});

function renderShopCard(s: any, isDM: boolean): string {
  const locName = s.location_name || 'Anywhere';
  return `<div class="col-md-6 col-lg-4">
    <div class="character-card" onclick="showShopDetail(${s.id})">
      <div class="d-flex justify-content-between align-items-start">
        <div class="char-name mb-1">${esc(s.name)}</div>
        ${isDM ? `<div class="d-flex gap-1" onclick="event.stopPropagation()">
          <button class="btn btn-sm btn-outline-primary py-0 px-1" onclick="showEditShop(${s.id})" title="Edit"><i class="fa-solid fa-pen" style="font-size:0.65rem"></i></button>
          <button class="btn btn-sm btn-outline-danger py-0 px-1" onclick="deleteShop(${s.id})" title="Delete"><i class="fa-solid fa-trash" style="font-size:0.65rem"></i></button>
        </div>` : ''}
      </div>
      ${s.description ? `<div class="char-detail small">${esc(s.description)}</div>` : ''}
      <div class="mt-1 d-flex gap-2 flex-wrap">
        <span class="badge badge-gold" style="font-size:0.6rem"><i class="fa-solid fa-location-dot me-1"></i>${esc(locName)}</span>
        <span class="badge badge-blood" style="font-size:0.6rem">Buy +${s.markup_percent}%</span>
        <span class="badge badge-muted" style="font-size:0.6rem">Sell ${s.markup_buy_percent}%</span>
      </div>
    </div>
  </div>`;
}

expose('filterShops', function () {
  const q = (document.getElementById('shopSearch') as HTMLInputElement)?.value?.toLowerCase() || '';
  document.querySelectorAll('#shopsGrid .character-card').forEach(card => {
    const parent = card.closest('.col-md-6, .col-lg-4') as HTMLElement;
    if (parent) parent.style.display = !q || card.textContent?.toLowerCase().includes(q) ? '' : 'none';
  });
});

// ─── Shop CRUD (DM) ───

expose('showNewShop', function () {
  const campaignOpts = allCampaigns.map((c: any) =>
    `<option value="${c.id}">${esc(c.name)}${c.party_name ? ' (' + esc(c.party_name) + ')' : ''}</option>`
  ).join('');

  showModal('New Shop', `
    <div class="mb-3"><label class="form-label">Shop Name</label><input class="form-control" id="shopName" placeholder="e.g. The Rusty Sword"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="shopDesc" rows="2" placeholder="What does this shop sell?"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Buy Markup %</label><input class="form-control" id="shopMarkup" type="number" value="0" step="1"><small class="text-muted">Added to item prices</small></div>
      <div class="col-6"><label class="form-label">Sell % of Value</label><input class="form-control" id="shopSellMarkup" type="number" value="50" step="1"><small class="text-muted">% of price paid to players</small></div>
    </div>
    <div class="mb-3"><label class="form-label">Campaign</label>
      <select class="form-select" id="shopCampaign">
        ${campaignOpts || '<option value="">No campaigns available</option>'}
      </select></div>
    <div class="mb-3"><label class="form-label">Location (optional)</label>
      <select class="form-select" id="shopLocation">
        <option value="">No specific location</option>
        ${allLocations.map((l: any) => `<option value="${l.id}">${esc(l.name)} (${esc(l.type)})</option>`).join('')}
      </select></div>
    <button class="btn btn-primary w-100" onclick="saveShop()"><i class="fa-solid fa-plus me-1"></i>Create Shop</button>
  `);
});

expose('saveShop', async function (shopId?: number) {
  const name = (document.getElementById('shopName') as HTMLInputElement).value;
  if (!name) { toast('Shop name is required', true); return; }

  const data: any = {
    name,
    description: (document.getElementById('shopDesc') as HTMLTextAreaElement).value,
    markup_percent: +(document.getElementById('shopMarkup') as HTMLInputElement).value || 0,
    markup_buy_percent: +(document.getElementById('shopSellMarkup') as HTMLInputElement).value || 50,
  };
  const locVal = (document.getElementById('shopLocation') as HTMLSelectElement).value;
  if (locVal) data.location_id = parseInt(locVal);

  try {
    if (shopId) {
      await api('PUT', `/api/shops/${shopId}`, data);
      toast('Shop updated');
    } else {
      const campaignId = (document.getElementById('shopCampaign') as HTMLSelectElement).value;
      if (!campaignId) { toast('Campaign is required', true); return; }
      await api('POST', `/api/campaigns/${campaignId}/shops`, data);
      toast('Shop created');
    }
    hideModal();
    (window as any).showShops();
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('showEditShop', async function (shopId: number) {
  const shop = selectedShopShops.find((s: any) => s.id === shopId);
  if (!shop) { toast('Shop not found', true); return; }

  showModal(`Edit: ${esc(shop.name)}`, `
    <div class="mb-3"><label class="form-label">Shop Name</label><input class="form-control" id="shopName" value="${esc(shop.name)}"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="shopDesc" rows="2">${esc(shop.description)}</textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Buy Markup %</label><input class="form-control" id="shopMarkup" type="number" value="${shop.markup_percent}" step="1"></div>
      <div class="col-6"><label class="form-label">Sell % of Value</label><input class="form-control" id="shopSellMarkup" type="number" value="${shop.markup_buy_percent}" step="1"></div>
    </div>
    <div class="mb-3"><label class="form-label">Location (optional)</label>
      <select class="form-select" id="shopLocation">
        <option value="">No specific location</option>
        ${allLocations.map((l: any) => `<option value="${l.id}" ${shop.location_id === l.id ? 'selected' : ''}>${esc(l.name)} (${esc(l.type)})</option>`).join('')}
      </select></div>
    <button class="btn btn-primary w-100" onclick="saveShop(${shopId})"><i class="fa-solid fa-save me-1"></i>Save Changes</button>
  `);
});

expose('deleteShop', async function (shopId: number) {
  if (!confirm('Delete this shop? All items will be removed.')) return;
  try {
    await api('DELETE', `/api/shops/${shopId}`);
    toast('Shop deleted');
    (window as any).showShops();
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Shop Items Detail ───

expose('showShopDetail', async function (shopId: number) {
  try {
    const [items, shops] = await Promise.all([
      api('GET', `/api/shops/${shopId}/items`),
      selectedShopShops.length ? Promise.resolve(selectedShopShops) : api('GET', '/api/shops'),
    ]);
    selectedShopShops = shops;
    const shop = shops.find((s: any) => s.id === shopId);
    if (!shop) { toast('Shop not found', true); return; }

    const isDM = currentUser?.role === 'dm' || currentUser?.role === 'admin';
    const hasChar = !!currentChar;

    showModal(esc(shop.name), `
      <div class="shop-detail-modal">
        ${shop.description ? `<p class="text-muted small">${esc(shop.description)}</p>` : ''}
        <div class="d-flex gap-2 mb-3 flex-wrap">
          <span class="badge badge-gold"><i class="fa-solid fa-location-dot me-1"></i>${esc(shop.location_name || 'Anywhere')}</span>
          ${items.length > 0 ? `<span class="badge badge-muted"><i class="fa-solid fa-box me-1"></i>${items.length} items</span>` : ''}
          <span class="badge badge-blood">Buy: +${shop.markup_percent}%</span>
          <span class="badge badge-blood">Sell: ${shop.markup_buy_percent}%</span>
        </div>

        <div class="d-flex justify-content-between align-items-center mb-2">
          <h6 class="mb-0">Stock</h6>
          ${isDM ? `<button class="btn btn-sm btn-gold" onclick="showAddShopItem(${shopId})"><i class="fa-solid fa-plus me-1"></i>Add Item</button>` : ''}
        </div>

        <div id="shopItemsList">
          ${items.length === 0 ? '<div class="text-muted small text-center py-3 fst-italic">No items in stock.</div>' :
            items.map((i: any) => {
              const effPrice = Math.round(i.price * (1 + shop.markup_percent / 100) * 100) / 100;
              return `<div class="inv-item">
                <div>
                  <span class="fw-bold">${esc(i.name)}</span>
                  ${i.category ? `<span class="badge badge-muted ms-1" style="font-size:0.6rem">${esc(i.category)}</span>` : ''}
                  ${i.quantity > 0
                    ? `<span class="badge badge-gold ms-1" style="font-size:0.6rem">x${i.quantity}</span>`
                    : `<span class="badge bg-danger ms-1" style="font-size:0.6rem">Out of stock</span>`}
                  ${i.description ? `<div class="small text-muted">${esc(i.description)}</div>` : ''}
                </div>
                <div class="d-flex align-items-center gap-1 flex-nowrap">
                  <span class="fw-bold text-nowrap" style="color:var(--gold);min-width:50px;text-align:right">${effPrice} gp</span>
                  ${i.quantity > 0 && hasChar ? `<button class="btn btn-sm btn-gold py-0 px-1" onclick="buyItem(${shopId}, ${i.id})" style="font-size:0.65rem">Buy</button>` : ''}
                  ${isDM ? `
                    <button class="btn btn-sm btn-outline-primary py-0 px-1" onclick="editShopItem(${i.id})" title="Edit" style="font-size:0.65rem"><i class="fa-solid fa-pen"></i></button>
                    <button class="btn btn-sm btn-outline-danger py-0 px-1" onclick="deleteShopItem(${i.id})" title="Remove" style="font-size:0.65rem"><i class="fa-solid fa-trash"></i></button>
                  ` : ''}
                </div>
              </div>`;
            }).join('')}
        </div>

        ${hasChar ? `<hr><button class="btn btn-outline-primary w-100" onclick="hideModal();sellItem(${shopId})"><i class="fa-solid fa-sack-dollar me-1"></i>Sell Items to Shop</button>` : ''}
        ${!hasChar ? '<p class="small text-muted mt-2 text-center fst-italic">Open a character sheet to buy or sell items.</p>' : ''}
      </div>
    `);
    // Dynamically add modal-lg class for wider shop view
    const dialog = document.querySelector('#genericModal .modal-dialog') as HTMLElement;
    if (dialog) dialog.classList.add('modal-lg');
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Shop Item CRUD (DM) ───

expose('showAddShopItem', async function (shopId: number) {
  await ensureCompendiumEquipment();
  showModal('Add Shop Item', `
    <div class="mb-3"><label class="form-label">Item Name</label>
      <input class="form-control" id="shopItemName" list="compEquipSuggestions" placeholder="Start typing for compendium suggestions...">
      <datalist id="compEquipSuggestions">
        ${compendiumEquipment.map((i: any) => `<option value="${esc(i.name)}">`).join('')}
      </datalist>
    </div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="shopItemDesc" rows="2"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-4"><label class="form-label">Price (gp)</label><input class="form-control" id="shopItemPrice" type="number" value="1" step="0.01"></div>
      <div class="col-4"><label class="form-label">Quantity</label><input class="form-control" id="shopItemQty" type="number" value="1" min="0"></div>
      <div class="col-4"><label class="form-label">Category</label>
        <select class="form-select" id="shopItemCat">
          <option value="">General</option><option value="weapon">Weapon</option><option value="armor">Armor</option>
          <option value="potion">Potion</option><option value="scroll">Scroll</option><option value="gear">Gear</option>
          <option value="ammunition">Ammunition</option><option value="food">Food/Drink</option><option value="other">Other</option>
        </select></div>
    </div>
    <div class="d-flex gap-2">
      <button class="btn btn-outline-gold btn-sm flex-grow-1" onclick="pickCompItem('shopItemName')"><i class="fa-solid fa-book me-1"></i>From Compendium</button>
    </div>
    <div class="mt-3">
      <button class="btn btn-primary w-100" onclick="saveShopItem(${shopId})"><i class="fa-solid fa-plus me-1"></i>Add Item</button>
    </div>
  `);
});

expose('pickCompItem', function (inputId: string) {
  showModal('Pick from Compendium', `
    <div class="mb-3"><input class="form-control" id="compPickSearch" placeholder="Search equipment..." oninput="filterCompPick()"></div>
    <div id="compPickList" style="max-height:300px;overflow-y:auto">
      ${compendiumEquipment.map((i: any) => `
        <div class="inv-item comp-pick-item" data-name="${esc(i.name)}" data-cost="${i.cost || i.price || 1}" data-cat="${i.category || ''}"
             onclick="selectCompPick('${inputId}')" style="cursor:pointer">
          <div><span class="fw-bold small">${esc(i.name)}</span>
            <span class="text-muted small">${i.cost ? i.cost + ' gp' : ''} ${i.category ? '· ' + esc(i.category) : ''}</span></div>
        </div>
      `).join('')}
    </div>
  `);
});

expose('filterCompPick', function () {
  const q = (document.getElementById('compPickSearch') as HTMLInputElement)?.value?.toLowerCase() || '';
  document.querySelectorAll('.comp-pick-item').forEach(el => {
    const name = el.getAttribute('data-name') || '';
    (el as HTMLElement).style.display = !q || name.toLowerCase().includes(q) ? '' : 'none';
  });
});

expose('selectCompPick', function (inputId: string) {
  const selected = document.querySelector('.comp-pick-item[style*="display: none"]') ? null : null;
  const el = document.querySelector('.comp-pick-item:not([style*="display: none"])') as HTMLElement;
  // Clicked the one visible item, fill the input
  const active = document.activeElement as HTMLElement;
  if (active && active.classList.contains('comp-pick-item')) {
    const name = active.getAttribute('data-name') || '';
    const cost = active.getAttribute('data-cost') || '1';
    const cat = active.getAttribute('data-cat') || '';
    const input = document.getElementById(inputId) as HTMLInputElement;
    if (input) input.value = name;
    const priceInput = document.getElementById('shopItemPrice') as HTMLInputElement;
    if (priceInput && priceInput.value === '1') priceInput.value = cost;
    const catSelect = document.getElementById('shopItemCat') as HTMLSelectElement;
    if (catSelect && cat) { const opt = Array.from(catSelect.options).find(o => o.value === cat); if (opt) opt.selected = true; }
    hideModal();
  }
});

// This is handled by onclick directly on the comp-pick-item elements
document.addEventListener('click', function compPickHandler(e) {
  const target = e.target as HTMLElement;
  const item = target.closest('.comp-pick-item') as HTMLElement;
  if (item) {
    const inputId = item.closest('[onclick*="selectCompPick"]')?.getAttribute('onclick')?.match(/'([^']+)'/)?.pop() || 'shopItemName';
    const name = item.getAttribute('data-name') || '';
    const cost = item.getAttribute('data-cost') || '1';
    const cat = item.getAttribute('data-cat') || '';
    const input = document.getElementById(inputId) as HTMLInputElement;
    if (input) input.value = name;
    const priceInput = document.getElementById('shopItemPrice') as HTMLInputElement;
    if (priceInput && priceInput.value === '1') priceInput.value = cost;
    const catSelect = document.getElementById('shopItemCat') as HTMLSelectElement;
    if (catSelect && cat) { const opt = Array.from(catSelect.options).find(o => o.value === cat); if (opt) opt.selected = true; }
    hideModal();
  }
});

expose('saveShopItem', async function (shopId: number, itemId?: number) {
  const name = (document.getElementById('shopItemName') as HTMLInputElement).value;
  if (!name) { toast('Item name is required', true); return; }

  const data: any = {
    name,
    description: (document.getElementById('shopItemDesc') as HTMLTextAreaElement).value,
    price: +(document.getElementById('shopItemPrice') as HTMLInputElement).value || 0,
    quantity: +(document.getElementById('shopItemQty') as HTMLInputElement).value || 0,
    category: (document.getElementById('shopItemCat') as HTMLSelectElement).value,
    properties: '{}',
  };

  try {
    if (itemId) {
      await api('PUT', `/api/shop-items/${itemId}`, data);
      toast('Item updated');
    } else {
      await api('POST', `/api/shops/${shopId}/items`, data);
      toast('Item added to shop');
    }
    hideModal();
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('editShopItem', async function (itemId: number) {
  // Find the item from any shop detail modal
  const itemsEl = document.getElementById('shopItemsList');
  if (!itemsEl) return;
  const itemRow = itemsEl.querySelector(`[onclick*="deleteShopItem(${itemId})"]`)?.closest('.inv-item');
  // We need to fetch the item data
  try {
    // Scan through shops to find this item
    for (const shop of selectedShopShops) {
      const items = await api('GET', `/api/shops/${shop.id}/items`);
      const item = items.find((i: any) => i.id === itemId);
      if (item) {
        showModal('Edit Shop Item', `
          <div class="mb-3"><label class="form-label">Item Name</label><input class="form-control" id="shopItemName" value="${esc(item.name)}"></div>
          <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="shopItemDesc" rows="2">${esc(item.description)}</textarea></div>
          <div class="row g-3 mb-3">
            <div class="col-4"><label class="form-label">Price (gp)</label><input class="form-control" id="shopItemPrice" type="number" value="${item.price}" step="0.01"></div>
            <div class="col-4"><label class="form-label">Quantity</label><input class="form-control" id="shopItemQty" type="number" value="${item.quantity}" min="0"></div>
            <div class="col-4"><label class="form-label">Category</label>
              <select class="form-select" id="shopItemCat">
                <option value="">General</option>
                ${['weapon','armor','potion','scroll','gear','ammunition','food','other'].map(c =>
                  `<option value="${c}" ${item.category === c ? 'selected' : ''}>${capitalize(c)}</option>`
                ).join('')}
              </select></div>
          </div>
          <button class="btn btn-primary w-100" onclick="saveShopItem(${shop.id}, ${itemId})"><i class="fa-solid fa-save me-1"></i>Save Changes</button>
        `);
        return;
      }
    }
    toast('Item not found', true);
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('deleteShopItem', async function (itemId: number) {
  if (!confirm('Remove this item from the shop?')) return;
  try {
    await api('DELETE', `/api/shop-items/${itemId}`);
    toast('Item removed from shop');
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Buy / Sell ───

expose('buyItem', async function (shopId: number, itemId: number) {
  if (!currentChar) { toast('Select a character first', true); return; }

  showModal('Buy Item', `
    <div class="text-center mb-3">
      <p class="mb-1">How many would you like to buy?</p>
      <div class="input-group" style="max-width:200px;margin:0 auto">
        <button class="btn btn-outline-secondary" onclick="qtyAdjust(-1)">−</button>
        <input class="form-control text-center" id="buyQty" type="number" value="1" min="1" max="99">
        <button class="btn btn-outline-secondary" onclick="qtyAdjust(1)">+</button>
      </div>
    </div>
    <button class="btn btn-gold w-100" onclick="confirmBuy(${shopId}, ${itemId})"><i class="fa-solid fa-cart-shopping me-1"></i>Purchase</button>
  `);
});

expose('qtyAdjust', function (delta: number) {
  const input = document.getElementById('buyQty') as HTMLInputElement;
  if (input) {
    const val = Math.max(1, Math.min(99, (parseInt(input.value) || 1) + delta));
    input.value = String(val);
  }
});

expose('confirmBuy', async function (shopId: number, itemId: number) {
  const qty = parseInt((document.getElementById('buyQty') as HTMLInputElement).value) || 1;
  try {
    const result = await api('POST', `/api/shops/${shopId}/buy`, {
      item_id: itemId,
      quantity: qty,
      character_id: currentChar!.id,
    });
    hideModal();
    toast(`Purchased ${esc(result.item_name || 'item')} x${qty} for ${result.total_cost} gp`);
    // Refresh shop items to update stock
    if (typeof (window as any).showShopDetail === 'function') {
      // Re-open shop detail with updated data
      (window as any).showShopDetail(shopId);
    }
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('sellItem', async function (shopId: number) {
  if (!currentChar) { toast('Select a character first', true); return; }

  try {
    const char = await api('GET', `/api/characters/${currentChar.id}`);
    const inv = char.inventory || [];
    if (!inv.length) {
      toast('No items to sell', true);
      return;
    }

    // Get shop details for sell markup
    const shops = selectedShopShops.length ? selectedShopShops : await api('GET', '/api/shops');
    const shop = shops.find((s: any) => s.id === shopId);
    const sellPct = shop ? shop.markup_buy_percent : 50;

    showModal('Sell Items', `
      <p class="text-muted small mb-2">Shop pays ${sellPct}% of base price. Select items to sell.</p>
      <div id="sellItemsList" style="max-height:350px;overflow-y:auto">
        ${inv.map((i: any) => `
          <div class="inv-item sell-inv-item" data-id="${i.id}" data-price="${i.cost || i.price || 0}" style="cursor:pointer" onclick="toggleSellItem(this)">
            <div>
              <span class="fw-bold small">${esc(i.name)}</span>
              ${i.quantity > 1 ? `<span class="badge badge-muted ms-1">x${i.quantity}</span>` : ''}
              <div class="small text-muted">Est. ${Math.round((i.cost || i.price || 0) * sellPct / 100)} gp each</div>
            </div>
            <div class="d-flex align-items-center gap-2">
              <input type="number" class="form-control form-control-sm" style="width:60px" value="1" min="1" max="${i.quantity}" onclick="event.stopPropagation()">
              <span class="sell-check"><i class="fa-regular fa-square"></i></span>
            </div>
          </div>
        `).join('')}
      </div>
      <button class="btn btn-gold w-100 mt-3" onclick="confirmSell(${shopId})"><i class="fa-solid fa-sack-dollar me-1"></i>Sell Selected Items</button>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('toggleSellItem', function (el: HTMLElement) {
  el.classList.toggle('selected');
  const check = el.querySelector('.sell-check i')!;
  if (el.classList.contains('selected')) {
    check.className = 'fa-solid fa-square-check text-success';
  } else {
    check.className = 'fa-regular fa-square';
  }
});

expose('confirmSell', async function (shopId: number) {
  const selected: Array<{ inventory_item_id: number; quantity: number }> = [];
  document.querySelectorAll('.sell-inv-item.selected').forEach((el: any) => {
    const id = parseInt(el.getAttribute('data-id'));
    const qtyInput = el.querySelector('input[type="number"]') as HTMLInputElement;
    const qty = parseInt(qtyInput?.value || '1');
    selected.push({ inventory_item_id: id, quantity: qty });
  });

  if (!selected.length) { toast('Select items to sell', true); return; }

  try {
    // Sell each item sequentially (the API sells one item at a time)
    let totalGold = 0;
    for (const s of selected) {
      const result = await api('POST', `/api/shops/${shopId}/sell`, s);
      totalGold += result.gold_earned || 0;
    }
    hideModal();
    toast(`Sold ${selected.length} item(s) for ${totalGold} gp`);
  } catch (e: any) {
    toast(e.message, true);
  }
});

// ─── Transactions ───

expose('showTransactions', async function () {
  try {
    const txns = await api('GET', '/api/shop-transactions');
    showModal('Transaction History', `
      <div style="max-height:400px;overflow-y:auto">
        ${txns.length === 0 ? '<div class="text-muted text-center py-4 fst-italic">No transactions yet.</div>' :
          txns.map((t: any) => {
            const date = t.created_at ? new Date(t.created_at).toLocaleDateString() : '';
            const typeIcon = t.transaction_type === 'buy' ? 'fa-cart-shopping text-success' : 'fa-sack-dollar text-warning';
            return `<div class="inv-item">
              <div>
                <span class="fw-bold small">${esc(t.item_name || 'Unknown')}</span>
                <span class="badge ${t.transaction_type === 'buy' ? 'badge-blood' : 'badge-gold'} ms-1" style="font-size:0.6rem">${esc(t.transaction_type)}</span>
                <div class="small text-muted">${esc(t.shop_name || '')} ${t.character_name ? '· ' + esc(t.character_name) : ''}</div>
              </div>
              <div class="text-nowrap">
                <span class="fw-bold" style="color:var(--gold)">${t.total_cost || t.gold_earned || 0} gp</span>
                <span class="text-muted small ms-2">x${t.quantity || 1}</span>
                <div class="text-muted small">${date}</div>
              </div>
            </div>`;
          }).join('')}
      </div>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
});

expose('showOneShots', function () {
  showView('oneshot');
  const el = document.getElementById('oneshotSection')!;
  el.setAttribute('hx-get', '/htmx/oneshot-adventures');
  el.setAttribute('hx-trigger', 'load');
  el.setAttribute('hx-swap', 'innerHTML');
  el.innerHTML = '<div class="ornament">✧ Loading one-shot adventures... ✧</div>';
  htmx.process(el);
});

expose('showCampaignOverview', function (campaignId: number) {
  showView('campaignOverview');
  const el = document.getElementById('campaignOverviewSection')!;
  el.setAttribute('hx-get', `/htmx/campaigns/${campaignId}/overview`);
  el.setAttribute('hx-trigger', 'load');
  el.setAttribute('hx-swap', 'innerHTML');
  el.innerHTML = '<div class="ornament">✧ Loading campaign overview... ✧</div>';
  htmx.process(el);
});

expose('loadFactionReputations', async function () {
  const charId = (document.getElementById('factionCharSel') as HTMLSelectElement).value;
  const area = document.getElementById('factionRepArea')!;
  const list = document.getElementById('factionRepList')!;
  if (!charId) { area.style.display = 'none'; return; }
  area.style.display = 'block';
  try {
    const reps = await api('GET', `/api/faction-reputation?character_id=${charId}`);
    const factions = await api('GET', '/api/factions');
    list.innerHTML = reps.length ? reps.map((r: any) => {
      const pct = ((r.standing + 100) / 200) * 100;
      const color = r.standing >= 50 ? '#2d6a2d' : r.standing >= 0 ? '#b8963e' : r.standing >= -50 ? '#8b4513' : '#8b0000';
      return `<div class="inv-item">
        <div>
          <span class="fw-bold">${esc(r.faction_name)}</span>
          <span class="badge badge-muted ms-1">${esc(r.faction_type)}</span>
          ${r.rank ? `<span class="badge badge-gold ms-1">${esc(r.rank)}</span>` : ''}
        </div>
        <div class="d-flex align-items-center gap-2">
          <div class="hp-bar" style="width:100px;height:8px;background:var(--parchment-dark)">
            <div class="hp-bar-fill" style="width:${pct}%;height:100%;background:${color}"></div>
          </div>
          <span class="fw-bold" style="color:${color}">${r.standing >= 0 ? '+' : ''}${r.standing}</span>
          <button class="btn btn-sm btn-outline-primary" onclick="editReputation(${r.character_id}, ${r.faction_id}, ${r.standing}, '${esc(r.rank)}', '${esc(r.notes)}')"><i class="fa-solid fa-pen"></i></button>
        </div>
      </div>`;
    }).join('') : '<p class="text-muted small">No reputation tracked for this character. Click a faction to set reputation.</p>';
  } catch {}
});

expose('editReputation', function (charId: number, factionId: number, standing: number, rank: string, notes: string) {
  showModal('Set Reputation', `
    <div class="mb-3"><label class="form-label">Standing (-100 to 100)</label>
      <input class="form-control" id="repStanding" type="number" value="${standing}" min="-100" max="100"></div>
    <div class="mb-3"><label class="form-label">Rank / Title</label><input class="form-control" id="repRank" value="${esc(rank)}"></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="repNotes" rows="2">${esc(notes)}</textarea></div>
    <button class="btn btn-primary w-100" onclick="saveReputation(${charId}, ${factionId})"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
});

expose('saveReputation', async function (charId: number, factionId: number) {
  await api('POST', '/api/faction-reputation', {
    character_id: charId, faction_id: factionId,
    standing: +(document.getElementById('repStanding') as HTMLInputElement).value || 0,
    rank: (document.getElementById('repRank') as HTMLInputElement).value,
    notes: (document.getElementById('repNotes') as HTMLTextAreaElement).value,
  });
  hideModal();
  (window as any).loadFactionReputations();
  toast('Reputation updated');
});

// ─── Crafting ───

async function renderCrafting() {
  const el = document.getElementById('craftingSection')!;
  if (!currentChar) return;
  try {
    const [recipes, projects] = await Promise.all([
      api('GET', '/api/crafting/recipes'),
      api('GET', `/api/characters/${currentChar.id}/crafting`),
    ]);

    let html = `<div class="d-flex justify-content-between align-items-center"><h5>Crafting</h5>
      <button class="btn btn-primary btn-sm" onclick="showStartCrafting()"><i class="fa-solid fa-hammer me-1"></i>Start Crafting</button>
    </div>`;

    // Active projects
    const active = projects.filter((p: any) => p.status === 'in-progress');
    if (active.length > 0) {
      html += `<h6 class="mt-3 text-muted">In Progress</h6>`;
      for (const p of active) {
        const pct = p.total_hours_required > 0 ? Math.min(100, Math.round((p.progress_hours / p.total_hours_required) * 100)) : 0;
        html += `<div class="card mb-2">
          <div class="card-body py-2 px-3">
            <div class="d-flex justify-content-between align-items-start">
              <div><span class="fw-bold">${esc(p.name)}</span>
                <span class="badge badge-gold ms-1">DC ${p.dc}</span>
                <span class="badge badge-muted ms-1">${p.progress_hours}/${p.total_hours_required}h</span>
              </div>
              <div class="d-flex gap-1">
                <button class="btn btn-sm btn-outline-primary" onclick="advanceCrafting(${p.id})" title="Advance 1 hour"><i class="fa-solid fa-forward"></i></button>
                <button class="btn btn-sm btn-outline-success" onclick="completeCrafting(${p.id})" title="Complete"><i class="fa-solid fa-check"></i></button>
                <button class="btn btn-sm btn-outline-danger" onclick="abandonCrafting(${p.id})" title="Abandon"><i class="fa-solid fa-xmark"></i></button>
              </div>
            </div>
            <div class="hp-bar mt-1" style="height:4px"><div class="hp-bar-fill" style="width:${pct}%;height:100%;background:var(--gold)"></div></div>
          </div>
        </div>`;
      }
    }

    // Completed projects
    const done = projects.filter((p: any) => p.status === 'complete');
    if (done.length > 0) {
      html += `<h6 class="mt-3 text-muted">Completed</h6>`;
      for (const p of done) {
        html += `<div class="card mb-1"><div class="card-body py-1 px-3 small text-muted">
          <i class="fa-solid fa-check-circle text-success me-1"></i>${esc(p.name)}
        </div></div>`;
      }
    }

    // Recipes
    html += `<h6 class="mt-3 text-muted">Known Recipes (${recipes.length})</h6>
      <div class="row g-2">`;
    for (const r of recipes) {
      html += `<div class="col-md-6"><div class="card">
        <div class="card-body py-2 px-3">
          <div class="d-flex justify-content-between">
            <span class="fw-bold small">${esc(r.name)}</span>
            <span class="badge ${r.category === 'potion' ? 'badge-blood' : r.category === 'scroll' ? 'badge-gold' : 'badge-muted'}" style="font-size:0.6rem">${r.category}</span>
          </div>
          <div class="small text-muted">${esc(r.description)}</div>
          <div class="small mt-1"><span class="text-muted">DC ${r.difficulty_dc}</span> · <span class="text-muted">${r.crafting_time_hours}h</span></div>
          <div class="mt-1"><button class="btn btn-sm btn-outline-gold py-0 px-1" style="font-size:0.65rem" onclick="startRecipe(${r.id})">Craft</button></div>
        </div>
      </div></div>`;
    }
    html += `</div>`;

    el.innerHTML = html;
  } catch (e: any) {
    el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(e.message)}</p></div>`;
  }
}
expose('renderCrafting', renderCrafting);

expose('startRecipe', async function (recipeId: number) {
  try {
    const recipes = await api('GET', '/api/crafting/recipes');
    const recipe = recipes.find((r: any) => r.id === recipeId);
    if (!recipe) return;
    const materials = JSON.parse(recipe.required_materials || '[]');
    const tools = JSON.parse(recipe.required_tools || '[]');
    showModal('Start Crafting', `
      <p class="mb-2"><strong>${esc(recipe.name)}</strong></p>
      <p class="small text-muted">${esc(recipe.description)}</p>
      <div class="mb-2"><span class="text-muted small">DC:</span> <strong>${recipe.difficulty_dc}</strong> &middot;
        <span class="text-muted small">Time:</span> <strong>${recipe.crafting_time_hours}h</strong></div>
      ${tools.length ? `<div class="mb-2"><span class="text-muted small">Tools:</span> ${tools.map((t: string) => `<span class="badge badge-muted me-1" style="font-size:0.6rem">${esc(t)}</span>`).join('')}</div>` : ''}
      ${materials.length ? `<div class="mb-2"><span class="text-muted small">Materials:</span><ul class="small mb-0">${materials.map((m: any) => `<li>${esc(m.name)} x${m.quantity}${m.consumed ? ' (consumed)' : ''}</li>`).join('')}</ul></div>` : ''}
      <div class="mb-2"><span class="text-muted small">Result:</span> <strong>${esc(recipe.result_item_name)}</strong> x${recipe.result_quantity}</div>
      <button class="btn btn-gold w-100 mt-2" onclick="confirmStartRecipe(${recipe.id},'${esc(recipe.name)}',${recipe.crafting_time_hours},${recipe.difficulty_dc})"><i class="fa-solid fa-hammer me-1"></i>Begin Crafting</button>
    `);
  } catch (e: any) { toast(e.message, true); }
});

expose('confirmStartRecipe', async function (recipeId: number, name: string, hours: number, dc: number) {
  try {
    await api('POST', `/api/characters/${currentChar.id}/crafting`, {
      recipe_id: recipeId,
      name: name,
      total_hours_required: hours,
      dc: dc,
      materials_allocated: '[]',
      notes: '',
    });
    hideModal();
    renderCrafting();
    toast('Crafting started!');
  } catch (e: any) { toast(e.message, true); }
});

expose('advanceCrafting', async function (id: number) {
  try {
    await api('PUT', `/api/crafting/${id}`, { progress_hours: 1 });
    renderCrafting();
    toast('Crafting advanced by 1 hour');
  } catch (e: any) { toast(e.message, true); }
});

expose('completeCrafting', async function (id: number) {
  try {
    await api('PUT', `/api/crafting/${id}`, { status: 'complete' });
    renderCrafting();
    toast('Crafting completed! Item added to inventory.');
  } catch (e: any) { toast(e.message, true); }
});

expose('abandonCrafting', async function (id: number) {
  if (!confirm('Abandon this project?')) return;
  try {
    await api('PUT', `/api/crafting/${id}`, { status: 'abandoned' });
    renderCrafting();
    toast('Project abandoned');
  } catch (e: any) { toast(e.message, true); }
});

expose('showStartCrafting', function () {
  showModal('Start Crafting', `
    <p class="small text-muted">Browse recipes from the Crafting tab, or create a custom project.</p>
    <div class="mb-3"><label class="form-label">Project Name</label><input class="form-control" id="custCraftName" placeholder="e.g. Brewing a custom potion"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Est. Hours</label><input class="form-control" id="custCraftHours" type="number" value="1" min="0.5" step="0.5"></div>
      <div class="col-6"><label class="form-label">DC</label><input class="form-control" id="custCraftDC" type="number" value="10"></div>
    </div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="custCraftNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="confirmCustomCraft()"><i class="fa-solid fa-hammer me-1"></i>Start</button>
  `);
});

expose('confirmCustomCraft', async function () {
  const name = (document.getElementById('custCraftName') as HTMLInputElement).value;
  if (!name) { toast('Enter a project name', true); return; }
  try {
    await api('POST', `/api/characters/${currentChar.id}/crafting`, {
      name: name,
      total_hours_required: +(document.getElementById('custCraftHours') as HTMLInputElement).value || 1,
      dc: +(document.getElementById('custCraftDC') as HTMLInputElement).value || 10,
      materials_allocated: '[]',
      notes: (document.getElementById('custCraftNotes') as HTMLTextAreaElement).value,
    });
    hideModal();
    renderCrafting();
    toast('Custom crafting started!');
  } catch (e: any) { toast(e.message, true); }
});

// ─── HP Auto-Calc in details ───

expose('calcHP', async function () {
  if (!currentChar) return;
  try {
    const result = await api('POST', `/api/characters/${currentChar.id}/calc-hp`);
    setCurrentChar(await api('GET', `/api/characters/${currentChar.id}`));
    renderSheet();
    toast(`HP calculated: ${result.hp_max} HP`);
  } catch (e: any) { toast(e.message, true); }
});

// ─── Random Character Generator ───

expose('generateRandomChar', async function () {
  try {
    const rc = await api('GET', '/api/generate/character');
    showModal('Random Character', `
      <div class="text-center mb-3">
        <span class="fw-bold fs-5">${esc(rc.name)}</span>
      </div>
      <div class="row g-2 mb-3">
        <div class="col-6"><span class="text-muted">Race:</span> ${esc(rc.race)}</div>
        <div class="col-6"><span class="text-muted">Class:</span> ${esc(rc.class)}</div>
        <div class="col-6"><span class="text-muted">Level:</span> ${rc.level}</div>
        <div class="col-6"><span class="text-muted">Background:</span> ${esc(rc.background)}</div>
        <div class="col-6"><span class="text-muted">Alignment:</span> ${esc(rc.alignment)}</div>
        <div class="col-6"><span class="text-muted">Personality:</span> ${esc(rc.personality)}</div>
      </div>
      <div class="row g-2 mb-3">
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">STR</div><div class="stat-value">${rc.str}</div></div></div>
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">DEX</div><div class="stat-value">${rc.dex}</div></div></div>
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">CON</div><div class="stat-value">${rc.con}</div></div></div>
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">INT</div><div class="stat-value">${rc.int}</div></div></div>
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">WIS</div><div class="stat-value">${rc.wis}</div></div></div>
        <div class="col-4 text-center"><div class="combat-stat"><div class="stat-label">CHA</div><div class="stat-value">${rc.cha}</div></div></div>
      </div>
      <div class="mb-2"><span class="text-muted small">Quirk:</span> <span class="small">${esc(rc.quirk)}</span></div>
      <div><span class="text-muted small">Backstory Hook:</span> <span class="small fst-italic">${esc(rc.backstory_hook)}</span></div>
      <hr>
      <p class="small text-muted">Use this as inspiration for your next character!</p>
    `);
  } catch (e: any) { toast(e.message, true); }
});

// ─── Character Comparison ───

expose('showComparison', async function () {
  const sel = document.getElementById('charCompareSelect') as HTMLSelectElement;
  if (!sel) return;
  const selected = Array.from(sel.selectedOptions).map(o => o.value).filter(v => v);
  if (selected.length < 2) { toast('Select at least 2 characters', true); return; }
  try {
    const chars = await api('GET', `/api/characters/compare?ids=${selected.join(',')}`);
    showModal('Character Comparison', `
      <div class="table-responsive"><table class="table table-sm table-bordered">
        <thead><tr><th></th>${chars.map((c: any) => `<th class="text-center">${esc(c.name)}</th>`).join('')}</tr></thead>
        <tbody>
          ${[['Race','race'],['Class','class'],['Level','level'],['Background','background'],['Alignment','alignment'],
             ['HP','hp_current + "/" + hp_max'],['AC','ac'],['Speed','speed'],['Initiative','initiative'],
             ['STR','str'],['DEX','dex'],['CON','con'],['INT','int'],['WIS','wis'],['CHA','cha'],['XP','xp']].map(([label, field]) => `
            <tr><td class="fw-bold">${label}</td>
              ${chars.map((c: any) => {
                if (field === 'hp_current + "/" + hp_max') {
                  return `<td class="text-center">${c.hp_current}/${c.hp_max}</td>`;
                }
                return `<td class="text-center">${c[field] ?? '-'}</td>`;
              }).join('')}
            </tr>`).join('')}
        </tbody>
      </table></div>
    `);
  } catch (e: any) { toast(e.message, true); }
});

// ─── Add character comparison to character list view ───

let compareMode = false;

expose('toggleCompareMode', function () {
  compareMode = !compareMode;
  const el = document.getElementById('charGrid')!;
  const btn = document.getElementById('compareBtn') as HTMLButtonElement;
  if (compareMode) {
    el.querySelectorAll('.character-card').forEach(card => card.classList.add('compare-selectable'));
    // Add compare bar
    let bar = document.getElementById('compareBar');
    if (!bar) {
      bar = document.createElement('div');
      bar.id = 'compareBar';
      bar.className = 'd-flex align-items-center gap-2 p-2 mb-2 border rounded';
      bar.style.background = 'var(--parchment)';
      bar.innerHTML = `
        <span class="small fw-bold me-2">Compare:</span>
        <select multiple class="form-select form-select-sm" id="charCompareSelect" style="height:2rem;width:auto;min-width:200px"></select>
        <button class="btn btn-sm btn-gold" onclick="showComparison()"><i class="fa-solid fa-arrow-right me-1"></i>Compare</button>
        <button class="btn btn-sm btn-outline-secondary" onclick="toggleCompareMode()">Done</button>`;
      document.getElementById('charactersView')?.insertBefore(bar, document.getElementById('charGrid'));
    }
    // Populate select
    const select = document.getElementById('charCompareSelect') as HTMLSelectElement;
    select.innerHTML = '';
    document.querySelectorAll('#charGrid .character-card').forEach(card => {
      const id = card.getAttribute('onclick')?.match(/\d+/)?.[0];
      const name = card.querySelector('.char-name')?.textContent;
      if (id && name) {
        select.innerHTML += `<option value="${id}">${esc(name)}</option>`;
      }
    });
    if (btn) btn.textContent = 'Cancel Compare';
  } else {
    el.querySelectorAll('.character-card').forEach(card => card.classList.remove('compare-selectable'));
    const bar = document.getElementById('compareBar');
    if (bar) bar.remove();
    if (btn) btn.textContent = 'Compare';
  }
});

// ─── Combat Tracker → extracted to ts/combat-tracker.ts ───
import './combat-tracker';

// ─── Shops ───

// ─── Wiki ───

expose('showWiki', async function (campaignId?: number) {
  showView('wiki');
  const el = document.getElementById('wikiContent')!;
  el.innerHTML = '<div class="ornament">✧ Loading wiki... ✧</div>';
  try {
    const campaigns = await api('GET', '/api/campaigns');
    if (!campaigns.length) {
      el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-book fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Campaigns</p><p class="small text-muted">Create a campaign to start building your campaign wiki.</p></div>';
      return;
    }
    const cid = campaignId || campaigns[0].id;
    const camp = campaigns.find((c: any) => c.id === cid);
    const pages = await api('GET', `/api/campaigns/${cid}/wiki`);

    const rootPages = pages.filter((p: any) => !p.parent_id);
    const childMap: Record<number, any[]> = {};
    pages.forEach((p: any) => {
      if (p.parent_id) {
        if (!childMap[p.parent_id]) childMap[p.parent_id] = [];
        childMap[p.parent_id].push(p);
      }
    });

    let sidebarHtml = '<div class="list-group list-group-flush">';
    for (const p of rootPages) {
      sidebarHtml += `<a href="#" class="list-group-item list-group-item-action py-1" onclick="loadWikiPage(${p.id});if(window.innerWidth<768){const o=document.getElementById('wikiOffcanvas');if(o){bootstrap.Offcanvas.getInstance(o)?.hide()}}">${esc(p.title)}</a>
        ${buildWikiChildren(p.id, childMap, 1)}`;
    }
    sidebarHtml += '</div>';

    if (!rootPages.length) {
      el.innerHTML = `
        <div class="d-flex justify-content-between align-items-center mb-3">
          <h4 class="mb-0"><i class="fa-solid fa-book me-2"></i>${esc(camp?.name || 'Wiki')}</h4>
          <div class="d-flex gap-1">
            <button class="btn btn-gold btn-sm" onclick="showAddWikiPage(${cid})"><i class="fa-solid fa-plus me-1"></i>New Page</button>
            <button class="btn btn-outline-gold btn-sm" onclick="showCampaignGraph(${cid})"><i class="fa-solid fa-project-diagram me-1"></i>Graph</button>
          </div>
        </div>
        <div class="empty-state"><i class="fa-solid fa-book-open fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">Empty Wiki</p><p class="small text-muted">Start building your campaign lore by creating pages.</p></div>`;
      return;
    }

    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center mb-3 flex-wrap gap-2">
        <h4 class="mb-0"><i class="fa-solid fa-book me-2"></i>${esc(camp?.name || 'Wiki')}</h4>
        <div class="d-flex gap-1">
          <button class="btn btn-gold btn-sm" onclick="showAddWikiPage(${cid})"><i class="fa-solid fa-plus me-1"></i>New Page</button>
          <button class="btn btn-outline-gold btn-sm" onclick="showCampaignGraph(${cid})"><i class="fa-solid fa-project-diagram me-1"></i>Graph</button>
        </div>
      </div>
      <div class="row g-0" style="min-height:500px">
        <div class="col-md-3 d-none d-md-block" style="overflow-y:auto;max-height:70vh;border-right:1px solid var(--border)">
          <div class="p-2"><small class="fw-bold text-muted">PAGES</small></div>
          ${sidebarHtml}
        </div>
        <div class="offcanvas offcanvas-start" id="wikiOffcanvas" tabindex="-1">
          <div class="offcanvas-header border-bottom">
            <h5 class="offcanvas-title">${esc(camp?.name || 'Wiki')} Pages</h5>
            <button type="button" class="btn-close" data-bs-dismiss="offcanvas"></button>
          </div>
          <div class="offcanvas-body p-0">
            <div class="p-2 border-bottom"><small class="fw-bold text-muted">PAGES</small></div>
            ${sidebarHtml}
          </div>
        </div>
        <div class="col-12 col-md-9" id="wikiPageContent">
          <div class="d-flex d-md-none gap-1 mb-2">
            <button class="btn btn-outline-primary btn-sm" onclick="toggleWikiSidebar()"><i class="fa-solid fa-bars me-1"></i> Pages</button>
          </div>
          <div class="p-3 text-center text-muted"><i class="fa-solid fa-book-open fa-2x mb-2 d-block"></i><p>Select a page from the sidebar</p></div>
        </div>
      </div>`;
  } catch (e: any) { el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(e.message)}</p></div>`; }
});

expose('toggleWikiSidebar', function () {
  const offcanvas = document.getElementById('wikiOffcanvas');
  if (offcanvas) bootstrap.Offcanvas.getOrCreateInstance(offcanvas).toggle();
});

function buildWikiChildren(parentId: number, childMap: Record<number, any[]>, depth: number): string {
  const children = childMap[parentId] || [];
  if (!children.length) return '';
  const pad = depth * 16;
  return children.map((c: any) =>
    `<a href="#" class="list-group-item list-group-item-action py-1 ps-${3 + depth}" style="padding-left:${pad + 16}px!important;font-size:0.9rem" onclick="loadWikiPage(${c.id});if(window.innerWidth<768){const o=document.getElementById('wikiOffcanvas');if(o){bootstrap.Offcanvas.getInstance(o)?.hide()}}">↳ ${esc(c.title)}</a>
    ${buildWikiChildren(c.id, childMap, depth + 1)}`
  ).join('');
}

expose('loadWikiPage', async function (pageId: number) {
  try {
    const page = await api('GET', `/api/wiki/${pageId}`);
    const el = document.getElementById('wikiPageContent')!;
    const renderContent = marked.parse(page.content);
    el.innerHTML = `
      <div class="p-3">
        <div class="d-flex justify-content-between align-items-start flex-wrap gap-2">
          <h3 class="mb-0">${esc(page.title)}</h3>
          <div class="d-flex gap-1">
            <button class="btn btn-sm btn-outline-primary" onclick="showEditWikiPage(${page.id},'${esc(page.title)}','${esc(page.content.replace(/'/g, "\\'"))}','${page.visibility}')"><i class="fa-solid fa-pen"></i></button>
            <button class="btn btn-sm btn-outline-danger" onclick="deleteWikiPage(${page.id})"><i class="fa-solid fa-trash"></i></button>
          </div>
        </div>
        <hr>
        <div class="wiki-content">${renderContent}</div>
        <div class="small text-muted mt-3">Updated: ${page.updated_at}</div>
      </div>`;
  } catch (e: any) { toast(e.message, true); }
});

expose('showAddWikiPage', function (campaignId: number) {
  showModal('New Wiki Page', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="wikiTitle"></div>
    <div class="mb-3"><label class="form-label">Content (Markdown)</label><textarea class="form-control" id="wikiContent" rows="8" placeholder="Write in Markdown..."></textarea></div>
    <div class="mb-3"><label class="form-label">Visibility</label>
      <select class="form-select" id="wikiVis"><option value="public">Public</option><option value="dm-only">DM Only</option></select></div>
    <button class="btn btn-primary w-100" onclick="saveWikiPage(${campaignId})">Create Page</button>
  `);
});

expose('saveWikiPage', async function (campaignId: number) {
  try {
    await api('POST', `/api/campaigns/${campaignId}/wiki`, {
      campaign_id: campaignId,
      title: (document.getElementById('wikiTitle') as HTMLInputElement).value,
      content: (document.getElementById('wikiContent') as HTMLTextAreaElement).value,
      visibility: (document.getElementById('wikiVis') as HTMLSelectElement).value,
      tags: '[]',
      sort_order: 0,
    });
    hideModal();
    (window as any).showWiki(campaignId);
    toast('Wiki page created');
  } catch (e: any) { toast(e.message, true); }
});

expose('showEditWikiPage', function (id: number, title: string, content: string, visibility: string) {
  showModal('Edit Wiki Page', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="wikiTitle" value="${esc(title)}"></div>
    <div class="mb-3"><label class="form-label">Content (Markdown)</label><textarea class="form-control" id="wikiContent" rows="8">${esc(content)}</textarea></div>
    <div class="mb-3"><label class="form-label">Visibility</label>
      <select class="form-select" id="wikiVis"><option value="public" ${visibility === 'public' ? 'selected' : ''}>Public</option><option value="dm-only" ${visibility === 'dm-only' ? 'selected' : ''}>DM Only</option></select></div>
    <button class="btn btn-primary w-100" onclick="saveEditWikiPage(${id})">Save</button>
  `);
});

expose('saveEditWikiPage', async function (id: number) {
  try {
    const page = await api('GET', `/api/wiki/${id}`);
    await api('PUT', `/api/wiki/${id}`, {
      ...page,
      title: (document.getElementById('wikiTitle') as HTMLInputElement).value,
      content: (document.getElementById('wikiContent') as HTMLTextAreaElement).value,
      visibility: (document.getElementById('wikiVis') as HTMLSelectElement).value,
    });
    hideModal();
    (window as any).loadWikiPage(id);
    toast('Wiki page updated');
  } catch (e: any) { toast(e.message, true); }
});

expose('deleteWikiPage', async function (id: number) {
  if (!confirm('Delete this wiki page?')) return;
  try {
    await api('DELETE', `/api/wiki/${id}`);
    const cid = await api('GET', '/api/campaigns').then((cs: any[]) => cs[0]?.id);
    (window as any).showWiki(cid);
    toast('Wiki page deleted');
  } catch (e: any) { toast(e.message, true); }
});

// ─── Campaign Graph ───

expose('showCampaignGraph', async function (campaignId: number) {
  const modalEl = document.getElementById('genericModal')!;
  const dialogEl = modalEl.querySelector('.modal-dialog') as HTMLElement;
  const origClass = dialogEl.className;
  dialogEl.className = 'modal-dialog modal-xl modal-dialog-scrollable';
  showModal('Campaign Web', `
    <div id="campaignGraphContainer" style="width:100%;height:600px;border:1px solid var(--border);border-radius:4px;background:var(--parchment-light)"></div>
    <div class="text-center mt-2"><small class="text-muted" id="campaignGraphStats">Loading all connections...</small></div>
  `);
  try {
    const data = await api('GET', `/api/campaigns/${campaignId}/graph`);
    const container = document.getElementById('campaignGraphContainer')!;
    createForceGraph(container, data, {
      campaign: { shape: 'ellipse', color: '#8b0000' },
      character: { shape: 'ellipse', color: '#8b0000' },
      location: { shape: 'square', color: '#b8963e' },
      npc: { shape: 'diamond', color: '#2d6a2d' },
      quest: { shape: 'star', color: '#8b4513' },
      session: { shape: 'dot', color: '#5c3a2a' },
      wiki: { shape: 'hexagon', color: '#b8963e' },
      faction: { shape: 'triangle', color: '#9b59b6' },
      encounter: { shape: 'dot', color: '#e67e22' },
      timeline: { shape: 'dot', color: '#5c3a2a' },
      calendar: { shape: 'dot', color: '#b8963e' },
    }, { linkDistance: 250, chargeStrength: -400 });
    document.getElementById('campaignGraphStats')!.innerHTML =
      `${data.nodes.length} entities &middot; ${data.edges.length} connections`;
  } catch (e:any) {
    const container = document.getElementById('campaignGraphContainer');
    if (container) container.innerHTML = `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">${esc(e.message)}</p></div>`;
  }
  modalEl.addEventListener('hidden.bs.modal', function restore() {
    dialogEl.className = origClass;
    modalEl.removeEventListener('hidden.bs.modal', restore);
  }, { once: true });
});

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
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="itemDesc" rows="2"></textarea></div>
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

expose('importCompendiumEquipmentToOneShot', async function (equipmentId: number, adventureId: number, quantity: number) {
  try {
    await api('POST', `/api/oneshot-adventures/${adventureId}/import/compendium-equipment`, {
      compendium_equipment_id: equipmentId,
      adventure_id: adventureId,
      quantity: quantity || 1,
    });
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
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="editItemDesc" rows="2"></textarea></div>
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
  showModal('Monster Compendium', `<div id="compendiumMonsterPickerContent" hx-get="/htmx/compendium-monsters/oneshot/${adventureId}" hx-trigger="load" hx-swap="innerHTML"><div class="text-center py-3"><i class="fa-solid fa-spinner fa-spin me-1"></i>Loading...</div></div>`);
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
            ? `<a href="javascript:void(0)" onclick="htmx.ajax('GET','/compendium/card/monster/${m.compendium_monster_id}',{target:'#cardContainer',swap:'beforeend'})" class="text-decoration-none"><strong>${esc(m.name)}</strong></a>`
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
            ? `<a href="javascript:void(0)" onclick="htmx.ajax('GET','/compendium/card/monster/${m.compendium_monster_id}',{target:'#cardContainer',swap:'beforeend'})" class="text-decoration-none"><strong>${esc(m.name)}</strong></a>`
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

// ─── Polymorphic File Uploads ───

expose('showUploadModal', function (ownerType: string, ownerId: number) {
  showModal('Upload File', `
    <div class="mb-3">
      <label class="form-label">Select file</label>
      <input type="file" class="form-control" id="uploadFileInput">
    </div>
    <button class="btn btn-primary w-100" onclick="doUpload('${ownerType}', ${ownerId})"><i class="fa-solid fa-upload me-1"></i>Upload</button>
  `);
});

expose('doUpload', async function (ownerType: string, ownerId: number) {
  const input = document.getElementById('uploadFileInput') as HTMLInputElement;
  if (!input?.files?.length) { toast('Select a file', true); return; }
  const form = new FormData();
  form.append('file', input.files[0]);
  form.append('owner_type', ownerType);
  form.append('owner_id', String(ownerId));
  try {
    const res = await fetch('/api/upload', { method: 'POST', body: form,
      headers: { 'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || getCsrfToken() }
    });
    if (!res.ok) throw new Error((await res.json()).error || 'Upload failed');
    hideModal();
    toast('File uploaded');
  } catch (e: any) { toast(e.message, true); }
});

// ─── Show combat nav for admin ───
// (handled in init by checking role)

// ═══════════════════════════════════════════
// Campaign Completeness Enhancements
// ═══════════════════════════════════════════

// ─── Campaign Dashboard ───

expose('showCampaignDashboard', async function (campaignId: number, campaignName: string) {
  showModal(`${esc(campaignName)} Dashboard`, `<div id="campaignDashContent"><div class="ornament">✧ Loading dashboard... ✧</div></div>`);
  try {
    const d = await api('GET', `/api/campaigns/${campaignId}/dashboard`);
    const hpPct = (h: number, m: number) => m > 0 ? Math.round((h / m) * 100) : 0;
    const avatarLetter = (n: string) => (n || '?').charAt(0).toUpperCase();

    const content = `
      <div class="dash-grid">
        <div class="dash-card">
          <h6>Characters</h6>
          ${(d.characters || []).map((ch: any) => `
            <div class="dash-char-card" onclick="openChar(${ch.id})" style="cursor:pointer">
              <div class="char-avatar">${avatarLetter(ch.name)}</div>
              <div class="char-info">
                <div class="char-name">${esc(ch.name)}</div>
                <div class="char-detail">${esc(ch.race)} ${esc(ch.class)} · Lvl ${ch.level}</div>
                <div class="dash-hp-bar"><div class="dash-hp-bar-fill${hpPct(ch.hp_current, ch.hp_max) < 30 ? ' low-hp' : ''}" style="width:${hpPct(ch.hp_current, ch.hp_max)}%"></div></div>
              </div>
              <span class="fw-bold" style="font-size:0.85rem">${ch.hp_current}/${ch.hp_max}</span>
            </div>
          `).join('') || '<div class="text-muted small">No characters yet.</div>'}
        </div>
        <div class="dash-card">
          <h6>Overview</h6>
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px">
            <div><div class="dash-value">${d.active_quests}</div><div class="dash-label">Active Quests</div></div>
            <div><div class="dash-value">${d.upcoming_sessions}</div><div class="dash-label">Upcoming Sessions</div></div>
            <div><div class="dash-value">${d.active_conditions}</div><div class="dash-label">Conditions</div></div>
            <div><div class="dash-value">${d.downtime_count}</div><div class="dash-label">Downtime Acts</div></div>
            <div><div class="dash-value">${d.recent_journal}</div><div class="dash-label">Journal (7d)</div></div>
            <div><div class="dash-value">${d.total_members}</div><div class="dash-label">Members</div></div>
          </div>
        </div>
        <div class="dash-card">
          <h6>Upcoming Events</h6>
          ${(d.upcoming_events || []).map((ev: any) => `
            <div class="dash-list-item">
              <span>${esc(ev.title)}</span>
              <span class="text-muted small">${ev.event_date || ''}</span>
            </div>
          `).join('') || '<div class="text-muted small">No upcoming events.</div>'}
        </div>
        <div class="dash-card">
          <h6>Recent Timeline</h6>
          ${(d.recent_timeline || []).map((tl: any) => `
            <div class="dash-list-item">
              <span>${esc(tl.title)}</span>
              <span class="text-muted small">${tl.event_date || ''}</span>
            </div>
          `).join('') || '<div class="text-muted small">No timeline events.</div>'}
        </div>
        <div class="dash-card">
          <h6>Recent Recaps</h6>
          ${(d.recent_recaps || []).map((r: any) => `
            <div class="dash-list-item">
              <span>${esc(r.title)}</span>
              <span class="text-muted small">${r.created_at || ''}</span>
            </div>
          `).join('') || '<div class="text-muted small">No recaps yet.</div>'}
        </div>
        <div class="dash-card">
          <h6>Recent Combats</h6>
          ${(d.recent_combats || []).map((cbt: any) => `
            <div class="dash-list-item">
              <span>${esc(cbt.name)}</span>
              <span class="text-muted small">Round ${cbt.round}</span>
            </div>
          `).join('') || '<div class="text-muted small">No combats yet.</div>'}
        </div>
        <div class="dash-card">
          <h6>Recent Dice Rolls</h6>
          ${(d.recent_dice_rolls || []).map((dr: any) => `
            <div class="dice-roll-mini">
              <span class="roll-expr">${esc(dr.expression)}</span>
              <span class="roll-total">${dr.total}</span>
            </div>
          `).join('') || '<div class="text-muted small">No dice rolls yet.</div>'}
        </div>
      </div>
      <div class="text-center mt-3">
        <button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>`;
    document.getElementById('campaignDashContent')!.innerHTML = content;
  } catch (e: any) {
    document.getElementById('campaignDashContent')!.innerHTML = `<div class="empty-state"><p class="text-danger">${esc(e.message)}</p></div>`;
  }
});

// ─── Party Inventory & Treasury ───

expose('showPartyInventory', async function (campaignId: number) {
  showModal('Party Inventory', `<div id="partyInvContent"><div class="ornament">✧ Loading... ✧</div></div>`);
  try {
    const items = await api('GET', `/api/campaigns/${campaignId}/party-items`);
    const content = `
      <button class="btn btn-gold btn-sm mb-2" onclick="addPartyItem(${campaignId})"><i class="fa-solid fa-plus me-1"></i>Add Item</button>
      ${items.length ? items.map((i: any) => `
        <div class="inv-item">
          <div>
            <strong>${esc(i.name)}</strong>
            <span class="badge badge-muted ms-1">×${i.quantity}</span>
            ${i.notes ? `<div class="small text-muted">${esc(i.notes)}</div>` : ''}
          </div>
          <button class="btn btn-sm btn-outline-danger" onclick="deletePartyItem(${campaignId}, ${i.id})"><i class="fa-solid fa-trash"></i></button>
        </div>
      `).join('') : '<div class="text-muted small fst-italic">No party items yet. Add some loot!</div>'}
      <div class="text-center mt-3">
        <button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>`;
    document.getElementById('partyInvContent')!.innerHTML = content;
  } catch (e: any) {
    document.getElementById('partyInvContent')!.innerHTML = `<p class="text-danger">${esc(e.message)}</p>`;
  }
});

expose('addPartyItem', async function (campaignId: number) {
  showModal('Add Party Item', `
    <div class="mb-2"><label class="form-label">Item Name</label><input class="form-control" id="piName"></div>
    <div class="mb-2"><label class="form-label">Quantity</label><input class="form-control" id="piQty" type="number" value="1"></div>
    <div class="mb-2"><label class="form-label">Notes</label><textarea class="form-control" id="piNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="savePartyItem(${campaignId})">Add</button>
  `);
});

expose('savePartyItem', async function (campaignId: number) {
  const name = (document.getElementById('piName') as HTMLInputElement).value.trim();
  if (!name) { toast('Name required', true); return; }
  await api('POST', `/api/campaigns/${campaignId}/party-items`, {
    name,
    quantity: parseInt((document.getElementById('piQty') as HTMLInputElement).value) || 1,
    notes: (document.getElementById('piNotes') as HTMLTextAreaElement).value,
  });
  hideModal();
  toast('Item added to party inventory');
  (window as any).showPartyInventory(campaignId);
});

expose('deletePartyItem', async function (campaignId: number, itemId: number) {
  if (!confirm('Remove this item?')) return;
  await api('DELETE', `/api/party-items/${itemId}`);
  toast('Item removed');
  (window as any).showPartyInventory(campaignId);
});

// ─── Session Planner ───

expose('showSessionPlanner', async function (campaignId: number) {
  showModal('Session Planner', `<div id="sessionPlanContent"><div class="ornament">✧ Loading sessions... ✧</div></div>`);
  try {
    const plans = await api('GET', `/api/campaigns/${campaignId}/session-plans`);
    const statusBadge = (s: string) => {
      const cls = s === 'planned' ? 'status-badge-planned' : s === 'ready' ? 'status-badge-ready' : s === 'in-progress' ? 'status-badge-in-progress' : 'status-badge-completed';
      return `<span class="${cls}">${esc(s)}</span>`;
    };
    const content = `
      <button class="btn btn-gold btn-sm mb-2" onclick="showSessionPlanForm(${campaignId})"><i class="fa-solid fa-plus me-1"></i>New Session Plan</button>
      ${plans.length ? plans.map((p: any) => `
        <div class="session-plan-card">
          <div class="d-flex justify-content-between align-items-start">
            <div>
              <div class="plan-title">${esc(p.title)}</div>
              <div class="plan-meta">
                ${p.session_date ? `<span><i class="fa-regular fa-calendar me-1"></i>${esc(p.session_date)}</span>` : ''}
                ${p.expected_duration ? `<span class="ms-2"><i class="fa-regular fa-clock me-1"></i>${esc(p.expected_duration)}</span>` : ''}
              </div>
            </div>
            <div class="d-flex gap-1 align-items-center">
              ${statusBadge(p.status)}
              <button class="btn btn-sm btn-outline-primary" onclick="showSessionPlanForm(${campaignId}, ${JSON.stringify(p).replace(/"/g, "'")})"><i class="fa-solid fa-pen"></i></button>
              <button class="btn btn-sm btn-outline-danger" onclick="deleteSessionPlan(${p.id}, ${campaignId})"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>
          ${p.dm_notes ? `<div class="small text-muted mt-1">${esc(p.dm_notes.substring(0, 200))}${p.dm_notes.length > 200 ? '...' : ''}</div>` : ''}
        </div>
      `).join('') : '<div class="text-muted small fst-italic">No session plans yet. Create one to get started!</div>'}
      <div class="text-center mt-3">
        <button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>`;
    document.getElementById('sessionPlanContent')!.innerHTML = content;
  } catch (e: any) {
    document.getElementById('sessionPlanContent')!.innerHTML = `<p class="text-danger">${esc(e.message)}</p>`;
  }
});

expose('showSessionPlanForm', function (campaignId: number, plan?: any) {
  const isEdit = !!plan;
  const title = isEdit ? 'Edit Session Plan' : 'New Session Plan';
  showModal(title, `
    <div class="mb-2"><label class="form-label">Title</label><input class="form-control" id="spTitle" value="${isEdit ? esc(plan.title) : ''}"></div>
    <div class="row g-2 mb-2">
      <div class="col-6"><label class="form-label">Session Date</label><input class="form-control" id="spDate" type="date" value="${isEdit && plan.session_date ? plan.session_date : ''}"></div>
      <div class="col-6"><label class="form-label">Expected Duration</label><input class="form-control" id="spDuration" placeholder="e.g. 3 hours" value="${isEdit ? esc(plan.expected_duration || '') : ''}"></div>
    </div>
    <div class="mb-2"><label class="form-label">Status</label>
      <select class="form-select" id="spStatus">
        <option value="planned" ${isEdit && plan.status === 'planned' ? 'selected' : ''}>Planned</option>
        <option value="ready" ${isEdit && plan.status === 'ready' ? 'selected' : ''}>Ready</option>
        <option value="in-progress" ${isEdit && plan.status === 'in-progress' ? 'selected' : ''}>In Progress</option>
        <option value="completed" ${isEdit && plan.status === 'completed' ? 'selected' : ''}>Completed</option>
      </select>
    </div>
    <div class="mb-2"><label class="form-label">DM Notes</label><textarea class="form-control" id="spNotes" rows="3">${isEdit ? esc(plan.dm_notes || '') : ''}</textarea></div>
    <div class="mb-2"><label class="form-label">Planned Encounters (one per line)</label><textarea class="form-control" id="spEncounters" rows="2" placeholder="Goblin ambush&#10;Bugbear leader">${isEdit && plan.planned_encounters ? (Array.isArray(plan.planned_encounters) ? plan.planned_encounters.join('\n') : plan.planned_encounters) : ''}</textarea></div>
    <div class="mb-2"><label class="form-label">Player Goals (one per line)</label><textarea class="form-control" id="spGoals" rows="2" placeholder="Rescue the prisoners&#10;Find the hidden passage">${isEdit && plan.player_goals ? (Array.isArray(plan.player_goals) ? plan.player_goals.join('\n') : plan.player_goals) : ''}</textarea></div>
    <button class="btn btn-primary w-100" onclick="saveSessionPlan(${campaignId}${isEdit ? `, ${plan.id}` : ''})"><i class="fa-solid fa-save me-1"></i>${isEdit ? 'Update' : 'Create'}</button>
  `);
});

expose('saveSessionPlan', async function (campaignId: number, planId?: number) {
  const title = (document.getElementById('spTitle') as HTMLInputElement).value.trim();
  if (!title) { toast('Title required', true); return; }
  const encounters = (document.getElementById('spEncounters') as HTMLTextAreaElement).value.split('\n').filter((l: string) => l.trim());
  const goals = (document.getElementById('spGoals') as HTMLTextAreaElement).value.split('\n').filter((l: string) => l.trim());
  const body = {
    title,
    session_date: (document.getElementById('spDate') as HTMLInputElement).value || '',
    status: (document.getElementById('spStatus') as HTMLSelectElement).value,
    dm_notes: (document.getElementById('spNotes') as HTMLTextAreaElement).value,
    planned_encounters: JSON.stringify(encounters),
    npc_ids: '[]',
    player_goals: JSON.stringify(goals),
    expected_duration: (document.getElementById('spDuration') as HTMLInputElement).value,
  };
  if (planId) {
    await api('PUT', `/api/session-plans/${planId}`, body);
  } else {
    await api('POST', `/api/campaigns/${campaignId}/session-plans`, body);
  }
  hideModal();
  toast(planId ? 'Session plan updated' : 'Session plan created');
  (window as any).showSessionPlanner(campaignId);
});

expose('deleteSessionPlan', async function (planId: number, campaignId: number) {
  if (!confirm('Delete this session plan?')) return;
  await api('DELETE', `/api/session-plans/${planId}`);
  toast('Session plan deleted');
  (window as any).showSessionPlanner(campaignId);
});

// ─── Encounter Difficulty Calculator ───

const CR_XP: Record<string, number> = {
  '0': 10, '1/8': 25, '1/4': 50, '1/2': 100, '1': 200, '2': 450, '3': 700,
  '4': 1100, '5': 1800, '6': 2300, '7': 2900, '8': 3900, '9': 5000, '10': 5900,
  '11': 7200, '12': 8400, '13': 10000, '14': 11500, '15': 13000, '16': 15000,
  '17': 18000, '18': 20000, '19': 22000, '20': 25000, '21': 33000, '22': 41000,
  '23': 50000, '24': 62000, '25': 75000, '30': 155000,
};

expose('showEncounterDifficulty', function () {
  showModal('Encounter Difficulty Calculator', `
    <div class="diff-calc-section">
      <h6>Party</h6>
      <div class="row g-2 mb-2">
        <div class="col-6"><label class="form-label"># Characters</label><input class="form-control" id="ecPartySize" type="number" value="4" min="1" max="10" oninput="calcEncounterDifficulty()"></div>
        <div class="col-6"><label class="form-label">Average Level</label><input class="form-control" id="ecAvgLevel" type="number" value="5" min="1" max="20" oninput="calcEncounterDifficulty()"></div>
      </div>
      <h6 class="mt-2">Monsters</h6>
      <div id="ecMonsterList"></div>
      <button class="btn btn-sm btn-outline-primary mt-1" onclick="addMonsterRow()"><i class="fa-solid fa-plus me-1"></i>Add Monster</button>
      <div id="ecResult" class="mt-3"></div>
    </div>
    <div class="text-center mt-2">
      <button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Close</button>
    </div>
  `);
  // Add first monster row
  addMonsterRow();
});

expose('addMonsterRow', function () {
  const list = document.getElementById('ecMonsterList');
  if (!list) return;
  const idx = list.children.length;
  const crOptions = Object.keys(CR_XP).map(cr => `<option value="${cr}">${cr}</option>`).join('');
  const row = document.createElement('div');
  row.className = 'row g-2 mb-1 align-items-center';
  row.innerHTML = `
    <div class="col-4"><input class="form-control form-control-sm" id="ecMonsterName${idx}" placeholder="Name"></div>
    <div class="col-3"><select class="form-select form-select-sm" id="ecMonsterCR${idx}" onchange="calcEncounterDifficulty()">${crOptions}</select></div>
    <div class="col-2"><input class="form-control form-control-sm" id="ecMonsterQty${idx}" type="number" value="1" min="1" oninput="calcEncounterDifficulty()"></div>
    <div class="col-3"><button class="btn btn-sm btn-outline-danger" onclick="this.closest('.row').remove();calcEncounterDifficulty()"><i class="fa-solid fa-xmark"></i></button></div>
  `;
  list.appendChild(row);
  calcEncounterDifficulty();
});

expose('calcEncounterDifficulty', function () {
  const partySize = parseInt((document.getElementById('ecPartySize') as HTMLInputElement)?.value) || 4;
  const avgLevel = parseInt((document.getElementById('ecAvgLevel') as HTMLInputElement)?.value) || 5;
  const resultEl = document.getElementById('ecResult');
  if (!resultEl) return;

  // Party XP thresholds (DMG)
  const thresholds = {
    easy: avgLevel * 25 * partySize,
    medium: avgLevel * 50 * partySize,
    hard: avgLevel * 75 * partySize,
    deadly: avgLevel * 100 * partySize,
  };

  // Sum monster XP
  const monsterList = document.getElementById('ecMonsterList');
  if (!monsterList) return;
  let totalXp = 0;
  let monsterCount = 0;
  const monsters: Array<{ name: string; cr: string; qty: number; xp: number }> = [];
  for (let i = 0; i < monsterList.children.length; i++) {
    const nameInput = document.getElementById(`ecMonsterName${i}`) as HTMLInputElement;
    const crSelect = document.getElementById(`ecMonsterCR${i}`) as HTMLSelectElement;
    const qtyInput = document.getElementById(`ecMonsterQty${i}`) as HTMLInputElement;
    if (nameInput && crSelect && qtyInput) {
      const cr = crSelect.value;
      const qty = parseInt(qtyInput.value) || 1;
      const xp = (CR_XP[cr] || 0) * qty;
      totalXp += xp;
      monsterCount += qty;
      monsters.push({ name: nameInput.value || `CR ${cr}`, cr, qty, xp });
    }
  }

  // Encounter multiplier
  let multiplier = 1;
  if (monsterCount >= 2) multiplier = 1.5;
  if (monsterCount >= 3) multiplier = 2;
  if (monsterCount >= 7) multiplier = 2.5;
  if (monsterCount >= 11) multiplier = 3;
  if (monsterCount >= 15) multiplier = 4;

  const adjustedXp = Math.round(totalXp * multiplier);

  // Determine difficulty
  let difficulty = 'easy';
  let badgeClass = 'diff-badge-easy';
  let pct = (adjustedXp / thresholds.deadly) * 100;
  if (adjustedXp >= thresholds.deadly) { difficulty = 'deadly'; badgeClass = 'diff-badge-deadly'; }
  else if (adjustedXp >= thresholds.hard) { difficulty = 'hard'; badgeClass = 'diff-badge-hard'; }
  else if (adjustedXp >= thresholds.medium) { difficulty = 'medium'; badgeClass = 'diff-badge-medium'; }
  pct = Math.min(100, pct);

  resultEl.innerHTML = `
    <div class="diff-meter position-relative" style="height:20px">
      <div class="diff-marker" style="left:${pct}%"></div>
    </div>
    <div class="d-flex justify-content-between small text-muted">
      <span>Easy (${thresholds.easy})</span>
      <span>Medium (${thresholds.medium})</span>
      <span>Hard (${thresholds.hard})</span>
      <span>Deadly (${thresholds.deadly})</span>
    </div>
    <div class="text-center mt-2">
      <span class="${badgeClass}">${difficulty.toUpperCase()}</span>
      <span class="ms-2 fw-bold">${adjustedXp.toLocaleString()} adjusted XP</span>
    </div>
    <div class="small text-muted mt-1">
      Total XP: ${totalXp.toLocaleString()} × ${multiplier} modifier
      ${monsterCount > 1 ? `(${monsterCount} monsters)` : ''}
      &middot; Per character: ${Math.round(adjustedXp / partySize).toLocaleString()} XP
    </div>
    ${monsters.filter(m => m.name).length ? `<div class="mt-2 small">${monsters.filter(m => m.name).map(m => `<div>${esc(m.name)} ×${m.qty} (${m.xp.toLocaleString()} XP)</div>`).join('')}</div>` : ''}
  `;
});

// ─── Treasure Generator ───

const TREASURE_TABLES: Record<string, Array<{ dice: string; coin: string; multiplier: number }>> = {
  easy: [
    { dice: '2d6', coin: 'CP', multiplier: 10 },
    { dice: '1d6', coin: 'SP', multiplier: 5 },
  ],
  medium: [
    { dice: '4d6', coin: 'CP', multiplier: 10 },
    { dice: '2d6', coin: 'SP', multiplier: 10 },
    { dice: '1d4', coin: 'GP', multiplier: 10 },
  ],
  hard: [
    { dice: '2d6', coin: 'CP', multiplier: 100 },
    { dice: '4d6', coin: 'SP', multiplier: 50 },
    { dice: '2d6', coin: 'GP', multiplier: 20 },
    { dice: '1d4', coin: 'PP', multiplier: 10 },
  ],
  deadly: [
    { dice: '4d6', coin: 'CP', multiplier: 100 },
    { dice: '6d6', coin: 'SP', multiplier: 100 },
    { dice: '4d6', coin: 'GP', multiplier: 100 },
    { dice: '2d6', coin: 'PP', multiplier: 20 },
  ],
};

const MAGIC_ITEMS: Record<string, string[]> = {
  common: ['Potion of Healing', 'Spell Scroll (Cantrip)', 'Cloak of Billowing', 'Candle of the Deep', 'Bag of Tricks (Grey)'],
  uncommon: ['Bag of Holding', 'Cloak of Protection', 'Boots of Striding', 'Wand of Magic Detection', 'Potion of Invisibility', '+1 Weapon'],
  rare: ['Flame Tongue', 'Cloak of Displacement', 'Ring of Protection', 'Belt of Hill Giant Strength', 'Potion of Greater Healing'],
  'very rare': ['Belt of Fire Giant Strength', 'Ring of Spell Turning', 'Cloak of Invisibility', 'Staff of the Magi', 'Potion of Supreme Healing'],
};

async function rollDice(dice: string): Promise<number> {
  try {
    const result = await api('POST', '/api/roll', { expression: dice });
    return result.total || 0;
  } catch {
    return 0;
  }
}

expose('showTreasureGenerator', function () {
  showModal('Treasure Generator', `
    <div class="diff-calc-section">
      <div class="row g-2 mb-2">
        <div class="col-6">
          <label class="form-label">Party Level</label>
          <select class="form-select" id="tgLevel">
            ${Array.from({length: 20}, (_, i) => `<option value="${i+1}" ${i+1 === 5 ? 'selected' : ''}>Level ${i+1}</option>`).join('')}
          </select>
        </div>
        <div class="col-6">
          <label class="form-label">Difficulty</label>
          <select class="form-select" id="tgDifficulty">
            <option value="easy">Easy</option>
            <option value="medium" selected>Medium</option>
            <option value="hard">Hard</option>
            <option value="deadly">Deadly</option>
          </select>
        </div>
      </div>
      <button class="btn btn-gold w-100" onclick="generateTreasure()"><i class="fa-solid fa-wand-sparkles me-1"></i>Generate Treasure</button>
      <div id="tgResult"></div>
    </div>
    <div class="text-center mt-2">
      <button class="btn btn-sm btn-outline-secondary" onclick="hideModal()">Close</button>
    </div>
  `);
});

expose('generateTreasure', async function () {
  const lvl = parseInt((document.getElementById('tgLevel') as HTMLSelectElement).value) || 5;
  const diff = (document.getElementById('tgDifficulty') as HTMLSelectElement).value;
  const resultEl = document.getElementById('tgResult');
  if (!resultEl) return;

  const table = TREASURE_TABLES[diff];
  const lines: string[] = [];
  let totalGp = 0;

  for (const entry of table) {
    const rolled = await rollDice(entry.dice);
    const amount = rolled * entry.multiplier;
    const line = `${rolled} × ${entry.multiplier} = ${amount.toLocaleString()} ${entry.coin}`;
    lines.push(line);

    // Convert to GP estimate
    const gpMultiplier: Record<string, number> = { CP: 0.01, SP: 0.1, EP: 0.5, GP: 1, PP: 10 };
    totalGp += amount * (gpMultiplier[entry.coin] || 0);
  }

  // Magic item tier based on level
  let magicTier = 'common';
  if (lvl >= 5) magicTier = 'uncommon';
  if (lvl >= 11) magicTier = 'rare';
  if (lvl >= 17) magicTier = 'very rare';

  const magicPool = MAGIC_ITEMS[magicTier] || [];
  const magicItem = magicPool[Math.floor(Math.random() * magicPool.length)];

  resultEl.innerHTML = `
    <div class="treasure-result">
      <div class="treasure-total">≈ ${totalGp.toLocaleString()} GP</div>
      ${lines.map(l => `<div class="treasure-line">${l}</div>`).join('')}
      <div class="treasure-line fw-bold mt-2">Magic Item: ${magicItem} (${magicTier})</div>
    </div>
    <button class="btn btn-sm btn-outline-primary mt-2 w-100" onclick="generateTreasure()"><i class="fa-solid fa-rotate me-1"></i>Generate Again</button>
  `;
});

// AI Generation → extracted to ts/ai.ts
import { initAIClickHandler } from './ai';

// PDF Viewer → extracted to ts/pdf-viewer.ts
import { initPdfViewerCleanup } from './pdf-viewer';

// Initialization → extracted to ts/init.ts
import { init } from './init';

// PWA → register service worker for offline support
import { registerSW } from './pwa';

// These are called from inline HTML onclick — register at window level
expose('openCampaignDashboard', function (campaignId: number, name: string) {
  (window as any).showCampaignDashboard(campaignId, name);
});

init();
registerSW();
