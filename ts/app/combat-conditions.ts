// @ts-nocheck — split from monolith
import { expose } from '../lib/expose';
import { esc, capitalize, showModal, hideModal, toast } from '../lib/dom';
import { api, getCsrfToken, getApiToken } from '../lib/api';
import { currentChar } from '../lib/state';
import { refreshChar } from '../lib/refresh';
import { renderError } from '../lib/errors';
import { renderSheet, updateField } from '../characters/sheet';
import { renderCombat } from '../characters/combat';
import { renderStats } from '../characters/stats';
import { renderInventory } from '../characters/inventory';
import { renderSpells } from '../characters/spells';
import { animateHpChange } from '../lib/animations';
import { FilePicker } from '../file-picker';

// Roll / Combat Actions
expose('applyDamage', async function () {
  if (!currentChar) return;
  const dmg = parseInt((document.getElementById('dmgInput') as HTMLInputElement)?.value || '0');
  if (!dmg) return;
  const oldHp = currentChar.hp_current;
  const newHp = Math.max(0, currentChar.hp_current - dmg);
  await updateField('hp_current', newHp);
  await (await import('../lib/save')).saveCharacter();
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
  const bar = document.getElementById('charHpBarFill');
  const hpText = document.getElementById('charHpText');
  if (bar && hpText) {
    bar.style.width = Math.max(0, Math.min(100, (oldHp / currentChar.hp_max) * 100)) + '%';
    animateHpChange(hpText, bar, oldHp, currentChar.hp_current, currentChar.hp_max);
  }
});

expose('adjustExhaustion', async function (delta: number) {
  if (!currentChar) return;
  const newLevel = Math.max(0, Math.min(6, (currentChar.exhaustion_level || 0) + delta));
  await api('PATCH', `/api/characters/${currentChar.id}/exhaustion`, { exhaustion_level: newLevel });
  await refreshChar();
  renderStats();
});

expose('toggleIdentify', async function (id: number) {
  const item = currentChar.inventory.find((i:any) => i.id === id);
  if (!item) return;
  const newVal = item.is_identified === false ? true : false;
  await api('PUT', `/api/inventory/${id}`, { ...item, is_identified: newVal });
  await refreshChar();
  renderInventory();
  toast(newVal ? 'Item identified' : 'Item marked unidentified');
});

// Conditions
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

// Feats
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
let editingFeatId: number | null = null;
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

// Companions
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
let compPortraitUrl: string = '';
expose('uploadCompPortrait', async function (id: number) {
  const input = document.getElementById('compPortraitUpload') as HTMLInputElement;
  if (!input.files || !input.files[0]) { toast('Select an image', true); return; }
  const form = new FormData();
  form.append('image', input.files[0]);
  try {
    const res = await fetch('/api/upload', {
      method: 'POST', headers: { 'X-CSRF-Token': getCsrfToken(), ...(getApiToken() ? { 'Authorization': `Bearer ${getApiToken()}` } : {}) }, credentials: 'include', body: form,
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Upload failed');
    compPortraitUrl = data.url;
    hideModal();
    toast('Portrait uploaded — re-open to confirm');
  } catch (e: any) { renderError(e); }
});
expose('browseCompPortrait', async function (id: number) {
  try {
    const url = await FilePicker.pick();
    compPortraitUrl = url;
    hideModal();
    toast('Portrait selected — re-open to confirm');
  } catch (e: any) { renderError(e); }
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
