let csrfToken = '';
let currentUser = null;
let currentView = 'characters';
let currentChar = null;
let currentTab = 'stats';
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
async function init() {
    try {
        const user = await api('GET', '/api/user/me');
        currentUser = user;
        const tokenRes = await api('GET', '/api/csrf-token');
        csrfToken = tokenRes.token;
        document.getElementById('userName').textContent = user.username;
        if (user.role === 'admin') {
            document.getElementById('adminLink').style.display = 'inline';
        }
        showView('characters');
        loadCharacters();
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
}
// ─── Character List ───
async function loadCharacters() {
    try {
        const chars = await api('GET', '/api/characters');
        const grid = document.getElementById('charGrid');
        grid.innerHTML = chars.map((c) => `
      <div class="character-card" onclick="openChar(${c.id})">
        <div class="char-name">${esc(c.name)}</div>
        <div class="char-detail">${esc(c.race)} ${esc(c.class)} · Level ${c.level}</div>
        <div class="char-hp">HP: ${c.hp_current}/${c.hp_max}</div>
      </div>
    `).join('');
    }
    catch (e) {
        toast(e.message, true);
    }
}
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
function renderSheet() {
    if (!currentChar)
        return;
    const c = currentChar;
    document.getElementById('sheetName').textContent = c.name;
    document.getElementById('sheetSubtitle').textContent =
        `${c.race} ${c.class}${c.subclass ? ' (' + c.subclass + ')' : ''} · Level ${c.level}`;
    // Tabs
    const sections = ['stats', 'combat', 'spells', 'inventory', 'features', 'details', 'dice'];
    const tabBar = document.getElementById('tabBar');
    tabBar.innerHTML = sections.map(s => `
    <button class="tab ${s === currentTab ? 'active' : ''}" onclick="switchTab('${s}')">${capitalize(s)}</button>
  `).join('');
    sections.forEach(s => {
        const el = document.getElementById(sectionId(s));
        el.style.display = s === currentTab ? 'block' : 'none';
    });
    renderStats();
    renderCombat();
    renderSpells();
    renderInventory();
    renderFeatures();
    renderDetails();
    renderDiceTab();
}
function sectionId(s) { return s + 'Section'; }
function switchTab(tab) {
    currentTab = tab;
    renderSheet();
}
window.switchTab = switchTab;
function renderStats() {
    const c = currentChar;
    const el = document.getElementById('statsSection');
    const abils = [
        { key: 'str', label: 'STR' },
        { key: 'dex', label: 'DEX' },
        { key: 'con', label: 'CON' },
        { key: 'int', label: 'INT' },
        { key: 'wis', label: 'WIS' },
        { key: 'cha', label: 'CHA' },
    ];
    el.innerHTML = `
    <div class="ability-grid">
      ${abils.map(a => {
        const val = c[a.key];
        const mod = Math.floor((val - 10) / 2);
        const modClass = mod > 0 ? 'mod-pos' : mod < 0 ? 'mod-neg' : 'mod-zero';
        return `<div class="ability-score">
          <div class="label">${a.label}</div>
          <div class="value">${val}</div>
          <div class="mod ${modClass}">${mod >= 0 ? '+' : ''}${mod}</div>
        </div>`;
    }).join('')}
    </div>
    <div class="ornament">✧</div>
    <div class="form-row" style="margin-top:16px">
      <div class="form-group">
        <label>Proficiency Bonus</label>
        <input type="number" value="${c.proficiency_bonus}" onchange="updateField('proficiency_bonus', +this.value)">
      </div>
      <div class="form-group">
        <label>Inspiration</label>
        <input type="number" value="${c.inspiration}" onchange="updateField('inspiration', +this.value)">
      </div>
      <div class="form-group">
        <label>Passive Perception</label>
        <input type="number" value="${c.passive_perception}" onchange="updateField('passive_perception', +this.value)">
      </div>
      <div class="form-group">
        <label>XP</label>
        <input type="number" value="${c.xp}" onchange="updateField('xp', +this.value)">
      </div>
    </div>
    <h3>Proficiencies</h3>
    <div id="profsArea">
      ${(c.proficiencies || []).map((p) => `<span class="badge badge-blood" style="margin:2px">${esc(p.name)} (${p.type}) <a href="#" onclick="deleteProf(${p.id});return false" style="color:white;text-decoration:none">×</a></span>`).join('')}
    </div>
    <button class="btn btn-sm" style="margin-top:8px" onclick="addProf()">+ Add Proficiency</button>
  `;
}
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
function renderCombat() {
    const c = currentChar;
    const el = document.getElementById('combatSection');
    const hpPct = c.hp_max > 0 ? Math.round((c.hp_current / c.hp_max) * 100) : 0;
    el.innerHTML = `
    <div class="combat-grid">
      <div class="combat-stat"><div class="label">AC</div><div class="value">${c.ac}</div></div>
      <div class="combat-stat"><div class="label">Initiative</div><div class="value">${c.initiative >= 0 ? '+' : ''}${c.initiative}</div></div>
      <div class="combat-stat"><div class="label">Speed</div><div class="value">${c.speed}</div></div>
    </div>
    <h3>Hit Points</h3>
    <div class="hp-bar">
      <div class="hp-bar-fill" style="width:${hpPct}%"></div>
      <div class="hp-bar-text">${c.hp_current} / ${c.hp_max}${c.temp_hp > 0 ? ' (+' + c.temp_hp + ' temp)' : ''}</div>
    </div>
    <div class="form-row" style="margin-top:12px">
      <div class="form-group"><label>HP Max</label><input type="number" value="${c.hp_max}" onchange="updateField('hp_max', +this.value)"></div>
      <div class="form-group"><label>HP Current</label><input type="number" value="${c.hp_current}" onchange="updateField('hp_current', +this.value)"></div>
      <div class="form-group"><label>Temp HP</label><input type="number" value="${c.temp_hp}" onchange="updateField('temp_hp', +this.value)"></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>Hit Dice</label><input value="${c.hit_dice}" onchange="updateField('hit_dice', this.value)"></div>
      <div class="form-group"><label>Remaining</label><input type="number" value="${c.hit_dice_current}" onchange="updateField('hit_dice_current', +this.value)"></div>
    </div>
    <h3>Weapons</h3>
    <div id="weaponArea">
      ${(c.inventory || []).filter((i) => i.category === 'weapon').map((w) => `
        <div class="inventory-item ${w.is_equipped ? 'equipped' : ''}">
          <span class="item-name">${esc(w.name)}</span>
          <span>${w.damage_dice} ${w.damage_type} ${w.is_equipped ? '[E]' : ''}</span>
        </div>
      `).join('') || '<div class="empty-state" style="padding:12px">No weapons. Add some in Inventory.</div>'}
    </div>
  `;
}
function renderSpells() {
    const c = currentChar;
    const el = document.getElementById('spellsSection');
    const sc = c.spellcasting;
    const levels = Array.from({ length: 10 }, (_, i) => i).reverse();
    const byLevel = { 0: [], 1: [], 2: [], 3: [], 4: [], 5: [], 6: [], 7: [], 8: [], 9: [] };
    (c.spells || []).forEach((s) => { if (byLevel[s.level])
        byLevel[s.level].push(s); });
    let html = '';
    if (sc && sc.ability) {
        html += `
      <div class="form-row">
        <div class="form-group"><label>Spellcasting Ability</label><input value="${sc.ability}" onchange="updateSpellcasting('ability', this.value)"></div>
        <div class="form-group"><label>Save DC</label><input type="number" value="${sc.save_dc}" onchange="updateSpellcasting('save_dc', +this.value)"></div>
        <div class="form-group"><label>Attack Bonus</label><input type="number" value="${sc.attack_bonus}" onchange="updateSpellcasting('attack_bonus', +this.value)"></div>
      </div>
      <h3>Spell Slots</h3>
      <div class="slot-tracker">`;
        for (let lv = 1; lv <= 9; lv++) {
            const maxKey = `slots_${lv}_max`;
            const usedKey = `slots_${lv}_used`;
            if (sc[maxKey] > 0) {
                html += `<div class="slot-level">
          <div class="level-label">Lv ${lv}</div>
          <div class="slot-dots">`;
                for (let i = 0; i < sc[maxKey]; i++) {
                    const filled = i < sc[usedKey];
                    html += `<div class="slot-dot ${filled ? 'filled' : 'empty'}" onclick="toggleSlot(${lv}, ${i})"></div>`;
                }
                html += `</div><small>${sc[usedKey]}/${sc[maxKey]}</small></div>`;
            }
        }
        html += `</div>`;
    }
    else {
        html += `<p style="color:var(--text-muted);font-style:italic">No spellcasting configured. <a href="#" onclick="enableSpellcasting();return false">Set up spellcasting</a></p>`;
    }
    html += `<h3>Spells</h3>
    <button class="btn btn-sm btn-primary" onclick="addSpell()">+ Add Spell</button>
    <div style="margin-top:8px">`;
    levels.forEach(lv => {
        const spells = byLevel[lv] || [];
        if (spells.length === 0)
            return;
        const label = lv === 0 ? 'Cantrips' : `Level ${lv}`;
        html += `<h4 style="margin-top:12px;color:var(--ink-light);font-size:0.95rem">${label}</h4>`;
        spells.forEach((s) => {
            html += `<div class="spell-item ${s.prepared ? 'prepared' : ''}">
        <strong>${esc(s.name)}</strong> <span style="color:var(--text-muted)">(${esc(s.school)})</span>
        <span style="float:right">${s.prepared ? '✓ Prepared' : ''}</span>
        <div style="font-size:0.85rem;color:var(--text-light)">
          ${esc(s.casting_time)} · ${esc(s.range)} · ${esc(s.components)} · ${esc(s.duration)}
        </div>
        <div style="margin-top:4px">${esc(s.description)}</div>
      </div>`;
        });
    });
    html += `</div>`;
    el.innerHTML = html;
}
async function toggleSlot(level, index) {
    const sc = currentChar.spellcasting;
    const usedKey = `slots_${level}_used`;
    sc[usedKey] = index + 1 === sc[usedKey] ? index : index + 1;
    await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, sc);
    renderSheet();
}
window.toggleSlot = toggleSlot;
async function enableSpellcasting() {
    currentChar.spellcasting = { ability: 'int', save_dc: 10, attack_bonus: 0, slots_1_max: 2, slots_1_used: 0 };
    await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, currentChar.spellcasting);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSheet();
}
window.enableSpellcasting = enableSpellcasting;
async function updateSpellcasting(field, value) {
    currentChar.spellcasting[field] = value;
    await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, currentChar.spellcasting);
}
window.updateSpellcasting = updateSpellcasting;
function renderInventory() {
    const c = currentChar;
    const el = document.getElementById('inventorySection');
    const inv = c.inventory || [];
    const cur = c.currency || { cp: 0, sp: 0, ep: 0, gp: 0, pp: 0 };
    el.innerHTML = `
    <h3>Currency</h3>
    <div class="form-row">
      <div class="form-group"><label>PP</label><input type="number" value="${cur.pp}" onchange="updateCurrency('pp', +this.value)"></div>
      <div class="form-group"><label>GP</label><input type="number" value="${cur.gp}" onchange="updateCurrency('gp', +this.value)"></div>
      <div class="form-group"><label>EP</label><input type="number" value="${cur.ep}" onchange="updateCurrency('ep', +this.value)"></div>
      <div class="form-group"><label>SP</label><input type="number" value="${cur.sp}" onchange="updateCurrency('sp', +this.value)"></div>
      <div class="form-group"><label>CP</label><input type="number" value="${cur.cp}" onchange="updateCurrency('cp', +this.value)"></div>
    </div>
    <div class="ornament">✧</div>
    <div style="display:flex;justify-content:space-between;align-items:center">
      <h3>Inventory</h3>
      <button class="btn btn-sm btn-primary" onclick="showAddItem()">+ Add Item</button>
    </div>
    <div style="margin-top:8px">
      ${inv.map((item) => `
        <div class="inventory-item ${item.is_equipped ? 'equipped' : ''}">
          <div>
            <span class="item-name">${esc(item.name)}</span>
            <span class="item-qty">×${item.quantity}</span>
            <span style="font-size:0.8rem;color:var(--text-muted)">${esc(item.category)}</span>
            ${item.is_equipped ? '<span class="badge badge-gold">E</span>' : ''}
            ${item.is_magical ? '<span class="badge badge-blood">★</span>' : ''}
          </div>
          <div>
            <span style="font-size:0.85rem;color:var(--text-light)">${item.damage_dice ? item.damage_dice + ' ' + item.damage_type : ''} ${item.ac_bonus ? 'AC+' + item.ac_bonus : ''}</span>
            <span style="margin-left:8px">
              <button class="btn btn-sm" onclick="toggleEquip(${item.id})">${item.is_equipped ? 'Unequip' : 'Equip'}</button>
              <button class="btn btn-sm btn-danger" onclick="deleteItem(${item.id})">×</button>
            </span>
          </div>
        </div>
      `).join('') || '<div class="empty-state" style="padding:16px">Empty handed. Add some items.</div>'}
    </div>
  `;
}
async function updateCurrency(type, value) {
    const cur = { ...currentChar.currency, [type]: value };
    await api('PUT', `/api/characters/${currentChar.id}/currency`, cur);
}
window.updateCurrency = updateCurrency;
async function toggleEquip(id) {
    const item = currentChar.inventory.find((i) => i.id === id);
    if (!item)
        return;
    item.is_equipped = !item.is_equipped;
    await api('PUT', `/api/inventory/${id}`, item);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSheet();
}
window.toggleEquip = toggleEquip;
async function deleteItem(id) {
    await api('DELETE', `/api/inventory/${id}`);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSheet();
}
window.deleteItem = deleteItem;
function renderFeatures() {
    const c = currentChar;
    const el = document.getElementById('featuresSection');
    const feats = c.features || [];
    el.innerHTML = `
    <button class="btn btn-sm btn-primary" onclick="addFeature()">+ Add Feature</button>
    <div style="margin-top:12px">
      ${feats.map((f) => `
        <div class="card" style="padding:16px;margin-bottom:8px">
          <div class="card-header" style="border:none;padding:0;margin:0">
            <strong>${esc(f.name)}</strong>
            <span><span class="badge badge-blood">Lv ${f.level_gained}</span> ${f.source ? '<span class="badge badge-gold">' + esc(f.source) + '</span>' : ''}</span>
          </div>
          <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">${esc(f.description)}</p>
        </div>
      `).join('') || '<div class="empty-state" style="padding:16px">No features yet.</div>'}
    </div>
  `;
}
function renderDetails() {
    const c = currentChar;
    const el = document.getElementById('detailsSection');
    el.innerHTML = `
    <div class="form-row">
      <div class="form-group"><label>Race</label><input value="${esc(c.race)}" onchange="updateField('race', this.value)"></div>
      <div class="form-group"><label>Class</label><input value="${esc(c.class)}" onchange="updateField('class', this.value)"></div>
      <div class="form-group"><label>Subclass</label><input value="${esc(c.subclass)}" onchange="updateField('subclass', this.value)"></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>Level</label><input type="number" value="${c.level}" onchange="updateField('level', +this.value)"></div>
      <div class="form-group"><label>Background</label><input value="${esc(c.background)}" onchange="updateField('background', this.value)"></div>
      <div class="form-group"><label>Alignment</label><input value="${esc(c.alignment)}" onchange="updateField('alignment', this.value)"></div>
    </div>
    <div class="form-group">
      <label>Personality Traits</label>
      <textarea onchange="updateField('personality_traits', this.value)">${esc(c.personality_traits)}</textarea>
    </div>
    <div class="form-group">
      <label>Ideals</label>
      <textarea onchange="updateField('ideals', this.value)">${esc(c.ideals)}</textarea>
    </div>
    <div class="form-group">
      <label>Bonds</label>
      <textarea onchange="updateField('bonds', this.value)">${esc(c.bonds)}</textarea>
    </div>
    <div class="form-group">
      <label>Flaws</label>
      <textarea onchange="updateField('flaws', this.value)">${esc(c.flaws)}</textarea>
    </div>
    <div class="form-group">
      <label>Appearance</label>
      <textarea onchange="updateField('appearance', this.value)">${esc(c.appearance)}</textarea>
    </div>
    <div class="form-group">
      <label>Backstory</label>
      <textarea style="min-height:150px" onchange="updateField('backstory', this.value)">${esc(c.backstory)}</textarea>
    </div>
  `;
}
// ─── Dice Tab ───
function renderDiceTab() {
    const el = document.getElementById('diceTabSection');
    if (!el)
        return;
    el.innerHTML = `
    <div class="dice-roller">
      <h3>Dice Roller</h3>
      <div class="form-row" style="max-width:400px;margin:0 auto">
        <div class="form-group">
          <label>Expression (e.g. 2d6+3)</label>
          <input id="diceExpr" value="1d20" placeholder="e.g. 1d20+5" style="text-align:center;font-size:1.2rem">
        </div>
      </div>
      <div id="diceResult" class="dice-result" style="display:none"></div>
      <button class="btn btn-gold" onclick="doRoll()">Roll the Bones</button>
      <div class="ornament">✧</div>
      <h3>Recent Rolls</h3>
      <div id="diceHistory" class="dice-history"></div>
    </div>
  `;
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
        el.innerHTML = `
      <div>${esc(expr)}</div>
      <div style="font-size:1.2rem;color:var(--ink-light)">${esc(result.text)}</div>
    `;
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
        el.innerHTML = rolls.slice(0, 20).map((r) => `
      <div class="dice-history-item">
        <span>${esc(r.expression)}</span>
        <span><strong>${r.total}</strong> <span style="color:var(--text-muted)">${esc(r.result)}</span></span>
      </div>
    `).join('') || '<div style="text-align:center;color:var(--text-muted);padding:12px">No rolls yet</div>';
    }
    catch { }
}
window.loadDiceHistory = loadDiceHistory;
// ─── Modals ───
function showModal(html) {
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';
    overlay.innerHTML = `<div class="modal">${html}</div>`;
    overlay.onclick = (e) => { if (e.target === overlay)
        overlay.remove(); };
    document.body.appendChild(overlay);
    return overlay.querySelector('.modal');
}
window.showAddItem = function () {
    const modal = showModal(`
    <h2>Add Item</h2>
    <div class="form-group">
      <label>Name</label>
      <input id="itemName" list="equipSuggestions">
      <datalist id="equipSuggestions"></datalist>
    </div>
    <div class="form-row">
      <div class="form-group"><label>Quantity</label><input id="itemQty" type="number" value="1"></div>
      <div class="form-group"><label>Weight</label><input id="itemWeight" type="number" value="0" step="0.1"></div>
    </div>
    <div class="form-group">
      <label>Category</label>
      <select id="itemCategory">
        <option value="gear">Gear</option>
        <option value="weapon">Weapon</option>
        <option value="armor">Armor</option>
        <option value="consumable">Consumable</option>
        <option value="tool">Tool</option>
        <option value="magic">Magic Item</option>
        <option value="ammunition">Ammunition</option>
      </select>
    </div>
    <div id="weaponFields" style="display:none">
      <div class="form-row">
        <div class="form-group"><label>Damage Dice</label><input id="itemDamage" placeholder="e.g. 1d8"></div>
        <div class="form-group"><label>Damage Type</label><input id="itemDmgType" placeholder="e.g. slashing"></div>
      </div>
    </div>
    <div id="armorFields" style="display:none">
      <div class="form-group"><label>AC Bonus</label><input id="itemAC" type="number" value="0"></div>
    </div>
    <div class="form-group">
      <label>Description</label>
      <textarea id="itemDesc"></textarea>
    </div>
    <div style="display:flex;gap:8px">
      <button class="btn btn-primary" onclick="saveItem(this)">Add</button>
      <button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button>
    </div>
  `);
    // Load equipment suggestions
    fetch('/api/compendium/equipment', { credentials: 'include' })
        .then(r => r.json())
        .then(items => {
        const dl = document.getElementById('equipSuggestions');
        dl.innerHTML = items.map((i) => `<option value="${esc(i.name)}">${esc(i.category)}</option>`).join('');
    }).catch(() => { });
    const catSelect = document.getElementById('itemCategory');
    catSelect.onchange = () => {
        document.getElementById('weaponFields').style.display = catSelect.value === 'weapon' ? 'block' : 'none';
        document.getElementById('armorFields').style.display = catSelect.value === 'armor' ? 'block' : 'none';
    };
};
window.saveItem = async function (btn) {
    const item = {
        name: document.getElementById('itemName').value,
        quantity: +document.getElementById('itemQty').value || 1,
        weight: +document.getElementById('itemWeight').value || 0,
        category: document.getElementById('itemCategory').value,
        damage_dice: document.getElementById('itemDamage')?.value || '',
        damage_type: document.getElementById('itemDmgType')?.value || '',
        weapon_properties: '',
        ac_bonus: +document.getElementById('itemAC')?.value || 0,
        armor_type: '',
        description: document.getElementById('itemDesc').value,
        is_equipped: false,
        is_magical: false,
        attunement: false,
        notes: '',
    };
    if (!item.name)
        return toast('Name required', true);
    try {
        await api('POST', `/api/characters/${currentChar.id}/inventory`, item);
        btn.closest('.modal-overlay')?.remove();
        currentChar = await api('GET', `/api/characters/${currentChar.id}`);
        renderSheet();
    }
    catch (e) {
        toast(e.message, true);
    }
};
window.addSpell = function () {
    const modal = showModal(`
    <h2>Add Spell</h2>
    <div class="form-group">
      <label>Spell Name</label>
      <input id="spellName" list="spellSuggestions">
      <datalist id="spellSuggestions"></datalist>
    </div>
    <div class="form-row">
      <div class="form-group"><label>Level</label><input id="spellLevel" type="number" value="1" min="0" max="9"></div>
      <div class="form-group"><label>School</label><input id="spellSchool" placeholder="e.g. Evocation"></div>
    </div>
    <div class="form-group"><label>Casting Time</label><input id="spellTime" placeholder="1 action"></div>
    <div class="form-group"><label>Range</label><input id="spellRange" placeholder="60 feet"></div>
    <div class="form-group"><label>Components</label><input id="spellComp" placeholder="V, S, M"></div>
    <div class="form-group"><label>Duration</label><input id="spellDur" placeholder="Instantaneous"></div>
    <div class="form-group"><label>Description</label><textarea id="spellDesc"></textarea></div>
    <label><input id="spellPrepared" type="checkbox"> Prepared</label>
    <div style="display:flex;gap:8px;margin-top:12px">
      <button class="btn btn-primary" onclick="saveSpell(this)">Add</button>
      <button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button>
    </div>
  `);
    fetch('/api/compendium/spells', { credentials: 'include' })
        .then(r => r.json())
        .then(spells => {
        const dl = document.getElementById('spellSuggestions');
        dl.innerHTML = spells.map((s) => `<option value="${esc(s.name)}">Lv${s.level} ${esc(s.school)}</option>`).join('');
    }).catch(() => { });
};
window.saveSpell = async function (btn) {
    const spell = {
        name: document.getElementById('spellName').value,
        level: +document.getElementById('spellLevel').value || 0,
        school: document.getElementById('spellSchool').value,
        casting_time: document.getElementById('spellTime').value,
        range: document.getElementById('spellRange').value,
        components: document.getElementById('spellComp').value,
        duration: document.getElementById('spellDur').value,
        description: document.getElementById('spellDesc').value,
        prepared: document.getElementById('spellPrepared').checked,
        always_prepared: false,
        source: '',
        notes: '',
    };
    if (!spell.name)
        return toast('Name required', true);
    try {
        await api('POST', `/api/characters/${currentChar.id}/spells`, spell);
        btn.closest('.modal-overlay')?.remove();
        currentChar = await api('GET', `/api/characters/${currentChar.id}`);
        renderSheet();
    }
    catch (e) {
        toast(e.message, true);
    }
};
window.addFeature = function () {
    const modal = showModal(`
    <h2>Add Feature</h2>
    <div class="form-group"><label>Name</label><input id="featName"></div>
    <div class="form-group"><label>Description</label><textarea id="featDesc"></textarea></div>
    <div class="form-row">
      <div class="form-group"><label>Source</label><input id="featSource" placeholder="e.g. Class, Race"></div>
      <div class="form-group"><label>Level Gained</label><input id="featLevel" type="number" value="1"></div>
    </div>
    <div style="display:flex;gap:8px">
      <button class="btn btn-primary" onclick="saveFeature(this)">Add</button>
      <button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button>
    </div>
  `);
};
window.saveFeature = async function (btn) {
    const feat = {
        name: document.getElementById('featName').value,
        description: document.getElementById('featDesc').value,
        source: document.getElementById('featSource').value,
        level_gained: +document.getElementById('featLevel').value || 1,
    };
    if (!feat.name)
        return toast('Name required', true);
    try {
        await api('POST', `/api/characters/${currentChar.id}/features`, feat);
        btn.closest('.modal-overlay')?.remove();
        currentChar = await api('GET', `/api/characters/${currentChar.id}`);
        renderSheet();
    }
    catch (e) {
        toast(e.message, true);
    }
};
window.addProf = function () {
    const modal = showModal(`
    <h2>Add Proficiency</h2>
    <div class="form-group"><label>Name</label><input id="profName" placeholder="e.g. Perception"></div>
    <div class="form-group">
      <label>Type</label>
      <select id="profType">
        <option value="skill">Skill</option>
        <option value="save">Saving Throw</option>
        <option value="tool">Tool</option>
        <option value="weapon">Weapon</option>
        <option value="armor">Armor</option>
        <option value="language">Language</option>
        <option value="other">Other</option>
      </select>
    </div>
    <div style="display:flex;gap:8px">
      <button class="btn btn-primary" onclick="saveProf(this)">Add</button>
      <button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button>
    </div>
  `);
};
window.saveProf = async function (btn) {
    const prof = {
        character_id: currentChar.id,
        name: document.getElementById('profName').value,
        type: document.getElementById('profType').value,
    };
    if (!prof.name)
        return toast('Name required', true);
    try {
        await api('POST', '/api/proficiencies', prof);
        btn.closest('.modal-overlay')?.remove();
        currentChar = await api('GET', `/api/characters/${currentChar.id}`);
        renderSheet();
    }
    catch (e) {
        toast(e.message, true);
    }
};
window.deleteProf = async function (id) {
    await api('DELETE', `/api/proficiencies/${id}`);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSheet();
};
// ─── Print ───
window.printChar = async function () {
    if (!currentChar)
        return;
    try {
        const res = await fetch(`/api/characters/${currentChar.id}/print`, {
            headers: { 'X-CSRF-Token': csrfToken },
            credentials: 'include',
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
// ─── New Character ───
window.newChar = function () {
    const modal = showModal(`
    <h2>New Character</h2>
    <div class="form-group"><label>Name</label><input id="newName" placeholder="Character name"></div>
    <div class="form-row">
      <div class="form-group"><label>Race</label><input id="newRace" list="raceSuggestions"></div>
      <div class="form-group"><label>Class</label><input id="newClass" list="classSuggestions"></div>
    </div>
    <datalist id="raceSuggestions"></datalist>
    <datalist id="classSuggestions"></datalist>
    <div style="display:flex;gap:8px">
      <button class="btn btn-primary" onclick="createChar(this)">Create</button>
      <button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button>
    </div>
  `);
    // Populate suggestions from compendium
    fetch('/api/compendium/races', { credentials: 'include' })
        .then(r => r.json())
        .then(races => {
        const dl = document.getElementById('raceSuggestions');
        dl.innerHTML = races.map((r) => `<option value="${esc(r.name)}">${esc(r.size)}</option>`).join('');
    }).catch(() => { });
    fetch('/api/compendium/classes', { credentials: 'include' })
        .then(r => r.json())
        .then(cls => {
        const dl = document.getElementById('classSuggestions');
        dl.innerHTML = cls.map((c) => `<option value="${esc(c.name)}">d${c.hit_die}</option>`).join('');
    }).catch(() => { });
};
window.createChar = async function (btn) {
    const name = document.getElementById('newName').value || 'Unnamed';
    const race = document.getElementById('newRace').value;
    const cls = document.getElementById('newClass').value;
    try {
        const char = await api('POST', '/api/characters', { name, race, class: cls });
        btn.closest('.modal-overlay')?.remove();
        if (char.id)
            await openChar(char.id);
        loadCharacters();
    }
    catch (e) {
        toast(e.message, true);
    }
};
// ─── Import ───
window.showImport = function () {
    const modal = showModal(`
    <h2>Import Character</h2>
    <p style="color:var(--text-muted);font-style:italic;margin-bottom:12px">Paste JSON or upload a file</p>
    <div class="form-group">
      <label>JSON Data</label>
      <textarea id="importJson" style="min-height:200px;font-family:monospace;font-size:0.8rem"></textarea>
    </div>
    <div class="form-group">
      <label>Or upload a file</label>
      <input type="file" id="importFile" accept=".json">
    </div>
    <div style="display:flex;gap:8px">
      <button class="btn btn-primary" onclick="doImport()">Import</button>
      <button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button>
    </div>
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
            const res = await fetch('/api/characters/import', {
                method: 'POST',
                headers: { 'X-CSRF-Token': csrfToken },
                credentials: 'include',
                body: form,
            });
            result = await res.json();
        }
        else if (jsonEl.value.trim()) {
            result = await api('POST', '/api/characters/import', JSON.parse(jsonEl.value));
        }
        else {
            toast('Provide JSON or a file', true);
            return;
        }
        const count = Array.isArray(result) ? result.length : 1;
        toast(`Imported ${count} character(s)`);
        document.querySelector('.modal-overlay')?.remove();
        loadCharacters();
    }
    catch (e) {
        toast('Import failed: ' + e.message, true);
    }
};
// ─── Compendium Browser ───
window.showCompendium = function () {
    showView('compendium');
    loadCompendiumRaces();
};
function loadCompendiumTab(tab) {
    document.querySelectorAll('.comp-tab').forEach(el => el.classList.remove('active'));
    document.getElementById('compTab' + capitalize(tab))?.classList.add('active');
    ['races', 'classes', 'spells', 'equipment'].forEach(s => {
        document.getElementById('comp' + capitalize(s)).style.display = s === tab ? 'block' : 'none';
    });
    switch (tab) {
        case 'races':
            loadCompendiumRaces();
            break;
        case 'classes':
            loadCompendiumClasses();
            break;
        case 'spells':
            loadCompendiumSpells();
            break;
        case 'equipment':
            loadCompendiumEquipment();
            break;
    }
}
window.loadCompendiumTab = loadCompendiumTab;
async function loadCompendiumRaces() {
    try {
        const races = await api('GET', '/api/compendium/races');
        const el = document.getElementById('compRaces');
        el.innerHTML = races.map((r) => `
      <div class="card" style="padding:16px;margin-bottom:8px">
        <div class="card-header" style="border:none;padding:0;margin:0">
          <strong>${esc(r.name)}</strong>
          <span><span class="badge badge-gold">${esc(r.size)}</span> Speed: ${r.speed}</span>
        </div>
        <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">${esc(r.description)}</p>
      </div>
    `).join('');
    }
    catch { }
}
async function loadCompendiumClasses() {
    try {
        const cls = await api('GET', '/api/compendium/classes');
        const el = document.getElementById('compClasses');
        el.innerHTML = cls.map((c) => `
      <div class="card" style="padding:16px;margin-bottom:8px">
        <div class="card-header" style="border:none;padding:0;margin:0">
          <strong>${esc(c.name)}</strong>
          <span>Hit Die: d${c.hit_die} · ${esc(c.primary_ability)}</span>
        </div>
        <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">${esc(c.description)}</p>
      </div>
    `).join('');
    }
    catch { }
}
async function loadCompendiumSpells() {
    try {
        const spells = await api('GET', '/api/compendium/spells');
        const el = document.getElementById('compSpells');
        el.innerHTML = spells.map((s) => `
      <div class="spell-item">
        <strong>${esc(s.name)}</strong>
        <span style="color:var(--text-muted)">Lv${s.level} ${esc(s.school)}</span>
        <div style="font-size:0.85rem;color:var(--text-light)">${esc(s.casting_time)} · ${esc(s.range)} · ${esc(s.duration)}</div>
      </div>
    `).join('');
    }
    catch { }
}
async function loadCompendiumEquipment() {
    try {
        const items = await api('GET', '/api/compendium/equipment');
        const el = document.getElementById('compEquipment');
        el.innerHTML = items.map((i) => `
      <div class="inventory-item">
        <span class="item-name">${esc(i.name)}</span>
        <span style="color:var(--text-muted)">${esc(i.category)} ${i.weight > 0 ? '· ' + i.weight + 'lb' : ''}</span>
      </div>
    `).join('');
    }
    catch { }
}
// ─── Export ───
window.exportChar = async function () {
    if (!currentChar)
        return;
    try {
        const data = await api('GET', `/api/characters/${currentChar.id}/export`);
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${currentChar.name.replace(/[^a-zA-Z0-9]/g, '_')}.json`;
        a.click();
        URL.revokeObjectURL(url);
    }
    catch (e) {
        toast(e.message, true);
    }
};
// ─── Utils ───
function esc(s) {
    const div = document.createElement('div');
    div.textContent = s;
    return div.innerHTML;
}
function capitalize(s) {
    return s.charAt(0).toUpperCase() + s.slice(1);
}
function toast(msg, isError = false) {
    const el = document.createElement('div');
    el.className = 'toast' + (isError ? ' error' : '');
    el.textContent = msg;
    document.body.appendChild(el);
    setTimeout(() => el.remove(), 4000);
}
window.logout = async function () {
    await api('POST', '/api/logout');
    window.location.href = '/login';
};
init();
export {};
//# sourceMappingURL=app.js.map