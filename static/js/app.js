let csrfToken = '';
let currentUser = null;
let currentView = 'characters';
let currentChar = null;
let currentTab = 'stats';
let allLocations = [];
let allNPCs = [];
// ─── Utilities ───
function esc(s) {
    if (!s)
        return '';
    const d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
}
function capitalize(s) {
    return s.charAt(0).toUpperCase() + s.slice(1);
}
// ─── API ───
async function api(method, path, body) {
    const headers = { 'Content-Type': 'application/json' };
    if (csrfToken)
        headers['X-CSRF-Token'] = csrfToken;
    const opts = { method, headers, credentials: 'include' };
    if (body !== undefined)
        opts.body = JSON.stringify(body);
    const res = await fetch(path, opts);
    if (!res.ok) {
        const err = await res.json().catch(() => ({ error: res.statusText }));
        throw new Error(err.error || 'Request failed');
    }
    return res.json();
}
// ─── Bootstrap Modal ───
let genericModal = null;
function getModal() {
    if (!genericModal) {
        genericModal = new bootstrap.Modal(document.getElementById('genericModal'));
    }
    return genericModal;
}
function showModal(title, bodyHtml) {
    document.getElementById('genericModalTitle').textContent = title;
    document.getElementById('genericModalBody').innerHTML = bodyHtml;
    getModal().show();
}
window.showModal = showModal;
function hideModal() {
    getModal().hide();
}
window.hideModal = hideModal;
// ─── Bootstrap Toast ───
function toast(msg, isError = false) {
    const container = document.getElementById('toastContainer');
    const id = 'toast-' + Date.now();
    const bg = isError ? 'bg-danger' : 'bg-success';
    container.innerHTML += `
    <div class="toast align-items-center text-white ${bg} border-0 mb-2" id="${id}" role="alert">
      <div class="d-flex">
        <div class="toast-body">${esc(msg)}</div>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
      </div>
    </div>`;
    const el = document.getElementById(id);
    new bootstrap.Toast(el, { autohide: true, delay: 5000 }).show();
    setTimeout(() => el.remove(), 6000);
}
// ─── Init ───
async function init() {
    try {
        const user = await api('GET', '/api/user/me');
        currentUser = user;
        const tokenRes = await api('GET', '/api/csrf-token');
        csrfToken = tokenRes.token;
        document.getElementById('userName').textContent = user.username;
        if (user.role === 'admin') {
            document.getElementById('adminNavItem').style.display = '';
        }
        showView('characters');
        loadCharacters();
        api('GET', '/api/locations').then(l => allLocations = l).catch(() => { });
        api('GET', '/api/npcs').then(n => allNPCs = n).catch(() => { });
    }
    catch {
        window.location.href = '/login';
    }
}
function showView(view) {
    currentView = view;
    document.getElementById('charactersView').style.display = view === 'characters' || view === 'sheet' ? 'block' : 'none';
    document.getElementById('sheetView').style.display = view === 'sheet' ? 'block' : 'none';
    document.getElementById('diceView').style.display = view === 'dice' ? 'block' : 'none';
    document.getElementById('compendiumView').style.display = view === 'compendium' ? 'block' : 'none';
    document.getElementById('partyView').style.display = view === 'party' ? 'block' : 'none';
}
window.showView = showView;
// ─── Character List ───
async function loadCharacters() {
    try {
        const chars = await api('GET', '/api/characters');
        const grid = document.getElementById('charGrid');
        grid.innerHTML = chars.map((c) => `
      <div class="col-md-6 col-lg-4">
        <div class="character-card" onclick="openChar(${c.id})">
          <div class="char-name">${esc(c.name)}</div>
          <div class="char-detail">${esc(c.race)} ${esc(c.class)} · Level ${c.level}</div>
          <div class="char-hp mt-1">HP: ${c.hp_current}/${c.hp_max}</div>
        </div>
      </div>
    `).join('');
    }
    catch (e) {
        toast(e.message, true);
    }
}
window.loadCharacters = loadCharacters;
async function openChar(id) {
    try {
        currentChar = await api('GET', `/api/characters/${id}`);
        currentTab = 'stats';
        showView('sheet');
        renderSheet();
    }
    catch (e) {
        toast(e.message, true);
    }
}
window.openChar = openChar;
// ─── Character Sheet ───
const sections = ['stats', 'combat', 'spells', 'inventory', 'features', 'locations', 'npcs', 'sessions', 'quests', 'journal', 'graph', 'analytics', 'details', 'dice'];
function renderSheet() {
    if (!currentChar)
        return;
    const c = currentChar;
    document.getElementById('sheetName').textContent = c.name;
    document.getElementById('sheetSubtitle').textContent =
        `${c.race} ${c.class}${c.subclass ? ' (' + c.subclass + ')' : ''} · Level ${c.level}`;
    const tabBar = document.getElementById('tabBar');
    tabBar.innerHTML = sections.map(s => `
    <li class="nav-item"><button class="nav-link ${s === currentTab ? 'active' : ''}" onclick="switchTab('${s}')">${capitalize(s)}</button></li>
  `).join('');
    sections.forEach(s => {
        const el = document.getElementById(s + 'Section');
        el.style.display = s === currentTab ? 'block' : 'none';
    });
    renderStats();
    renderCombat();
    renderSpells();
    renderInventory();
    renderFeatures();
    if (currentTab === 'locations')
        renderLocations();
    if (currentTab === 'npcs')
        renderNPCs();
    if (currentTab === 'sessions')
        renderSessions();
    if (currentTab === 'quests')
        renderQuests();
    if (currentTab === 'journal')
        renderJournal();
    if (currentTab === 'graph')
        renderGraph();
    if (currentTab === 'analytics')
        renderAnalytics();
    renderDetails();
    renderDiceTab();
}
function switchTab(tab) {
    currentTab = tab;
    renderSheet();
}
window.switchTab = switchTab;
// ─── Roll / Combat Actions ───
async function rollCheck(type, name, adv) {
    if (!currentChar)
        return;
    try {
        const result = await api('POST', '/api/roll/check', {
            character_id: currentChar.id, type, name, advantage: adv,
        });
        toast(result.text);
    }
    catch (e) {
        toast(e.message, true);
    }
}
window.rollCheck = rollCheck;
async function applyDamage() {
    if (!currentChar)
        return;
    const dmg = parseInt(document.getElementById('dmgInput')?.value || '0');
    if (!dmg)
        return;
    const newHp = Math.max(0, currentChar.hp_current - dmg);
    await updateField('hp_current', newHp);
}
window.applyDamage = applyDamage;
async function applyHeal() {
    if (!currentChar)
        return;
    const heal = parseInt(document.getElementById('healInput')?.value || '0');
    if (!heal)
        return;
    const newHp = Math.min(currentChar.hp_max, currentChar.hp_current + heal);
    await updateField('hp_current', newHp);
}
window.applyHeal = applyHeal;
async function doRest(type) {
    if (!currentChar)
        return;
    try {
        const result = await api('POST', `/api/characters/${currentChar.id}/rest`, { rest_type: type, hit_dice_count: type === 'short' ? 1 : 0 });
        toast(`${type} rest: healed ${result.hp_healed} HP`);
        currentChar = await api('GET', `/api/characters/${currentChar.id}`);
        renderSheet();
    }
    catch (e) {
        toast(e.message, true);
    }
}
window.doRest = doRest;
async function doLevelUp() {
    if (!currentChar)
        return;
    try {
        const result = await api('POST', `/api/characters/${currentChar.id}/levelup`);
        toast(`Level Up! Now level ${result.new_level} (+${result.hp_gain} HP)`);
        currentChar = await api('GET', `/api/characters/${currentChar.id}`);
        renderSheet();
    }
    catch (e) {
        toast(e.message, true);
    }
}
window.doLevelUp = doLevelUp;
async function updateField(field, value) {
    if (!currentChar)
        return;
    currentChar[field] = value;
    try {
        await api('PUT', `/api/characters/${currentChar.id}`, currentChar);
    }
    catch (e) {
        toast(e.message, true);
    }
}
window.updateField = updateField;
// ─── Stats ───
function renderStats() {
    const c = currentChar;
    const el = document.getElementById('statsSection');
    const abils = ['str', 'dex', 'con', 'int', 'wis', 'cha'].map(k => ({ key: k, label: k.toUpperCase() }));
    el.innerHTML = `
    <div class="row g-3">
      ${abils.map(a => {
        const val = c[a.key];
        const mod = c[`${a.key}_mod`];
        const cls = mod > 0 ? 'text-success' : mod < 0 ? 'text-danger' : 'text-muted';
        return `<div class="col-4 col-md-2">
          <div class="ability-box" onclick="rollCheck('check','${a.key}','normal')">
            <div class="abil-label">${a.label}</div>
            <div class="abil-value">${val}</div>
            <div class="abil-mod ${cls}">${mod >= 0 ? '+' : ''}${mod}</div>
          </div>
        </div>`;
    }).join('')}
    </div>
    <div class="d-flex gap-2 mt-3">
      <button class="btn btn-sm btn-outline-primary" onclick="rollCheck('check','str','advantage')"><i class="fa-solid fa-chevron-up me-1"></i>Advantage</button>
      <button class="btn btn-sm btn-outline-primary" onclick="rollCheck('check','str','disadvantage')"><i class="fa-solid fa-chevron-down me-1"></i>Disadvantage</button>
    </div>
    <div class="ornament my-2">✧</div>
    <div class="row g-3 mt-2">
      <div class="col-6 col-md-3"><label class="form-label">Proficiency</label><input type="number" class="form-control form-control-sm" value="${c.proficiency_bonus}" onchange="updateField('proficiency_bonus',+this.value)"></div>
      <div class="col-6 col-md-3"><label class="form-label">Inspiration</label><input type="number" class="form-control form-control-sm" value="${c.inspiration}" onchange="updateField('inspiration',+this.value)"></div>
      <div class="col-6 col-md-3"><label class="form-label">Passive Percep.</label><input type="number" class="form-control form-control-sm" value="${c.passive_perception}" onchange="updateField('passive_perception',+this.value)"></div>
      <div class="col-6 col-md-3"><label class="form-label">XP</label><input type="number" class="form-control form-control-sm" value="${c.xp}" onchange="updateField('xp',+this.value)"></div>
    </div>
    <h5 class="mt-3">Skills <small class="text-muted fw-normal">(click to roll)</small></h5>
    <div id="skillsArea">${renderSkills(c)}</div>
    <h5 class="mt-3">Proficiencies</h5>
    <div id="profsArea">${(c.proficiencies || []).map((p) => `<span class="badge badge-blood me-1 mb-1">${esc(p.name)} (${p.type}) <a href="#" onclick="deleteProf(${p.id});return false" class="text-white text-decoration-none">×</a></span>`).join('')}</div>
    <button class="btn btn-sm btn-outline-primary mt-2" onclick="addProf()"><i class="fa-solid fa-plus me-1"></i>Add Proficiency</button>
  `;
}
function renderSkills(c) {
    const skls = [
        { name: 'Athletics', abil: 'str' }, { name: 'Acrobatics', abil: 'dex' }, { name: 'Sleight of Hand', abil: 'dex' }, { name: 'Stealth', abil: 'dex' },
        { name: 'Arcana', abil: 'int' }, { name: 'History', abil: 'int' }, { name: 'Investigation', abil: 'int' }, { name: 'Nature', abil: 'int' }, { name: 'Religion', abil: 'int' },
        { name: 'Animal Handling', abil: 'wis' }, { name: 'Insight', abil: 'wis' }, { name: 'Medicine', abil: 'wis' }, { name: 'Perception', abil: 'wis' }, { name: 'Survival', abil: 'wis' },
        { name: 'Deception', abil: 'cha' }, { name: 'Intimidation', abil: 'cha' }, { name: 'Performance', abil: 'cha' }, { name: 'Persuasion', abil: 'cha' },
    ];
    const profs = (c.proficiencies || []).filter((p) => p.type === 'skill').map((p) => p.name.toLowerCase());
    return skls.map(s => {
        const isProf = profs.includes(s.name.toLowerCase());
        const mod = c[`${s.abil}_mod`];
        const total = isProf ? mod + c.proficiency_bonus : mod;
        const sign = total >= 0 ? '+' : '';
        return `<div class="skill-row d-flex justify-content-between" onclick="rollCheck('skill','${s.name}','normal')">
      <span class="skill-name">${s.name}${isProf ? ' <span class="text-primary">★</span>' : ''}</span>
      <span class="fw-bold">${sign}${total}</span>
    </div>`;
    }).join('');
}
// ─── Combat ───
function renderCombat() {
    const c = currentChar;
    const el = document.getElementById('combatSection');
    const pct = c.hp_max > 0 ? Math.round((c.hp_current / c.hp_max) * 100) : 0;
    el.innerHTML = `
    <div class="row g-3">
      <div class="col-4"><div class="combat-stat"><div class="stat-label">AC</div><div class="stat-value">${c.ac}</div></div></div>
      <div class="col-4"><div class="combat-stat"><div class="stat-label">Initiative</div><div class="stat-value">${c.initiative >= 0 ? '+' : ''}${c.initiative}</div></div></div>
      <div class="col-4"><div class="combat-stat"><div class="stat-label">Speed</div><div class="stat-value">${c.speed}</div></div></div>
    </div>
    <h5 class="mt-3">Hit Points</h5>
    <div class="hp-bar position-relative mb-2">
      <div class="hp-bar-fill" style="width:${pct}%"></div>
      <div class="position-absolute top-0 start-0 end-0 bottom-0 d-flex align-items-center justify-content-center text-white small fw-bold" style="font-size:0.8rem">${c.hp_current} / ${c.hp_max}${c.temp_hp > 0 ? ' (+' + c.temp_hp + ' temp)' : ''}</div>
    </div>
    <div class="row g-2">
      <div class="col-4"><label class="form-label small">HP Max</label><input type="number" class="form-control form-control-sm" value="${c.hp_max}" onchange="updateField('hp_max',+this.value)"></div>
      <div class="col-4"><label class="form-label small">Current</label><input type="number" class="form-control form-control-sm" value="${c.hp_current}" onchange="updateField('hp_current',+this.value)"></div>
      <div class="col-4"><label class="form-label small">Temp HP</label><input type="number" class="form-control form-control-sm" value="${c.temp_hp}" onchange="updateField('temp_hp',+this.value)"></div>
    </div>
    <div class="row g-2 mt-2">
      <div class="col-6">
        <label class="form-label small">Damage</label>
        <div class="input-group input-group-sm"><input type="number" class="form-control" id="dmgInput" value="0"><button class="btn btn-danger" onclick="applyDamage()">Apply</button></div>
      </div>
      <div class="col-6">
        <label class="form-label small">Heal</label>
        <div class="input-group input-group-sm"><input type="number" class="form-control" id="healInput" value="0"><button class="btn btn-success" onclick="applyHeal()">Apply</button></div>
      </div>
    </div>
    <div class="d-flex gap-2 mt-3">
      <button class="btn btn-sm btn-outline-primary" onclick="doRest('short')"><i class="fa-solid fa-campground me-1"></i>Short Rest</button>
      <button class="btn btn-sm btn-outline-primary" onclick="doRest('long')"><i class="fa-solid fa-moon me-1"></i>Long Rest</button>
      <button class="btn btn-sm btn-gold" onclick="doLevelUp()"><i class="fa-solid fa-arrow-up me-1"></i>Level Up</button>
    </div>
    <h5 class="mt-3">Saving Throws <small class="text-muted fw-normal">(click to roll)</small></h5>
    <div class="d-flex flex-wrap gap-1 mb-3">
      ${['str', 'dex', 'con', 'int', 'wis', 'cha'].map(a => {
        const mod = c[`${a}_mod`];
        const total = c.proficiency_bonus + mod;
        const sign = total >= 0 ? '+' : '';
        return `<span class="badge badge-gold" style="cursor:pointer" onclick="rollCheck('save','${a}','normal')">${a.toUpperCase()} ${sign}${total}</span>`;
    }).join('')}
    </div>
    <h5 class="mt-3">Death Saves</h5>
    <div class="row g-2">
      <div class="col-6"><label class="form-label small">Successes</label><input type="number" class="form-control form-control-sm" value="${c.death_save_successes}" onchange="updateField('death_save_successes',+this.value)" min="0" max="3"></div>
      <div class="col-6"><label class="form-label small">Failures</label><input type="number" class="form-control form-control-sm" value="${c.death_save_failures}" onchange="updateField('death_save_failures',+this.value)" min="0" max="3"></div>
    </div>
    <h5 class="mt-3">Concentration</h5>
    <div class="form-check"><input type="checkbox" class="form-check-input" id="concentrationCb" ${c.concentrating ? 'checked' : ''} onchange="updateField('concentrating',this.checked)"><label class="form-check-label" for="concentrationCb">Concentrating on a spell</label></div>
    <div class="mt-2">
      <label class="form-label small">Concentrating On</label>
      <input class="form-control form-control-sm" value="${esc(c.concentrating_on)}" onchange="updateField('concentrating_on',this.value)" placeholder="e.g. Hunter's Mark">
    </div>
    <h5 class="mt-3">Hit Dice</h5>
    <div class="row g-2">
      <div class="col-6"><label class="form-label small">Total</label><input type="number" class="form-control form-control-sm" value="${c.hit_dice_total}" onchange="updateField('hit_dice_total',+this.value)"></div>
      <div class="col-6"><label class="form-label small">Used</label><input type="number" class="form-control form-control-sm" value="${c.hit_dice_used}" onchange="updateField('hit_dice_used',+this.value)"></div>
    </div>`;
}
// ─── Currency ───
async function updateCurrency() {
    if (!currentChar)
        return;
    const coins = ['cp', 'sp', 'ep', 'gp', 'pp'];
    const updates = {};
    coins.forEach(c => { updates[c] = +document.getElementById('coin' + c)?.value || 0; });
    await api('PUT', `/api/characters/${currentChar.id}/currency`, updates);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    toast('Currency updated');
}
window.updateCurrency = updateCurrency;
// ─── Inventory ───
function renderInventory() {
    const inv = currentChar.inventory || [];
    const categories = { weapon: [], armor: [], gear: [], potion: [], scroll: [], tool: [], wondrous: [], other: [] };
    inv.forEach((i) => { if (categories[i.category])
        categories[i.category].push(i);
    else
        categories.other.push(i); });
    const total = inv.reduce((s, i) => s + (i.weight || 0) * (i.quantity || 1), 0);
    document.getElementById('inventorySection').innerHTML = `
    <div class="d-flex justify-content-between align-items-center">
      <h5>Inventory <span class="text-muted small">(Total: ${total} lbs)</span></h5>
      <div><button class="btn btn-primary btn-sm" onclick="addInventory()"><i class="fa-solid fa-plus me-1"></i>Add Item</button></div>
    </div>
    <div class="mt-2" id="invList">
      ${Object.entries(categories).filter(([, items]) => items.length).map(([cat, items]) => `
        <h6 class="mt-3 text-muted">${capitalize(cat)}</h6>
        ${items.map((i) => `
          <div class="inv-item${i.equipped ? ' equipped' : ''}">
            <div><span class="fw-bold">${esc(i.name)}</span> ${i.quantity > 1 ? `<span class="badge badge-muted">x${i.quantity}</span>` : ''}
              ${i.equipped ? '<span class="badge badge-gold">Equipped</span>' : ''}
              ${i.damage_dice ? `<span class="badge badge-blood ms-1">${esc(i.damage_dice)} ${esc(i.damage_type)}</span>` : ''}
              ${i.ac_bonus > 0 ? `<span class="badge badge-gold ms-1">AC+${i.ac_bonus}</span>` : ''}</div>
            <div class="d-flex gap-1">
              <button class="btn btn-sm btn-outline-primary" onclick="editInventory(${i.id},'${esc(i.name)}',${i.quantity},'${esc(i.category)}',${i.weight},${i.equipped})" title="Edit"><i class="fa-solid fa-pen"></i></button>
              <button class="btn btn-sm btn-outline-secondary" onclick="toggleEquip(${i.id})" title="${i.equipped ? 'Unequip' : 'Equip'}"><i class="fa-solid fa-shield-halved"></i></button>
              <button class="btn btn-sm btn-outline-danger" onclick="deleteInventory(${i.id})" title="Remove"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>`).join('')}
      `).join('') || '<div class="empty-state"><i class="fa-solid fa-backpack fa-2x mb-2 d-block text-muted"></i>No items. Add gear to your inventory.</div>'}
    </div>`;
}
window.addInventory = function () {
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
};
window.editInventory = function (id, name, qty, cat, weight, equipped) {
    showModal('Edit Item', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="invName" value="${esc(name)}"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Quantity</label><input class="form-control" id="invQty" type="number" value="${qty}"></div>
      <div class="col-6"><label class="form-label">Weight (lbs)</label><input class="form-control" id="invWeight" type="number" value="${weight}" step="0.1"></div>
    </div>
    <div class="mb-3"><label class="form-label">Category</label>
      <select class="form-select" id="invCat">${['gear', 'weapon', 'armor', 'potion', 'scroll', 'tool', 'wondrous', 'other'].map(c => `<option value="${c}"${c === cat ? ' selected' : ''}>${capitalize(c)}</option>`).join('')}</select></div>
    <div class="mb-3"><div class="form-check"><input type="checkbox" class="form-check-input" id="invEquip"${equipped ? ' checked' : ''}><label class="form-check-label">Equipped</label></div></div>
    <button class="btn btn-primary w-100" onclick="saveEditInventory(${id},this)">Save</button>
  `);
};
window.saveInventory = async function (btn) {
    await api('POST', `/api/characters/${currentChar.id}/inventory`, {
        name: document.getElementById('invName').value,
        quantity: +document.getElementById('invQty').value || 1,
        weight: +document.getElementById('invWeight').value || 0,
        category: document.getElementById('invCat').value,
    });
    hideModal();
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderInventory();
    toast('Item added');
};
window.saveEditInventory = async function (id, btn) {
    await api('PUT', `/api/inventory/${id}`, {
        name: document.getElementById('invName').value,
        quantity: +document.getElementById('invQty').value || 1,
        weight: +document.getElementById('invWeight').value || 0,
        category: document.getElementById('invCat').value,
        equipped: document.getElementById('invEquip').checked,
    });
    hideModal();
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderInventory();
    toast('Item updated');
};
window.deleteInventory = async function (id) {
    if (!confirm('Remove this item?'))
        return;
    await api('DELETE', `/api/inventory/${id}`);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderInventory();
    toast('Item removed');
};
window.toggleEquip = async function (id) {
    const item = currentChar.inventory.find((i) => i.id === id);
    if (!item)
        return;
    item.equipped = !item.equipped;
    await api('PUT', `/api/inventory/${id}`, item);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderInventory();
    toast(item.equipped ? 'Equipped' : 'Unequipped');
};
// ─── Spells ───
function renderSpells() {
    const spells = currentChar.spells || [];
    const sc = currentChar.spellcasting || {};
    document.getElementById('spellsSection').innerHTML = sc.spellcasting_ability ? `
    <h5>Spellcasting</h5>
    <div class="row g-3 mb-3">
      <div class="col-md-4"><label class="form-label">Ability</label><input class="form-control form-control-sm" value="${esc(sc.spellcasting_ability)}" onchange="updateSpellcasting('spellcasting_ability',this.value)"></div>
      <div class="col-md-4"><label class="form-label">Save DC</label><input class="form-control form-control-sm" type="number" value="${sc.spell_save_dc || 0}" onchange="updateSpellcasting('spell_save_dc',+this.value)"></div>
      <div class="col-md-4"><label class="form-label">Atk Bonus</label><input class="form-control form-control-sm" type="number" value="${sc.spell_attack_bonus || 0}" onchange="updateSpellcasting('spell_attack_bonus',+this.value)"></div>
    </div>
    <h6>Spell Slots</h6>
    <div class="d-flex gap-3 flex-wrap mb-3">
      ${[1, 2, 3, 4, 5, 6, 7, 8, 9].map(lv => {
        const mx = sc[`slots_${lv}_max`] || 0;
        if (!mx)
            return '';
        return `<div class="text-center">
          <div class="text-muted small">Lv ${lv}</div>
          <input type="number" class="form-control form-control-sm text-center" style="width:55px" id="slotUse${lv}" value="${sc[`slots_${lv}_used`] || 0}" onchange="updateSpellSlot(${lv})" min="0" max="${mx}">
          <div class="text-muted small">/ ${mx}</div>
        </div>`;
    }).join('')}
    </div>
    <div class="d-flex justify-content-between align-items-center mt-3">
      <h6>Known Spells</h6>
      <button class="btn btn-primary btn-sm" onclick="addSpell()"><i class="fa-solid fa-plus me-1"></i>Add Spell</button>
    </div>
    <div class="mt-2">
      ${spells.map((s) => `
        <div class="inv-item ${s.prepared ? 'equipped' : ''}">
          <div><span class="fw-bold">${esc(s.name)}</span> <span class="text-muted small">${esc(s.level > 0 ? 'Lv' + s.level : 'Cantrip')} ${esc(s.school)}</span></div>
          <div class="d-flex gap-1">
            <button class="btn btn-sm btn-outline-primary" onclick="editSpell(${s.id},'${esc(s.name)}',${s.level},'${esc(s.school)}',${s.prepared},'${esc(s.components || '')}','${esc(s.range || '')}','${esc(s.casting_time || '')}','${esc(s.duration || '')}','${esc(s.description || '')}')"><i class="fa-solid fa-pen"></i></button>
            <button class="btn btn-sm btn-outline-danger" onclick="deleteSpell(${s.id})"><i class="fa-solid fa-trash"></i></button>
          </div>
        </div>`).join('') || '<div class="empty-state"><i class="fa-solid fa-wand-sparkles fa-2x mb-2 d-block text-muted"></i>No spells known. Learn some magic!</div>'}
    </div>` : `
    <div class="empty-state"><i class="fa-solid fa-wand-sparkles fa-2x mb-2 d-block text-muted"></i>
    <p class="text-muted fst-italic">No spellcasting.</p>
    <button class="btn btn-outline-primary btn-sm" onclick="enableSpellcasting()"><i class="fa-solid fa-magic me-1"></i>Set Up Spellcasting</button></div>`;
}
async function updateSpellcasting(field, value) {
    if (!currentChar)
        return;
    const sc = currentChar.spellcasting || {};
    sc[field] = value;
    await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, sc);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSpells();
}
window.updateSpellcasting = updateSpellcasting;
async function updateSpellSlot(level) {
    if (!currentChar)
        return;
    const sc = currentChar.spellcasting || {};
    sc[`slots_${level}_used`] = +document.getElementById(`slotUse${level}`).value || 0;
    await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, sc);
}
window.updateSpellSlot = updateSpellSlot;
window.enableSpellcasting = async function () {
    currentChar.spellcasting = {
        spellcasting_ability: 'int', spell_save_dc: 10, spell_attack_bonus: 0,
        slots_1_max: 2, slots_1_used: 0,
    };
    await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, currentChar.spellcasting);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSpells();
};
window.addSpell = function () {
    showModal('Add Spell', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="spellName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="spellLevel" type="number" value="0" min="0" max="9"></div>
      <div class="col-6"><label class="form-label">School</label>
        <select class="form-select" id="spellSchool">${['Abjuration', 'Conjuration', 'Divination', 'Enchantment', 'Evocation', 'Illusion', 'Necromancy', 'Transmutation'].map(s => `<option value="${s}">${s}</option>`).join('')}</select></div>
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
};
window.editSpell = function (id, name, level, school, prepared, comp, range, cast, dur, desc) {
    showModal('Edit Spell', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="spellName" value="${esc(name)}"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Level</label><input class="form-control" id="spellLevel" type="number" value="${level}" min="0" max="9"></div>
      <div class="col-6"><label class="form-label">School</label>
        <select class="form-select" id="spellSchool">${['Abjuration', 'Conjuration', 'Divination', 'Enchantment', 'Evocation', 'Illusion', 'Necromancy', 'Transmutation'].map(s => `<option value="${s}"${s === school ? ' selected' : ''}>${s}</option>`).join('')}</select></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Casting Time</label><input class="form-control" id="spellCast" value="${esc(cast)}"></div>
      <div class="col-6"><label class="form-label">Range</label><input class="form-control" id="spellRange" value="${esc(range)}"></div>
    </div>
    <div class="mb-3"><label class="form-label">Components</label><input class="form-control" id="spellComp" value="${esc(comp)}"></div>
    <div class="mb-3"><label class="form-label">Duration</label><input class="form-control" id="spellDur" value="${esc(dur)}"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="spellDesc" rows="3">${esc(desc)}</textarea></div>
    <div class="mb-3"><div class="form-check"><input type="checkbox" class="form-check-input" id="spellPrep"${prepared ? ' checked' : ''}><label class="form-check-label">Prepared</label></div></div>
    <button class="btn btn-primary w-100" onclick="saveEditSpell(${id},this)">Save Spell</button>
  `);
};
window.saveSpell = async function (btn) {
    await api('POST', `/api/characters/${currentChar.id}/spells`, {
        name: document.getElementById('spellName').value,
        level: +document.getElementById('spellLevel').value || 0,
        school: document.getElementById('spellSchool').value,
        casting_time: document.getElementById('spellCast').value,
        range: document.getElementById('spellRange').value,
        components: document.getElementById('spellComp').value,
        duration: document.getElementById('spellDur').value,
        description: document.getElementById('spellDesc').value,
    });
    hideModal();
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSpells();
    toast('Spell added');
};
window.saveEditSpell = async function (id, btn) {
    await api('PUT', `/api/spells/${id}`, {
        name: document.getElementById('spellName').value,
        level: +document.getElementById('spellLevel').value || 0,
        school: document.getElementById('spellSchool').value,
        casting_time: document.getElementById('spellCast').value,
        range: document.getElementById('spellRange').value,
        components: document.getElementById('spellComp').value,
        duration: document.getElementById('spellDur').value,
        description: document.getElementById('spellDesc').value,
        prepared: document.getElementById('spellPrep').checked,
    });
    hideModal();
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSpells();
    toast('Spell updated');
};
window.deleteSpell = async function (id) {
    if (!confirm('Remove this spell?'))
        return;
    await api('DELETE', `/api/spells/${id}`);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSpells();
    toast('Spell removed');
};
// ─── Features ───
function renderFeatures() {
    const feats = currentChar.features || [];
    document.getElementById('featuresSection').innerHTML = `
    <div class="d-flex justify-content-between align-items-center">
      <h5>Features & Proficiencies</h5>
      <button class="btn btn-primary btn-sm" onclick="addFeature()"><i class="fa-solid fa-plus me-1"></i>Add Feature</button>
    </div>
    <div class="mt-2">
      ${feats.map((f) => `
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
        </div>`).join('') || '<div class="empty-state"><i class="fa-solid fa-star fa-2x mb-2 d-block text-muted"></i>No features added yet.</div>'}
    </div>`;
}
window.addFeature = function () {
    showModal('Add Feature', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="featName"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="featDesc" rows="3"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Source</label><input class="form-control" id="featSource" placeholder="Class, Race, etc."></div>
      <div class="col-6"><label class="form-label">Level Gained</label><input class="form-control" id="featLevel" type="number" value="1"></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveFeature(this)">Add Feature</button>
  `);
};
window.saveFeature = async function (btn) {
    await api('POST', `/api/characters/${currentChar.id}/features`, {
        name: document.getElementById('featName').value,
        description: document.getElementById('featDesc').value,
        source: document.getElementById('featSource').value,
        level_gained: +document.getElementById('featLevel').value || 1,
    });
    hideModal();
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderFeatures();
    toast('Feature added');
};
window.deleteFeature = async function (id) {
    await api('DELETE', `/api/features/${id}`);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderFeatures();
    toast('Feature removed');
};
// ─── Proficiencies ───
window.addProf = function () {
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
};
window.saveProf = async function (btn) {
    await api('POST', '/api/proficiencies', {
        character_id: currentChar.id,
        type: document.getElementById('profType').value,
        name: document.getElementById('profName').value,
    });
    hideModal();
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSheet();
    toast('Proficiency added');
};
window.deleteProf = async function (id) {
    await api('DELETE', `/api/proficiencies/${id}`);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSheet();
    toast('Proficiency removed');
};
// ─── Details ───
function renderDetails() {
    const c = currentChar;
    const el = document.getElementById('detailsSection');
    el.innerHTML = `
    <div class="row g-3">
      <div class="col-md-4"><label class="form-label">Race</label><input class="form-control form-control-sm" value="${esc(c.race)}" onchange="updateField('race',this.value)"></div>
      <div class="col-md-4"><label class="form-label">Class</label><input class="form-control form-control-sm" value="${esc(c.class)}" onchange="updateField('class',this.value)"></div>
      <div class="col-md-4"><label class="form-label">Subclass</label><input class="form-control form-control-sm" value="${esc(c.subclass)}" onchange="updateField('subclass',this.value)"></div>
    </div>
    <div class="row g-3 mt-1">
      <div class="col-md-4"><label class="form-label">Level</label><input class="form-control form-control-sm" type="number" value="${c.level}" onchange="updateField('level',+this.value)"></div>
      <div class="col-md-4"><label class="form-label">Background</label><input class="form-control form-control-sm" value="${esc(c.background)}" onchange="updateField('background',this.value)"></div>
      <div class="col-md-4"><label class="form-label">Alignment</label><input class="form-control form-control-sm" value="${esc(c.alignment)}" onchange="updateField('alignment',this.value)"></div>
    </div>
    <hr class="my-3">
    ${['personality_traits', 'ideals', 'bonds', 'flaws', 'appearance'].map(f => `
      <div class="mb-3"><label class="form-label">${capitalize(f.replace(/_/g, ' '))}</label>
      <textarea class="form-control form-control-sm" rows="2" onchange="updateField('${f}',this.value)">${esc(c[f])}</textarea></div>
    `).join('')}
    <div class="mb-3"><label class="form-label">Backstory</label>
    <textarea class="form-control form-control-sm" rows="4" onchange="updateField('backstory',this.value)">${esc(c.backstory)}</textarea></div>
    <h5 class="mt-4">Currency</h5>
    <div class="row g-3">
      ${['cp', 'sp', 'ep', 'gp', 'pp'].map(coin => `
        <div class="col-4 col-md-2"><label class="form-label small">${coin.toUpperCase()}</label>
        <input class="form-control form-control-sm" id="coin${coin}" value="${c.currency?.[coin] || 0}" type="number"></div>
      `).join('')}
      <div class="col-4 col-md-2 d-flex align-items-end"><button class="btn btn-gold btn-sm w-100" onclick="updateCurrency()">Save</button></div>
    </div>`;
}
// ─── Locations ───
async function renderLocations() {
    const el = document.getElementById('locationsSection');
    try {
        const links = await api('GET', `/api/characters/${currentChar.id}/locations`);
        el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center"><h5>Linked Locations</h5>
        <button class="btn btn-primary btn-sm" onclick="showLinkLocation()"><i class="fa-solid fa-link me-1"></i>Link Location</button>
      </div>
      <div class="mt-2">${links.length ? links.map((l) => `
        <div class="inv-item">
          <div><span class="fw-bold">${esc(l.location_name)}</span> <span class="text-muted small">(${esc(l.location_type)})</span>
            ${l.notes ? `<br><small class="text-muted">${esc(l.notes)}</small>` : ''}</div>
          <div><span class="badge badge-gold me-1">${esc(l.relationship)}</span>
            <button class="btn btn-sm btn-outline-danger" onclick="unlinkLocation(${l.id})"><i class="fa-solid fa-trash"></i></button></div>
        </div>`).join('')
            : '<div class="empty-state">No locations linked.</div>'}</div>
      <hr class="my-3">
      <div class="d-flex justify-content-between align-items-center"><h5>All Locations</h5>
        <button class="btn btn-outline-primary btn-sm" onclick="showCreateLocation()"><i class="fa-solid fa-plus me-1"></i>New Location</button>
      </div>
      <div class="mt-2">${allLocations.map((l) => `
        <div class="inv-item">
          <div><span class="fw-bold">${esc(l.name)}</span> <span class="text-muted small">(${esc(l.type)})</span>
            <br><small class="text-muted">${esc(l.description).substring(0, 80)}</small></div>
        </div>`).join('')}&nbsp;</div>`;
    }
    catch {
        el.innerHTML = '<div class="empty-state">Could not load locations.</div>';
    }
}
window.showLinkLocation = function () {
    showModal('Link Location', `
    <div class="mb-3"><label class="form-label">Location</label>
      <select class="form-select" id="linkLocId">${allLocations.map((l) => `<option value="${l.id}">${esc(l.name)} (${esc(l.type)})</option>`).join('')}</select></div>
    <div class="mb-3"><label class="form-label">Relationship</label>
      <select class="form-select" id="linkLocRel">
        <option value="current">Current Location</option><option value="hometown">Hometown</option><option value="visited">Visited</option>
        <option value="headquarters">Headquarters</option><option value="quest">Quest Location</option><option value="other">Other</option>
      </select></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="linkLocNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveLinkLocation()"><i class="fa-solid fa-link me-1"></i>Link</button>
  `);
};
window.saveLinkLocation = async function () {
    await api('POST', `/api/characters/${currentChar.id}/locations`, {
        location_id: +document.getElementById('linkLocId').value,
        relationship: document.getElementById('linkLocRel').value,
        notes: document.getElementById('linkLocNotes').value,
    });
    hideModal();
    renderLocations();
    toast('Location linked');
};
window.unlinkLocation = async function (id) {
    await api('DELETE', `/api/locations/link/${id}`);
    renderLocations();
};
window.showCreateLocation = function () {
    showModal('New Location', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="newLocName"></div>
    <div class="mb-3"><label class="form-label">Type</label>
      <select class="form-select" id="newLocType">
        <option value="region">Region</option><option value="city">City</option><option value="town">Town</option>
        <option value="dungeon">Dungeon</option><option value="tavern">Tavern</option><option value="temple">Temple</option>
        <option value="shop">Shop</option><option value="wilderness">Wilderness</option><option value="other">Other</option>
      </select></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newLocDesc" rows="3"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveNewLocation()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
};
window.saveNewLocation = async function () {
    await api('POST', '/api/locations', {
        name: document.getElementById('newLocName').value,
        type: document.getElementById('newLocType').value,
        description: document.getElementById('newLocDesc').value,
    });
    hideModal();
    allLocations = await api('GET', '/api/locations');
    renderLocations();
    toast('Location created');
};
// ─── NPCs ───
async function renderNPCs() {
    const el = document.getElementById('npcsSection');
    try {
        const links = await api('GET', `/api/characters/${currentChar.id}/npcs`);
        el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center"><h5>Related NPCs</h5>
        <button class="btn btn-primary btn-sm" onclick="showLinkNPC()"><i class="fa-solid fa-link me-1"></i>Link NPC</button>
      </div>
      <div class="mt-2">${links.length ? links.map((n) => `
        <div class="inv-item">
          <div><span class="fw-bold">${esc(n.npc_name)}</span>
            <span class="text-muted small">${esc(n.npc_race)} ${esc(n.npc_class)}</span>
            ${!n.npc_is_alive ? '<span class="badge badge-blood ms-1">Deceased</span>' : ''}</div>
          <div>
            <span class="badge badge-gold">${esc(n.relationship)}</span>
            ${n.interaction_count > 0 ? `<span class="badge badge-blood ms-1">${n.interaction_count} talks</span>` : ''}
            <button class="btn btn-sm btn-outline-primary" onclick="logNPCInteraction(${n.id})"><i class="fa-solid fa-comment"></i></button>
            <button class="btn btn-sm btn-outline-danger" onclick="unlinkNPC(${n.id})"><i class="fa-solid fa-trash"></i></button>
          </div>
        </div>`).join('')
            : '<div class="empty-state">No NPCs linked yet.</div>'}</div>
      <hr class="my-3">
      <div class="d-flex justify-content-between align-items-center"><h5>All NPCs</h5>
        <button class="btn btn-outline-primary btn-sm" onclick="showCreateNPC()"><i class="fa-solid fa-plus me-1"></i>New NPC</button>
      </div>
      <div class="mt-2">${allNPCs.map((n) => `
        <div class="inv-item">
          <div><span class="fw-bold">${esc(n.name)}</span>
            <span class="text-muted small">${esc(n.race)} ${esc(n.class)}</span></div>
          <div class="text-muted small">HP: ${n.hp_current}/${n.hp_max}</div>
        </div>`).join('')}&nbsp;</div>`;
    }
    catch {
        el.innerHTML = '<div class="empty-state">Could not load NPCs.</div>';
    }
}
window.showLinkNPC = function () {
    showModal('Link NPC', `
    <div class="mb-3"><label class="form-label">NPC</label>
      <select class="form-select" id="linkNPCId">${allNPCs.map((n) => `<option value="${n.id}">${esc(n.name)} (${esc(n.race)} ${esc(n.class)})</option>`).join('')}</select></div>
    <div class="mb-3"><label class="form-label">Relationship</label>
      <select class="form-select" id="linkNPCRel">
        <option value="ally">Ally</option><option value="enemy">Enemy</option><option value="family">Family</option>
        <option value="contact">Contact</option><option value="acquaintance">Acquaintance</option>
        <option value="pet">Pet/Mount</option><option value="deity">Deity/Patron</option><option value="other">Other</option>
      </select></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="linkNPCNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveLinkNPC()"><i class="fa-solid fa-link me-1"></i>Link</button>
  `);
};
window.saveLinkNPC = async function () {
    await api('POST', `/api/characters/${currentChar.id}/npcs`, {
        npc_id: +document.getElementById('linkNPCId').value,
        relationship: document.getElementById('linkNPCRel').value,
        notes: document.getElementById('linkNPCNotes').value,
    });
    hideModal();
    renderNPCs();
    toast('NPC linked');
};
window.logNPCInteraction = async function (id) {
    await api('POST', `/api/npcs/link/${id}/interact`, {});
    renderNPCs();
    toast('Interaction logged');
};
window.unlinkNPC = async function (id) {
    await api('DELETE', `/api/npcs/link/${id}`);
    renderNPCs();
};
window.showCreateNPC = function () {
    showModal('New NPC', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="newNPCName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Race</label><input class="form-control" id="newNPCRace"></div>
      <div class="col-6"><label class="form-label">Class</label><input class="form-control" id="newNPCClass"></div>
    </div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newNPCDesc" rows="3"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveNewNPC()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
};
window.saveNewNPC = async function () {
    await api('POST', '/api/npcs', {
        name: document.getElementById('newNPCName').value,
        race: document.getElementById('newNPCRace').value,
        class: document.getElementById('newNPCClass').value,
        description: document.getElementById('newNPCDesc').value,
    });
    hideModal();
    allNPCs = await api('GET', '/api/npcs');
    renderNPCs();
    toast('NPC created');
};
// ─── Sessions ───
async function renderSessions() {
    const el = document.getElementById('sessionsSection');
    try {
        const sessions = await api('GET', `/api/characters/${currentChar.id}/sessions`);
        el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center"><h5>Session Log</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddSession()"><i class="fa-solid fa-plus me-1"></i>Log Session</button>
      </div>
      <div class="mt-3">
        ${sessions.map((s) => `
          <div class="card mb-2">
            <div class="card-body py-2 px-3">
              <div class="d-flex justify-content-between align-items-start">
                <div><span class="fw-bold">${esc(s.title) || 'Session'}</span>
                  <span class="badge badge-gold ms-2">${s.session_date}</span>
                  ${s.xp_earned > 0 ? `<span class="badge badge-blood ms-1">+${s.xp_earned} XP</span>` : ''}
                  ${s.gold_earned > 0 ? `<span class="badge badge-gold ms-1">+${s.gold_earned} GP</span>` : ''}</div>
                <button class="btn btn-sm btn-outline-danger" onclick="deleteSession(${s.id})"><i class="fa-solid fa-trash"></i></button>
              </div>
              <p class="mb-0 mt-1 small text-muted">${esc(s.notes).substring(0, 200)}</p>
              ${s.important_events ? `<p class="mb-0 mt-1 small fst-italic text-muted">${esc(s.important_events).substring(0, 150)}</p>` : ''}
            </div>
          </div>`).join('') || '<div class="empty-state"><i class="fa-solid fa-calendar fa-2x mb-2 d-block text-muted"></i>No sessions logged yet.</div>'}
      </div>`;
    }
    catch {
        el.innerHTML = '<div class="empty-state">Could not load sessions.</div>';
    }
}
window.showAddSession = function () {
    showModal('Log Session', `
    <div class="mb-3"><label class="form-label">Date</label><input class="form-control" id="sessDate" type="date" value="${new Date().toISOString().split('T')[0]}"></div>
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="sessTitle" placeholder="Session 1: The Adventure Begins"></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="sessNotes" rows="3" placeholder="What happened?"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">XP Earned</label><input class="form-control" id="sessXP" type="number" value="0"></div>
      <div class="col-6"><label class="form-label">Gold Earned</label><input class="form-control" id="sessGold" type="number" value="0"></div>
    </div>
    <div class="mb-3"><label class="form-label">Important Events</label><textarea class="form-control" id="sessEvents" rows="2" placeholder="Key moments, NPCs met, revelations..."></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveSession()"><i class="fa-solid fa-save me-1"></i>Log Session</button>
  `);
};
window.saveSession = async function () {
    await api('POST', `/api/characters/${currentChar.id}/sessions`, {
        session_date: document.getElementById('sessDate').value,
        title: document.getElementById('sessTitle').value,
        notes: document.getElementById('sessNotes').value,
        xp_earned: +document.getElementById('sessXP').value || 0,
        gold_earned: +document.getElementById('sessGold').value || 0,
        important_events: document.getElementById('sessEvents').value,
    });
    hideModal();
    renderSessions();
    toast('Session logged');
};
window.deleteSession = async function (id) {
    if (!confirm('Delete this session?'))
        return;
    await api('DELETE', `/api/sessions/${id}`);
    renderSessions();
    toast('Session deleted');
};
// ─── Quests ───
async function renderQuests() {
    const el = document.getElementById('questsSection');
    try {
        const quests = await api('GET', `/api/characters/${currentChar.id}/quests`);
        const groups = { active: [], available: [], complete: [], failed: [], abandoned: [] };
        quests.forEach((q) => { if (groups[q.status])
            groups[q.status].push(q); });
        let html = '<div class="d-flex justify-content-between align-items-center"><h5>Quests</h5><button class="btn btn-primary btn-sm" onclick="showAddQuest()"><i class="fa-solid fa-plus me-1"></i>New Quest</button></div>';
        const labels = { active: 'Active', available: 'Available', complete: 'Complete', failed: 'Failed', abandoned: 'Abandoned' };
        for (const st of ['active', 'available', 'complete', 'failed', 'abandoned']) {
            const qs = groups[st] || [];
            if (!qs.length)
                continue;
            html += `<h6 class="mt-3 text-muted">${labels[st]}</h6>`;
            for (const q of qs) {
                const opts = ['active', 'available', 'complete', 'failed', 'abandoned'].map(s => `<option value="${s}"${s === q.status ? ' selected' : ''}>${capitalize(s)}</option>`).join('');
                html += `<div class="card mb-2">
          <div class="card-body py-2 px-3">
            <div class="d-flex justify-content-between align-items-start">
              <div><span class="fw-bold">${esc(q.name)}</span></div>
              <div class="d-flex gap-1">
                <select class="form-select form-select-sm" style="width:auto" onchange="updateQuestStatus(${q.id},this.value)">${opts}</select>
                <button class="btn btn-sm btn-outline-danger" onclick="deleteQuest(${q.id})"><i class="fa-solid fa-trash"></i></button>
              </div>
            </div>
            <p class="mb-0 mt-1 small text-muted">${esc(q.description).substring(0, 200)}</p>
            ${q.objectives ? `<div class="mt-1 small text-muted"><strong>Objectives:</strong> ${esc(q.objectives).substring(0, 150)}</div>` : ''}
            ${q.rewards ? `<div class="mt-1 small text-success"><strong>Reward:</strong> ${esc(q.rewards).substring(0, 150)}</div>` : ''}
          </div>
        </div>`;
            }
        }
        if (quests.length === 0)
            html += '<div class="empty-state"><i class="fa-solid fa-scroll fa-2x mb-2 d-block text-muted"></i>No quests yet.</div>';
        el.innerHTML = html;
    }
    catch {
        el.innerHTML = '<div class="empty-state">Could not load quests.</div>';
    }
}
window.showAddQuest = function () {
    showModal('New Quest', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="questName" placeholder="e.g. Retrieve the Lost Artifact"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="questDesc" rows="3"></textarea></div>
    <div class="mb-3"><label class="form-label">Objectives</label><textarea class="form-control" id="questObj" rows="2" placeholder="1. Travel to the Temple\n2. Defeat the guardian\n3. Retrieve the artifact"></textarea></div>
    <div class="mb-3"><label class="form-label">Rewards</label><textarea class="form-control" id="questRewards" rows="2" placeholder="500 XP, +1 Longsword, 200 GP"></textarea></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="questNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveQuest()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
};
window.saveQuest = async function () {
    await api('POST', `/api/characters/${currentChar.id}/quests`, {
        name: document.getElementById('questName').value,
        description: document.getElementById('questDesc').value,
        objectives: document.getElementById('questObj').value,
        rewards: document.getElementById('questRewards').value,
        notes: document.getElementById('questNotes').value,
    });
    hideModal();
    renderQuests();
    toast('Quest created');
};
window.updateQuestStatus = async function (id, status) {
    const quests = await api('GET', `/api/characters/${currentChar.id}/quests`);
    const q = quests.find((x) => x.id === id);
    if (!q)
        return;
    q.status = status;
    await api('PUT', `/api/quests/${id}`, q);
    renderQuests();
    toast('Quest status updated');
};
window.deleteQuest = async function (id) {
    if (!confirm('Delete this quest?'))
        return;
    await api('DELETE', `/api/quests/${id}`);
    renderQuests();
    toast('Quest deleted');
};
// ─── Journal ───
async function renderJournal() {
    const el = document.getElementById('journalSection');
    try {
        const entries = await api('GET', `/api/characters/${currentChar.id}/journal`);
        el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center"><h5>Character Journal</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddJournal()"><i class="fa-solid fa-plus me-1"></i>Write Entry</button>
      </div>
      <div class="mt-3">
        ${entries.map((j) => `
          <div class="card mb-2">
            <div class="card-body py-2 px-3">
              <div class="d-flex justify-content-between align-items-start">
                <div><span class="fw-bold">${esc(j.title) || 'Untitled'}</span>
                  <span class="badge badge-gold ms-2">${j.entry_date}</span></div>
                <button class="btn btn-sm btn-outline-danger" onclick="deleteJournal(${j.id})"><i class="fa-solid fa-trash"></i></button>
              </div>
              <div class="mt-2 small text-muted" style="white-space:pre-wrap">${esc(j.entry)}</div>
            </div>
          </div>`).join('') || '<div class="empty-state"><i class="fa-solid fa-book-open fa-2x mb-2 d-block text-muted"></i>No journal entries yet.</div>'}
      </div>`;
    }
    catch {
        el.innerHTML = '<div class="empty-state">Could not load journal.</div>';
    }
}
window.showAddJournal = function () {
    showModal('Journal Entry', `
    <div class="mb-3"><label class="form-label">Date</label><input class="form-control" id="journalDate" type="date" value="${new Date().toISOString().split('T')[0]}"></div>
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="journalTitle" placeholder="Day 1: Arrival in Waterdeep"></div>
    <div class="mb-3"><label class="form-label">Entry</label><textarea class="form-control" id="journalEntry" rows="6" placeholder="Write your character's thoughts..."></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveJournal()"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
};
window.saveJournal = async function () {
    await api('POST', `/api/characters/${currentChar.id}/journal`, {
        entry_date: document.getElementById('journalDate').value,
        title: document.getElementById('journalTitle').value,
        entry: document.getElementById('journalEntry').value,
    });
    hideModal();
    renderJournal();
    toast('Journal entry saved');
};
window.deleteJournal = async function (id) {
    if (!confirm('Delete this journal entry?'))
        return;
    await api('DELETE', `/api/journal/${id}`);
    renderJournal();
    toast('Journal entry deleted');
};
// ─── Graph ───
async function renderGraph() {
    const el = document.getElementById('graphSection');
    el.innerHTML = `<div class="ornament mb-3">✧ Drawing your web of fate ✧</div>
    <div id="graphContainer" style="width:100%;height:600px;border:1px solid var(--border);border-radius:4px;background:var(--parchment-light)"></div>`;
    try {
        const data = await api('GET', `/api/characters/${currentChar.id}/graph`);
        if (typeof vis !== 'undefined') {
            const container = document.getElementById('graphContainer');
            const nodes = new vis.DataSet(data.nodes.map((n) => ({
                id: n.id, label: n.label, group: n.group,
                color: { background: n.color, border: '#2c1810' },
                font: { face: 'Playfair Display', color: '#2c1810', size: n.size > 20 ? 14 : 11 },
                size: n.size,
                borderWidth: 2,
            })));
            const edges = new vis.DataSet(data.edges.map((e) => ({
                from: e.from, to: e.to, label: e.label,
                dashes: e.dashes, width: e.width,
                color: { color: '#8b7355', highlight: '#8b0000' },
                font: { face: 'Vollkorn', size: 10, color: '#5c3a2a', align: 'middle' },
                smooth: { type: 'curvedCW', roundness: 0.15 },
            })));
            new vis.Network(container, { nodes, edges }, {
                physics: { solver: 'forceAtlas2Based', forceAtlas2Based: { gravitationalConstant: -80, centralGravity: 0.005, springLength: 200, springConstant: 0.02, damping: 0.4 }, stabilization: { iterations: 100 } },
                interaction: { hover: true, tooltipDelay: 200, navigationButtons: true, keyboard: true },
                groups: {
                    character: { shape: 'ellipse', color: { background: '#8b0000', border: '#5c0000' }, font: { color: '#fff', size: 16 } },
                    location: { shape: 'square', color: { background: '#b8963e', border: '#8a7020' } },
                    npc: { shape: 'diamond', color: { background: '#2d6a2d', border: '#1a4a1a' } },
                    quest: { shape: 'star', color: { background: '#8b4513', border: '#5c2e0d' } },
                    session: { shape: 'dot', color: { background: '#5c3a2a', border: '#3c2010' } },
                },
                edges: { smooth: true },
            });
        }
        else {
            el.innerHTML += `<div class="p-3 small">
        <h5>Character Web</h5>
        <p>${data.nodes.map((n) => `${n.label} [${n.group}]`).join(' &rarr; ')}</p>
        <p class="text-muted fst-italic mt-2">${data.nodes.length} connections &middot; ${data.edges.length} relationships</p></div>`;
        }
    }
    catch (e) {
        el.innerHTML += `<div class="empty-state">Could not load graph: ${esc(e.message)}</div>`;
    }
}
// ─── Analytics ───
async function renderAnalytics() {
    const el = document.getElementById('analyticsSection');
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
            ? stats.top_npcs.map((n) => `<p class="mb-1 small text-muted">&loz; ${esc(n)}</p>`).join('')
            : '<p class="mb-0 small text-muted fst-italic">No NPC interactions yet</p>'}
            </div>
          </div>
        </div>
      </div>
      <div id="questChartContainer" style="height:200px;max-width:400px;margin:0 auto"></div>`;
        if ((typeof Chart !== 'undefined') && stats.quests.total > 0) {
            const ctx = document.createElement('canvas');
            document.getElementById('questChartContainer').appendChild(ctx);
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
    }
    catch (e) {
        el.innerHTML = `<div class="empty-state">Could not load analytics: ${esc(e.message)}</div>`;
    }
}
// ─── Dice ───
function renderDiceTab() {
    const targetId = currentView === 'dice' ? 'diceViewSection' : 'diceSection';
    const el = document.getElementById(targetId);
    if (!el) return;
    el.innerHTML = `
    <div class="text-center">
      <h5>Dice Roller</h5>
      <div class="row justify-content-center mb-3">
        <div class="col-md-6"><label class="form-label">Expression (e.g. 2d6+3)</label>
          <input class="form-control text-center" id="diceExpr" value="1d20" placeholder="e.g. 1d20+5" style="font-size:1.2rem">
        </div>
      </div>
      <div id="diceResult" class="mb-3" style="display:none"></div>
      <button class="btn btn-gold" onclick="doRoll()"><i class="fa-solid fa-dice me-2"></i>Roll the Bones</button>
      <div class="ornament my-3">✧</div>
      <h5>Recent Rolls</h5>
      <div id="diceHistory"></div>
    </div>`;
    loadDiceHistory();
}
async function doRoll() {
    const expr = document.getElementById('diceExpr').value;
    if (!expr)
        return;
    try {
        const result = await api('POST', '/api/roll', { expression: expr, character_id: currentChar?.id });
        const el = document.getElementById('diceResult');
        el.style.display = 'block';
        el.innerHTML = `<div class="text-muted">${esc(expr)}</div><div class="h4">${esc(result.text)}</div>`;
        loadDiceHistory();
    }
    catch (e) {
        toast(e.message, true);
    }
}
window.doRoll = doRoll;
async function loadDiceHistory() {
    const el = document.getElementById('diceHistory');
    if (!el)
        return;
    try {
        const rolls = await api('GET', '/api/dice-rolls' + (currentChar ? `?character_id=${currentChar.id}` : ''));
        el.innerHTML = rolls.slice(0, 20).map((r) => `<div class="d-flex justify-content-between py-1 border-bottom dice-history-item">
        <span>${esc(r.expression)}</span>
        <span><strong>${r.total}</strong> <span class="text-muted small">${esc(r.result)}</span></span>
      </div>`).join('') || '<div class="text-center text-muted py-3">No rolls yet</div>';
    }
    catch { }
}
window.loadDiceHistory = loadDiceHistory;
// ─── New Character ───
window.newChar = function () {
    showModal('New Character', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="newName" placeholder="Character name"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Race</label><input class="form-control" id="newRace" list="raceSuggestions"><datalist id="raceSuggestions"></datalist></div>
      <div class="col-6"><label class="form-label">Class</label><input class="form-control" id="newClass" list="classSuggestions"><datalist id="classSuggestions"></datalist></div>
    </div>
    <button class="btn btn-primary w-100" onclick="createChar()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
    fetch('/api/compendium/races', { credentials: 'include' }).then(r => r.json()).then((races) => {
        document.getElementById('raceSuggestions').innerHTML = races.map((r) => `<option value="${esc(r.name)}">`).join('');
    }).catch(() => { });
    fetch('/api/compendium/classes', { credentials: 'include' }).then(r => r.json()).then((cls) => {
        document.getElementById('classSuggestions').innerHTML = cls.map((c) => `<option value="${esc(c.name)}">`).join('');
    }).catch(() => { });
};
window.createChar = async function () {
    const name = document.getElementById('newName').value || 'Unnamed';
    const race = document.getElementById('newRace').value;
    const cls = document.getElementById('newClass').value;
    try {
        const char = await api('POST', '/api/characters', { name, race, class: cls });
        hideModal();
        if (char.id)
            await openChar(char.id);
        loadCharacters();
    }
    catch (e) {
        toast(e.message, true);
    }
};
// ─── Import / Export ───
window.showImport = function () {
    showModal('Import Character', `
    <p class="text-muted fst-italic small mb-3">Paste JSON or upload a file</p>
    <div class="mb-3"><label class="form-label">JSON</label><textarea class="form-control" id="importJson" rows="6" style="font-family:monospace;font-size:0.8rem"></textarea></div>
    <div class="mb-3"><label class="form-label">File</label><input class="form-control" type="file" id="importFile" accept=".json"></div>
    <button class="btn btn-primary w-100" onclick="doImport()"><i class="fa-solid fa-file-import me-1"></i>Import</button>
  `);
};
window.doImport = async function () {
    const jsonEl = document.getElementById('importJson');
    const fileEl = document.getElementById('importFile');
    try {
        let result;
        if (fileEl.files && fileEl.files[0]) {
            const form = new FormData();
            form.append('file', fileEl.files[0]);
            const res = await fetch('/api/characters/import', { method: 'POST', headers: { 'X-CSRF-Token': csrfToken }, credentials: 'include', body: form });
            result = await res.json();
        }
        else if (jsonEl.value.trim()) {
            result = await api('POST', '/api/characters/import', JSON.parse(jsonEl.value));
        }
        else {
            toast('Provide JSON or a file', true);
            return;
        }
        toast(`Imported ${Array.isArray(result) ? result.length : 1} character(s)`);
        hideModal();
        loadCharacters();
    }
    catch (e) {
        toast('Import failed: ' + e.message, true);
    }
};
window.exportChar = async function () {
    if (!currentChar)
        return;
    try {
        const data = await api('GET', `/api/characters/${currentChar.id}/export`);
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const a = document.createElement('a');
        const url = URL.createObjectURL(blob);
        a.href = url;
        a.download = currentChar.name.replace(/[^a-zA-Z0-9]/g, '_') + '.json';
        a.click();
        URL.revokeObjectURL(url);
    }
    catch (e) {
        toast(e.message, true);
    }
};
// ─── Print ───
window.printChar = async function () {
    if (!currentChar)
        return;
    try {
        const res = await fetch(`/api/characters/${currentChar.id}/print`, {
            headers: { 'X-CSRF-Token': csrfToken }, credentials: 'include',
        });
        const text = await res.text();
        const win = window.open('', '_blank');
        if (win) {
            win.document.write(`<pre style="font-family:monospace;font-size:12px;line-height:1.4">${esc(text)}</pre>`);
            win.document.close();
            win.print();
        }
    }
    catch (e) {
        toast(e.message, true);
    }
};
// ─── Party View ───
window.showParty = async function () {
    showView('party');
    const el = document.getElementById('partyContent');
    el.innerHTML = '<div class="ornament mb-3">✧ Assembling the party... ✧</div>';
    try {
        const groups = await api('GET', '/api/party');
        el.innerHTML = groups.map((g) => `
      <div class="card mb-3">
        <div class="card-header d-flex justify-content-between align-items-center">
          <strong>${esc(g.name || 'Unnamed Campaign')}</strong>
          <span class="badge badge-gold">${g.members.length} members</span>
        </div>
        <div class="card-body">
          <div class="row g-3">
            ${g.members.map((m) => {
            const pct = m.hp_max > 0 ? Math.round((m.hp_current / m.hp_max) * 100) : 0;
            const sc = m.status === 'down' ? 'var(--danger)' : m.status === 'injured' ? 'var(--gold)' : 'var(--success)';
            return `<div class="col-md-6 col-lg-4">
                <div class="character-card" onclick="openChar(${m.id})">
                  <div class="char-name">${esc(m.name)}</div>
                  <div class="char-detail">${esc(m.race)} ${esc(m.class)} · Level ${m.level}</div>
                  <div class="d-flex gap-3 mt-1 small text-muted">
                    <span>AC: ${m.ac}</span><span style="color:${sc}">${esc(m.status)}</span>
                  </div>
                  <div class="hp-bar position-relative mt-2" style="height:12px">
                    <div class="hp-bar-fill" style="width:${pct}%;height:100%"></div>
                    <div class="position-absolute top-0 start-0 end-0 bottom-0 d-flex align-items-center justify-content-center text-white" style="font-size:0.65rem">${m.hp_current}/${m.hp_max}</div>
                  </div>
                </div>
              </div>`;
        }).join('')}
          </div>
        </div>
      </div>
    `).join('') || '<div class="empty-state"><i class="fa-solid fa-flag fa-2x mb-2 d-block text-muted"></i>No characters yet.</div>';
    }
    catch (e) {
        el.innerHTML = `<div class="empty-state">Failed: ${esc(e.message)}</div>`;
    }
};
// ─── Compendium ───
window.showCompendium = function () {
    showView('compendium');
    loadCompendiumRaces();
};
window.loadCompendiumTab = function (tab) {
    document.getElementById('compTabRaces').classList.remove('active');
    document.getElementById('compTabClasses').classList.remove('active');
    document.getElementById('compTabSpells').classList.remove('active');
    document.getElementById('compTabEquipment').classList.remove('active');
    const tabEl = document.getElementById('compTab' + capitalize(tab));
    if (tabEl)
        tabEl.classList.add('active');
    ['races', 'classes', 'spells', 'equipment'].forEach(s => {
        const el = document.getElementById('comp' + capitalize(s));
        if (el)
            el.style.display = s === tab ? 'block' : 'none';
    });
    if (tab === 'races')
        loadCompendiumRaces();
    if (tab === 'classes')
        loadCompendiumClasses();
    if (tab === 'spells')
        loadCompendiumSpells();
    if (tab === 'equipment')
        loadCompendiumEquipment();
};
async function loadCompendiumRaces() {
    try {
        const races = await api('GET', '/api/compendium/races');
        document.getElementById('compRaces').innerHTML = races.map((r) => `
      <div class="card mb-2">
        <div class="card-body py-2 px-3">
          <div class="d-flex justify-content-between"><strong>${esc(r.name)}</strong>
            <span><span class="badge badge-gold">${esc(r.size)}</span> Speed: ${r.speed}</span></div>
          <p class="mb-0 mt-1 small text-muted">${esc(r.description)}</p>
        </div>
      </div>`).join('');
    }
    catch { }
}
async function loadCompendiumClasses() {
    try {
        const cls = await api('GET', '/api/compendium/classes');
        document.getElementById('compClasses').innerHTML = cls.map((c) => `
      <div class="card mb-2">
        <div class="card-body py-2 px-3">
          <div class="d-flex justify-content-between"><strong>${esc(c.name)}</strong>
            <span class="text-muted small">d${c.hit_die} · ${esc(c.primary_ability)}</span></div>
          <p class="mb-0 mt-1 small text-muted">${esc(c.description)}</p>
        </div>
      </div>`).join('');
    }
    catch { }
}
async function loadCompendiumSpells() {
    try {
        const spells = await api('GET', '/api/compendium/spells');
        document.getElementById('compSpells').innerHTML = spells.map((s) => `
      <div class="inv-item">
        <div><span class="fw-bold">${esc(s.name)}</span> <span class="text-muted small">Lv${s.level} ${esc(s.school)}</span></div>
        <div class="text-muted small">${esc(s.casting_time)} · ${esc(s.range)} · ${esc(s.duration)}</div>
      </div>`).join('');
    }
    catch { }
}
async function loadCompendiumEquipment() {
    try {
        const items = await api('GET', '/api/compendium/equipment');
        document.getElementById('compEquipment').innerHTML = items.map((i) => `
      <div class="inv-item">
        <span class="fw-bold">${esc(i.name)}</span>
        <span class="text-muted small">${esc(i.category)}${i.weight ? ' · ' + i.weight + 'lb' : ''}</span>
      </div>`).join('');
    }
    catch { }
}
// ─── Delete Character ───
window.deleteChar = async function () {
    if (!currentChar) return;
    if (!confirm('Delete this character?')) return;
    try {
        await api('DELETE', `/api/characters/${currentChar.id}`);
        currentChar = null;
        showView('characters');
        loadCharacters();
        toast('Character deleted');
    } catch (e) {
        toast(e.message, true);
    }
};
// ─── Logout ───
window.logout = async function () {
    await api('POST', '/api/logout');
    window.location.href = '/login';
};
init();

//# sourceMappingURL=app.js.map