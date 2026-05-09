let csrfToken = '';
let currentUser = null;
let currentView = 'characters';
let currentChar = null;
let currentTab = 'stats';
let allLocations = [];
let allNPCs = [];
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
        // Preload global lists
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
    const sections = ['stats', 'combat', 'spells', 'inventory', 'features', 'locations', 'npcs', 'sessions', 'quests', 'journal', 'graph', 'details', 'dice'];
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
    const abils = ['str', 'dex', 'con', 'int', 'wis', 'cha'].map(k => ({ key: k, label: k.toUpperCase() }));
    el.innerHTML = `
    <div class="ability-grid">
      ${abils.map(a => {
        const val = c[a.key];
        const mod = Math.floor((val - 10) / 2);
        const cls = mod > 0 ? 'mod-pos' : mod < 0 ? 'mod-neg' : 'mod-zero';
        return `<div class="ability-score">
          <div class="label">${a.label}</div>
          <div class="value">${val}</div>
          <div class="mod ${cls}">${mod >= 0 ? '+' : ''}${mod}</div>
        </div>`;
    }).join('')}
    </div>
    <div class="ornament">✧</div>
    <div class="form-row" style="margin-top:16px">
      <div class="form-group"><label>Proficiency</label><input type="number" value="${c.proficiency_bonus}" onchange="updateField('proficiency_bonus',+this.value)"></div>
      <div class="form-group"><label>Inspiration</label><input type="number" value="${c.inspiration}" onchange="updateField('inspiration',+this.value)"></div>
      <div class="form-group"><label>Passive Percep.</label><input type="number" value="${c.passive_perception}" onchange="updateField('passive_perception',+this.value)"></div>
      <div class="form-group"><label>XP</label><input type="number" value="${c.xp}" onchange="updateField('xp',+this.value)"></div>
    </div>
    <h3>Proficiencies</h3>
    <div id="profsArea">${(c.proficiencies || []).map((p) => `<span class="badge badge-blood" style="margin:2px">${esc(p.name)} (${p.type}) <a href="#" onclick="deleteProf(${p.id});return false" style="color:white;text-decoration:none">×</a></span>`).join('')}</div>
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
    const pct = c.hp_max > 0 ? Math.round((c.hp_current / c.hp_max) * 100) : 0;
    el.innerHTML = `
    <div class="combat-grid">
      <div class="combat-stat"><div class="label">AC</div><div class="value">${c.ac}</div></div>
      <div class="combat-stat"><div class="label">Initiative</div><div class="value">${c.initiative >= 0 ? '+' : ''}${c.initiative}</div></div>
      <div class="combat-stat"><div class="label">Speed</div><div class="value">${c.speed}</div></div>
    </div>
    <h3>Hit Points</h3>
    <div class="hp-bar"><div class="hp-bar-fill" style="width:${pct}%"></div><div class="hp-bar-text">${c.hp_current} / ${c.hp_max}${c.temp_hp > 0 ? ' (+' + c.temp_hp + ' temp)' : ''}</div></div>
    <div class="form-row" style="margin-top:12px">
      <div class="form-group"><label>HP Max</label><input type="number" value="${c.hp_max}" onchange="updateField('hp_max',+this.value)"></div>
      <div class="form-group"><label>Current</label><input type="number" value="${c.hp_current}" onchange="updateField('hp_current',+this.value)"></div>
      <div class="form-group"><label>Temp HP</label><input type="number" value="${c.temp_hp}" onchange="updateField('temp_hp',+this.value)"></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>Hit Dice</label><input value="${c.hit_dice}" onchange="updateField('hit_dice',this.value)"></div>
      <div class="form-group"><label>Remaining</label><input type="number" value="${c.hit_dice_current}" onchange="updateField('hit_dice_current',+this.value)"></div>
    </div>
  `;
}
function renderSpells() {
    const c = currentChar;
    const el = document.getElementById('spellsSection');
    const sc = c.spellcasting;
    const byLevel = {};
    for (let i = 0; i <= 9; i++)
        byLevel[i] = [];
    (c.spells || []).forEach((s) => { if (byLevel[s.level])
        byLevel[s.level].push(s); });
    let html = '';
    if (sc && sc.ability) {
        html += `<div class="form-row">
      <div class="form-group"><label>Ability</label><input value="${sc.ability}" onchange="updateSpellcasting('ability',this.value)"></div>
      <div class="form-group"><label>Save DC</label><input type="number" value="${sc.save_dc}" onchange="updateSpellcasting('save_dc',+this.value)"></div>
      <div class="form-group"><label>Attack</label><input type="number" value="${sc.attack_bonus}" onchange="updateSpellcasting('attack_bonus',+this.value)"></div>
    </div><h3>Spell Slots</h3><div class="slot-tracker">`;
        for (let lv = 1; lv <= 9; lv++) {
            const mx = sc['slots_' + lv + '_max'];
            const us = sc['slots_' + lv + '_used'];
            if (mx > 0) {
                html += `<div class="slot-level"><div class="level-label">Lv ${lv}</div><div class="slot-dots">`;
                for (let i = 0; i < mx; i++)
                    html += `<div class="slot-dot ${i < us ? 'filled' : 'empty'}" onclick="toggleSlot(${lv},${i})"></div>`;
                html += `</div><small>${us}/${mx}</small></div>`;
            }
        }
        html += `</div>`;
    }
    else {
        html += `<p style="color:var(--text-muted);font-style:italic">No spellcasting. <a href="#" onclick="enableSpellcasting();return false">Set up</a></p>`;
    }
    html += `<h3>Spells</h3><button class="btn btn-sm btn-primary" onclick="addSpell()">+ Add Spell</button><div style="margin-top:8px">`;
    for (let lv = 9; lv >= 0; lv--) {
        const spells = byLevel[lv] || [];
        if (!spells.length)
            continue;
        html += `<h4 style="margin-top:12px;color:var(--ink-light);font-size:0.95rem">${lv === 0 ? 'Cantrips' : 'Level ' + lv}</h4>`;
        spells.forEach((s) => {
            html += `<div class="spell-item ${s.prepared ? 'prepared' : ''}">
        <strong>${esc(s.name)}</strong> <span style="color:var(--text-muted)">(${esc(s.school)})</span>
        <span style="float:right">${s.prepared ? '✓ Prepared' : ''}</span>
        <div style="font-size:0.85rem;color:var(--text-light)">${esc(s.casting_time)} · ${esc(s.range)} · ${esc(s.components)} · ${esc(s.duration)}</div>
        <div style="margin-top:4px">${esc(s.description)}</div>
      </div>`;
        });
    }
    html += `</div>`;
    el.innerHTML = html;
}
async function toggleSlot(level, index) {
    const sc = currentChar.spellcasting;
    const k = 'slots_' + level + '_used';
    sc[k] = index + 1 === sc[k] ? index : index + 1;
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
      ${['PP', 'GP', 'EP', 'SP', 'CP'].map(u => `<div class="form-group"><label>${u}</label><input type="number" value="${cur[u.toLowerCase()]}" onchange="updateCurrency('${u.toLowerCase()}',+this.value)"></div>`).join('')}
    </div>
    <div class="ornament">✧</div>
    <div style="display:flex;justify-content:space-between;align-items:center"><h3>Inventory</h3><button class="btn btn-sm btn-primary" onclick="showAddItem()">+ Add</button></div>
    <div style="margin-top:8px">
      ${inv.map((item) => `<div class="inventory-item ${item.is_equipped ? 'equipped' : ''}">
        <div><span class="item-name">${esc(item.name)}</span> <span class="item-qty">×${item.quantity}</span> <span style="font-size:0.8rem;color:var(--text-muted)">${esc(item.category)}</span>
          ${item.is_equipped ? '<span class="badge badge-gold">E</span>' : ''} ${item.is_magical ? '<span class="badge badge-blood">★</span>' : ''}
        </div>
        <div><span style="font-size:0.85rem;color:var(--text-light)">${item.damage_dice ? item.damage_dice + ' ' + item.damage_type : ''} ${item.ac_bonus ? 'AC+' + item.ac_bonus : ''}</span>
          <span style="margin-left:8px">
            <button class="btn btn-sm" onclick="toggleEquip(${item.id})">${item.is_equipped ? 'Unequip' : 'Equip'}</button>
            <button class="btn btn-sm btn-danger" onclick="deleteItem(${item.id})">×</button>
          </span>
        </div>
      </div>`).join('') || '<div class="empty-state" style="padding:16px">Empty handed.</div>'}
    </div>`;
}
async function updateCurrency(t, v) {
    await api('PUT', `/api/characters/${currentChar.id}/currency`, { ...currentChar.currency, [t]: v });
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
    const feats = currentChar.features || [];
    document.getElementById('featuresSection').innerHTML = `
    <button class="btn btn-sm btn-primary" onclick="addFeature()">+ Add</button>
    <div style="margin-top:12px">${feats.map((f) => `
      <div class="card" style="padding:16px;margin-bottom:8px">
        <div class="card-header" style="border:none;padding:0;margin:0">
          <strong>${esc(f.name)}</strong>
          <span><span class="badge badge-blood">Lv ${f.level_gained}</span> ${f.source ? '<span class="badge badge-gold">' + esc(f.source) + '</span>' : ''}</span>
        </div>
        <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">${esc(f.description)}</p>
      </div>`).join('') || '<div class="empty-state" style="padding:16px">No features.</div>'}
    </div>`;
}
function renderDetails() {
    const c = currentChar;
    const el = document.getElementById('detailsSection');
    el.innerHTML = `
    <div class="form-row">
      <div class="form-group"><label>Race</label><input value="${esc(c.race)}" onchange="updateField('race',this.value)"></div>
      <div class="form-group"><label>Class</label><input value="${esc(c.class)}" onchange="updateField('class',this.value)"></div>
      <div class="form-group"><label>Subclass</label><input value="${esc(c.subclass)}" onchange="updateField('subclass',this.value)"></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>Level</label><input type="number" value="${c.level}" onchange="updateField('level',+this.value)"></div>
      <div class="form-group"><label>Background</label><input value="${esc(c.background)}" onchange="updateField('background',this.value)"></div>
      <div class="form-group"><label>Alignment</label><input value="${esc(c.alignment)}" onchange="updateField('alignment',this.value)"></div>
    </div>
    ${['personality_traits', 'ideals', 'bonds', 'flaws', 'appearance'].map(f => `
      <div class="form-group"><label>${capitalize(f.replace(/_/g, ' '))}</label>
      <textarea onchange="updateField('${f}',this.value)">${esc(c[f])}</textarea></div>
    `).join('')}
    <div class="form-group"><label>Backstory</label>
    <textarea style="min-height:150px" onchange="updateField('backstory',this.value)">${esc(c.backstory)}</textarea></div>`;
}
// ─── Locations ───
async function renderLocations() {
    const el = document.getElementById('locationsSection');
    try {
        const links = await api('GET', `/api/characters/${currentChar.id}/locations`);
        el.innerHTML = `
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>Known Locations</h3>
        <button class="btn btn-sm btn-primary" onclick="showLinkLocation()">+ Link Location</button>
      </div>
      ${links.length ? links.map((l) => `
        <div class="inventory-item">
          <div><span class="item-name">${esc(l.location_name)}</span> <span style="color:var(--text-muted)">(${esc(l.location_type)})</span></div>
          <div><span class="badge badge-gold">${esc(l.relationship)}</span>
            <button class="btn btn-sm btn-danger" onclick="unlinkLocation(${l.id})">×</button>
          </div>
        </div>`).join('')
            : '<div class="empty-state" style="padding:16px">No locations linked. <a href="#" onclick="showLinkLocation();return false">Link one</a></div>'}
      <hr style="border-color:var(--border-light);margin:16px 0">
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>All Your Locations</h3>
        <button class="btn btn-sm" onclick="showCreateLocation()">+ New Location</button>
      </div>
      ${allLocations.map((l) => `
        <div class="inventory-item">
          <div><span class="item-name">${esc(l.name)}</span> <span style="color:var(--text-muted)">(${esc(l.type)})</span></div>
          <div><span style="font-size:0.85rem;color:var(--text-light)">${esc(l.description).substring(0, 60)}</span></div>
        </div>`).join('')}`;
    }
    catch { }
}
window.showLinkLocation = function () {
    const modal = showModal(`
    <h2>Link Location to Character</h2>
    <div class="form-group"><label>Location</label>
      <select id="linkLocId">${allLocations.map((l) => `<option value="${l.id}">${esc(l.name)} (${esc(l.type)})</option>`).join('')}</select>
    </div>
    <div class="form-group"><label>Relationship</label>
      <select id="linkLocRel">
        <option value="current">Current Location</option>
        <option value="hometown">Hometown</option>
        <option value="visited">Visited</option>
        <option value="headquarters">Headquarters</option>
        <option value="quest">Quest Location</option>
        <option value="other">Other</option>
      </select>
    </div>
    <div class="form-group"><label>Notes</label><textarea id="linkLocNotes"></textarea></div>
    <button class="btn btn-primary" onclick="saveLinkLocation(this)">Link</button>
  `);
};
window.saveLinkLocation = async function (btn) {
    await api('POST', `/api/characters/${currentChar.id}/locations`, {
        location_id: +document.getElementById('linkLocId').value,
        relationship: document.getElementById('linkLocRel').value,
        notes: document.getElementById('linkLocNotes').value,
    });
    btn.closest('.modal-overlay')?.remove();
    renderLocations();
};
window.unlinkLocation = async function (id) {
    await api('DELETE', `/api/locations/link/${id}`);
    renderLocations();
};
window.showCreateLocation = function () {
    const modal = showModal(`
    <h2>New Location</h2>
    <div class="form-group"><label>Name</label><input id="newLocName"></div>
    <div class="form-group"><label>Type</label>
      <select id="newLocType">
        <option value="region">Region</option>
        <option value="city">City</option>
        <option value="town">Town</option>
        <option value="dungeon">Dungeon</option>
        <option value="tavern">Tavern</option>
        <option value="temple">Temple</option>
        <option value="shop">Shop</option>
        <option value="wilderness">Wilderness</option>
        <option value="other">Other</option>
      </select>
    </div>
    <div class="form-group"><label>Description</label><textarea id="newLocDesc"></textarea></div>
    <button class="btn btn-primary" onclick="saveNewLocation(this)">Create</button>
  `);
};
window.saveNewLocation = async function (btn) {
    await api('POST', '/api/locations', {
        name: document.getElementById('newLocName').value,
        type: document.getElementById('newLocType').value,
        description: document.getElementById('newLocDesc').value,
    });
    btn.closest('.modal-overlay')?.remove();
    allLocations = await api('GET', '/api/locations');
    renderLocations();
};
// ─── NPCs ───
async function renderNPCs() {
    const el = document.getElementById('npcsSection');
    try {
        const links = await api('GET', `/api/characters/${currentChar.id}/npcs`);
        el.innerHTML = `
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>Related NPCs</h3>
        <button class="btn btn-sm btn-primary" onclick="showLinkNPC()">+ Link NPC</button>
      </div>
      ${links.length ? links.map((n) => `
        <div class="inventory-item">
          <div><span class="item-name">${esc(n.npc_name)}</span>
            <span style="color:var(--text-muted)">${esc(n.npc_race)} ${esc(n.npc_class)}</span>
            ${n.npc_is_alive ? '' : '<span class="badge badge-blood">Deceased</span>'}
          </div>
          <div><span class="badge badge-gold">${esc(n.relationship)}</span>
            <button class="btn btn-sm btn-danger" onclick="unlinkNPC(${n.id})">×</button>
          </div>
        </div>`).join('')
            : '<div class="empty-state" style="padding:16px">No NPCs linked yet.</div>'}
      <hr style="border-color:var(--border-light);margin:16px 0">
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>All Your NPCs</h3>
        <button class="btn btn-sm" onclick="showCreateNPC()">+ New NPC</button>
      </div>
      ${allNPCs.map((n) => `
        <div class="inventory-item">
          <div><span class="item-name">${esc(n.name)}</span>
            <span style="color:var(--text-muted)">${esc(n.race)} ${esc(n.class)}</span>
          </div>
          <div style="font-size:0.85rem;color:var(--text-light)">HP: ${n.hp_current}/${n.hp_max}</div>
        </div>`).join('')}`;
    }
    catch { }
}
window.showLinkNPC = function () {
    const modal = showModal(`
    <h2>Link NPC to Character</h2>
    <div class="form-group"><label>NPC</label>
      <select id="linkNPCId">${allNPCs.map((n) => `<option value="${n.id}">${esc(n.name)} (${esc(n.race)} ${esc(n.class)})</option>`).join('')}</select>
    </div>
    <div class="form-group"><label>Relationship</label>
      <select id="linkNPCRel">
        <option value="ally">Ally</option>
        <option value="enemy">Enemy</option>
        <option value="family">Family</option>
        <option value="contact">Contact</option>
        <option value="acquaintance">Acquaintance</option>
        <option value="pet">Pet/Mount</option>
        <option value="deity">Deity/Patron</option>
        <option value="other">Other</option>
      </select>
    </div>
    <div class="form-group"><label>Notes</label><textarea id="linkNPCNotes"></textarea></div>
    <button class="btn btn-primary" onclick="saveLinkNPC(this)">Link</button>
  `);
};
window.saveLinkNPC = async function (btn) {
    await api('POST', `/api/characters/${currentChar.id}/npcs`, {
        npc_id: +document.getElementById('linkNPCId').value,
        relationship: document.getElementById('linkNPCRel').value,
        notes: document.getElementById('linkNPCNotes').value,
    });
    btn.closest('.modal-overlay')?.remove();
    renderNPCs();
};
window.unlinkNPC = async function (id) {
    await api('DELETE', `/api/npcs/link/${id}`);
    renderNPCs();
};
window.showCreateNPC = function () {
    const modal = showModal(`
    <h2>New NPC</h2>
    <div class="form-group"><label>Name</label><input id="newNPCName"></div>
    <div class="form-row">
      <div class="form-group"><label>Race</label><input id="newNPCRace"></div>
      <div class="form-group"><label>Class</label><input id="newNPCClass"></div>
    </div>
    <div class="form-group"><label>Description</label><textarea id="newNPCDesc"></textarea></div>
    <button class="btn btn-primary" onclick="saveNewNPC(this)">Create</button>
  `);
};
window.saveNewNPC = async function (btn) {
    await api('POST', '/api/npcs', {
        name: document.getElementById('newNPCName').value,
        race: document.getElementById('newNPCRace').value,
        class: document.getElementById('newNPCClass').value,
        description: document.getElementById('newNPCDesc').value,
    });
    btn.closest('.modal-overlay')?.remove();
    allNPCs = await api('GET', '/api/npcs');
    renderNPCs();
};
// ─── Sessions ───
async function renderSessions() {
    const el = document.getElementById('sessionsSection');
    try {
        const sessions = await api('GET', `/api/characters/${currentChar.id}/sessions`);
        el.innerHTML = `
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>Session Log</h3>
        <button class="btn btn-sm btn-primary" onclick="showAddSession()">+ Log Session</button>
      </div>
      <div style="margin-top:12px">
        ${sessions.map((s) => `
          <div class="card" style="padding:16px;margin-bottom:8px">
            <div class="card-header" style="border:none;padding:0;margin:0">
              <strong>${esc(s.title) || 'Session ' + s.session_date}</strong>
              <span><span class="badge badge-gold">${s.session_date}</span>
                ${s.xp_earned > 0 ? '<span class="badge badge-blood">+' + s.xp_earned + ' XP</span>' : ''}
                ${s.gold_earned > 0 ? '<span class="badge badge-gold">+' + s.gold_earned + ' GP</span>' : ''}
              </span>
            </div>
            <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px;">${esc(s.notes).substring(0, 200)}</p>
            ${s.important_events ? `<p style="font-size:0.85rem;color:var(--text-muted);margin-top:4px"><em>${esc(s.important_events).substring(0, 150)}</em></p>` : ''}
            <button class="btn btn-sm btn-danger" style="margin-top:4px" onclick="deleteSession(${s.id})">×</button>
          </div>`).join('') || '<div class="empty-state">No sessions logged yet. Start your campaign!</div>'}
      </div>`;
    }
    catch { }
}
window.showAddSession = function () {
    const modal = showModal(`
    <h2>Log Session</h2>
    <div class="form-group"><label>Date</label><input id="sessDate" type="date" value="${new Date().toISOString().split('T')[0]}"></div>
    <div class="form-group"><label>Title</label><input id="sessTitle" placeholder="Session 1: The Adventure Begins"></div>
    <div class="form-group"><label>Notes</label><textarea id="sessNotes" style="min-height:100px" placeholder="What happened?"></textarea></div>
    <div class="form-row">
      <div class="form-group"><label>XP Earned</label><input id="sessXP" type="number" value="0"></div>
      <div class="form-group"><label>Gold Earned</label><input id="sessGold" type="number" value="0"></div>
    </div>
    <div class="form-group"><label>Important Events</label><textarea id="sessEvents" placeholder="Key moments, NPCs met, revelations..."></textarea></div>
    <button class="btn btn-primary" onclick="saveSession(this)">Log Session</button>
  `);
};
window.saveSession = async function (btn) {
    await api('POST', `/api/characters/${currentChar.id}/sessions`, {
        session_date: document.getElementById('sessDate').value,
        title: document.getElementById('sessTitle').value,
        notes: document.getElementById('sessNotes').value,
        xp_earned: +document.getElementById('sessXP').value || 0,
        gold_earned: +document.getElementById('sessGold').value || 0,
        important_events: document.getElementById('sessEvents').value,
    });
    btn.closest('.modal-overlay')?.remove();
    renderSessions();
};
window.deleteSession = async function (id) {
    await api('DELETE', `/api/sessions/${id}`);
    renderSessions();
};
// ─── Quests ───
async function renderQuests() {
    const el = document.getElementById('questsSection');
    try {
        const quests = await api('GET', `/api/characters/${currentChar.id}/quests`);
        const groups = { available: [], active: [], complete: [], failed: [], abandoned: [] };
        quests.forEach((q) => { if (groups[q.status])
            groups[q.status].push(q); });
        let questHtml = '<div style="display:flex;justify-content:space-between;align-items:center"><h3>Quests</h3><button class="btn btn-sm btn-primary" onclick="showAddQuest()">+ New Quest</button></div>';
        const labels = { active: 'Active', available: 'Available', complete: 'Complete', failed: 'Failed', abandoned: 'Abandoned' };
        for (const st of ['active', 'available', 'complete', 'failed', 'abandoned']) {
            const qs = groups[st] || [];
            if (!qs.length)
                continue;
            questHtml += '<h4 style="margin-top:12px;color:var(--ink-light)">' + labels[st] + '</h4>';
            for (const q of qs) {
                const opts = ['active', 'available', 'complete', 'failed', 'abandoned'].map(s => '<option value="' + s + '"' + (s === q.status ? ' selected' : '') + '>' + s + '</option>').join('');
                questHtml += '<div class="card" style="padding:16px;margin-bottom:8px">';
                questHtml += '<div class="card-header" style="border:none;padding:0;margin:0"><strong>' + esc(q.name) + '</strong>';
                questHtml += '<span><select onchange="updateQuestStatus(' + q.id + ',this.value)" style="width:auto;font-size:0.8rem;padding:2px 6px">' + opts + '</select>';
                questHtml += '<button class="btn btn-sm btn-danger" onclick="deleteQuest(' + q.id + ')">&times;</button></span></div>';
                questHtml += '<p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">' + esc(q.description).substring(0, 200) + '</p>';
                if (q.objectives)
                    questHtml += '<div style="font-size:0.85rem;color:var(--text-muted);margin-top:4px"><strong>Objectives:</strong> ' + esc(q.objectives).substring(0, 150) + '</div>';
                if (q.rewards)
                    questHtml += '<div style="font-size:0.85rem;color:var(--success);margin-top:2px"><strong>Reward:</strong> ' + esc(q.rewards).substring(0, 150) + '</div>';
                questHtml += '</div>';
            }
        }
        if (quests.length === 0) {
            questHtml += '<div class="empty-state">No quests. <a href="#" onclick="showAddQuest();return false">Start one</a></div>';
        }
        el.innerHTML = questHtml;
    }
    catch { }
}
window.showAddQuest = function () {
    const modal = showModal(`
    <h2>New Quest</h2>
    <div class="form-group"><label>Name</label><input id="questName" placeholder="e.g. Retrieve the Lost Artifact"></div>
    <div class="form-group"><label>Description</label><textarea id="questDesc"></textarea></div>
    <div class="form-group"><label>Objectives</label><textarea id="questObj" placeholder="1. Travel to the Forgotten Temple\n2. Defeat the guardian\n3. Retrieve the artifact"></textarea></div>
    <div class="form-group"><label>Rewards</label><textarea id="questRewards" placeholder="500 XP, +1 Longsword, 200 GP"></textarea></div>
    <div class="form-group"><label>Notes</label><textarea id="questNotes"></textarea></div>
    <button class="btn btn-primary" onclick="saveQuest(this)">Create</button>
  `);
};
window.saveQuest = async function (btn) {
    await api('POST', `/api/characters/${currentChar.id}/quests`, {
        name: document.getElementById('questName').value,
        description: document.getElementById('questDesc').value,
        objectives: document.getElementById('questObj').value,
        rewards: document.getElementById('questRewards').value,
        notes: document.getElementById('questNotes').value,
    });
    btn.closest('.modal-overlay')?.remove();
    renderQuests();
};
window.updateQuestStatus = async function (id, status) {
    const quests = await api('GET', `/api/characters/${currentChar.id}/quests`);
    const q = quests.find((x) => x.id === id);
    if (!q)
        return;
    q.status = status;
    await api('PUT', `/api/quests/${id}`, q);
    renderQuests();
};
window.deleteQuest = async function (id) {
    await api('DELETE', `/api/quests/${id}`);
    renderQuests();
};
// ─── Journal ───
async function renderJournal() {
    const el = document.getElementById('journalSection');
    try {
        const entries = await api('GET', `/api/characters/${currentChar.id}/journal`);
        el.innerHTML = `
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>Character Journal</h3>
        <button class="btn btn-sm btn-primary" onclick="showAddJournal()">+ Write Entry</button>
      </div>
      <div style="margin-top:12px">
        ${entries.map((j) => `
          <div class="card" style="padding:16px;margin-bottom:8px">
            <div class="card-header" style="border:none;padding:0;margin:0">
              <strong>${esc(j.title) || 'Untitled'}</strong>
              <span><span class="badge badge-gold">${j.entry_date}</span>
                <button class="btn btn-sm btn-danger" onclick="deleteJournal(${j.id})">×</button>
              </span>
            </div>
            <div style="font-size:0.9rem;color:var(--text-light);margin-top:8px;white-space:pre-wrap">${esc(j.entry)}</div>
          </div>`).join('') || '<div class="empty-state">No journal entries yet. Record your adventures!</div>'}
      </div>`;
    }
    catch { }
}
window.showAddJournal = function () {
    const modal = showModal(`
    <h2>Journal Entry</h2>
    <div class="form-group"><label>Date</label><input id="journalDate" type="date" value="${new Date().toISOString().split('T')[0]}"></div>
    <div class="form-group"><label>Title</label><input id="journalTitle" placeholder="Day 1: Arrival in Waterdeep"></div>
    <div class="form-group"><label>Entry</label><textarea id="journalEntry" style="min-height:200px" placeholder="Write your character's thoughts..."></textarea></div>
    <button class="btn btn-primary" onclick="saveJournal(this)">Save</button>
  `);
};
window.saveJournal = async function (btn) {
    await api('POST', `/api/characters/${currentChar.id}/journal`, {
        entry_date: document.getElementById('journalDate').value,
        title: document.getElementById('journalTitle').value,
        entry: document.getElementById('journalEntry').value,
    });
    btn.closest('.modal-overlay')?.remove();
    renderJournal();
};
window.deleteJournal = async function (id) {
    await api('DELETE', `/api/journal/${id}`);
    renderJournal();
};
// ─── Graph ───
async function renderGraph() {
    const el = document.getElementById('graphSection');
    el.innerHTML = `<div class="ornament">✧ Drawing your web of fate ✧</div>
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
            // Fallback: simple text representation
            el.innerHTML += `<div style="padding:20px;font-size:0.9rem">
        <h3>Character Web</h3>
        <p>${data.nodes.map((n) => `${n.label} [${n.group}]`).join(' <span style="color:var(--text-muted)">→</span> ')}</p>
        <p style="color:var(--text-muted);font-style:italic;margin-top:8px">
          ${data.nodes.length} connections · ${data.edges.length} relationships</p>
      </div>`;
        }
    }
    catch (e) {
        el.innerHTML += `<div class="empty-state">Could not load graph: ${e.message}</div>`;
    }
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
        <div class="form-group"><label>Expression (e.g. 2d6+3)</label>
          <input id="diceExpr" value="1d20" placeholder="e.g. 1d20+5" style="text-align:center;font-size:1.2rem"></div>
      </div>
      <div id="diceResult" class="dice-result" style="display:none"></div>
      <button class="btn btn-gold" onclick="doRoll()">Roll the Bones</button>
      <div class="ornament">✧</div>
      <h3>Recent Rolls</h3>
      <div id="diceHistory" class="dice-history"></div>
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
        el.innerHTML = `<div>${esc(expr)}</div><div style="font-size:1.2rem;color:var(--ink-light)">${esc(result.text)}</div>`;
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
        el.innerHTML = rolls.slice(0, 20).map((r) => `<div class="dice-history-item"><span>${esc(r.expression)}</span><span><strong>${r.total}</strong> <span style="color:var(--text-muted)">${esc(r.result)}</span></span></div>`).join('') || '<div style="text-align:center;color:var(--text-muted);padding:12px">No rolls yet</div>';
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
    <div class="form-group"><label>Name</label><input id="itemName" list="equipSuggestions"><datalist id="equipSuggestions"></datalist></div>
    <div class="form-row"><div class="form-group"><label>Qty</label><input id="itemQty" type="number" value="1"></div><div class="form-group"><label>Weight</label><input id="itemWeight" type="number" value="0" step="0.1"></div></div>
    <div class="form-group"><label>Category</label><select id="itemCategory">${['gear', 'weapon', 'armor', 'consumable', 'tool', 'magic', 'ammunition'].map(c => `<option value="${c}">${c}</option>`).join('')}</select></div>
    <div id="weaponFields" style="display:none"><div class="form-row"><div class="form-group"><label>Damage</label><input id="itemDamage" placeholder="1d8"></div><div class="form-group"><label>Type</label><input id="itemDmgType" placeholder="slashing"></div></div></div>
    <div id="armorFields" style="display:none"><div class="form-group"><label>AC Bonus</label><input id="itemAC" type="number" value="0"></div></div>
    <div class="form-group"><label>Description</label><textarea id="itemDesc"></textarea></div>
    <div style="display:flex;gap:8px"><button class="btn btn-primary" onclick="saveItem(this)">Add</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
    fetch('/api/compendium/equipment', { credentials: 'include' }).then(r => r.json()).then(items => {
        document.getElementById('equipSuggestions').innerHTML = items.map((i) => `<option value="${esc(i.name)}">`).join('');
    }).catch(() => { });
    const catSel = document.getElementById('itemCategory');
    catSel.addEventListener('change', () => {
        document.getElementById('weaponFields').style.display = catSel.value === 'weapon' ? 'block' : 'none';
        document.getElementById('armorFields').style.display = catSel.value === 'armor' ? 'block' : 'none';
    });
};
window.saveItem = async function (btn) {
    const item = {
        name: document.getElementById('itemName').value,
        quantity: +document.getElementById('itemQty').value || 1,
        weight: +document.getElementById('itemWeight').value || 0,
        category: document.getElementById('itemCategory').value,
        damage_dice: document.getElementById('itemDamage')?.value || '',
        damage_type: document.getElementById('itemDmgType')?.value || '',
        weapon_properties: '', ac_bonus: +document.getElementById('itemAC')?.value || 0,
        armor_type: '', description: document.getElementById('itemDesc').value,
        is_equipped: false, is_magical: false, attunement: false, notes: '',
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
    <div class="form-group"><label>Name</label><input id="spellName" list="spellSuggestions"><datalist id="spellSuggestions"></datalist></div>
    <div class="form-row"><div class="form-group"><label>Level</label><input id="spellLevel" type="number" value="1" min="0" max="9"></div><div class="form-group"><label>School</label><input id="spellSchool"></div></div>
    <div class="form-group"><label>Casting Time</label><input id="spellTime" placeholder="1 action"></div>
    <div class="form-group"><label>Range</label><input id="spellRange" placeholder="60 feet"></div>
    <div class="form-group"><label>Components</label><input id="spellComp" placeholder="V, S, M"></div>
    <div class="form-group"><label>Duration</label><input id="spellDur" placeholder="Instantaneous"></div>
    <div class="form-group"><label>Description</label><textarea id="spellDesc"></textarea></div>
    <label><input id="spellPrepared" type="checkbox"> Prepared</label>
    <div style="display:flex;gap:8px;margin-top:12px"><button class="btn btn-primary" onclick="saveSpell(this)">Add</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
    fetch('/api/compendium/spells', { credentials: 'include' }).then(r => r.json()).then(spells => {
        document.getElementById('spellSuggestions').innerHTML = spells.map((s) => `<option value="${esc(s.name)}">Lv${s.level} ${esc(s.school)}</option>`).join('');
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
        always_prepared: false, source: '', notes: '',
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
    showModal(`<h2>Add Feature</h2>
    <div class="form-group"><label>Name</label><input id="featName"></div>
    <div class="form-group"><label>Description</label><textarea id="featDesc"></textarea></div>
    <div class="form-row"><div class="form-group"><label>Source</label><input id="featSource" placeholder="e.g. Class"></div><div class="form-group"><label>Level</label><input id="featLevel" type="number" value="1"></div></div>
    <div style="display:flex;gap:8px"><button class="btn btn-primary" onclick="saveFeature(this)">Add</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
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
    showModal(`<h2>Add Proficiency</h2>
    <div class="form-group"><label>Name</label><input id="profName" placeholder="e.g. Perception"></div>
    <div class="form-group"><label>Type</label><select id="profType">${['skill', 'save', 'tool', 'weapon', 'armor', 'language', 'other'].map(t => `<option value="${t}">${t}</option>`).join('')}</select></div>
    <div style="display:flex;gap:8px"><button class="btn btn-primary" onclick="saveProf(this)">Add</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
};
window.saveProf = async function (btn) {
    const prof = { character_id: currentChar.id, name: document.getElementById('profName').value, type: document.getElementById('profType').value };
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
// ─── New Character ───
window.newChar = function () {
    const modal = showModal(`
    <h2>New Character</h2>
    <div class="form-group"><label>Name</label><input id="newName" placeholder="Character name"></div>
    <div class="form-row"><div class="form-group"><label>Race</label><input id="newRace" list="raceSuggestions"></div><div class="form-group"><label>Class</label><input id="newClass" list="classSuggestions"></div></div>
    <datalist id="raceSuggestions"></datalist><datalist id="classSuggestions"></datalist>
    <div style="display:flex;gap:8px"><button class="btn btn-primary" onclick="createChar(this)">Create</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
    fetch('/api/compendium/races', { credentials: 'include' }).then(r => r.json()).then(races => {
        document.getElementById('raceSuggestions').innerHTML = races.map((r) => `<option value="${esc(r.name)}">`).join('');
    }).catch(() => { });
    fetch('/api/compendium/classes', { credentials: 'include' }).then(r => r.json()).then(cls => {
        document.getElementById('classSuggestions').innerHTML = cls.map((c) => `<option value="${esc(c.name)}">`).join('');
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
    showModal(`
    <h2>Import Character</h2>
    <p style="color:var(--text-muted);font-style:italic;margin-bottom:12px">Paste JSON or upload a file</p>
    <div class="form-group"><label>JSON</label><textarea id="importJson" style="min-height:200px;font-family:monospace;font-size:0.8rem"></textarea></div>
    <div class="form-group"><label>File</label><input type="file" id="importFile" accept=".json"></div>
    <div style="display:flex;gap:8px"><button class="btn btn-primary" onclick="doImport()">Import</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
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
        document.querySelector('.modal-overlay')?.remove();
        loadCharacters();
    }
    catch (e) {
        toast('Import failed: ' + e.message, true);
    }
};
// ─── Compendium ───
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
    if (tab === 'races')
        loadCompendiumRaces();
    if (tab === 'classes')
        loadCompendiumClasses();
    if (tab === 'spells')
        loadCompendiumSpells();
    if (tab === 'equipment')
        loadCompendiumEquipment();
}
window.loadCompendiumTab = loadCompendiumTab;
async function loadCompendiumRaces() {
    try {
        const races = await api('GET', '/api/compendium/races');
        document.getElementById('compRaces').innerHTML = races.map((r) => `
      <div class="card" style="padding:16px;margin-bottom:8px">
        <div class="card-header" style="border:none;padding:0;margin:0"><strong>${esc(r.name)}</strong>
          <span><span class="badge badge-gold">${esc(r.size)}</span> Speed: ${r.speed}</span></div>
        <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">${esc(r.description)}</p></div>`).join('');
    }
    catch { }
}
async function loadCompendiumClasses() {
    try {
        const cls = await api('GET', '/api/compendium/classes');
        document.getElementById('compClasses').innerHTML = cls.map((c) => `
      <div class="card" style="padding:16px;margin-bottom:8px">
        <div class="card-header" style="border:none;padding:0;margin:0"><strong>${esc(c.name)}</strong>
          <span>d${c.hit_die} · ${esc(c.primary_ability)}</span></div>
        <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">${esc(c.description)}</p></div>`).join('');
    }
    catch { }
}
async function loadCompendiumSpells() {
    try {
        const spells = await api('GET', '/api/compendium/spells');
        document.getElementById('compSpells').innerHTML = spells.map((s) => `
      <div class="spell-item"><strong>${esc(s.name)}</strong> <span style="color:var(--text-muted)">Lv${s.level} ${esc(s.school)}</span>
        <div style="font-size:0.85rem;color:var(--text-light)">${esc(s.casting_time)} · ${esc(s.range)} · ${esc(s.duration)}</div></div>`).join('');
    }
    catch { }
}
async function loadCompendiumEquipment() {
    try {
        const items = await api('GET', '/api/compendium/equipment');
        document.getElementById('compEquipment').innerHTML = items.map((i) => `
      <div class="inventory-item"><span class="item-name">${esc(i.name)}</span>
        <span style="color:var(--text-muted)">${esc(i.category)}${i.weight ? ' · ' + i.weight + 'lb' : ''}</span></div>`).join('');
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
// ─── Utils ───
function esc(s) { const d = document.createElement('div'); d.textContent = s; return d.innerHTML; }
function capitalize(s) { return s.charAt(0).toUpperCase() + s.slice(1); }
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