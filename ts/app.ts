declare const vis: any;
declare const Chart: any;
declare const bootstrap: any;

(() => {

let csrfToken = '';
let currentUser: { id: number; username: string; role: string } | null = null;
let currentView = 'characters';
let currentChar: any = null;
let currentTab = 'stats';
let allLocations: any[] = [];
let allNPCs: any[] = [];
let loadingCount = 0;

// ─── Utilities ───

function esc(s: string | null | undefined): string {
  if (!s) return '';
  const d = document.createElement('div'); d.textContent = s; return d.innerHTML;
}

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

// ─── FAB ───

function toggleFabMenu() {
  const menu = document.getElementById('fabMenu');
  if (menu) menu.classList.toggle('open');
}
(window as any).toggleFabMenu = toggleFabMenu;

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
(window as any).sortList = sortList;

// ─── Keyboard Shortcuts ───

function initShortcuts() {
  document.addEventListener('keydown', (e) => {
    const target = e.target as HTMLElement;
    const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT';

    if (e.key === 'Escape') {
      hideModal();
      return;
    }

    if (isInput) return;

    if (e.key === '?') {
      showShortcutsHelp();
      return;
    }

    if (e.key === 't' || e.key === 'T') {
      toggleTheme();
      return;
    }

    if (e.key === 'n' && currentView === 'characters') {
      (window as any).newChar();
      return;
    }

    if (e.key === 'd' && currentView !== 'sheet') {
      showView('dice');
      renderDiceTab();
      setTimeout(() => {
        const input = document.getElementById('diceExpr');
        if (input) input.focus();
      }, 100);
      return;
    }

    if (e.key === 'p') {
      (window as any).showParty();
      return;
    }

    if (e.key === 'c') {
      (window as any).showCompendium();
      return;
    }

    if (e.key === '/' && currentView === 'characters') {
      e.preventDefault();
      const search = document.querySelector<HTMLInputElement>('#charSearch');
      if (search) search.focus();
      return;
    }

    if (currentView === 'sheet') {
      const num = parseInt(e.key);
      if (num >= 1 && num <= 9) {
        const idx = num - 1;
        if (idx < sections.length) {
          switchTab(sections[idx]);
        }
      }
    }
  });
}

function showShortcutsHelp() {
  showModal('Keyboard Shortcuts', `
    <div class="shortcut-grid">
      <div class="d-flex justify-content-between py-1"><span><kbd>n</kbd> New Character</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>d</kbd> Dice Roller</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>p</kbd> Party View</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>c</kbd> Compendium</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>/</kbd> Search Characters</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>1</kbd>-<kbd>9</kbd> Sheet Tabs</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>Esc</kbd> Close Modal</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>?</kbd> This Help</span></div>
      <div class="d-flex justify-content-between py-1"><span><kbd>T</kbd> Toggle Theme</span></div>
    </div>
  `);
}
(window as any).showShortcutsHelp = showShortcutsHelp;

// ─── Loading ───

function showLoading() {
  loadingCount++;
  const overlay = document.getElementById('loadingOverlay');
  if (overlay) overlay.classList.remove('d-none');
}
function hideLoading() {
  loadingCount = Math.max(0, loadingCount - 1);
  const overlay = document.getElementById('loadingOverlay');
  if (overlay && loadingCount === 0) overlay.classList.add('d-none');
}

// ─── API ───

async function api(method: string, path: string, body?: any): Promise<any> {
  showLoading();
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (csrfToken) headers['X-CSRF-Token'] = csrfToken;
  const opts: RequestInit = { method, headers, credentials: 'include' };
  if (body !== undefined) opts.body = JSON.stringify(body);
  try {
    const res = await fetch(path, opts);
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }));
      throw new Error(err.error || 'Request failed');
    }
    return res.json();
  } finally {
    hideLoading();
  }
}

// ─── Theme ───

function toggleTheme() {
  const html = document.documentElement;
  const isDark = html.getAttribute('data-theme') === 'dark';
  const newTheme = isDark ? 'light' : 'dark';
  html.setAttribute('data-theme', newTheme);
  localStorage.setItem('villum-theme', newTheme);
  updateThemeIcon();
}
(window as any).toggleTheme = toggleTheme;

function updateThemeIcon() {
  const icon = document.getElementById('themeIcon');
  if (!icon) return;
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
  icon.className = isDark ? 'fa-solid fa-sun' : 'fa-solid fa-moon';
}

function initTheme() {
  const saved = localStorage.getItem('villum-theme') || 'light';
  document.documentElement.setAttribute('data-theme', saved);
  updateThemeIcon();
}

// ─── Bootstrap Modal ───

let modalEl: HTMLElement | null = null;
function getModal(): any {
  if (!modalEl) modalEl = document.getElementById('genericModal');
  let inst = bootstrap.Modal.getInstance(modalEl);
  if (!inst) inst = new bootstrap.Modal(modalEl, { backdrop: true, keyboard: true });
  return inst;
}

function showModal(title: string, bodyHtml: string) {
  const modal = getModal();
  document.getElementById('genericModalTitle')!.textContent = title;
  document.getElementById('genericModalBody')!.innerHTML = bodyHtml;
  modal.show();
}
(window as any).showModal = showModal;

function hideModal() {
  getModal().hide();
}
(window as any).hideModal = hideModal;

// ─── Bootstrap Toast ───

function toast(msg: string, isError = false) {
  const container = document.getElementById('toastContainer')!;
  const id = 'toast-' + Date.now();
  const bg = isError ? 'bg-danger' : 'bg-success';
  container.innerHTML += `
    <div class="toast align-items-center text-white ${bg} border-0 mb-2" id="${id}" role="alert">
      <div class="d-flex">
        <div class="toast-body">${esc(msg)}</div>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
      </div>
    </div>`;
  const el = document.getElementById(id)!;
  new bootstrap.Toast(el, { autohide: true, delay: 5000 }).show();
  setTimeout(() => el.remove(), 6000);
}

// ─── Init ───

async function init() {
  initTheme();
  initShortcuts();
  try {
    const user = await api('GET', '/api/user/me');
    currentUser = user;
    const tokenRes = await api('GET', '/api/csrf-token');
    csrfToken = tokenRes.token;
    document.getElementById('userName')!.textContent = user.username;
    if (user.role === 'admin') {
      document.getElementById('adminNavItem')!.style.display = '';
    }
    showView('characters');
    loadCharacters();
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
(window as any).showView = showView;

// ─── Character List ───

async function loadCharacters() {
  try {
    const chars = await api('GET', '/api/characters');
    const grid = document.getElementById('charGrid')!;
    grid.innerHTML = chars.map((c: any) => `
      <div class="col-md-6 col-lg-4">
        <div class="character-card" onclick="openChar(${c.id})">
          <div class="char-name">${esc(c.name)}</div>
          <div class="char-detail">${esc(c.race)} ${esc(c.class)} · Level ${c.level}</div>
          <div class="char-hp mt-1">HP: ${c.hp_current}/${c.hp_max}</div>
        </div>
      </div>
    `).join('');
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).loadCharacters = loadCharacters;

function filterCharacters() {
  const q = (document.getElementById('charSearch') as HTMLInputElement)?.value?.toLowerCase() || '';
  document.querySelectorAll('#charGrid .character-card').forEach(card => {
    const parent = card.closest('.col-md-6') as HTMLElement;
    if (parent) {
      parent.style.display = !q || card.textContent?.toLowerCase().includes(q) ? '' : 'none';
    }
  });
}
(window as any).filterCharacters = filterCharacters;

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

const sections = ['stats', 'combat', 'spells', 'inventory', 'features', 'locations', 'npcs', 'sessions', 'quests', 'journal', 'graph', 'analytics', 'details', 'dice'];

function renderSheet() {
  if (!currentChar) return;
  const c = currentChar;
  document.getElementById('sheetName')!.textContent = c.name;
  document.getElementById('sheetSubtitle')!.textContent =
    `${c.race} ${c.class}${c.subclass ? ' (' + c.subclass + ')' : ''} · Level ${c.level}`;

  const tabBar = document.getElementById('tabBar')!;
  tabBar.innerHTML = sections.map(s => `
    <li class="nav-item"><button class="nav-link ${s === currentTab ? 'active' : ''}" onclick="switchTab('${s}')">${capitalize(s)}</button></li>
  `).join('');

  sections.forEach(s => {
    const el = document.getElementById(s + 'Section')!;
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

function switchTab(tab: string) {
  currentTab = tab;
  renderSheet();
}
(window as any).switchTab = switchTab;

// ─── Roll / Combat Actions ───

async function rollCheck(type: string, name: string, adv: string) {
  if (!currentChar) return;
  try {
    const result = await api('POST', '/api/roll/check', {
      character_id: currentChar.id, type, name, advantage: adv,
    });
    toast(result.text);
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).rollCheck = rollCheck;

async function applyDamage() {
  if (!currentChar) return;
  const dmg = parseInt((document.getElementById('dmgInput') as HTMLInputElement)?.value || '0');
  if (!dmg) return;
  const newHp = Math.max(0, currentChar.hp_current - dmg);
  await updateField('hp_current', newHp);
}
(window as any).applyDamage = applyDamage;

async function applyHeal() {
  if (!currentChar) return;
  const heal = parseInt((document.getElementById('healInput') as HTMLInputElement)?.value || '0');
  if (!heal) return;
  const newHp = Math.min(currentChar.hp_max, currentChar.hp_current + heal);
  await updateField('hp_current', newHp);
}
(window as any).applyHeal = applyHeal;

async function doRest(type: string) {
  if (!currentChar) return;
  try {
    const result = await api('POST', `/api/characters/${currentChar.id}/rest`, { rest_type: type, hit_dice_count: type === 'short' ? 1 : 0 });
    toast(`${type} rest: healed ${result.hp_healed} HP`);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSheet();
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).doRest = doRest;

async function doLevelUp() {
  if (!currentChar) return;
  try {
    const result = await api('POST', `/api/characters/${currentChar.id}/levelup`);
    toast(`Level Up! Now level ${result.new_level} (+${result.hp_gain} HP)`);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSheet();
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).doLevelUp = doLevelUp;

let autoSaveTimer: number | null = null;

function autoSaveField(field: string, el: HTMLElement) {
  const input = el as HTMLInputElement;
  const isCheckbox = input.type === 'checkbox';
  const isTextarea = el.tagName === 'TEXTAREA';
  const raw = isCheckbox ? input.checked : (el as any).value;
  const num = parseFloat(String(raw));
  const finalVal = !isNaN(num) && !isCheckbox && !isTextarea ? num : raw;
  if (!currentChar) return;
  currentChar[field] = finalVal;
  if (autoSaveTimer) clearTimeout(autoSaveTimer);
  autoSaveTimer = window.setTimeout(async () => {
    try {
      await api('PUT', `/api/characters/${currentChar.id}`, currentChar);
    } catch (e: any) {
      toast(e.message, true);
    }
  }, 800);
}
(window as any).autoSaveField = autoSaveField;

async function updateField(field: string, value: any) {
  if (!currentChar) return;
  currentChar[field] = value;
  try { await api('PUT', `/api/characters/${currentChar.id}`, currentChar); } catch (e: any) { toast(e.message, true); }
}
(window as any).updateField = updateField;

function updateXPBar() {
  const container = document.getElementById('xpBarContainer');
  if (container && currentChar) container.innerHTML = renderXPBar(currentChar);
}
(window as any).updateXPBar = updateXPBar;

// ─── Stats ───

function renderStats() {
  const c = currentChar;
  const el = document.getElementById('statsSection')!;
  const abils = ['str','dex','con','int','wis','cha'].map((k, i) => ({ key: k, label: k.toUpperCase(), desc: ['Strength','Dexterity','Constitution','Intelligence','Wisdom','Charisma'][i], skills: ['Athletics','Acrobatics, Sleight of Hand, Stealth','','Arcana, History, Investigation, Nature, Religion','Animal Handling, Insight, Medicine, Perception, Survival','Deception, Intimidation, Performance, Persuasion'][i] }));
  el.innerHTML = `
    <div class="row g-3">
      ${abils.map(a => {
        const val = (c as any)[a.key];
        const mod = (c as any)[`${a.key}_mod`];
        const cls = mod > 0 ? 'text-success' : mod < 0 ? 'text-danger' : 'text-muted';
        return `<div class="col-4 col-md-2">
          <div class="ability-box" title="${a.desc} (${a.label})\nModifier: ${mod >= 0 ? '+' : ''}${mod}\nSkills: ${a.skills || 'None'}">
            <div class="abil-label" onclick="rollCheck('check','${a.key}','normal')" style="cursor:pointer">${a.label}</div>
            <input type="number" class="form-control form-control-sm text-center abil-value-input" value="${val}" oninput="autoSaveField('${a.key}',this)" onfocus="this.select()">
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
      <div class="col-6 col-md-3"><label class="form-label">Proficiency</label><input type="number" class="form-control form-control-sm" value="${c.proficiency_bonus}" oninput="autoSaveField('proficiency_bonus',this)"></div>
      <div class="col-6 col-md-3"><label class="form-label">Inspiration</label><input type="number" class="form-control form-control-sm" value="${c.inspiration}" oninput="autoSaveField('inspiration',this)"></div>
      <div class="col-6 col-md-3"><label class="form-label">Passive Percep.</label><input type="number" class="form-control form-control-sm" value="${c.passive_perception}" oninput="autoSaveField('passive_perception',this)"></div>
      <div class="col-6 col-md-3"><label class="form-label">XP</label><input type="number" class="form-control form-control-sm" value="${c.xp}" oninput="autoSaveField('xp',this);updateXPBar()"></div>
    </div>
    <div class="mt-2" id="xpBarContainer">${renderXPBar(c)}</div>
    </div>
    <h5 class="mt-3">Skills <small class="text-muted fw-normal">(click to roll)</small></h5>
    <div id="skillsArea">${renderSkills(c)}</div>
    <h5 class="mt-3">Proficiencies</h5>
    <div id="profsArea">${(c.proficiencies||[]).map((p:any) =>
      `<span class="badge badge-blood me-1 mb-1">${esc(p.name)} (${p.type}) <a href="#" onclick="deleteProf(${p.id});return false" class="text-white text-decoration-none">×</a></span>`
    ).join('')}</div>
    <button class="btn btn-sm btn-outline-primary mt-2" onclick="addProf()"><i class="fa-solid fa-plus me-1"></i>Add Proficiency</button>
  `;
}

function renderSkills(c: any) {
  const skls = [
    {name:'Athletics',abil:'str'},{name:'Acrobatics',abil:'dex'},{name:'Sleight of Hand',abil:'dex'},{name:'Stealth',abil:'dex'},
    {name:'Arcana',abil:'int'},{name:'History',abil:'int'},{name:'Investigation',abil:'int'},{name:'Nature',abil:'int'},{name:'Religion',abil:'int'},
    {name:'Animal Handling',abil:'wis'},{name:'Insight',abil:'wis'},{name:'Medicine',abil:'wis'},{name:'Perception',abil:'wis'},{name:'Survival',abil:'wis'},
    {name:'Deception',abil:'cha'},{name:'Intimidation',abil:'cha'},{name:'Performance',abil:'cha'},{name:'Persuasion',abil:'cha'},
  ];
  const profs = (c.proficiencies||[]).filter((p:any) => p.type === 'skill').map((p:any) => p.name.toLowerCase());
  return skls.map(s => {
    const isProf = profs.includes(s.name.toLowerCase());
    const mod = (c as any)[`${s.abil}_mod`];
    const total = isProf ? mod + c.proficiency_bonus : mod;
    const sign = total >= 0 ? '+' : '';
    const breakdown = isProf ? `${s.abil.toUpperCase()} ${mod >= 0 ? '+' : ''}${mod} + Prof ${c.proficiency_bonus} = ${sign}${total}` : `${s.abil.toUpperCase()} ${mod >= 0 ? '+' : ''}${mod} = ${sign}${total}`;
    return `<div class="skill-row d-flex justify-content-between" onclick="rollCheck('skill','${s.name}','normal')" title="${breakdown}">
      <span class="skill-name">${s.name}${isProf ? ' <span class="text-primary">★</span>' : ''}</span>
      <span class="fw-bold">${sign}${total}</span>
    </div>`;
  }).join('');
}

// ─── XP Progress Bar ───

const XP_TABLE = [0, 300, 900, 2700, 6500, 14000, 23000, 34000, 48000, 64000, 85000, 100000, 120000, 140000, 165000, 195000, 225000, 265000, 305000, 355000];

function renderXPBar(c: any) {
  const level = c.level || 1;
  const xp = c.xp || 0;
  const idx = Math.min(level - 1, XP_TABLE.length - 2);
  const currentMilestone = XP_TABLE[idx];
  const nextMilestone = XP_TABLE[idx + 1] || currentMilestone + 10000;
  if (level >= 20) {
    return `<div class="small text-muted fst-italic">Maximum level reached</div>`;
  }
  const progress = nextMilestone > currentMilestone ? Math.min(100, Math.max(0, ((xp - currentMilestone) / (nextMilestone - currentMilestone)) * 100)) : 0;
  return `
    <div class="d-flex justify-content-between small mb-1">
      <span class="text-muted">Level ${level}</span>
      <span class="text-muted">${xp.toLocaleString()} / ${nextMilestone.toLocaleString()} XP</span>
      <span class="text-muted">Level ${level + 1}</span>
    </div>
    <div class="hp-bar" style="height:8px" title="${Math.round(progress)}% to next level">
      <div class="hp-bar-fill" style="width:${progress}%;height:100%;background:linear-gradient(90deg,var(--gold),var(--gold-light))"></div>
    </div>`;
}

// ─── Combat ───

function renderCombat() {
  const c = currentChar;
  const el = document.getElementById('combatSection')!;
  const pct = c.hp_max > 0 ? Math.round((c.hp_current / c.hp_max) * 100) : 0;
  el.innerHTML = `
    <div class="row g-3">
      <div class="col-4"><div class="combat-stat" title="Armor Class — how hard you are to hit"><div class="stat-label">AC</div><div class="stat-value">${c.ac}</div></div></div>
      <div class="col-4"><div class="combat-stat" title="Initiative modifier — added to d20 for turn order"><div class="stat-label">Initiative</div><div class="stat-value">${c.initiative >= 0 ? '+' : ''}${c.initiative}</div></div></div>
      <div class="col-4"><div class="combat-stat" title="Movement speed in feet per round"><div class="stat-label">Speed</div><div class="stat-value">${c.speed}</div></div></div>
    </div>
    <h5 class="mt-3">Hit Points</h5>
    <div class="hp-bar position-relative mb-2" title="${c.hp_current} / ${c.hp_max} HP${c.temp_hp > 0 ? ' (+' + c.temp_hp + ' temporary)' : ''}">
      <div class="hp-bar-fill" style="width:${pct}%"></div>
      <div class="position-absolute top-0 start-0 end-0 bottom-0 d-flex align-items-center justify-content-center text-white small fw-bold" style="font-size:0.8rem">${c.hp_current} / ${c.hp_max}${c.temp_hp > 0 ? ' (+' + c.temp_hp + ' temp)' : ''}</div>
    </div>
    <div class="row g-2">
      <div class="col-4"><label class="form-label small">HP Max</label><input type="number" class="form-control form-control-sm" value="${c.hp_max}" oninput="autoSaveField('hp_max',this)"></div>
      <div class="col-4"><label class="form-label small">Current</label><input type="number" class="form-control form-control-sm" value="${c.hp_current}" oninput="autoSaveField('hp_current',this)"></div>
      <div class="col-4"><label class="form-label small">Temp HP</label><input type="number" class="form-control form-control-sm" value="${c.temp_hp}" oninput="autoSaveField('temp_hp',this)"></div>
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
      ${['str','dex','con','int','wis','cha'].map(a => {
        const mod = (c as any)[`${a}_mod`];
        const total = c.proficiency_bonus + mod;
        const sign = total >= 0 ? '+' : '';
        return `<span class="badge badge-gold" style="cursor:pointer" onclick="rollCheck('save','${a}','normal')">${a.toUpperCase()} ${sign}${total}</span>`;
      }).join('')}
    </div>
    <h5 class="mt-3">Death Saves</h5>
    <div class="row g-2">
      <div class="col-6"><label class="form-label small">Successes</label><input type="number" class="form-control form-control-sm" value="${c.death_save_successes}" oninput="autoSaveField('death_save_successes',this)" min="0" max="3"></div>
      <div class="col-6"><label class="form-label small">Failures</label><input type="number" class="form-control form-control-sm" value="${c.death_save_failures}" oninput="autoSaveField('death_save_failures',this)" min="0" max="3"></div>
    </div>
    <h5 class="mt-3">Concentration</h5>
    <div class="form-check"><input type="checkbox" class="form-check-input" id="concentrationCb" ${c.concentrating ? 'checked' : ''} onchange="autoSaveField('concentrating',this)"><label class="form-check-label" for="concentrationCb">Concentrating on a spell</label></div>
    <div class="mt-2">
      <label class="form-label small">Concentrating On</label>
      <input class="form-control form-control-sm" value="${esc(c.concentrating_on)}" oninput="autoSaveField('concentrating_on',this)" placeholder="e.g. Hunter's Mark">
    </div>
    <h5 class="mt-3">Hit Dice</h5>
    <div class="row g-2">
      <div class="col-6"><label class="form-label small">Total</label><input type="number" class="form-control form-control-sm" value="${c.hit_dice_total}" oninput="autoSaveField('hit_dice_total',this)"></div>
      <div class="col-6"><label class="form-label small">Used</label><input type="number" class="form-control form-control-sm" value="${c.hit_dice_used}" oninput="autoSaveField('hit_dice_used',this)"></div>
    </div>`;
}

// ─── Currency ───

async function updateCurrency() {
  if (!currentChar) return;
  const coins = ['cp','sp','ep','gp','pp'];
  const updates: Record<string,number> = {};
  coins.forEach(c => { updates[c] = +(document.getElementById('coin' + c) as HTMLInputElement)?.value || 0; });
  await api('PUT', `/api/characters/${currentChar.id}/currency`, updates);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  toast('Currency updated');
}
(window as any).updateCurrency = updateCurrency;

// ─── Inventory ───

function renderInventory() {
  const inv = currentChar.inventory || [];
  const categories: Record<string, any[]> = { weapon: [], armor: [], gear: [], potion: [], scroll: [], tool: [], wondrous: [], other: [] };
  inv.forEach((i:any) => { if (categories[i.category]) categories[i.category].push(i); else categories.other.push(i); });
  const total = inv.reduce((s:number,i:any)=>s+(i.weight||0)*(i.quantity||1),0);

  document.getElementById('inventorySection')!.innerHTML = `
    <div class="d-flex justify-content-between align-items-center">
      <h5>Inventory <span class="text-muted small">(Total: ${total} lbs)</span></h5>
      <div><button class="btn btn-primary btn-sm" onclick="addInventory()"><i class="fa-solid fa-plus me-1"></i>Add Item</button></div>
    </div>
    <div class="mt-2" id="invList">
      ${Object.entries(categories).filter(([,items]) => items.length).map(([cat, items]) => `
        <h6 class="mt-3 text-muted">${capitalize(cat)}</h6>
        ${(items as any[]).map((i:any) => `
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
      `).join('') || '<div class="empty-state"><i class="fa-solid fa-backpack fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">Empty Pockets</p><p class="small text-muted">No items yet. Add gear to your inventory.</p></div>'}
    </div>`;
}

(window as any).addInventory = function () {
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

(window as any).editInventory = function (id:number,name:string,qty:number,cat:string,weight:number,equipped:boolean) {
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
};

(window as any).saveInventory = async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/inventory`, {
    name: (document.getElementById('invName') as HTMLInputElement).value,
    quantity: +(document.getElementById('invQty') as HTMLInputElement).value || 1,
    weight: +(document.getElementById('invWeight') as HTMLInputElement).value || 0,
    category: (document.getElementById('invCat') as HTMLSelectElement).value,
  });
  hideModal();
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderInventory();
  toast('Item added');
};

(window as any).saveEditInventory = async function (id:number, btn:HTMLElement) {
  await api('PUT', `/api/inventory/${id}`, {
    name: (document.getElementById('invName') as HTMLInputElement).value,
    quantity: +(document.getElementById('invQty') as HTMLInputElement).value || 1,
    weight: +(document.getElementById('invWeight') as HTMLInputElement).value || 0,
    category: (document.getElementById('invCat') as HTMLSelectElement).value,
    equipped: (document.getElementById('invEquip') as HTMLInputElement).checked,
  });
  hideModal();
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderInventory();
  toast('Item updated');
};

(window as any).deleteInventory = async function (id:number) {
  if (!confirm('Remove this item?')) return;
  await api('DELETE', `/api/inventory/${id}`);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderInventory();
  toast('Item removed');
};

(window as any).toggleEquip = async function (id:number) {
  const item = currentChar.inventory.find((i:any) => i.id === id);
  if (!item) return;
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
  document.getElementById('spellsSection')!.innerHTML = sc.spellcasting_ability ? `
    <h5>Spellcasting</h5>
    <div class="row g-3 mb-3">
      <div class="col-md-4"><label class="form-label">Ability</label><input class="form-control form-control-sm" value="${esc(sc.spellcasting_ability)}" onchange="updateSpellcasting('spellcasting_ability',this.value)"></div>
      <div class="col-md-4"><label class="form-label">Save DC</label><input class="form-control form-control-sm" type="number" value="${sc.spell_save_dc||0}" onchange="updateSpellcasting('spell_save_dc',+this.value)"></div>
      <div class="col-md-4"><label class="form-label">Atk Bonus</label><input class="form-control form-control-sm" type="number" value="${sc.spell_attack_bonus||0}" onchange="updateSpellcasting('spell_attack_bonus',+this.value)"></div>
    </div>
    <h6>Spell Slots</h6>
    <div class="d-flex gap-3 flex-wrap mb-3">
      ${[1,2,3,4,5,6,7,8,9].map(lv => {
        const mx = sc[`slots_${lv}_max`] || 0;
        if (!mx) return '';
        return `<div class="text-center">
          <div class="text-muted small">Lv ${lv}</div>
          <input type="number" class="form-control form-control-sm text-center" style="width:55px" id="slotUse${lv}" value="${sc[`slots_${lv}_used`]||0}" onchange="updateSpellSlot(${lv})" min="0" max="${mx}">
          <div class="text-muted small">/ ${mx}</div>
        </div>`;
      }).join('')}
    </div>
    <div class="d-flex justify-content-between align-items-center mt-3">
      <h6>Known Spells</h6>
      <button class="btn btn-primary btn-sm" onclick="addSpell()"><i class="fa-solid fa-plus me-1"></i>Add Spell</button>
    </div>
    <div class="row g-2 mt-2">
      ${spells.map((s:any) => `
        <div class="col-md-6">
          <div class="card spell-card ${s.prepared ? 'border-gold' : ''}">
            <div class="card-body py-2 px-3">
              <div class="d-flex justify-content-between align-items-start">
                <div>
                  <span class="fw-bold">${esc(s.name)}</span>
                  <span class="badge ${s.level === 0 ? 'badge-muted' : 'badge-blood'} ms-1">${s.level > 0 ? 'Lv' + s.level : 'Cantrip'}</span>
                  <span class="badge badge-gold ms-1">${esc(s.school)}</span>
                </div>
                <div class="d-flex gap-1">
                  <button class="btn btn-sm btn-outline-primary" onclick="editSpell(${s.id},'${esc(s.name)}',${s.level},'${esc(s.school)}',${s.prepared},'${esc(s.components||'')}','${esc(s.range||'')}','${esc(s.casting_time||'')}','${esc(s.duration||'')}','${esc(s.description||'')}')"><i class="fa-solid fa-pen"></i></button>
                  <button class="btn btn-sm btn-outline-danger" onclick="deleteSpell(${s.id})"><i class="fa-solid fa-trash"></i></button>
                </div>
              </div>
              <div class="small text-muted mt-1">
                ${s.casting_time ? `<span class="me-2"><i class="fa-regular fa-clock me-1"></i>${esc(s.casting_time)}</span>` : ''}
                ${s.range ? `<span class="me-2"><i class="fa-solid fa-bullseye me-1"></i>${esc(s.range)}</span>` : ''}
                ${s.duration ? `<span><i class="fa-regular fa-hourglass me-1"></i>${esc(s.duration)}</span>` : ''}
              </div>
              ${s.description ? `<p class="mb-0 mt-1 small text-muted">${esc(s.description).substring(0, 150)}${s.description.length > 150 ? '...' : ''}</p>` : ''}
            </div>
          </div>
        </div>`).join('') || '<div class="empty-state"><i class="fa-solid fa-wand-sparkles fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Spells Known</p><p class="small text-muted">Add spells to your spellbook using the button above.</p></div>'}
    </div>` : `
    <div class="empty-state"><i class="fa-solid fa-wand-sparkles fa-2x mb-2 d-block text-muted"></i>
    <p class="text-muted fst-italic">No spellcasting.</p>
    <button class="btn btn-outline-primary btn-sm" onclick="enableSpellcasting()"><i class="fa-solid fa-magic me-1"></i>Set Up Spellcasting</button></div>`;
}

async function updateSpellcasting(field:string, value:any) {
  if (!currentChar) return;
  const sc = currentChar.spellcasting || {};
  sc[field] = value;
  await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, sc);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSpells();
}
(window as any).updateSpellcasting = updateSpellcasting;

async function updateSpellSlot(level:number) {
  if (!currentChar) return;
  const sc = currentChar.spellcasting || {};
  sc[`slots_${level}_used`] = +(document.getElementById(`slotUse${level}`) as HTMLInputElement).value || 0;
  await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, sc);
}
(window as any).updateSpellSlot = updateSpellSlot;

(window as any).enableSpellcasting = async function () {
  currentChar.spellcasting = {
    spellcasting_ability: 'int', spell_save_dc: 10, spell_attack_bonus: 0,
    slots_1_max: 2, slots_1_used: 0,
  };
  await api('PUT', `/api/characters/${currentChar.id}/spellcasting`, currentChar.spellcasting);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSpells();
};

(window as any).addSpell = function () {
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
};

(window as any).editSpell = function (id:number,name:string,level:number,school:string,prepared:boolean,comp:string,range:string,cast:string,dur:string,desc:string) {
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
};

(window as any).saveSpell = async function (btn:HTMLElement) {
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
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSpells();
  toast('Spell added');
};

(window as any).saveEditSpell = async function (id:number, btn:HTMLElement) {
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
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSpells();
  toast('Spell updated');
};

(window as any).deleteSpell = async function (id:number) {
  if (!confirm('Remove this spell?')) return;
  await api('DELETE', `/api/spells/${id}`);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSpells();
  toast('Spell removed');
};

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

(window as any).addFeature = function () {
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

(window as any).saveFeature = async function (btn:HTMLElement) {
  await api('POST', `/api/characters/${currentChar.id}/features`, {
    name: (document.getElementById('featName') as HTMLInputElement).value,
    description: (document.getElementById('featDesc') as HTMLTextAreaElement).value,
    source: (document.getElementById('featSource') as HTMLInputElement).value,
    level_gained: +(document.getElementById('featLevel') as HTMLInputElement).value || 1,
  });
  hideModal();
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderFeatures();
  toast('Feature added');
};

(window as any).deleteFeature = async function (id:number) {
  await api('DELETE', `/api/features/${id}`);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderFeatures();
  toast('Feature removed');
};

// ─── Proficiencies ───

(window as any).addProf = function () {
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

(window as any).saveProf = async function (btn:HTMLElement) {
  await api('POST', '/api/proficiencies', {
    character_id: currentChar.id,
    type: (document.getElementById('profType') as HTMLSelectElement).value,
    name: (document.getElementById('profName') as HTMLInputElement).value,
  });
  hideModal();
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSheet();
  toast('Proficiency added');
};

(window as any).deleteProf = async function (id:number) {
  await api('DELETE', `/api/proficiencies/${id}`);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSheet();
  toast('Proficiency removed');
};

// ─── Details ───

function renderDetails() {
  const c = currentChar;
  const el = document.getElementById('detailsSection')!;
  el.innerHTML = `
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
    <hr class="my-3">
    ${['personality_traits','ideals','bonds','flaws','appearance'].map(f => `
      <div class="mb-3"><label class="form-label">${capitalize(f.replace(/_/g,' '))}</label>
      <textarea class="form-control form-control-sm" rows="2" oninput="autoSaveField('${f}',this)">${esc((c as any)[f])}</textarea></div>
    `).join('')}
    <div class="mb-3"><label class="form-label">Backstory</label>
    <textarea class="form-control form-control-sm" rows="4" oninput="autoSaveField('backstory',this)">${esc(c.backstory)}</textarea></div>
    <h5 class="mt-4">Currency</h5>
    <div class="row g-3">
      ${['cp','sp','ep','gp','pp'].map(coin => `
        <div class="col-4 col-md-2"><label class="form-label small">${coin.toUpperCase()}</label>
        <input class="form-control form-control-sm" id="coin${coin}" value="${c.currency?.[coin]||0}" type="number"></div>
      `).join('')}
      <div class="col-4 col-md-2 d-flex align-items-end"><button class="btn btn-gold btn-sm w-100" onclick="updateCurrency()">Save</button></div>
    </div>`;
}

// ─── Locations ───

async function renderLocations() {
  const el = document.getElementById('locationsSection')!;
  try {
    const links = await api('GET', `/api/characters/${currentChar.id}/locations`);
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center"><h5>Linked Locations</h5>
        <button class="btn btn-primary btn-sm" onclick="showLinkLocation()"><i class="fa-solid fa-link me-1"></i>Link Location</button>
      </div>
      <div class="mt-2">${links.length ? links.map((l:any) => `
        <div class="inv-item">
          <div><span class="fw-bold">${esc(l.location_name)}</span> <span class="text-muted small">(${esc(l.location_type)})</span>
            ${l.notes ? `<br><small class="text-muted">${esc(l.notes)}</small>` : ''}</div>
          <div><span class="badge badge-gold me-1">${esc(l.relationship)}</span>
            <button class="btn btn-sm btn-outline-danger" onclick="unlinkLocation(${l.id})"><i class="fa-solid fa-trash"></i></button></div>
        </div>`).join('')
        : '<div class="empty-state"><i class="fa-solid fa-map fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Linked Locations</p><p class="small text-muted">Link locations from your campaign to this character.</p></div>'}</div>
      <hr class="my-3">
      <div class="d-flex justify-content-between align-items-center"><h5>All Locations</h5>
        <button class="btn btn-outline-primary btn-sm" onclick="showCreateLocation()"><i class="fa-solid fa-plus me-1"></i>New Location</button>
      </div>
      <div class="mt-2">${allLocations.map((l:any) => `
        <div class="inv-item">
          <div><span class="fw-bold">${esc(l.name)}</span> <span class="text-muted small">(${esc(l.type)})</span>
            <br><small class="text-muted">${esc(l.description).substring(0, 80)}</small></div>
        </div>`).join('')}&nbsp;</div>`;
  } catch { el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load locations. Try again later.</p></div>'; }
}

(window as any).showLinkLocation = function () {
  showModal('Link Location', `
    <div class="mb-3"><label class="form-label">Location</label>
      <select class="form-select" id="linkLocId">${allLocations.map((l:any) => `<option value="${l.id}">${esc(l.name)} (${esc(l.type)})</option>`).join('')}</select></div>
    <div class="mb-3"><label class="form-label">Relationship</label>
      <select class="form-select" id="linkLocRel">
        <option value="current">Current Location</option><option value="hometown">Hometown</option><option value="visited">Visited</option>
        <option value="headquarters">Headquarters</option><option value="quest">Quest Location</option><option value="other">Other</option>
      </select></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="linkLocNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveLinkLocation()"><i class="fa-solid fa-link me-1"></i>Link</button>
  `);
};

(window as any).saveLinkLocation = async function () {
  await api('POST', `/api/characters/${currentChar.id}/locations`, {
    location_id: +(document.getElementById('linkLocId') as HTMLSelectElement).value,
    relationship: (document.getElementById('linkLocRel') as HTMLSelectElement).value,
    notes: (document.getElementById('linkLocNotes') as HTMLTextAreaElement).value,
  });
  hideModal();
  renderLocations();
  toast('Location linked');
};

(window as any).unlinkLocation = async function (id:number) {
  await api('DELETE', `/api/locations/link/${id}`);
  renderLocations();
};

(window as any).showCreateLocation = function () {
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

(window as any).saveNewLocation = async function () {
  await api('POST', '/api/locations', {
    name: (document.getElementById('newLocName') as HTMLInputElement).value,
    type: (document.getElementById('newLocType') as HTMLSelectElement).value,
    description: (document.getElementById('newLocDesc') as HTMLTextAreaElement).value,
  });
  hideModal();
  allLocations = await api('GET', '/api/locations');
  renderLocations();
  toast('Location created');
};

// ─── NPCs ───

async function renderNPCs() {
  const el = document.getElementById('npcsSection')!;
  try {
    const links = await api('GET', `/api/characters/${currentChar.id}/npcs`);
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center"><h5>Related NPCs</h5>
        <button class="btn btn-primary btn-sm" onclick="showLinkNPC()"><i class="fa-solid fa-link me-1"></i>Link NPC</button>
      </div>
      <div class="mt-2">${links.length ? links.map((n:any) => `
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
        : '<div class="empty-state"><i class="fa-solid fa-user-group fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No NPCs Linked</p><p class="small text-muted">Link NPCs to track relationships and interactions.</p></div>'}</div>
      <hr class="my-3">
      <div class="d-flex justify-content-between align-items-center"><h5>All NPCs</h5>
        <button class="btn btn-outline-primary btn-sm" onclick="showCreateNPC()"><i class="fa-solid fa-plus me-1"></i>New NPC</button>
      </div>
      <div class="mt-2">${allNPCs.map((n:any) => `
        <div class="inv-item">
          <div><span class="fw-bold">${esc(n.name)}</span>
            <span class="text-muted small">${esc(n.race)} ${esc(n.class)}</span></div>
          <div class="text-muted small">HP: ${n.hp_current}/${n.hp_max}</div>
        </div>`).join('')}&nbsp;</div>`;
  } catch { el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load NPCs. Try again later.</p></div>'; }
}

(window as any).showLinkNPC = function () {
  showModal('Link NPC', `
    <div class="mb-3"><label class="form-label">NPC</label>
      <select class="form-select" id="linkNPCId">${allNPCs.map((n:any) => `<option value="${n.id}">${esc(n.name)} (${esc(n.race)} ${esc(n.class)})</option>`).join('')}</select></div>
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

(window as any).saveLinkNPC = async function () {
  await api('POST', `/api/characters/${currentChar.id}/npcs`, {
    npc_id: +(document.getElementById('linkNPCId') as HTMLSelectElement).value,
    relationship: (document.getElementById('linkNPCRel') as HTMLSelectElement).value,
    notes: (document.getElementById('linkNPCNotes') as HTMLTextAreaElement).value,
  });
  hideModal();
  renderNPCs();
  toast('NPC linked');
};

(window as any).logNPCInteraction = async function (id:number) {
  await api('POST', `/api/npcs/link/${id}/interact`, {});
  renderNPCs();
  toast('Interaction logged');
};

(window as any).unlinkNPC = async function (id:number) {
  await api('DELETE', `/api/npcs/link/${id}`);
  renderNPCs();
};

(window as any).showCreateNPC = function () {
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

(window as any).saveNewNPC = async function () {
  await api('POST', '/api/npcs', {
    name: (document.getElementById('newNPCName') as HTMLInputElement).value,
    race: (document.getElementById('newNPCRace') as HTMLInputElement).value,
    class: (document.getElementById('newNPCClass') as HTMLInputElement).value,
    description: (document.getElementById('newNPCDesc') as HTMLTextAreaElement).value,
  });
  hideModal();
  allNPCs = await api('GET', '/api/npcs');
  renderNPCs();
  toast('NPC created');
};

// ─── Sessions ───

async function renderSessions() {
  const el = document.getElementById('sessionsSection')!;
  try {
    const sessions = await api('GET', `/api/characters/${currentChar.id}/sessions`);
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center"><h5>Session Log</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddSession()"><i class="fa-solid fa-plus me-1"></i>Log Session</button>
      </div>
      <div class="mt-3">
        ${sessions.map((s:any) => `
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
  } catch { el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load sessions. Try again later.</p></div>'; }
}

(window as any).showAddSession = function () {
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

(window as any).saveSession = async function () {
  await api('POST', `/api/characters/${currentChar.id}/sessions`, {
    session_date: (document.getElementById('sessDate') as HTMLInputElement).value,
    title: (document.getElementById('sessTitle') as HTMLInputElement).value,
    notes: (document.getElementById('sessNotes') as HTMLTextAreaElement).value,
    xp_earned: +(document.getElementById('sessXP') as HTMLInputElement).value || 0,
    gold_earned: +(document.getElementById('sessGold') as HTMLInputElement).value || 0,
    important_events: (document.getElementById('sessEvents') as HTMLTextAreaElement).value,
  });
  hideModal();
  renderSessions();
  toast('Session logged');
};

(window as any).deleteSession = async function (id:number) {
  if (!confirm('Delete this session?')) return;
  await api('DELETE', `/api/sessions/${id}`);
  renderSessions();
  toast('Session deleted');
};

// ─── Quests ───

async function renderQuests() {
  const el = document.getElementById('questsSection')!;
  try {
    const quests = await api('GET', `/api/characters/${currentChar.id}/quests`);
    const groups: Record<string, any[]> = { active: [], available: [], complete: [], failed: [], abandoned: [] };
    quests.forEach((q:any) => { if (groups[q.status]) groups[q.status].push(q); });
    let html = '<div class="d-flex justify-content-between align-items-center"><h5>Quests</h5><button class="btn btn-primary btn-sm" onclick="showAddQuest()"><i class="fa-solid fa-plus me-1"></i>New Quest</button></div>';
    const labels: Record<string,string> = { active: 'Active', available: 'Available', complete: 'Complete', failed: 'Failed', abandoned: 'Abandoned' };
    for (const st of ['active', 'available', 'complete', 'failed', 'abandoned']) {
      const qs = groups[st] || [];
      if (!qs.length) continue;
      html += `<h6 class="mt-3 text-muted">${labels[st]}</h6>`;
      for (const q of qs) {
        const opts = ['active','available','complete','failed','abandoned'].map(s => `<option value="${s}"${s===q.status?' selected':''}>${capitalize(s)}</option>`).join('');
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
    if (quests.length === 0) html += '<div class="empty-state"><i class="fa-solid fa-scroll fa-2x mb-2 d-block text-muted"></i>No quests yet.</div>';
    el.innerHTML = html;
  } catch { el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load quests. Try again later.</p></div>'; }
}

(window as any).showAddQuest = function () {
  showModal('New Quest', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="questName" placeholder="e.g. Retrieve the Lost Artifact"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="questDesc" rows="3"></textarea></div>
    <div class="mb-3"><label class="form-label">Objectives</label><textarea class="form-control" id="questObj" rows="2" placeholder="1. Travel to the Temple\n2. Defeat the guardian\n3. Retrieve the artifact"></textarea></div>
    <div class="mb-3"><label class="form-label">Rewards</label><textarea class="form-control" id="questRewards" rows="2" placeholder="500 XP, +1 Longsword, 200 GP"></textarea></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="questNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveQuest()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
};

(window as any).saveQuest = async function () {
  await api('POST', `/api/characters/${currentChar.id}/quests`, {
    name: (document.getElementById('questName') as HTMLInputElement).value,
    description: (document.getElementById('questDesc') as HTMLTextAreaElement).value,
    objectives: (document.getElementById('questObj') as HTMLTextAreaElement).value,
    rewards: (document.getElementById('questRewards') as HTMLTextAreaElement).value,
    notes: (document.getElementById('questNotes') as HTMLTextAreaElement).value,
  });
  hideModal();
  renderQuests();
  toast('Quest created');
};

(window as any).updateQuestStatus = async function (id:number, status:string) {
  const quests = await api('GET', `/api/characters/${currentChar.id}/quests`);
  const q = quests.find((x:any) => x.id === id);
  if (!q) return;
  q.status = status;
  await api('PUT', `/api/quests/${id}`, q);
  renderQuests();
  toast('Quest status updated');
};

(window as any).deleteQuest = async function (id:number) {
  if (!confirm('Delete this quest?')) return;
  await api('DELETE', `/api/quests/${id}`);
  renderQuests();
  toast('Quest deleted');
};

// ─── Journal ───

async function renderJournal() {
  const el = document.getElementById('journalSection')!;
  try {
    const entries = await api('GET', `/api/characters/${currentChar.id}/journal`);
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center"><h5>Character Journal</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddJournal()"><i class="fa-solid fa-plus me-1"></i>Write Entry</button>
      </div>
      <div class="mt-3">
        ${entries.map((j:any) => `
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
  } catch { el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load journal. Try again later.</p></div>'; }
}

(window as any).showAddJournal = function () {
  showModal('Journal Entry', `
    <div class="mb-3"><label class="form-label">Date</label><input class="form-control" id="journalDate" type="date" value="${new Date().toISOString().split('T')[0]}"></div>
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="journalTitle" placeholder="Day 1: Arrival in Waterdeep"></div>
    <div class="mb-3"><label class="form-label">Entry</label><textarea class="form-control" id="journalEntry" rows="6" placeholder="Write your character's thoughts..."></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveJournal()"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
};

(window as any).saveJournal = async function () {
  await api('POST', `/api/characters/${currentChar.id}/journal`, {
    entry_date: (document.getElementById('journalDate') as HTMLInputElement).value,
    title: (document.getElementById('journalTitle') as HTMLInputElement).value,
    entry: (document.getElementById('journalEntry') as HTMLTextAreaElement).value,
  });
  hideModal();
  renderJournal();
  toast('Journal entry saved');
};

(window as any).deleteJournal = async function (id:number) {
  if (!confirm('Delete this journal entry?')) return;
  await api('DELETE', `/api/journal/${id}`);
  renderJournal();
  toast('Journal entry deleted');
};

// ─── Graph ───

async function renderGraph() {
  const el = document.getElementById('graphSection')!;
  el.innerHTML = `<div class="ornament mb-3">✧ Drawing your web of fate ✧</div>
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
      el.innerHTML += `<div class="p-3 small">
        <h5>Character Web</h5>
        <p>${data.nodes.map((n:any) => `${n.label} [${n.group}]`).join(' &rarr; ')}</p>
        <p class="text-muted fst-italic mt-2">${data.nodes.length} connections &middot; ${data.edges.length} relationships</p></div>`;
    }
  } catch (e:any) {
    el.innerHTML += `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load graph: ${esc(e.message)}</p></div>`;
  }
}

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

// ─── 3D Dice ───

const DICE_PRESETS = ['d4', 'd6', 'd8', 'd10', 'd12', 'd20', 'd100'];

// Pip positions for d6 faces (1-6)
const D6_PIPS: Record<number, Array<{top: string; left: string}>> = {
  1: [{top:'50%',left:'50%'}],
  2: [{top:'25%',left:'25%'},{top:'75%',left:'75%'}],
  3: [{top:'25%',left:'25%'},{top:'50%',left:'50%'},{top:'75%',left:'75%'}],
  4: [{top:'25%',left:'25%'},{top:'25%',left:'75%'},{top:'75%',left:'25%'},{top:'75%',left:'75%'}],
  5: [{top:'25%',left:'25%'},{top:'25%',left:'75%'},{top:'50%',left:'50%'},{top:'75%',left:'25%'},{top:'75%',left:'75%'}],
  6: [{top:'25%',left:'25%'},{top:'25%',left:'75%'},{top:'50%',left:'25%'},{top:'50%',left:'75%'},{top:'75%',left:'25%'},{top:'75%',left:'75%'}],
};

function setDiceExpr(die: string) {
  const input = document.getElementById('diceExpr') as HTMLInputElement;
  const m = input.value.match(/^(\d*)d\d+/);
  const count = m ? m[1] || '1' : '1';
  input.value = count + die;
  doRoll();
}
(window as any).setDiceExpr = setDiceExpr;

async function rollWithAdvantage(isAdv: boolean) {
  const input = document.getElementById('diceExpr') as HTMLInputElement;
  const expr = input.value.trim();
  if (!expr.match(/^\d*d\d+/)) return;
  try {
    const result = await api('POST', '/api/roll', {
      expression: expr,
      character_id: currentChar?.id,
      advantage: isAdv ? 'advantage' : 'disadvantage',
    });
    const container = document.getElementById('dice3dContainer');
    const resultDiv = document.getElementById('diceResult');
    if (container) container.innerHTML = '';
    if (resultDiv) {
      const rolls = result.breakdown?.[0]?.rolls || [];
      const chosen = result.total;
      let badge = '';
      if (rolls.some((r: number) => r === 20)) badge = '<span class="badge bg-success ms-2">Critical Hit!</span>';
      else if (rolls.some((r: number) => r === 1)) badge = '<span class="badge bg-danger ms-2">Critical Fail!</span>';
      resultDiv.style.display = 'block';
      resultDiv.innerHTML = `
        <div class="dice-result-box text-center">
          <div class="roll-expression">${esc(result.expression)} (${isAdv ? 'advantage' : 'disadvantage'})</div>
          <div class="d-flex justify-content-center gap-3 mb-2">
            ${rolls.map((r: number, i: number) => {
              const isChosen = (i === 0 && r === chosen) || (i === 1 && r === chosen);
              return `<span class="die-face" style="${isChosen ? 'border-color:var(--gold);box-shadow:0 0 0 2px var(--gold)' : 'opacity:0.5'}">${r}</span>`;
            }).join('')}
          </div>
          <div class="roll-total-anim">${chosen}</div>
          ${badge}
          <div class="roll-text text-muted">${esc(result.text)}</div>
        </div>`;
      animateDiceRoll(result.breakdown);
    }
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).rollWithAdvantage = rollWithAdvantage;

function renderDiceTab() {
  const targetId = currentView === 'dice' ? 'diceViewSection' : 'diceSection';
  const el = document.getElementById(targetId);
  if (!el) return;
  el.innerHTML = `
    <div class="text-center dice-roller">
      <h5>Dice Roller</h5>
      <div class="row justify-content-center mb-2">
        <div class="col-md-8">
          <label class="form-label">Expression</label>
          <input class="form-control text-center" id="diceExpr" value="1d20" placeholder="e.g. 2d6+3" style="font-size:1.3rem;font-weight:700">
        </div>
      </div>
      <div class="dice-quick-btns mb-3">
        ${DICE_PRESETS.map(d => `<button class="btn btn-sm dice-btn" onclick="setDiceExpr('${d}')">${d}</button>`).join('')}
      </div>
      <div class="mb-3">
        <button class="btn btn-outline-gold btn-sm me-1" onclick="rollWithAdvantage(true)" title="Roll with advantage"><i class="fa-solid fa-angles-up me-1"></i>Advantage</button>
        <button class="btn btn-outline-gold btn-sm" onclick="rollWithAdvantage(false)" title="Roll with disadvantage"><i class="fa-solid fa-angles-down me-1"></i>Disadvantage</button>
      </div>
      <div id="dice3dContainer" class="dice-3d-container"></div>
      <div id="diceResult" class="mb-3" style="display:none"></div>
      <button class="btn btn-gold" onclick="doRoll()"><i class="fa-solid fa-dice me-2"></i>Roll the Bones</button>
      <div class="ornament my-3">✧</div>
      <h5>Recent Rolls</h5>
      <div id="diceHistory"></div>
    </div>`;
  const input = document.getElementById('diceExpr') as HTMLInputElement;
  input.addEventListener('keydown', (e) => { if (e.key === 'Enter') doRoll(); });
  loadDiceHistory();
}
(window as any).renderDiceTab = renderDiceTab;

function rotateToFace(face: number, sides: number): string {
  // Map result value to the correct 3D rotation so it lands face-up
  if (sides === 6) {
    // Standard d6 opposite faces sum to 7
    const rotations: Record<number, string> = {
      1: 'rotateX(0deg) rotateY(0deg)',
      2: 'rotateX(-90deg) rotateY(0deg)',
      3: 'rotateY(90deg) rotateX(0deg)',
      4: 'rotateY(-90deg) rotateX(0deg)',
      5: 'rotateX(90deg) rotateY(0deg)',
      6: 'rotateX(180deg) rotateY(0deg)',
    };
    return rotations[face] || 'rotateX(0deg) rotateY(0deg)';
  }
  return '';
}

function build3DCube(value: number, sides: number, dieLabel: string): string {
  if (sides === 6) {
    const faceValues = [1, 2, 3, 4, 5, 6];
    const faceNames = ['front', 'back', 'right', 'left', 'top', 'bottom'];
    const faces = faceValues.map(v => {
      const pips = D6_PIPS[v] || [];
      const pipHtml = pips.map(p => `<span class="pip" style="top:${p.top};left:${p.left};transform:translate(-50%,-50%)"></span>`).join('');
      return `<div class="dice-3d-face ${faceNames[faceValues.indexOf(v)]}">${pipHtml}</div>`;
    }).join('');
    return `<div class="dice-3d-cube d6" data-sides="6" data-value="${value}" style="transform:${rotateToFace(value, 6)}">${faces}</div>`;
  }
  // 2D fallback for other dice types
  const cls = 'd' + sides;
  return `<div class="dice-2d ${cls}">${value}</div>`;
}

function animateDiceRoll(breakdown: any[]) {
  const container = document.getElementById('dice3dContainer');
  if (!container) return;

  // Build dice with rolling animation
  container.innerHTML = breakdown.map((b: any) => {
    if (!b.rolls || b.rolls.length === 0) return '';
    const sides = parseInt(b.die.replace('d', ''));
    const dieLabel = b.die;
    return b.rolls.map((r: number) => {
      const dieHtml = sides === 6
        ? `<div class="dice-3d-cube d6 rolling" data-sides="6" data-value="${r}">${[1,2,3,4,5,6].map((v,i) => {
            const pips = D6_PIPS[v] || [];
            const pipHtml = pips.map(p => `<span class="pip" style="top:${p.top};left:${p.left};transform:translate(-50%,-50%)"></span>`).join('');
            return `<div class="dice-3d-face ${['front','back','right','left','top','bottom'][i]}">${pipHtml}</div>`;
          }).join('')}</div>`
        : `<div class="dice-2d d${sides} rolling">?</div>`;
      return `<div class="dice-3d-wrapper"><span class="dice-3d-label">${dieLabel}</span>${dieHtml}</div>`;
    }).join('');
  }).join('');
}

function settleDice(breakdown: any[]) {
  const container = document.getElementById('dice3dContainer');
  if (!container) return;

  setTimeout(() => {
    container.innerHTML = breakdown.map((b: any) => {
      if (!b.rolls || b.rolls.length === 0) return '';
      const sides = parseInt(b.die.replace('d', ''));
      const dieLabel = b.die;
      return b.rolls.map((r: number) => {
        const dieHtml = build3DCube(r, sides, dieLabel);

        // Check for crits on d20
        let extraClass = '';
        if (sides === 20) {
          if (r === 20) extraClass = ' dice-crit-success';
          else if (r === 1) extraClass = ' dice-crit-fail';
        }

        const wrapper = document.createElement('div');
        wrapper.className = 'dice-3d-wrapper' + extraClass;
        wrapper.innerHTML = `<span class="dice-3d-label">${dieLabel}</span>${dieHtml}`;
        return wrapper.outerHTML;
      }).join('');
    }).join('');
  }, 900);
}

async function doRoll() {
  const expr = (document.getElementById('diceExpr') as HTMLInputElement).value;
  if (!expr) return;

  // Show rolling animation immediately
  const resultEl = document.getElementById('diceResult')!;
  resultEl.style.display = 'none';
  const container = document.getElementById('dice3dContainer');
  if (container) {
    const m = expr.match(/(\d*)d(\d+)/g);
    if (m) {
      const parts = m.map(s => {
        const [, count, sides] = s.match(/(\d*)d(\d+)/)!;
        return { die: s, count: parseInt(count || '1'), sides: parseInt(sides) };
      });
      const breakdown = parts.flatMap(p =>
        Array.from({ length: p.count }, () => ({ die: p.die, rolls: [0], total: 0 }))
      );
      animateDiceRoll(breakdown);
    }
  }

  try {
    const result = await api('POST', '/api/roll', { expression: expr, character_id: currentChar?.id });
    resultEl.style.display = 'block';

    // Settle dice with actual results
    if (result.breakdown) {
      settleDice(result.breakdown);
    }

    let facesHtml = '';
    if (result.breakdown) {
      facesHtml = result.breakdown.map((b: any) => {
        if (b.rolls) {
          const dieLabel = b.die;
          const rolls = b.rolls.map((r: number) => `<span class="die-face" data-value="${r}">${r}</span>`).join('');
          return `<div class="die-group"><span class="die-label">${dieLabel}:</span> <span class="die-faces">${rolls}</span></div>`;
        }
        return '';
      }).join('');
    }

    // Check for crits on d20
    let critBadge = '';
    if (result.breakdown) {
      for (const b of result.breakdown) {
        if (b.die === 'd20' && b.rolls) {
          for (const r of b.rolls) {
            if (r === 20) critBadge = '<span class="badge bg-success ms-2"><i class="fa-solid fa-bolt me-1"></i>Critical Hit!</span>';
            else if (r === 1) critBadge = '<span class="badge bg-danger ms-2"><i class="fa-solid fa-skull me-1"></i>Critical Fail!</span>';
          }
        }
      }
    }

    // Delay result text to sync with dice animation
    setTimeout(() => {
      resultEl.innerHTML = `
        <div class="dice-result-box">
          <div class="roll-total-anim">${result.total} ${critBadge}</div>
          <div class="roll-expression">${esc(result.expression)}</div>
          <div class="roll-breakdown">${facesHtml}</div>
          <div class="roll-text text-muted small">${esc(result.text)}</div>
        </div>`;
    }, 500);
    loadDiceHistory();
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).doRoll = doRoll;

async function loadDiceHistory() {
  const el = document.getElementById('diceHistory');
  if (!el) return;
  try {
    const rolls = await api('GET', '/api/dice-rolls' + (currentChar ? `?character_id=${currentChar.id}` : ''));
    el.innerHTML = rolls.slice(0, 20).map((r:any) =>
      `<div class="d-flex justify-content-between py-1 border-bottom dice-history-item">
        <span class="small">${esc(r.expression)}</span>
        <span><strong>${r.total}</strong> <span class="text-muted small">${esc(r.result)}</span></span>
      </div>`
    ).join('') || '<div class="text-center text-muted py-3">No rolls yet</div>';
  } catch {}
}
(window as any).loadDiceHistory = loadDiceHistory;

// ─── New Character ───

(window as any).newChar = function () {
  showModal('New Character', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="newName" placeholder="Character name"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Race</label><input class="form-control" id="newRace" list="raceSuggestions"><datalist id="raceSuggestions"></datalist></div>
      <div class="col-6"><label class="form-label">Class</label><input class="form-control" id="newClass" list="classSuggestions"><datalist id="classSuggestions"></datalist></div>
    </div>
    <button class="btn btn-primary w-100" onclick="createChar()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
  fetch('/api/compendium/races', { credentials: 'include' }).then(r => r.json()).then((races:any[]) => {
    document.getElementById('raceSuggestions')!.innerHTML = races.map((r:any) => `<option value="${esc(r.name)}">`).join('');
  }).catch(() => {});
  fetch('/api/compendium/classes', { credentials: 'include' }).then(r => r.json()).then((cls:any[]) => {
    document.getElementById('classSuggestions')!.innerHTML = cls.map((c:any) => `<option value="${esc(c.name)}">`).join('');
  }).catch(() => {});
};

(window as any).createChar = async function () {
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
};

// ─── Import / Export ───

(window as any).showImport = function () {
  showModal('Import Character', `
    <p class="text-muted fst-italic small mb-3">Paste JSON or upload a file</p>
    <div class="mb-3"><label class="form-label">JSON</label><textarea class="form-control" id="importJson" rows="6" style="font-family:monospace;font-size:0.8rem"></textarea></div>
    <div class="mb-3"><label class="form-label">File</label><input class="form-control" type="file" id="importFile" accept=".json"></div>
    <button class="btn btn-primary w-100" onclick="doImport()"><i class="fa-solid fa-file-import me-1"></i>Import</button>
  `);
};

(window as any).doImport = async function () {
  const jsonEl = document.getElementById('importJson') as HTMLTextAreaElement;
  const fileEl = document.getElementById('importFile') as HTMLInputElement;
  try {
    let result;
    if (fileEl.files && fileEl.files[0]) {
      const form = new FormData();
      form.append('file', fileEl.files[0]);
      const res = await fetch('/api/characters/import', { method: 'POST', headers: { 'X-CSRF-Token': csrfToken }, credentials: 'include', body: form });
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
};

(window as any).exportChar = async function () {
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
    if (win) {
      win.document.write(`<pre style="font-family:monospace;font-size:12px;line-height:1.4">${esc(text)}</pre>`);
      win.document.close();
      win.print();
    }
  } catch (e:any) {
    toast(e.message, true);
  }
};

// ─── Party View & Campaign Management ───

(window as any).showParty = async function () {
  showView('party');
  const el = document.getElementById('partyContent')!;
  el.innerHTML = '<div class="ornament mb-3">✧ Assembling the party... ✧</div>';
  try {
    const [groups, campaigns] = await Promise.all([
      api('GET', '/api/party'),
      api('GET', '/api/campaigns'),
    ]);

    const getCampaign = (campaignId: number) => campaigns.find((c: any) => c.id === campaignId);
    const isOwner = (campaignId: number) => { const c = getCampaign(campaignId); return c && c.user_id === currentUser?.id; };
    const isDm = (campaignId: number) => { const c = getCampaign(campaignId); return c && (c.my_role === 'dm' || c.user_id === currentUser?.id); };

    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center mb-3">
        <h1 class="h2 mb-0"><i class="fa-solid fa-flag me-2"></i>Party View</h1>
        <button class="btn btn-gold btn-sm" onclick="showCreateCampaign()"><i class="fa-solid fa-plus me-1"></i>New Campaign</button>
      </div>
      ${groups.map((g:any) => {
        const own = g.id ? isOwner(g.id) : false;
        const dm = g.id ? isDm(g.id) : false;
        const canOpen = (userId: number) => userId === currentUser?.id || currentUser?.role === 'admin' || dm;
        return `<div class="card mb-3">
          <div class="card-header d-flex justify-content-between align-items-center">
            <div>
              <strong>${esc(g.name || 'Unnamed Campaign')}</strong>
              ${g.owner_name ? `<span class="ms-2 small text-muted">DM: ${esc(g.owner_name)}</span>` : ''}
            </div>
            <div class="d-flex align-items-center gap-2">
              <span class="badge badge-gold">${g.members.length} members</span>
              ${g.id && (own || dm) ? `<button class="btn btn-outline-primary btn-sm" onclick="showManageCampaign(${g.id},'${esc(g.name)}')"><i class="fa-solid fa-users-gear"></i></button>` : ''}
              ${g.id && own ? `<button class="btn btn-outline-danger btn-sm" onclick="deleteCampaign(${g.id})"><i class="fa-solid fa-trash"></i></button>` : ''}
            </div>
          </div>
          <div class="card-body">
            <div class="row g-3">
              ${g.members.map((m:any) => {
                const pct = m.hp_max > 0 ? Math.round((m.hp_current / m.hp_max) * 100) : 0;
                const sc = m.status === 'down' ? 'var(--danger)' : m.status === 'injured' ? 'var(--gold)' : 'var(--success)';
                return `<div class="col-md-6 col-lg-4">
                  <div class="character-card" ${canOpen(m.user_id) ? `onclick="openChar(${m.id})"` : ''} style="${canOpen(m.user_id) ? '' : 'cursor:default;opacity:0.75'}">
                    <div class="char-name">${esc(m.name)}</div>
                    <div class="char-detail">${esc(m.race)} ${esc(m.class)} · Level ${m.level}</div>
                    ${m.owner_name && m.owner_name !== currentUser?.username ? `<div class="small text-muted"><i class="fa-solid fa-user me-1"></i>${esc(m.owner_name)}</div>` : ''}
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
        </div>`;
      }).join('') || '<div class="empty-state"><i class="fa-solid fa-flag fa-2x mb-2 d-block text-muted"></i>No characters yet. Create a campaign and add members to build your party!</div>'}`;
  } catch (e:any) {
    el.innerHTML = `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Failed: ${esc(e.message)}</p></div>`;
  }
};

(window as any).showCreateCampaign = function () {
  showModal('Create Campaign', `
    <div class="mb-3"><label class="form-label">Campaign Name</label><input class="form-control" id="newCampaignName"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newCampaignDesc" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="doCreateCampaign()">Create</button>
  `);
};

(window as any).doCreateCampaign = async function () {
  try {
    const name = (document.getElementById('newCampaignName') as HTMLInputElement).value;
    if (!name) { toast('Name required', true); return; }
    await api('POST', '/api/campaigns', { name, description: (document.getElementById('newCampaignDesc') as HTMLTextAreaElement).value });
    hideModal();
    toast('Campaign created');
    (window as any).showParty();
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).showManageCampaign = async function (campaignId: number, name: string) {
  let membersHtml = '<p class="text-muted">Loading members...</p>';
  try {
    const members = await api('GET', `/api/campaigns/${campaignId}/members`);
    membersHtml = members.length
      ? `<ul class="list-group mb-3">${members.map((m: any) => {
          const isDmMember = m.role === 'dm';
          return `<li class="list-group-item d-flex justify-content-between align-items-center">
            <span>
              <i class="fa-solid ${isDmMember ? 'fa-crown text-gold' : 'fa-user'} me-2"></i>
              ${esc(m.username)}
              ${isDmMember ? '<span class="badge badge-gold ms-2">DM</span>' : ''}
            </span>
            <div class="d-flex gap-1">
              ${m.username !== currentUser?.username ? `
                <button class="btn btn-sm ${isDmMember ? 'btn-outline-secondary' : 'btn-outline-gold'}" onclick="doToggleDm(${campaignId}, ${m.user_id}, '${isDmMember ? 'player' : 'dm'}')" title="${isDmMember ? 'Remove DM' : 'Make DM'}">
                  <i class="fa-solid ${isDmMember ? 'fa-user' : 'fa-crown'}"></i>
                </button>
                <button class="btn btn-outline-danger btn-sm" onclick="doRemoveMember(${campaignId}, ${m.user_id})"><i class="fa-solid fa-xmark"></i></button>
              ` : '<span class="text-muted small">(you)</span>'}
            </div>
          </li>`;
        }).join('')}</ul>`
      : '<p class="text-muted mb-3">No members yet. Add players by username.</p>';
  } catch {}
  showModal(`Manage: ${esc(name)}`, `
    ${membersHtml}
    <div class="input-group mb-3">
      <input class="form-control" id="addMemberUsername" placeholder="Username to add">
      <button class="btn btn-gold" onclick="doAddMember(${campaignId})"><i class="fa-solid fa-plus"></i></button>
    </div>
    <div id="userSuggestions" class="mb-2"></div>
    <button class="btn btn-outline-secondary w-100" onclick="(window as any).showParty();hideModal()">Done</button>
  `);
  const input = document.getElementById('addMemberUsername') as HTMLInputElement;
  if (input) {
    input.addEventListener('input', () => searchUsers(input.value));
  }
};

let searchTimeout: any = null;
async function searchUsers(q: string) {
  clearTimeout(searchTimeout);
  if (q.length < 2) { document.getElementById('userSuggestions')!.innerHTML = ''; return; }
  searchTimeout = setTimeout(async () => {
    try {
      const users = await api('GET', `/api/users/search?q=${encodeURIComponent(q)}`);
      const el = document.getElementById('userSuggestions')!;
      el.innerHTML = users.map((u: any) =>
        `<div class="d-flex justify-content-between align-items-center p-1 border-bottom" style="cursor:pointer" onclick="document.getElementById('addMemberUsername')!.value='${esc(u.username)}';el.innerHTML=''">
          <span>${esc(u.username)}</span>
        </div>`
      ).join('');
    } catch {}
  }, 300);
}

(window as any).doAddMember = async function (campaignId: number) {
  const username = (document.getElementById('addMemberUsername') as HTMLInputElement).value.trim();
  if (!username) return;
  try {
    await api('POST', `/api/campaigns/${campaignId}/members`, { username });
    toast('Member added');
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).doToggleDm = async function (campaignId: number, userId: number, newRole: string) {
  try {
    await api('PUT', `/api/campaigns/${campaignId}/members/${userId}`, { role: newRole });
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).doRemoveMember = async function (campaignId: number, userId: number) {
  if (!confirm('Remove this member?')) return;
  try {
    await api('DELETE', `/api/campaigns/${campaignId}/members/${userId}`);
    (window as any).showManageCampaign(campaignId, '');
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).deleteCampaign = async function (campaignId: number) {
  if (!confirm('Delete this campaign? Characters will be unlinked.')) return;
  try {
    await api('DELETE', `/api/campaigns/${campaignId}`);
    toast('Campaign deleted');
    (window as any).showParty();
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── Compendium ───

(window as any).showCompendium = function () {
  showView('compendium');
  loadCompendiumRaces();
};

(window as any).loadCompendiumTab = function (tab:string) {
  document.getElementById('compTabRaces')!.classList.remove('active');
  document.getElementById('compTabClasses')!.classList.remove('active');
  document.getElementById('compTabSpells')!.classList.remove('active');
  document.getElementById('compTabEquipment')!.classList.remove('active');
  const tabEl = document.getElementById('compTab' + capitalize(tab));
  if (tabEl) tabEl.classList.add('active');
  ['races','classes','spells','equipment'].forEach(s => {
    const el = document.getElementById('comp' + capitalize(s));
    if (el) el.style.display = s === tab ? 'block' : 'none';
  });
  if (tab === 'races') loadCompendiumRaces();
  if (tab === 'classes') loadCompendiumClasses();
  if (tab === 'spells') loadCompendiumSpells();
  if (tab === 'equipment') loadCompendiumEquipment();
};

async function loadCompendiumRaces() {
  try {
    const races = await api('GET', '/api/compendium/races');
    document.getElementById('compRaces')!.innerHTML = races.map((r:any) => `
      <div class="card mb-2">
        <div class="card-body py-2 px-3">
          <div class="d-flex justify-content-between"><strong>${esc(r.name)}</strong>
            <span><span class="badge badge-gold">${esc(r.size)}</span> Speed: ${r.speed}</span></div>
          <p class="mb-0 mt-1 small text-muted">${esc(r.description)}</p>
        </div>
      </div>`).join('');
  } catch {}
}

async function loadCompendiumClasses() {
  try {
    const cls = await api('GET', '/api/compendium/classes');
    document.getElementById('compClasses')!.innerHTML = cls.map((c:any) => `
      <div class="card mb-2">
        <div class="card-body py-2 px-3">
          <div class="d-flex justify-content-between"><strong>${esc(c.name)}</strong>
            <span class="text-muted small">d${c.hit_die} · ${esc(c.primary_ability)}</span></div>
          <p class="mb-0 mt-1 small text-muted">${esc(c.description)}</p>
        </div>
      </div>`).join('');
  } catch {}
}

async function loadCompendiumSpells() {
  try {
    const spells = await api('GET', '/api/compendium/spells');
    document.getElementById('compSpells')!.innerHTML = spells.map((s:any) => `
      <div class="inv-item">
        <div><span class="fw-bold">${esc(s.name)}</span> <span class="text-muted small">Lv${s.level} ${esc(s.school)}</span></div>
        <div class="text-muted small">${esc(s.casting_time)} · ${esc(s.range)} · ${esc(s.duration)}</div>
      </div>`).join('');
  } catch {}
}

async function loadCompendiumEquipment() {
  try {
    const items = await api('GET', '/api/compendium/equipment');
    document.getElementById('compEquipment')!.innerHTML = items.map((i:any) => `
      <div class="inv-item">
        <span class="fw-bold">${esc(i.name)}</span>
        <span class="text-muted small">${esc(i.category)}${i.weight ? ' · ' + i.weight + 'lb' : ''}</span>
      </div>`).join('');
  } catch {}
}

// ─── Delete Character ───

(window as any).deleteChar = async function () {
  if (!currentChar) return;
  if (!confirm('Delete this character?')) return;
  try {
    await api('DELETE', `/api/characters/${currentChar.id}`);
    currentChar = null;
    showView('characters');
    loadCharacters();
    toast('Character deleted');
  } catch (e: any) {
    toast(e.message, true);
  }
};

// ─── Logout ───

(window as any).logout = async function () {
  await api('POST', '/api/logout');
  window.location.href = '/login';
};

init();

})();
