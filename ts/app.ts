// @ts-nocheck — legacy monolith being split into modules, pre-existing type errors
import { expose } from './lib/expose';
import { renderMarkdown } from './lib/markdown';
import L from 'leaflet';
import * as bootstrap from 'bootstrap';
expose('bootstrap', bootstrap);
import { showView, setCurrentView, getCurrentView } from './navigation';
import { toggleFabMenu, updateFabForView } from './fab';
import './fab';
import './dice';
import './characters/resources';
import { initBridge } from './lib/bridge';
import { esc, attrEscape, capitalize, showModal, hideModal, toast, openCompendiumPicker } from './lib/dom';
import { FilePicker } from './file-picker';
import { initTheme } from './lib/theme';
import { api, setCsrfToken, getCsrfToken, getApiToken, clearApiToken } from './lib/api';
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
import { compendiumSearchModal } from './compendium-search';
import './compendium';
import './combat-tracker';
import './encounter';
import './party';
import './party-subtabs';
import './timeline';
import './factions';
import './share';
import './character-sheet';
import './selection';
import { currentUser, currentChar, currentTab, allLocations, allNPCs, currentCampaign, setCurrentChar, setCurrentTab, setAllLocations, setAllNPCs, setCurrentCampaign } from './lib/state';

// Expose API helper globally for E2E tests (window.api check)
expose('api', api);
try { initBridge(); } catch (e) { console.warn('bridge init failed', e); }

declare const htmx: any;

// ─── FAB ───
// (moved to ts/fab.ts)

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
expose('loadCharacters', loadCharacters);
expose('filterCharacters', filterCharacters);

async function openChar(id: number) {
  try {
    setCurrentChar(await api('GET', `/api/characters/${id}`));
    expose('currentChar', currentChar);
    expose('canEditCharacter', !!(currentChar as any).can_edit);
    setCurrentTab('stats');
    showView('sheet');
    renderSheet();
  } catch (e: any) {
    toast(e.message, true);
  }
}
expose('openChar', openChar);

// ─── Campaign / Character switching ───

expose('switchCampaign', function () {
  setCurrentChar(null);
  (window as any).loadCampaignPicker();
});

expose('switchCharacter', function () {
  if (!currentCampaign) {
    (window as any).loadCampaignPicker();
    return;
  }
  (window as any).loadCharacterPicker(currentCampaign.id);
});

expose('getCurrentCampaign', () => currentCampaign);

// ─── D3 Force Graph / Graph / Analytics → extracted to ts/app/graph-analytics.ts ───
import './app/graph-analytics';
import './app/character-ops';
import './app/character-details';
import './app/combat-conditions';
import './app/notes';
import './app/sort-switch';

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
                  <button class="btn btn-sm btn-outline-primary js-edit-feat" data-id="${f.id}" data-name="${esc(f.name)}" data-desc="${esc(f.description)}" data-prereq="${esc(f.prerequisites)}" data-source="${esc(f.source)}" data-level="${f.level_gained}"><i class="fa-solid fa-pen"></i></button>
                  <button class="btn btn-sm btn-outline-danger" onclick="deleteFeat(${f.id})"><i class="fa-solid fa-trash"></i></button>
                </div>
              </div>
            </div>
          </div>`).join('')
          : '<div class="empty-state"><i class="fa-solid fa-star fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Feats</p><p class="small text-muted">Track your character feats here (distinct from class/race features).</p></div>'}
      </div>`;
    el.querySelectorAll<HTMLButtonElement>('.js-edit-feat').forEach(btn => {
      btn.addEventListener('click', () => (window as any).showEditFeat(Number(btn.dataset.id), btn.dataset.name || '', btn.dataset.desc || '', btn.dataset.prereq || '', btn.dataset.source || '', Number(btn.dataset.level)));
    });
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

// ─── Factions → extracted to ts/factions.ts ───
// ─── Shops → extracted to ts/app/shops.ts ───
import './app/shops';

// ─── Inventory / Spells → extracted to ts/app/inventory-spells.ts ───
import './app/inventory-spells';

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
          <button class="btn btn-sm btn-outline-primary js-edit-rep" data-cid="${r.character_id}" data-fid="${r.faction_id}" data-standing="${r.standing}" data-rank="${esc(r.rank)}" data-notes="${esc(r.notes)}"><i class="fa-solid fa-pen"></i></button>
        </div>
      </div>`;
    }).join('') : '<p class="text-muted small">No reputation tracked for this character. Click a faction to set reputation.</p>';
    list.querySelectorAll<HTMLButtonElement>('.js-edit-rep').forEach(btn => {
      btn.addEventListener('click', () => (window as any).editReputation(Number(btn.dataset.cid), Number(btn.dataset.fid), Number(btn.dataset.standing), btn.dataset.rank || '', btn.dataset.notes || ''));
    });
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
      <button class="btn btn-gold w-100 mt-2 js-confirm-recipe" data-id="${recipe.id}" data-name="${esc(recipe.name)}" data-hours="${recipe.crafting_time_hours}" data-dc="${recipe.difficulty_dc}"><i class="fa-solid fa-hammer me-1"></i>Begin Crafting</button>
    `);
    setTimeout(() => {
      const btn = document.querySelector<HTMLButtonElement>('.js-confirm-recipe');
      if (btn) btn.addEventListener('click', () => (window as any).confirmStartRecipe(Number(btn.dataset.id), btn.dataset.name || '', Number(btn.dataset.hours), Number(btn.dataset.dc)));
    }, 0);
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

// ─── Random Gen / Comparison → extracted to ts/app/random-gen.ts ───
import './app/random-gen';

// ─── Combat Tracker → extracted to ts/combat-tracker.ts ───
import './combat-tracker';

// ─── Wiki / Campaign Graph → extracted to ts/app/wiki.ts ───
import './app/wiki';

// ─── One-Shot Tree/Items/Shops/Monsters/NPCs → extracted to ts/app/oneshot.ts ───
import './app/oneshot';

// ─── Polymorphic File Uploads → extracted to ts/app/uploads.ts ───
import './app/uploads';

// ─── Show combat nav for admin ───
// (handled in init by checking role)

// ─── Campaign Dashboard / Party Inventory / Session Planner → extracted to ts/app/campaign-dashboard.ts ───
import './app/campaign-dashboard';

// ─── Encounter Difficulty / Treasure → extracted to ts/app/encounter-treasure.ts ───
import './app/encounter-treasure';

// ═══════════════════════════════════════════


// AI Generation → extracted to ts/ai.ts
import { initAIClickHandler } from './ai';

// PDF Viewer → extracted to ts/pdf-viewer.ts
import { initPdfViewerCleanup } from './pdf-viewer';

// Initialization → extracted to ts/init.ts
import { init } from './init';

// PWA → register service worker for offline support
import { registerSW } from './pwa';

// Centralized auto-save → dirty tracking, save button state, interval settings
import { startAutoSave, isDirty, isSaving, saveCharacter } from './lib/save';

// These are called from inline HTML onclick — register at window level
expose('openCampaignDashboard', function (campaignId: number, name: string) {
  (window as any).showCampaignDashboard(campaignId, name);
});

function updateSaveBtnState() {
  const btn = document.getElementById('saveCharBtn');
  if (!btn) return;
  btn.className = 'btn btn-sm ' + (isSaving() ? 'btn-outline-primary btn-save-dirty' : isDirty() ? 'btn-gold btn-save-dirty' : 'btn-outline-primary');
  const icon = btn.querySelector('i');
  if (icon) {
    if (isSaving()) { icon.className = 'fa-solid fa-spinner fa-spin me-1'; }
    else { icon.className = 'fa-solid fa-floppy-disk me-1'; }
  }
  btn.disabled = isSaving();
}
expose('updateSaveBtnState', updateSaveBtnState);
window.addEventListener('villum-savestate', updateSaveBtnState);

init();
registerSW();
startAutoSave();
