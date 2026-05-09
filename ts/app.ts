export {};

declare const vis: any;
declare const Chart: any;

let csrfToken = '';
let currentUser: { id: number; username: string; role: string } | null = null;
let currentView = 'characters';
let currentChar: any = null;
let currentTab = 'stats';
let allLocations: any[] = [];
let allNPCs: any[] = [];

async function api(method: string, path: string, body?: any): Promise<any> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (csrfToken) headers['X-CSRF-Token'] = csrfToken;
  const opts: RequestInit = { method, headers, credentials: 'include' };
  if (body !== undefined) opts.body = JSON.stringify(body);
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
    document.getElementById('userName')!.textContent = user.username;
    if (user.role === 'admin') {
      (document.getElementById('adminLink') as HTMLElement).style.display = 'inline';
    }
    showView('characters');
    loadCharacters();
    // Preload global lists
    api('GET', '/api/locations').then(l => allLocations = l).catch(() => {});
    api('GET', '/api/npcs').then(n => allNPCs = n).catch(() => {});
  } catch {
    window.location.href = '/login';
  }
}

function showView(view: string) {
  currentView = view;
  document.getElementById('charactersView')!.style.display = view === 'characters' || view === 'sheet' ? 'block' : 'none';
  document.getElementById('sheetView')!.style.display = view === 'sheet' ? 'block' : 'none';
  document.getElementById('diceView')!.style.display = view === 'dice' ? 'block' : 'none';
  document.getElementById('compendiumView')!.style.display = view === 'compendium' ? 'block' : 'none';
  document.getElementById('partyView')!.style.display = view === 'party' ? 'block' : 'none';
}

// ─── Character List ───

async function loadCharacters() {
  try {
    const chars = await api('GET', '/api/characters');
    const grid = document.getElementById('charGrid')!;
    grid.innerHTML = chars.map((c: any) => `
      <div class="character-card" onclick="openChar(${c.id})">
        <div class="char-name">${esc(c.name)}</div>
        <div class="char-detail">${esc(c.race)} ${esc(c.class)} · Level ${c.level}</div>
        <div class="char-hp">HP: ${c.hp_current}/${c.hp_max}</div>
      </div>
    `).join('');
  } catch (e: any) {
    toast(e.message, true);
  }
}

async function openChar(id: number) {
  try {
    currentChar = await api('GET', `/api/characters/${id}`);
    currentTab = 'stats';
    showView('sheet');
    renderSheet();
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).openChar = openChar;

// ─── Character Sheet ───

function renderSheet() {
  if (!currentChar) return;
  const c = currentChar;
  document.getElementById('sheetName')!.textContent = c.name;
  document.getElementById('sheetSubtitle')!.textContent =
    `${c.race} ${c.class}${c.subclass ? ' (' + c.subclass + ')' : ''} · Level ${c.level}`;

  const sections = ['stats', 'combat', 'spells', 'inventory', 'features', 'locations', 'npcs', 'sessions', 'quests', 'journal', 'graph', 'analytics', 'details', 'dice'];
  const tabBar = document.getElementById('tabBar')!;
  tabBar.innerHTML = sections.map(s => `
    <button class="tab ${s === currentTab ? 'active' : ''}" onclick="switchTab('${s}')">${capitalize(s)}</button>
  `).join('');

  sections.forEach(s => {
    const el = document.getElementById(sectionId(s))!;
    el.style.display = s === currentTab ? 'block' : 'none';
  });

  renderStats();
  renderCombat();
  renderSpells();
  renderInventory();
  renderFeatures();
  if (currentTab === 'locations') renderLocations();
  if (currentTab === 'npcs') renderNPCs();
  if (currentTab === 'sessions') renderSessions();
  if (currentTab === 'quests') renderQuests();
  if (currentTab === 'journal') renderJournal();
  if (currentTab === 'graph') renderGraph();
  if (currentTab === 'analytics') renderAnalytics();
  renderDetails();
  renderDiceTab();
}

function sectionId(s: string): string { return s + 'Section'; }

function switchTab(tab: string) {
  currentTab = tab;
  renderSheet();
}
(window as any).switchTab = switchTab;

function renderStats() {
  const c = currentChar;
  const el = document.getElementById('statsSection')!;
  const abils = ['str','dex','con','int','wis','cha'].map(k => ({ key: k, label: k.toUpperCase() }));
  el.innerHTML = `
    <div class="ability-grid">
      ${abils.map(a => {
        const val = (c as any)[a.key];
        const mod = (c as any)[`${a.key}_mod`];
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
    <div id="profsArea">${(c.proficiencies||[]).map((p:any) =>
      `<span class="badge badge-blood" style="margin:2px">${esc(p.name)} (${p.type}) <a href="#" onclick="deleteProf(${p.id});return false" style="color:white;text-decoration:none">×</a></span>`
    ).join('')}</div>
    <button class="btn btn-sm" style="margin-top:8px" onclick="addProf()">+ Add Proficiency</button>
  `;
}

async function updateField(field: string, value: any) {
  if (!currentChar) return;
  currentChar[field] = value;
  try { await api('PUT', `/api/characters/${currentChar.id}`, currentChar); } catch (e: any) { toast(e.message, true); }
}
(window as any).updateField = updateField;

function renderCombat() {
  const c = currentChar;
  const el = document.getElementById('combatSection')!;
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
    <h3>Death Saves</h3>
    <div style="display:flex;gap:16px;margin-bottom:12px">
      <div><strong>Successes:</strong> ${[1,2,3].map(i =>
        `<span class="slot-dot ${i <= (c.death_saves_successes||0) ? 'filled' : 'empty'}" onclick="toggleDeathSave('successes',${i})" style="display:inline-block"></span>`
      ).join('')}</div>
      <div><strong>Failures:</strong> ${[1,2,3].map(i =>
        `<span class="slot-dot ${i <= (c.death_saves_failures||0) ? 'filled' : 'empty'}" onclick="toggleDeathSave('failures',${i})" style="display:inline-block;background:${i<=(c.death_saves_failures||0)?'var(--danger)':'transparent'};border-color:${i<=(c.death_saves_failures||0)?'var(--danger)':'var(--border)'}"></span>`
      ).join('')}</div>
    </div>
    <div class="form-group">
      <label>Concentrating On</label>
      <input value="${esc(c.concentrating_on||'')}" onchange="updateField('concentrating_on',this.value)" placeholder="e.g. Hunter's Mark">
    </div>
  `;
}

function renderSpells() {
  const c = currentChar;
  const el = document.getElementById('spellsSection')!;
  const sc = c.spellcasting;
  const byLevel: Record<number,any[]> = {};
  for (let i=0;i<=9;i++) byLevel[i] = [];
  (c.spells||[]).forEach((s:any) => { if (byLevel[s.level]) byLevel[s.level].push(s); });

  let html = '';
  if (sc && sc.ability) {
    html += `<div class="form-row">
      <div class="form-group"><label>Ability</label><input value="${sc.ability}" onchange="updateSpellcasting('ability',this.value)"></div>
      <div class="form-group"><label>Save DC</label><input type="number" value="${sc.save_dc}" onchange="updateSpellcasting('save_dc',+this.value)"></div>
      <div class="form-group"><label>Attack</label><input type="number" value="${sc.attack_bonus}" onchange="updateSpellcasting('attack_bonus',+this.value)"></div>
    </div><h3>Spell Slots</h3><div class="slot-tracker">`;
    for (let lv=1;lv<=9;lv++) {
      const mx = (sc as any)['slots_'+lv+'_max'];
      const us = (sc as any)['slots_'+lv+'_used'];
      if (mx > 0) {
        html += `<div class="slot-level"><div class="level-label">Lv ${lv}</div><div class="slot-dots">`;
        for (let i=0;i<mx;i++) html += `<div class="slot-dot ${i<us?'filled':'empty'}" onclick="toggleSlot(${lv},${i})"></div>`;
        html += `</div><small>${us}/${mx}</small></div>`;
      }
    }
    html += `</div>`;
  } else {
    html += `<p style="color:var(--text-muted);font-style:italic">No spellcasting. <a href="#" onclick="enableSpellcasting();return false">Set up</a></p>`;
  }
  html += `<h3>Spells</h3><button class="btn btn-sm btn-primary" onclick="addSpell()">+ Add Spell</button><div style="margin-top:8px">`;
  for (let lv=9;lv>=0;lv--) {
    const spells = byLevel[lv]||[];
    if (!spells.length) continue;
    html += `<h4 style="margin-top:12px;color:var(--ink-light);font-size:0.95rem">${lv===0?'Cantrips':'Level '+lv}</h4>`;
    spells.forEach((s:any) => {
      html += `<div class="spell-item ${s.prepared?'prepared':''}">
        <strong>${esc(s.name)}</strong> <span style="color:var(--text-muted)">(${esc(s.school)})</span>
        <span style="float:right">${s.prepared?'✓ Prepared':''}</span>
        <div style="font-size:0.85rem;color:var(--text-light)">${esc(s.casting_time)} · ${esc(s.range)} · ${esc(s.components)} · ${esc(s.duration)}</div>
        <div style="margin-top:4px">${esc(s.description)}</div>
      </div>`;
    });
  }
  html += `</div>`;
  el.innerHTML = html;
}

async function toggleSlot(level:number,index:number) {
  const sc = currentChar.spellcasting;
  const k = 'slots_'+level+'_used';
  sc[k] = index+1 === sc[k] ? index : index+1;
  await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, sc);
  renderSheet();
}
(window as any).toggleSlot = toggleSlot;

async function enableSpellcasting() {
  currentChar.spellcasting = { ability:'int', save_dc:10, attack_bonus:0, slots_1_max:2, slots_1_used:0 };
  await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, currentChar.spellcasting);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSheet();
}
(window as any).enableSpellcasting = enableSpellcasting;

(window as any).toggleDeathSave = async function (field:string, val:number) {
  if (!currentChar) return;
  if (currentChar['death_saves_' + field] === val) {
    currentChar['death_saves_' + field] = val - 1;
  } else {
    currentChar['death_saves_' + field] = val;
  }
  await api('PUT', `/api/characters/${currentChar.id}`, currentChar);
  renderSheet();
};

async function updateSpellcasting(field:string,value:any) {
  currentChar.spellcasting[field] = value;
  await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, currentChar.spellcasting);
}
(window as any).updateSpellcasting = updateSpellcasting;

function renderInventory() {
  const c = currentChar;
  const el = document.getElementById('inventorySection')!;
  const inv = c.inventory||[];
  const cur = c.currency||{cp:0,sp:0,ep:0,gp:0,pp:0};
  el.innerHTML = `
    <h3>Currency</h3>
    <div class="form-row">
      ${['PP','GP','EP','SP','CP'].map(u => `<div class="form-group"><label>${u}</label><input type="number" value="${cur[u.toLowerCase()]}" onchange="updateCurrency('${u.toLowerCase()}',+this.value)"></div>`).join('')}
    </div>
    <div class="ornament">✧</div>
    <div style="display:flex;justify-content:space-between;align-items:center"><h3>Inventory</h3><button class="btn btn-sm btn-primary" onclick="showAddItem()">+ Add</button></div>
    <div style="margin-top:8px">
      ${inv.map((item:any) => `<div class="inventory-item ${item.is_equipped?'equipped':''}">
        <div><span class="item-name">${esc(item.name)}</span> <span class="item-qty">×${item.quantity}</span> <span style="font-size:0.8rem;color:var(--text-muted)">${esc(item.category)}</span>
          ${item.is_equipped?'<span class="badge badge-gold">E</span>':''} ${item.is_magical?'<span class="badge badge-blood">★</span>':''}
        </div>
        <div><span style="font-size:0.85rem;color:var(--text-light)">${item.damage_dice?item.damage_dice+' '+item.damage_type:''} ${item.ac_bonus?'AC+'+item.ac_bonus:''}</span>
          <span style="margin-left:8px">
            <button class="btn btn-sm" onclick="toggleEquip(${item.id})">${item.is_equipped?'Unequip':'Equip'}</button>
            <button class="btn btn-sm btn-danger" onclick="deleteItem(${item.id})">×</button>
          </span>
        </div>
      </div>`).join('')||'<div class="empty-state" style="padding:16px">Empty handed.</div>'}
    </div>`;
}

async function updateCurrency(t:string,v:number) {
  await api('PUT', `/api/characters/${currentChar.id}/currency`, {...currentChar.currency,[t]:v});
}
(window as any).updateCurrency = updateCurrency;

async function toggleEquip(id:number) {
  const item = currentChar.inventory.find((i:any)=>i.id===id);
  if (!item) return; item.is_equipped = !item.is_equipped;
  await api('PUT', `/api/inventory/${id}`, item);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`); renderSheet();
}
(window as any).toggleEquip = toggleEquip;

async function deleteItem(id:number) {
  await api('DELETE', `/api/inventory/${id}`);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`); renderSheet();
}
(window as any).deleteItem = deleteItem;

function renderFeatures() {
  const feats = currentChar.features||[];
  document.getElementById('featuresSection')!.innerHTML = `
    <button class="btn btn-sm btn-primary" onclick="addFeature()">+ Add</button>
    <div style="margin-top:12px">${feats.map((f:any)=>`
      <div class="card" style="padding:16px;margin-bottom:8px">
        <div class="card-header" style="border:none;padding:0;margin:0">
          <strong>${esc(f.name)}</strong>
          <span><span class="badge badge-blood">Lv ${f.level_gained}</span> ${f.source?'<span class="badge badge-gold">'+esc(f.source)+'</span>':''}</span>
        </div>
        <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">${esc(f.description)}</p>
      </div>`).join('')||'<div class="empty-state" style="padding:16px">No features.</div>'}
    </div>`;
}

function renderDetails() {
  const c = currentChar;
  const el = document.getElementById('detailsSection')!;
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
    ${['personality_traits','ideals','bonds','flaws','appearance'].map(f => `
      <div class="form-group"><label>${capitalize(f.replace(/_/g,' '))}</label>
      <textarea onchange="updateField('${f}',this.value)">${esc((c as any)[f])}</textarea></div>
    `).join('')}
    <div class="form-group"><label>Backstory</label>
    <textarea style="min-height:150px" onchange="updateField('backstory',this.value)">${esc(c.backstory)}</textarea></div>`;
}

// ─── Locations ───

async function renderLocations() {
  const el = document.getElementById('locationsSection')!;
  try {
    const links = await api('GET', `/api/characters/${currentChar.id}/locations`);
    el.innerHTML = `
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>Known Locations</h3>
        <button class="btn btn-sm btn-primary" onclick="showLinkLocation()">+ Link Location</button>
      </div>
      ${links.length ? links.map((l:any) => `
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
      ${allLocations.map((l:any) => `
        <div class="inventory-item">
          <div><span class="item-name">${esc(l.name)}</span> <span style="color:var(--text-muted)">(${esc(l.type)})</span></div>
          <div><span style="font-size:0.85rem;color:var(--text-light)">${esc(l.description).substring(0,60)}</span></div>
        </div>`).join('')}`;
  } catch {}
}

(window as any).showLinkLocation = function () {
  const modal = showModal(`
    <h2>Link Location to Character</h2>
    <div class="form-group"><label>Location</label>
      <select id="linkLocId">${allLocations.map((l:any) => `<option value="${l.id}">${esc(l.name)} (${esc(l.type)})</option>`).join('')}</select>
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

(window as any).saveLinkLocation = async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/locations`, {
    location_id: +(document.getElementById('linkLocId') as HTMLSelectElement).value,
    relationship: (document.getElementById('linkLocRel') as HTMLSelectElement).value,
    notes: (document.getElementById('linkLocNotes') as HTMLTextAreaElement).value,
  });
  btn.closest('.modal-overlay')?.remove();
  renderLocations();
};

(window as any).unlinkLocation = async function (id:number) {
  await api('DELETE', `/api/locations/link/${id}`);
  renderLocations();
};

(window as any).showCreateLocation = function () {
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

(window as any).saveNewLocation = async function (btn:HTMLElement) {
  await api('POST', '/api/locations', {
    name: (document.getElementById('newLocName') as HTMLInputElement).value,
    type: (document.getElementById('newLocType') as HTMLSelectElement).value,
    description: (document.getElementById('newLocDesc') as HTMLTextAreaElement).value,
  });
  btn.closest('.modal-overlay')?.remove();
  allLocations = await api('GET', '/api/locations');
  renderLocations();
};

// ─── NPCs ───

async function renderNPCs() {
  const el = document.getElementById('npcsSection')!;
  try {
    const links = await api('GET', `/api/characters/${currentChar.id}/npcs`);
    el.innerHTML = `
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>Related NPCs</h3>
        <button class="btn btn-sm btn-primary" onclick="showLinkNPC()">+ Link NPC</button>
      </div>
      ${links.length ? links.map((n:any) => `
        <div class="inventory-item">
          <div><span class="item-name">${esc(n.npc_name)}</span>
            <span style="color:var(--text-muted)">${esc(n.npc_race)} ${esc(n.npc_class)}</span>
            ${n.npc_is_alive ? '' : '<span class="badge badge-blood">Deceased</span>'}
          </div>
          <div>
            <span class="badge badge-gold">${esc(n.relationship)}</span>
            ${n.interaction_count > 0 ? `<span class="badge badge-blood">${n.interaction_count} interactions</span>` : ''}
            <button class="btn btn-sm" onclick="logNPCInteraction(${n.id})">+ Speak</button>
            <button class="btn btn-sm btn-danger" onclick="unlinkNPC(${n.id})">×</button>
          </div>
        </div>`).join('')
        : '<div class="empty-state" style="padding:16px">No NPCs linked yet.</div>'}
      <hr style="border-color:var(--border-light);margin:16px 0">
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>All Your NPCs</h3>
        <button class="btn btn-sm" onclick="showCreateNPC()">+ New NPC</button>
      </div>
      ${allNPCs.map((n:any) => `
        <div class="inventory-item">
          <div><span class="item-name">${esc(n.name)}</span>
            <span style="color:var(--text-muted)">${esc(n.race)} ${esc(n.class)}</span>
          </div>
          <div style="font-size:0.85rem;color:var(--text-light)">HP: ${n.hp_current}/${n.hp_max}</div>
        </div>`).join('')}`;
  } catch {}
}

(window as any).showLinkNPC = function () {
  const modal = showModal(`
    <h2>Link NPC to Character</h2>
    <div class="form-group"><label>NPC</label>
      <select id="linkNPCId">${allNPCs.map((n:any) => `<option value="${n.id}">${esc(n.name)} (${esc(n.race)} ${esc(n.class)})</option>`).join('')}</select>
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

(window as any).saveLinkNPC = async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/npcs`, {
    npc_id: +(document.getElementById('linkNPCId') as HTMLSelectElement).value,
    relationship: (document.getElementById('linkNPCRel') as HTMLSelectElement).value,
    notes: (document.getElementById('linkNPCNotes') as HTMLTextAreaElement).value,
  });
  btn.closest('.modal-overlay')?.remove();
  renderNPCs();
};

(window as any).logNPCInteraction = async function (id:number) {
  await api('POST', `/api/npcs/link/${id}/interact`, {});
  renderNPCs();
};

(window as any).unlinkNPC = async function (id:number) {
  await api('DELETE', `/api/npcs/link/${id}`);
  renderNPCs();
};

(window as any).showCreateNPC = function () {
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

(window as any).saveNewNPC = async function (btn:HTMLElement) {
  await api('POST', '/api/npcs', {
    name: (document.getElementById('newNPCName') as HTMLInputElement).value,
    race: (document.getElementById('newNPCRace') as HTMLInputElement).value,
    class: (document.getElementById('newNPCClass') as HTMLInputElement).value,
    description: (document.getElementById('newNPCDesc') as HTMLTextAreaElement).value,
  });
  btn.closest('.modal-overlay')?.remove();
  allNPCs = await api('GET', '/api/npcs');
  renderNPCs();
};

// ─── Sessions ───

async function renderSessions() {
  const el = document.getElementById('sessionsSection')!;
  try {
    const sessions = await api('GET', `/api/characters/${currentChar.id}/sessions`);
    el.innerHTML = `
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>Session Log</h3>
        <button class="btn btn-sm btn-primary" onclick="showAddSession()">+ Log Session</button>
      </div>
      <div style="margin-top:12px">
        ${sessions.map((s:any) => `
          <div class="card" style="padding:16px;margin-bottom:8px">
            <div class="card-header" style="border:none;padding:0;margin:0">
              <strong>${esc(s.title) || 'Session ' + s.session_date}</strong>
              <span><span class="badge badge-gold">${s.session_date}</span>
                ${s.xp_earned > 0 ? '<span class="badge badge-blood">+' + s.xp_earned + ' XP</span>' : ''}
                ${s.gold_earned > 0 ? '<span class="badge badge-gold">+' + s.gold_earned + ' GP</span>' : ''}
              </span>
            </div>
            <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px;">${esc(s.notes).substring(0,200)}</p>
            ${s.important_events ? `<p style="font-size:0.85rem;color:var(--text-muted);margin-top:4px"><em>${esc(s.important_events).substring(0,150)}</em></p>` : ''}
            <button class="btn btn-sm btn-danger" style="margin-top:4px" onclick="deleteSession(${s.id})">×</button>
          </div>`).join('') || '<div class="empty-state">No sessions logged yet. Start your campaign!</div>'}
      </div>`;
  } catch {}
}

(window as any).showAddSession = function () {
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

(window as any).saveSession = async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/sessions`, {
    session_date: (document.getElementById('sessDate') as HTMLInputElement).value,
    title: (document.getElementById('sessTitle') as HTMLInputElement).value,
    notes: (document.getElementById('sessNotes') as HTMLTextAreaElement).value,
    xp_earned: +(document.getElementById('sessXP') as HTMLInputElement).value || 0,
    gold_earned: +(document.getElementById('sessGold') as HTMLInputElement).value || 0,
    important_events: (document.getElementById('sessEvents') as HTMLTextAreaElement).value,
  });
  btn.closest('.modal-overlay')?.remove();
  renderSessions();
};

(window as any).deleteSession = async function (id:number) {
  await api('DELETE', `/api/sessions/${id}`);
  renderSessions();
};

// ─── Quests ───

async function renderQuests() {
  const el = document.getElementById('questsSection')!;
  try {
    const quests = await api('GET', `/api/characters/${currentChar.id}/quests`);
    const groups: Record<string,any[]> = { available:[], active:[], complete:[], failed:[], abandoned:[] };
    quests.forEach((q:any) => { if (groups[q.status]) groups[q.status].push(q); });

    let questHtml = '<div style="display:flex;justify-content:space-between;align-items:center"><h3>Quests</h3><button class="btn btn-sm btn-primary" onclick="showAddQuest()">+ New Quest</button></div>';
    const labels: Record<string,string> = { active:'Active', available:'Available', complete:'Complete', failed:'Failed', abandoned:'Abandoned' };

    for (const st of ['active','available','complete','failed','abandoned']) {
      const qs = groups[st]||[];
      if (!qs.length) continue;
      questHtml += '<h4 style="margin-top:12px;color:var(--ink-light)">' + labels[st] + '</h4>';
      for (const q of qs) {
        const opts = ['active','available','complete','failed','abandoned'].map(s =>
          '<option value="' + s + '"' + (s===q.status?' selected':'') + '>' + s + '</option>'
        ).join('');
        questHtml += '<div class="card" style="padding:16px;margin-bottom:8px">';
        questHtml += '<div class="card-header" style="border:none;padding:0;margin:0"><strong>' + esc(q.name) + '</strong>';
        questHtml += '<span><select onchange="updateQuestStatus(' + q.id + ',this.value)" style="width:auto;font-size:0.8rem;padding:2px 6px">' + opts + '</select>';
        questHtml += '<button class="btn btn-sm btn-danger" onclick="deleteQuest(' + q.id + ')">&times;</button></span></div>';
        questHtml += '<p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">' + esc(q.description).substring(0,200) + '</p>';
        if (q.objectives) questHtml += '<div style="font-size:0.85rem;color:var(--text-muted);margin-top:4px"><strong>Objectives:</strong> ' + esc(q.objectives).substring(0,150) + '</div>';
        if (q.rewards) questHtml += '<div style="font-size:0.85rem;color:var(--success);margin-top:2px"><strong>Reward:</strong> ' + esc(q.rewards).substring(0,150) + '</div>';
        questHtml += '</div>';
      }
    }
    if (quests.length === 0) {
      questHtml += '<div class="empty-state">No quests. <a href="#" onclick="showAddQuest();return false">Start one</a></div>';
    }
    el.innerHTML = questHtml;
  } catch {}
}

(window as any).showAddQuest = function () {
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

(window as any).saveQuest = async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/quests`, {
    name: (document.getElementById('questName') as HTMLInputElement).value,
    description: (document.getElementById('questDesc') as HTMLTextAreaElement).value,
    objectives: (document.getElementById('questObj') as HTMLTextAreaElement).value,
    rewards: (document.getElementById('questRewards') as HTMLTextAreaElement).value,
    notes: (document.getElementById('questNotes') as HTMLTextAreaElement).value,
  });
  btn.closest('.modal-overlay')?.remove();
  renderQuests();
};

(window as any).updateQuestStatus = async function (id:number, status:string) {
  const quests = await api('GET', `/api/characters/${currentChar.id}/quests`);
  const q = quests.find((x:any)=>x.id===id);
  if (!q) return;
  q.status = status;
  await api('PUT', `/api/quests/${id}`, q);
  renderQuests();
};

(window as any).deleteQuest = async function (id:number) {
  await api('DELETE', `/api/quests/${id}`);
  renderQuests();
};

// ─── Journal ───

async function renderJournal() {
  const el = document.getElementById('journalSection')!;
  try {
    const entries = await api('GET', `/api/characters/${currentChar.id}/journal`);
    el.innerHTML = `
      <div style="display:flex;justify-content:space-between;align-items:center"><h3>Character Journal</h3>
        <button class="btn btn-sm btn-primary" onclick="showAddJournal()">+ Write Entry</button>
      </div>
      <div style="margin-top:12px">
        ${entries.map((j:any) => `
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
  } catch {}
}

(window as any).showAddJournal = function () {
  const modal = showModal(`
    <h2>Journal Entry</h2>
    <div class="form-group"><label>Date</label><input id="journalDate" type="date" value="${new Date().toISOString().split('T')[0]}"></div>
    <div class="form-group"><label>Title</label><input id="journalTitle" placeholder="Day 1: Arrival in Waterdeep"></div>
    <div class="form-group"><label>Entry</label><textarea id="journalEntry" style="min-height:200px" placeholder="Write your character's thoughts..."></textarea></div>
    <button class="btn btn-primary" onclick="saveJournal(this)">Save</button>
  `);
};

(window as any).saveJournal = async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/journal`, {
    entry_date: (document.getElementById('journalDate') as HTMLInputElement).value,
    title: (document.getElementById('journalTitle') as HTMLInputElement).value,
    entry: (document.getElementById('journalEntry') as HTMLTextAreaElement).value,
  });
  btn.closest('.modal-overlay')?.remove();
  renderJournal();
};

(window as any).deleteJournal = async function (id:number) {
  await api('DELETE', `/api/journal/${id}`);
  renderJournal();
};

// ─── Graph ───

async function renderGraph() {
  const el = document.getElementById('graphSection')!;
  el.innerHTML = `<div class="ornament">✧ Drawing your web of fate ✧</div>
    <div id="graphContainer" style="width:100%;height:600px;border:1px solid var(--border);border-radius:4px;background:var(--parchment-light)"></div>`;

  try {
    const data = await api('GET', `/api/characters/${currentChar.id}/graph`);

    if (typeof vis !== 'undefined') {
      const container = document.getElementById('graphContainer')!;
      const nodes = new vis.DataSet(data.nodes.map((n:any) => ({
        id: n.id, label: n.label, group: n.group,
        color: { background: n.color, border: '#2c1810' },
        font: { face: 'Playfair Display', color: '#2c1810', size: n.size > 20 ? 14 : 11 },
        size: n.size,
        borderWidth: 2,
      })));
      const edges = new vis.DataSet(data.edges.map((e:any) => ({
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
    } else {
      // Fallback: simple text representation
      el.innerHTML += `<div style="padding:20px;font-size:0.9rem">
        <h3>Character Web</h3>
        <p>${data.nodes.map((n:any) => `${n.label} [${n.group}]`).join(' <span style="color:var(--text-muted)">→</span> ')}</p>
        <p style="color:var(--text-muted);font-style:italic;margin-top:8px">
          ${data.nodes.length} connections · ${data.edges.length} relationships</p>
      </div>`;
    }
  } catch (e: any) {
    el.innerHTML += `<div class="empty-state">Could not load graph: ${e.message}</div>`;
  }
}

// ─── Analytics / Statistics ───

async function renderAnalytics() {
  const el = document.getElementById('analyticsSection')!;
  if (!el) return;
  el.innerHTML = '<div class="ornament">✧ Loading analytics... ✧</div>';

  try {
    const stats = await api('GET', `/api/characters/${currentChar.id}/stats`);

    const sc = (n: number) => n === 0 ? 'var(--text-muted)' : 'var(--blood)';
    el.innerHTML = `
      <h3>Campaign Overview</h3>
      <div class="combat-grid" style="margin-bottom:16px">
        <div class="combat-stat"><div class="label">Sessions</div><div class="value">${stats.session_count}</div></div>
        <div class="combat-stat"><div class="label">Level</div><div class="value">${stats.level}</div></div>
        <div class="combat-stat" style="color:var(--success)"><div class="label">Total XP</div><div class="value">${stats.total_xp_earned}</div></div>
        <div class="combat-stat" style="color:var(--gold)"><div class="label">Gold Earned</div><div class="value">${stats.total_gold_earned}</div></div>
      </div>

      <div class="form-row" style="margin-bottom:16px">
        <div class="card" style="padding:16px">
          <h3 style="margin-bottom:8px">Quests (${stats.quests.total})</h3>
          <div style="display:flex;gap:8px;flex-wrap:wrap">
            ${stats.quests.active > 0 ? `<span class="badge badge-blood">${stats.quests.active} Active</span>` : ''}
            ${stats.quests.complete > 0 ? `<span class="badge" style="background:var(--success);color:white">${stats.quests.complete} Complete</span>` : ''}
            ${stats.quests.failed > 0 ? `<span class="badge" style="background:#666;color:white">${stats.quests.failed} Failed</span>` : ''}
            ${stats.quests.available > 0 ? `<span class="badge badge-gold">${stats.quests.available} Available</span>` : ''}
          </div>
        </div>
        <div class="card" style="padding:16px">
          <h3 style="margin-bottom:8px">Rests</h3>
          <div style="display:flex;gap:8px;flex-wrap:wrap">
            <span class="badge badge-gold">${stats.rests.short} Short</span>
            <span class="badge badge-blood">${stats.rests.long} Long</span>
            ${stats.rests.total_healed > 0 ? `<span class="badge" style="background:var(--success);color:white">${stats.rests.total_healed} HP Healed</span>` : ''}
          </div>
        </div>
      </div>

      <div class="form-row" style="margin-bottom:16px">
        <div class="card" style="padding:16px">
          <h3 style="margin-bottom:8px">World</h3>
          <p style="color:var(--text-light)">${stats.locations_count} Locations explored</p>
          <p style="color:var(--text-light)">${stats.npc_interactions} NPC interactions</p>
          <p style="color:var(--text-light)">${stats.journal_count} Journal entries</p>
          <p style="color:var(--text-light)">${stats.dice_rolls.total_rolls} Dice rolls (avg ${stats.dice_rolls.average.toFixed(1)})</p>
        </div>
        <div class="card" style="padding:16px">
          <h3 style="margin-bottom:8px">Notable NPCs</h3>
          ${stats.top_npcs && stats.top_npcs.length > 0
            ? stats.top_npcs.map((n:string) => `<p style="color:var(--text-light)">✦ ${esc(n)}</p>`).join('')
            : '<p style="color:var(--text-muted);font-style:italic">No NPC interactions yet</p>'}
        </div>
      </div>
      <div id="questChartContainer" style="height:200px;max-width:400px;margin:0 auto"></div>`;

    // Draw quest pie chart if Chart.js available
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
  } catch (e: any) {
    el.innerHTML = '<div class="empty-state">Could not load analytics: ' + esc(e.message) + '</div>';
  }
}

// ─── Dice Tab ───

function renderDiceTab() {
  const el = document.getElementById('diceTabSection')!;
  if (!el) return;
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
  const expr = (document.getElementById('diceExpr') as HTMLInputElement).value;
  if (!expr) return;
  try {
    const result = await api('POST', '/api/roll', { expression: expr, character_id: currentChar?.id });
    const el = document.getElementById('diceResult')!;
    el.style.display = 'block';
    el.innerHTML = `<div>${esc(expr)}</div><div style="font-size:1.2rem;color:var(--ink-light)">${esc(result.text)}</div>`;
    loadDiceHistory();
  } catch (e: any) { toast(e.message, true); }
}
(window as any).doRoll = doRoll;

async function loadDiceHistory() {
  const el = document.getElementById('diceHistory');
  if (!el) return;
  try {
    const rolls = await api('GET', '/api/dice-rolls' + (currentChar ? `?character_id=${currentChar.id}` : ''));
    el.innerHTML = rolls.slice(0,20).map((r:any) =>
      `<div class="dice-history-item"><span>${esc(r.expression)}</span><span><strong>${r.total}</strong> <span style="color:var(--text-muted)">${esc(r.result)}</span></span></div>`
    ).join('') || '<div style="text-align:center;color:var(--text-muted);padding:12px">No rolls yet</div>';
  } catch {}
}
(window as any).loadDiceHistory = loadDiceHistory;

// ─── Modals ───

function showModal(html: string): HTMLElement {
  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay';
  overlay.innerHTML = `<div class="modal">${html}</div>`;
  overlay.onclick = (e) => { if (e.target === overlay) overlay.remove(); };
  document.body.appendChild(overlay);
  return overlay.querySelector('.modal') as HTMLElement;
}

(window as any).showAddItem = function () {
  const modal = showModal(`
    <h2>Add Item</h2>
    <div class="form-group"><label>Name</label><input id="itemName" list="equipSuggestions"><datalist id="equipSuggestions"></datalist></div>
    <div class="form-row"><div class="form-group"><label>Qty</label><input id="itemQty" type="number" value="1"></div><div class="form-group"><label>Weight</label><input id="itemWeight" type="number" value="0" step="0.1"></div></div>
    <div class="form-group"><label>Category</label><select id="itemCategory">${['gear','weapon','armor','consumable','tool','magic','ammunition'].map(c=>`<option value="${c}">${c}</option>`).join('')}</select></div>
    <div id="weaponFields" style="display:none"><div class="form-row"><div class="form-group"><label>Damage</label><input id="itemDamage" placeholder="1d8"></div><div class="form-group"><label>Type</label><input id="itemDmgType" placeholder="slashing"></div></div></div>
    <div id="armorFields" style="display:none"><div class="form-group"><label>AC Bonus</label><input id="itemAC" type="number" value="0"></div></div>
    <div class="form-group"><label>Description</label><textarea id="itemDesc"></textarea></div>
    <div style="display:flex;gap:8px"><button class="btn btn-primary" onclick="saveItem(this)">Add</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
  fetch('/api/compendium/equipment',{credentials:'include'}).then(r=>r.json()).then(items => {
    (document.getElementById('equipSuggestions') as HTMLDataListElement).innerHTML = items.map((i:any)=>`<option value="${esc(i.name)}">`).join('');
  }).catch(()=>{});
  const catSel = document.getElementById('itemCategory') as HTMLSelectElement;
  catSel.addEventListener('change', () => {
    document.getElementById('weaponFields')!.style.display = catSel.value === 'weapon' ? 'block' : 'none';
    document.getElementById('armorFields')!.style.display = catSel.value === 'armor' ? 'block' : 'none';
  });
};

(window as any).saveItem = async function (btn:HTMLElement) {
  const item = {
    name: (document.getElementById('itemName') as HTMLInputElement).value,
    quantity: +(document.getElementById('itemQty') as HTMLInputElement).value || 1,
    weight: +(document.getElementById('itemWeight') as HTMLInputElement).value || 0,
    category: (document.getElementById('itemCategory') as HTMLSelectElement).value,
    damage_dice: (document.getElementById('itemDamage') as HTMLInputElement)?.value || '',
    damage_type: (document.getElementById('itemDmgType') as HTMLInputElement)?.value || '',
    weapon_properties: '', ac_bonus: +(document.getElementById('itemAC') as HTMLInputElement)?.value || 0,
    armor_type: '', description: (document.getElementById('itemDesc') as HTMLTextAreaElement).value,
    is_equipped: false, is_magical: false, attunement: false, notes: '',
  };
  if (!item.name) return toast('Name required', true);
  try {
    await api('POST', `/api/characters/${currentChar.id}/inventory`, item);
    btn.closest('.modal-overlay')?.remove();
    currentChar = await api('GET', `/api/characters/${currentChar.id}`); renderSheet();
  } catch (e:any) { toast(e.message, true); }
};

(window as any).addSpell = function () {
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
  fetch('/api/compendium/spells',{credentials:'include'}).then(r=>r.json()).then(spells => {
    (document.getElementById('spellSuggestions') as HTMLDataListElement).innerHTML = spells.map((s:any)=>`<option value="${esc(s.name)}">Lv${s.level} ${esc(s.school)}</option>`).join('');
  }).catch(()=>{});
};

(window as any).saveSpell = async function (btn:HTMLElement) {
  const spell = {
    name: (document.getElementById('spellName') as HTMLInputElement).value,
    level: +(document.getElementById('spellLevel') as HTMLInputElement).value || 0,
    school: (document.getElementById('spellSchool') as HTMLInputElement).value,
    casting_time: (document.getElementById('spellTime') as HTMLInputElement).value,
    range: (document.getElementById('spellRange') as HTMLInputElement).value,
    components: (document.getElementById('spellComp') as HTMLInputElement).value,
    duration: (document.getElementById('spellDur') as HTMLInputElement).value,
    description: (document.getElementById('spellDesc') as HTMLTextAreaElement).value,
    prepared: (document.getElementById('spellPrepared') as HTMLInputElement).checked,
    always_prepared: false, source: '', notes: '',
  };
  if (!spell.name) return toast('Name required', true);
  try {
    await api('POST', `/api/characters/${currentChar.id}/spells`, spell);
    btn.closest('.modal-overlay')?.remove();
    currentChar = await api('GET', `/api/characters/${currentChar.id}`); renderSheet();
  } catch (e:any) { toast(e.message, true); }
};

(window as any).addFeature = function () {
  showModal(`<h2>Add Feature</h2>
    <div class="form-group"><label>Name</label><input id="featName"></div>
    <div class="form-group"><label>Description</label><textarea id="featDesc"></textarea></div>
    <div class="form-row"><div class="form-group"><label>Source</label><input id="featSource" placeholder="e.g. Class"></div><div class="form-group"><label>Level</label><input id="featLevel" type="number" value="1"></div></div>
    <div style="display:flex;gap:8px"><button class="btn btn-primary" onclick="saveFeature(this)">Add</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
};

(window as any).saveFeature = async function (btn:HTMLElement) {
  const feat = {
    name: (document.getElementById('featName') as HTMLInputElement).value,
    description: (document.getElementById('featDesc') as HTMLTextAreaElement).value,
    source: (document.getElementById('featSource') as HTMLInputElement).value,
    level_gained: +(document.getElementById('featLevel') as HTMLInputElement).value || 1,
  };
  if (!feat.name) return toast('Name required', true);
  try {
    await api('POST', `/api/characters/${currentChar.id}/features`, feat);
    btn.closest('.modal-overlay')?.remove();
    currentChar = await api('GET', `/api/characters/${currentChar.id}`); renderSheet();
  } catch (e:any) { toast(e.message, true); }
};

(window as any).addProf = function () {
  showModal(`<h2>Add Proficiency</h2>
    <div class="form-group"><label>Name</label><input id="profName" placeholder="e.g. Perception"></div>
    <div class="form-group"><label>Type</label><select id="profType">${['skill','save','tool','weapon','armor','language','other'].map(t=>`<option value="${t}">${t}</option>`).join('')}</select></div>
    <div style="display:flex;gap:8px"><button class="btn btn-primary" onclick="saveProf(this)">Add</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
};

(window as any).saveProf = async function (btn:HTMLElement) {
  const prof = { character_id: currentChar.id, name: (document.getElementById('profName') as HTMLInputElement).value, type: (document.getElementById('profType') as HTMLSelectElement).value };
  if (!prof.name) return toast('Name required', true);
  try {
    await api('POST', '/api/proficiencies', prof);
    btn.closest('.modal-overlay')?.remove();
    currentChar = await api('GET', `/api/characters/${currentChar.id}`); renderSheet();
  } catch (e:any) { toast(e.message, true); }
};

(window as any).deleteProf = async function (id:number) {
  await api('DELETE', `/api/proficiencies/${id}`);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`); renderSheet();
};

// ─── Print ───

(window as any).printChar = async function () {
  if (!currentChar) return;
  try {
    const res = await fetch(`/api/characters/${currentChar.id}/print`, {
      headers: { 'X-CSRF-Token': csrfToken }, credentials: 'include',
    });
    const text = await res.text();
    const win = window.open('', '_blank');
    if (win) { win.document.write(`<pre style="font-family:monospace;font-size:12px;line-height:1.4">${esc(text)}</pre>`); win.document.close(); win.print(); }
  } catch (e:any) { toast(e.message, true); }
};

// ─── New Character ───

(window as any).newChar = function () {
  const modal = showModal(`
    <h2>New Character</h2>
    <div class="form-group"><label>Name</label><input id="newName" placeholder="Character name"></div>
    <div class="form-row"><div class="form-group"><label>Race</label><input id="newRace" list="raceSuggestions"></div><div class="form-group"><label>Class</label><input id="newClass" list="classSuggestions"></div></div>
    <datalist id="raceSuggestions"></datalist><datalist id="classSuggestions"></datalist>
    <div style="display:flex;gap:8px"><button class="btn btn-primary" onclick="createChar(this)">Create</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
  fetch('/api/compendium/races',{credentials:'include'}).then(r=>r.json()).then(races => {
    (document.getElementById('raceSuggestions') as HTMLDataListElement).innerHTML = races.map((r:any)=>`<option value="${esc(r.name)}">`).join('');
  }).catch(()=>{});
  fetch('/api/compendium/classes',{credentials:'include'}).then(r=>r.json()).then(cls => {
    (document.getElementById('classSuggestions') as HTMLDataListElement).innerHTML = cls.map((c:any)=>`<option value="${esc(c.name)}">`).join('');
  }).catch(()=>{});
};

(window as any).createChar = async function (btn:HTMLElement) {
  const name = (document.getElementById('newName') as HTMLInputElement).value || 'Unnamed';
  const race = (document.getElementById('newRace') as HTMLInputElement).value;
  const cls = (document.getElementById('newClass') as HTMLInputElement).value;
  try {
    const char = await api('POST', '/api/characters', { name, race, class: cls });
    btn.closest('.modal-overlay')?.remove();
    if (char.id) await openChar(char.id);
    loadCharacters();
  } catch (e:any) { toast(e.message, true); }
};

// ─── Import ───

(window as any).showImport = function () {
  showModal(`
    <h2>Import Character</h2>
    <p style="color:var(--text-muted);font-style:italic;margin-bottom:12px">Paste JSON or upload a file</p>
    <div class="form-group"><label>JSON</label><textarea id="importJson" style="min-height:200px;font-family:monospace;font-size:0.8rem"></textarea></div>
    <div class="form-group"><label>File</label><input type="file" id="importFile" accept=".json"></div>
    <div style="display:flex;gap:8px"><button class="btn btn-primary" onclick="doImport()">Import</button><button class="btn" onclick="this.closest('.modal-overlay').remove()">Cancel</button></div>`);
};

(window as any).doImport = async function () {
  const jsonEl = document.getElementById('importJson') as HTMLTextAreaElement;
  const fileEl = document.getElementById('importFile') as HTMLInputElement;
  try {
    let result;
    if (fileEl.files && fileEl.files[0]) {
      const form = new FormData(); form.append('file', fileEl.files[0]);
      const res = await fetch('/api/characters/import', { method:'POST', headers:{'X-CSRF-Token':csrfToken}, credentials:'include', body:form });
      result = await res.json();
    } else if (jsonEl.value.trim()) {
      result = await api('POST', '/api/characters/import', JSON.parse(jsonEl.value));
    } else { toast('Provide JSON or a file', true); return; }
    toast(`Imported ${Array.isArray(result)?result.length:1} character(s)`);
    document.querySelector('.modal-overlay')?.remove();
    loadCharacters();
  } catch (e:any) { toast('Import failed: '+e.message, true); }
};

// ─── Party View ───

(window as any).showParty = async function () {
  showView('party');
  const el = document.getElementById('partyContent')!;
  el.innerHTML = '<div class="ornament">✧ Assembling the party... ✧</div>';
  try {
    const groups: any[] = await api('GET', '/api/party');
    el.innerHTML = groups.map((g: any) => `
      <div class="card" style="margin-bottom:16px">
        <div class="card-header"><strong>${esc(g.name || 'Unnamed Campaign')}</strong>
          <span>${g.members.length} members</span>
        </div>
        <div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(250px,1fr));gap:12px;margin-top:8px">
          ${g.members.map((m: any) => {
            const pct = m.hp_max > 0 ? Math.round((m.hp_current / m.hp_max) * 100) : 0;
            const sc = m.status === 'down' ? 'var(--danger)' : m.status === 'injured' ? 'var(--gold)' : 'var(--success)';
            return '<div class="character-card" onclick="openChar(' + m.id + ')" style="cursor:pointer">' +
              '<div class="char-name">' + esc(m.name) + '</div>' +
              '<div class="char-detail">' + esc(m.race) + ' ' + esc(m.class) + ' &middot; Level ' + m.level + '</div>' +
              '<div style="display:flex;gap:12px;margin-top:8px;font-size:0.85rem;color:var(--text-light)">' +
              '<span>AC: ' + m.ac + '</span>' +
              '<span style="color:' + sc + '">' + esc(m.status) + '</span></div>' +
              '<div class="hp-bar" style="margin-top:6px;height:12px">' +
              '<div class="hp-bar-fill" style="width:' + pct + '%;height:100%"></div>' +
              '<div class="hp-bar-text" style="font-size:0.7rem">' + m.hp_current + '/' + m.hp_max + '</div></div></div>';
          }).join('')}
        </div>
      </div>
    `).join('') || '<div class="empty-state">No characters yet.</div>';
  } catch (e: any) {
    el.innerHTML = '<div class="empty-state">Failed: ' + esc(e.message) + '</div>';
  }
};

// ─── Compendium ───

(window as any).showCompendium = function () {
  showView('compendium'); loadCompendiumRaces();
};

function loadCompendiumTab(tab:string) {
  document.querySelectorAll('.comp-tab').forEach(el => el.classList.remove('active'));
  (document.getElementById('compTab'+capitalize(tab)) as HTMLElement)?.classList.add('active');
  ['races','classes','spells','equipment'].forEach(s => {
    (document.getElementById('comp'+capitalize(s)) as HTMLElement).style.display = s === tab ? 'block' : 'none';
  });
  if (tab === 'races') loadCompendiumRaces();
  if (tab === 'classes') loadCompendiumClasses();
  if (tab === 'spells') loadCompendiumSpells();
  if (tab === 'equipment') loadCompendiumEquipment();
}
(window as any).loadCompendiumTab = loadCompendiumTab;

async function loadCompendiumRaces() {
  try {
    const races = await api('GET','/api/compendium/races');
    document.getElementById('compRaces')!.innerHTML = races.map((r:any)=>`
      <div class="card" style="padding:16px;margin-bottom:8px">
        <div class="card-header" style="border:none;padding:0;margin:0"><strong>${esc(r.name)}</strong>
          <span><span class="badge badge-gold">${esc(r.size)}</span> Speed: ${r.speed}</span></div>
        <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">${esc(r.description)}</p></div>`).join('');
  } catch {}
}

async function loadCompendiumClasses() {
  try {
    const cls = await api('GET','/api/compendium/classes');
    document.getElementById('compClasses')!.innerHTML = cls.map((c:any)=>`
      <div class="card" style="padding:16px;margin-bottom:8px">
        <div class="card-header" style="border:none;padding:0;margin:0"><strong>${esc(c.name)}</strong>
          <span>d${c.hit_die} · ${esc(c.primary_ability)}</span></div>
        <p style="font-size:0.9rem;color:var(--text-light);margin-top:4px">${esc(c.description)}</p></div>`).join('');
  } catch {}
}

async function loadCompendiumSpells() {
  try {
    const spells = await api('GET','/api/compendium/spells');
    document.getElementById('compSpells')!.innerHTML = spells.map((s:any)=>`
      <div class="spell-item"><strong>${esc(s.name)}</strong> <span style="color:var(--text-muted)">Lv${s.level} ${esc(s.school)}</span>
        <div style="font-size:0.85rem;color:var(--text-light)">${esc(s.casting_time)} · ${esc(s.range)} · ${esc(s.duration)}</div></div>`).join('');
  } catch {}
}

async function loadCompendiumEquipment() {
  try {
    const items = await api('GET','/api/compendium/equipment');
    document.getElementById('compEquipment')!.innerHTML = items.map((i:any)=>`
      <div class="inventory-item"><span class="item-name">${esc(i.name)}</span>
        <span style="color:var(--text-muted)">${esc(i.category)}${i.weight?' · '+i.weight+'lb':''}</span></div>`).join('');
  } catch {}
}

// ─── Export ───

(window as any).exportChar = async function () {
  if (!currentChar) return;
  try {
    const data = await api('GET', `/api/characters/${currentChar.id}/export`);
    const blob = new Blob([JSON.stringify(data,null,2)], {type:'application/json'});
    const a = document.createElement('a');
    const url = URL.createObjectURL(blob);
    a.href = url;
    a.download = currentChar.name.replace(/[^a-zA-Z0-9]/g,'_')+'.json';
    a.click(); URL.revokeObjectURL(url);
  } catch (e:any) { toast(e.message, true); }
};

// ─── Utils ───

function esc(s:string):string { const d=document.createElement('div'); d.textContent=s; return d.innerHTML; }
function capitalize(s:string):string { return s.charAt(0).toUpperCase()+s.slice(1); }

function toast(msg:string, isError=false) {
  const el = document.createElement('div');
  el.className = 'toast' + (isError?' error':'');
  el.textContent = msg;
  document.body.appendChild(el);
  setTimeout(() => el.remove(), 4000);
}

(window as any).logout = async function () {
  await api('POST','/api/logout');
  window.location.href = '/login';
};

init();
