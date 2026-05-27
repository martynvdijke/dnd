import * as d3 from 'd3';
import Chart from 'chart.js/auto';
import { marked } from 'marked';
import L from 'leaflet';
import * as bootstrap from 'bootstrap';
(window as any).bootstrap = bootstrap;
import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';

declare const htmx: any;

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

// ─── Global Search ───

function showSearchOverlay() {
  let overlay = document.getElementById('searchOverlay');
  if (!overlay) {
    overlay = document.createElement('div');
    overlay.id = 'searchOverlay';
    overlay.className = 'search-overlay';
    overlay.addEventListener('click', (e) => { if (e.target === overlay) hideSearchOverlay(); });
    document.body.appendChild(overlay);
    const panel = document.createElement('div');
    panel.id = 'searchPanel';
    panel.className = 'search-panel';
    overlay.appendChild(panel);
  }
  overlay.style.display = 'flex';
}

function hideSearchOverlay() {
  const overlay = document.getElementById('searchOverlay');
  if (overlay) overlay.style.display = 'none';
}
(window as any).hideSearchOverlay = hideSearchOverlay;

async function doSearch() {
  const q = (document.getElementById('searchInput') as HTMLInputElement)?.value?.trim();
  if (!q) return;
  try {
    const results = await api('GET', '/api/search?q=' + encodeURIComponent(q));
    let html = '';
    let total = 0;
    const sections: Record<string, { label: string; icon: string; items: any[] }> = {
      characters:  { label: 'Characters',  icon: 'fa-users',     items: results.characters },
      npcs:        { label: 'NPCs',         icon: 'fa-user-group', items: results.npcs },
      notes:       { label: 'Notes',        icon: 'fa-note-sticky', items: results.notes },
      quests:      { label: 'Quests',       icon: 'fa-scroll',    items: results.quests },
      journal:     { label: 'Journal',      icon: 'fa-book-open', items: results.journal },
      sessions:    { label: 'Sessions',     icon: 'fa-calendar',  items: results.sessions },
      campaigns:   { label: 'Campaigns',    icon: 'fa-flag',      items: results.campaigns },
      spells:      { label: 'Spells',       icon: 'fa-wand-sparkles', items: results.spells },
      equipment:   { label: 'Equipment',    icon: 'fa-backpack',  items: results.equipment },
      races:       { label: 'Races',        icon: 'fa-person',    items: results.races },
      classes:     { label: 'Classes',      icon: 'fa-graduation-cap', items: results.classes },
      feats:       { label: 'Feats',        icon: 'fa-star',      items: results.feats },
      backgrounds: { label: 'Backgrounds',  icon: 'fa-address-card', items: results.backgrounds },
    };
    for (const [key, sec] of Object.entries(sections)) {
      if (sec.items.length === 0) continue;
      total += sec.items.length;
      html += `<h6 class="mt-3 mb-2"><i class="fa-solid ${sec.icon} me-2 text-muted"></i>${sec.label} (${sec.items.length})</h6>`;
      for (const item of sec.items) {
        html += `<div class="search-result-item" onclick="navigateSearchResult('${key}',${item.id},'${esc(item.name)}');hideSearchOverlay()">
          <div class="fw-bold small">${esc(item.name)}</div>
          ${item.snippet ? `<div class="text-muted small">${item.snippet}</div>` : ''}
        </div>`;
      }
    }
    if (total === 0) {
      html = `<div class="empty-state"><i class="fa-solid fa-search fa-2x mb-2 d-block text-muted"></i><p class="fw-bold">No Results</p><p class="small text-muted">No matches found for "${esc(q)}".</p></div>`;
    }
    showSearchOverlay();
    const panel = document.getElementById('searchPanel');
    if (panel) {
      panel.innerHTML = `<div class="d-flex justify-content-between align-items-center mb-2"><h5 class="mb-0">Search Results${total > 0 ? ` (${total})` : ''}</h5><button class="btn btn-sm btn-outline-secondary" onclick="hideSearchOverlay()"><i class="fa-solid fa-xmark"></i></button></div>${html}`;
    }
  } catch (e: any) {
    toast(e.message, true);
  }
};
function initSearch() {
  const input = document.getElementById('searchInput');
  if (!input) return;
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') doSearch();
  });
  const btn = document.getElementById('searchBtn');
  if (btn) btn.addEventListener('click', doSearch);
}

(window as any).navigateSearchResult = function (type: string, id: number, name: string) {
  if (type === 'characters') {
    openChar(id);
  } else if (type === 'campaigns') {
    showView('characters');
    toast('Campaign: ' + name);
  } else if (['spells','equipment','races','classes','feats','backgrounds'].includes(type)) {
    (window as any).showCompendium();
  } else if (type === 'npcs') {
    showView('characters');
    toast('NPC: ' + name);
  } else {
    showView('characters');
    toast(name);
  }
};
(window as any).doSearch = doSearch;

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
(window as any).api = api;

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
  document.querySelectorAll('.modal-backdrop').forEach(el => el.remove());
  document.body.classList.remove('modal-open');
  document.body.style.removeProperty('padding-right');
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

// ─── WebSocket ───

let ws: WebSocket | null = null;
let wsReconnectTimer: any = null;

function connectWS() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return;
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  ws = new WebSocket(`${proto}//${window.location.host}/api/ws`);
  ws.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data);
      if (msg.type === 'character_update' && currentView === 'sheet' && currentChar && msg.payload.character_id === currentChar.id) {
        openChar(currentChar.id);
      }
      if (msg.type === 'party_update' && currentView === 'party') {
        (window as any).showParty();
      }
    } catch {}
  };
  ws.onclose = () => {
    ws = null;
    wsReconnectTimer = setTimeout(connectWS, 5000);
  };
  ws.onerror = () => ws?.close();
}

// ─── Init ───

async function init() {
  initTheme();
  initShortcuts();
  initSearch();
  try {
    const user = await api('GET', '/api/user/me');
    currentUser = user;
    const tokenRes = await api('GET', '/api/csrf-token');
    csrfToken = tokenRes.token;
    document.querySelector('meta[name="csrf-token"]')?.setAttribute('content', csrfToken);
    document.body.addEventListener('htmx:configRequest', (e: any) => {
      e.detail.headers['X-CSRF-Token'] = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || csrfToken;
    });
    document.getElementById('userName')!.textContent = user.username;
    if (user.role === 'admin') {
      document.getElementById('adminNavItem')!.style.display = '';
      document.getElementById('combatNavItem')!.style.display = '';
    }
    if (user.role === 'admin' || user.role === 'dm') {
      document.getElementById('oneshotNavItem')!.style.display = '';
      document.getElementById('factionsNavItem')!.style.display = '';
      document.getElementById('shopsNavItem')!.style.display = '';
    }
    showView('characters');
    loadCharacters();
    connectWS();
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
  document.getElementById('encounterView')!.style.display = view === 'encounter' ? 'block' : 'none';
  document.getElementById('timelineView')!.style.display = view === 'timeline' ? 'block' : 'none';
  document.getElementById('singleEncounterView')!.style.display = view === 'singleEncounter' ? 'block' : 'none';
  document.getElementById('combatTrackerView')!.style.display = view === 'combatTracker' ? 'block' : 'none';
  document.getElementById('wikiView')!.style.display = view === 'wiki' ? 'block' : 'none';
  document.getElementById('oneshotView')!.style.display = view === 'oneshot' ? 'block' : 'none';
  document.getElementById('factionsView')!.style.display = view === 'factions' ? 'block' : 'none';
  document.getElementById('shopsView')!.style.display = view === 'shops' ? 'block' : 'none';
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
    (window as any).currentChar = currentChar;
    currentTab = 'stats';
    showView('sheet');
    renderSheet();
  } catch (e: any) {
    toast(e.message, true);
  }
}
(window as any).openChar = openChar;

// ─── Character Sheet ───

const sections = ['stats', 'combat', 'spells', 'inventory', 'features', 'feats', 'companions', 'crafting', 'locations', 'npcs', 'sessions', 'quests', 'journal', 'notes', 'graph', 'analytics', 'details', 'dice'];

function renderSheet() {
  if (!currentChar) return;
  const c = currentChar;
  const multi = c.classes && c.classes.length > 0;
  const classStr = multi
    ? c.classes.map((cc: any) => `${cc.class}${cc.subclass ? ' (' + cc.subclass + ')' : ''} ${cc.level}`).join(' / ')
    : `${c.class}${c.subclass ? ' (' + c.subclass + ')' : ''}`;
  document.getElementById('sheetName')!.innerHTML = c.portrait_url
    ? `<img src="${esc(c.portrait_url)}" class="character-portrait me-2" alt="">${esc(c.name)}`
    : esc(c.name);
  document.getElementById('sheetSubtitle')!.textContent =
    `${c.race} ${classStr} · Level ${c.level}`;

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
  renderGraph();
  renderAnalytics();
  renderCrafting();
  renderDetails();
  renderDiceTab();
}

const htmxTabs = ['spells', 'features', 'feats', 'companions', 'crafting', 'notes'];

function switchTab(tab: string) {
  currentTab = tab;
  renderSheet();
  if (htmxTabs.includes(tab) && currentChar) {
    const el = document.getElementById(tab + 'Section');
    if (el) {
      el.setAttribute('hx-get', `/htmx/${tab}?character_id=${currentChar.id}`);
      el.setAttribute('hx-trigger', 'load');
      el.setAttribute('hx-swap', 'innerHTML');
      el.innerHTML = '<div class="ornament">✧ Loading... ✧</div>';
      htmx.process(el);
    }
  }
  // Client-side rendering for tabs
  if (currentChar) {
    switch (tab) {
      case 'inventory': renderInventory(); break;
      case 'locations': renderLocations(); break;
      case 'npcs': renderNPCs(); break;
      case 'sessions': renderSessions(); break;
      case 'quests': renderQuests(); break;
      case 'journal': renderJournal(); break;
    }
  }
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

(window as any).applyDamage = async function () {
  if (!currentChar) return;
  const dmg = parseInt((document.getElementById('dmgInput') as HTMLInputElement)?.value || '0');
  if (!dmg) return;
  const newHp = Math.max(0, currentChar.hp_current - dmg);
  await updateField('hp_current', newHp);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
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
};

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
    <!-- Passive Investigation & Insight -->
    <div class="row g-2 mt-2">
      <div class="col-4 col-md-2">
        <div class="passive-score-box" title="10 + WIS modifier + proficiency if proficient">
          <div class="score-value">${(c.wis_mod||0) + ((c.proficiencies||[]).some((p:any)=>p.name==='Insight')?c.proficiency_bonus:0) + 10}</div>
          <div class="score-label">Passive Insight</div>
          <div class="score-breakdown">10 + ${c.wis_mod||0} WIS${(c.proficiencies||[]).some((p:any)=>p.name==='Insight')?' + '+c.proficiency_bonus+' Prof':''}</div>
        </div>
      </div>
      <div class="col-4 col-md-2">
        <div class="passive-score-box" title="10 + INT modifier + proficiency if proficient">
          <div class="score-value">${(c.int_mod||0) + ((c.proficiencies||[]).some((p:any)=>p.name==='Investigation')?c.proficiency_bonus:0) + 10}</div>
          <div class="score-label">Passive Investigation</div>
          <div class="score-breakdown">10 + ${c.int_mod||0} INT${(c.proficiencies||[]).some((p:any)=>p.name==='Investigation')?' + '+c.proficiency_bonus+' Prof':''}</div>
        </div>
      </div>
      <div class="col-4 col-md-3">
        <div class="exhaustion-display">
          <span class="exhaustion-level ex-${c.exhaustion_level||0}">${c.exhaustion_level||0}</span>
          <div>
            <div class="exhaustion-label">Exhaustion</div>
            <div class="exhaustion-effect">${['-','Disadvantage on ability checks','Speed halved','Disadvantage on attacks & saves','HP max halved','Speed reduced to 0','Death'][c.exhaustion_level||0]||''}</div>
          </div>
          <div class="ms-auto d-flex gap-1">
            <button class="exhaustion-btn" onclick="adjustExhaustion(-1)" title="Reduce exhaustion">−</button>
            <button class="exhaustion-btn" onclick="adjustExhaustion(1)" title="Increase exhaustion">+</button>
          </div>
        </div>
      </div>
    </div>
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

// ─── Exhaustion ───

(window as any).adjustExhaustion = async function (delta: number) {
  if (!currentChar) return;
  const newLevel = Math.max(0, Math.min(6, (currentChar.exhaustion_level || 0) + delta));
  await api('PATCH', `/api/characters/${currentChar.id}/exhaustion`, { exhaustion_level: newLevel });
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderStats();
};

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
  const c = currentChar;
  const categories: Record<string, any[]> = { weapon: [], armor: [], gear: [], potion: [], scroll: [], tool: [], wondrous: [], other: [] };
  inv.forEach((i:any) => { if (categories[i.category]) categories[i.category].push(i); else categories.other.push(i); });
  const total = inv.reduce((s:number,i:any)=>s+(i.weight||0)*(i.quantity||1),0);

  // Encumbrance: STR x5 (light), x10 (encumbered), x15 (heavy)
  const str = c.str || 10;
  const lightMax = str * 5;
  const encumberedMax = str * 10;
  const heavyMax = str * 15;
  const encPct = heavyMax > 0 ? Math.min(100, (total / heavyMax) * 100) : 0;
  let encState = 'light';
  let encLabel = 'Light Load';
  if (total > heavyMax) { encState = 'over'; encLabel = 'Over Capacity'; }
  else if (total > encumberedMax) { encState = 'heavy'; encLabel = 'Heavily Encumbered'; }
  else if (total > lightMax) { encState = 'encumbered'; encLabel = 'Encumbered'; }

  // Attunement count (equipped items with attunement flag)
  const attuneItems = inv.filter((i:any) => i.equipped && i.attunement);
  const attuneCount = attuneItems.length;
  let attuneState = 'attune-ok';
  if (attuneCount >= 3) attuneState = 'attune-full';
  else if (attuneCount >= 2) attuneState = 'attune-warn';

  // Equipped items grouped for loadout
  const equipped = inv.filter((i:any) => i.equipped);
  const loadoutGroups: Record<string, any[]> = { weapon: [], armor: [], shield: [], ring: [], wondrous: [], other: [] };
  equipped.forEach((i:any) => {
    if (i.category === 'weapon') loadoutGroups.weapon.push(i);
    else if (i.category === 'armor' && i.ac_bonus > 0 && i.name.toLowerCase().includes('shield')) loadoutGroups.shield.push(i);
    else if (i.category === 'armor') loadoutGroups.armor.push(i);
    else if (i.category === 'wondrous') loadoutGroups.wondrous.push(i);
    else loadoutGroups.other.push(i);
  });

  document.getElementById('inventorySection')!.innerHTML = `
    <div class="d-flex justify-content-between align-items-center">
      <h5>Inventory <span class="text-muted small">(Total: ${total} / ${heavyMax} lbs)</span></h5>
      <div class="d-flex gap-2 align-items-center">
        <span class="attune-counter ${attuneState}" title="Attuned items">🔗 ${attuneCount}/3</span>
        <button class="btn btn-primary btn-sm" onclick="addInventory()"><i class="fa-solid fa-plus me-1"></i>Add Item</button>
      </div>
    </div>
    <div class="encumbrance-bar" title="${total} lbs / ${heavyMax} lbs max">
      <div class="encumbrance-bar-fill enc-${encState}" style="width:${encPct}%"></div>
    </div>
    <div class="d-flex justify-content-between">
      <span class="encumbrance-state enc-${encState}">${esc(encLabel)}</span>
      <span class="text-muted small">${encLabel === 'Light Load' ? 'Speed normal' : encLabel === 'Encumbered' ? 'Speed -10' : encLabel === 'Heavily Encumbered' ? 'Speed -20, Disadvantage on checks' : 'Speed 0'}</span>
    </div>
    <!-- Loadout Panel -->
    <div class="loadout-panel mt-2">
      <div class="loadout-header" onclick="this.nextElementSibling.classList.toggle('d-none')">
        <h6><i class="fa-solid fa-shield-halved me-1"></i>Loadout (${equipped.length} equipped)</h6>
        <span class="text-muted small"><i class="fa-solid fa-chevron-down"></i></span>
      </div>
      <div class="loadout-body">
        ${Object.entries(loadoutGroups).filter(([,items]) => items.length).map(([cat, items]) => `
          <div class="loadout-category">
            <div class="loadout-category-label">${capitalize(cat)}</div>
            ${(items as any[]).map((i:any) => `
              <div class="loadout-item">
                <span class="item-name">${esc(i.name)}</span>
                <span class="item-detail">${i.damage_dice ? esc(i.damage_dice) + (i.damage_type ? ' ' + esc(i.damage_type) : '') : i.ac_bonus > 0 ? 'AC +' + i.ac_bonus : ''}</span>
              </div>
            `).join('')}
          </div>
        `).join('') || '<div class="text-muted small fst-italic">No items equipped.</div>'}
      </div>
    </div>
    <div class="mt-2" id="invList">
      ${Object.entries(categories).filter(([,items]) => items.length).map(([cat, items]) => `
        <h6 class="mt-3 text-muted">${capitalize(cat)}</h6>
        ${(items as any[]).map((i:any) => `
          <div class="inv-item${i.equipped ? ' equipped' : ''}${i.is_identified === false ? ' unidentified' : ''}">
            <div>
              <span class="fw-bold">${esc(i.name)}</span>
              ${i.quantity > 1 ? `<span class="badge badge-muted">x${i.quantity}</span>` : ''}
              ${i.equipped ? '<span class="badge badge-gold">Equipped</span>' : ''}
              ${i.attunement ? '<span class="badge-attunement" title="Requires Attunement">Attune</span>' : ''}
              ${i.is_identified === false ? '<span class="badge-unidentified">Unidentified</span>' : ''}
              ${i.damage_dice && (i.is_identified !== false) ? `<span class="badge badge-blood ms-1">${esc(i.damage_dice)} ${esc(i.damage_type)}</span>` : ''}
              ${i.ac_bonus > 0 && (i.is_identified !== false) ? `<span class="badge badge-gold ms-1">AC+${i.ac_bonus}</span>` : ''}
              ${i.is_identified === false && i.damage_dice ? `<span class="badge badge-muted ms-1">???</span>` : ''}
            </div>
            <div class="d-flex gap-1">
              ${i.is_identified === false ? `<button class="btn-identify" onclick="toggleIdentify(${i.id})" title="Identify item">🔍 ID</button>` : ''}
              ${i.magic && i.is_identified !== false ? `<button class="btn-identify" onclick="toggleIdentify(${i.id})" title="Mark unidentified">🔮</button>` : ''}
              <button class="btn btn-sm btn-outline-primary" onclick="editInventory(${i.id},'${esc(i.name)}',${i.quantity},'${esc(i.category)}',${i.weight},${i.equipped})" title="Edit"><i class="fa-solid fa-pen"></i></button>
              <button class="btn btn-sm btn-outline-secondary" onclick="toggleEquip(${i.id})" title="${i.equipped ? 'Unequip' : 'Equip'}"><i class="fa-solid fa-shield-halved"></i></button>
              <button class="btn btn-sm btn-outline-danger" onclick="deleteInventory(${i.id})" title="Remove"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>`).join('')}
      `).join('') || '<div class="empty-state"><i class="fa-solid fa-backpack fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">Empty Pockets</p><p class="small text-muted">No items yet. Add gear to your inventory.</p></div>'}
    </div>`;
}

// ─── Identify Toggle ───

(window as any).toggleIdentify = async function (id: number) {
  const item = currentChar.inventory.find((i:any) => i.id === id);
  if (!item) return;
  const newVal = item.is_identified === false ? true : false;
  await api('PUT', `/api/inventory/${id}`, { ...item, is_identified: newVal });
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderInventory();
  toast(newVal ? 'Item identified' : 'Item marked unidentified');
};

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
  document.getElementById('spellsSection')!.innerHTML = sc.ability ? `
    <h5>Spellcasting</h5>
    <div class="row g-3 mb-3">
      <div class="col-md-4"><label class="form-label">Ability</label><input class="form-control form-control-sm" value="${esc(sc.ability)}" onchange="updateSpellcasting('ability',this.value)"></div>
      <div class="col-md-4"><label class="form-label">Save DC</label><input class="form-control form-control-sm" type="number" value="${sc.save_dc||0}" onchange="updateSpellcasting('save_dc',+this.value)"></div>
      <div class="col-md-4"><label class="form-label">Atk Bonus</label><input class="form-control form-control-sm" type="number" value="${sc.attack_bonus||0}" onchange="updateSpellcasting('attack_bonus',+this.value)"></div>
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
      <h6>Known Spells <span class="text-muted small fw-normal">${spells.filter((s:any)=>s.prepared).length}/${spells.filter((s:any)=>s.level>0 && spells.filter((ss:any)=>ss.level===s.level).length > 0).length + 3} prepared</span></h6>
      <div class="d-flex gap-2">
        <button class="btn btn-sm btn-outline-gold" onclick="showPrepareSpells()"><i class="fa-solid fa-book-open me-1"></i>Prepare Spells</button>
        <button class="btn btn-primary btn-sm" onclick="addSpell()"><i class="fa-solid fa-plus me-1"></i>Add Spell</button>
      </div>
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
    ability: 'int', save_dc: 10, attack_bonus: 0,
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

// ─── Spell Preparation Modal ───

(window as any).showPrepareSpells = function () {
  const spells = currentChar.spells || [];
  const sc = currentChar.spellcasting || {};
  const maxPrepared = currentChar.level > 0 ? (currentChar.class_mod || 0) + currentChar.level : 0; // WIS/INT/CHA mod + level
  const currentPrepared = spells.filter((s:any) => s.prepared).length;

  // Group by level
  const byLevel: Record<number, any[]> = {};
  spells.forEach((s:any) => {
    const lv = s.level || 0;
    if (!byLevel[lv]) byLevel[lv] = [];
    byLevel[lv].push(s);
  });

  const bodyHtml = `
    <div class="mb-2 d-flex justify-content-between">
      <span class="fw-bold">Prepared: ${currentPrepared} / ${maxPrepared}</span>
      <span class="text-muted small">Max = spellcasting mod + level</span>
    </div>
    <div class="spell-prep-list">
      ${Object.keys(byLevel).sort((a,b)=>+a - +b).map(lv => `
        <div class="spell-prep-group">
          <h6>${lv === '0' ? 'Cantrips' : 'Level ' + lv}</h6>
          ${byLevel[+lv].map((s:any) => `
            <div class="spell-prep-item">
              <input type="checkbox" class="form-check-input" id="prep-${s.id}" ${s.prepared ? 'checked' : ''}>
              <label for="prep-${s.id}">${esc(s.name)} <span class="text-muted">(${esc(s.school)})</span></label>
            </div>
          `).join('')}
        </div>
      `).join('')}
    </div>
    <button class="btn btn-gold w-100 mt-3" onclick="saveSpellPrep()"><i class="fa-solid fa-book-open me-1"></i>Save Preparation</button>
  `;
  showModal('Prepare Spells', bodyHtml);
};

(window as any).saveSpellPrep = async function () {
  const spellIds: number[] = [];
  (currentChar.spells || []).forEach((s:any) => {
    const cb = document.getElementById(`prep-${s.id}`) as HTMLInputElement;
    if (cb && cb.checked) spellIds.push(s.id);
  });
  await api('PUT', `/api/characters/${currentChar.id}/spells/prepare`, { spell_ids: spellIds });
  hideModal();
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSpells();
  toast('Spell preparation saved');
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
      <div class="col-md-12 mb-2">
        <label class="form-label">Portrait</label>
        <div class="d-flex align-items-center gap-2">
          ${c.portrait_url ? `<img src="${esc(c.portrait_url)}" class="character-portrait-lg me-2" alt="">` : ''}
          <input type="file" class="form-control form-control-sm" id="portraitUpload" accept="image/*">
          <button class="btn btn-primary btn-sm" onclick="uploadPortrait()"><i class="fa-solid fa-upload me-1"></i>Upload</button>
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
    <h5 class="mt-4">Currency</h5>
    <div class="row g-3">
      ${['cp','sp','ep','gp','pp'].map(coin => `
        <div class="col-4 col-md-2"><label class="form-label small">${coin.toUpperCase()}</label>
        <input class="form-control form-control-sm" id="coin${coin}" value="${c.currency?.[coin]||0}" type="number"></div>
      `).join('')}
      <div class="col-4 col-md-2 d-flex align-items-end"><button class="btn btn-gold btn-sm w-100" onclick="updateCurrency()">Save</button></div>
    </div>
    <div class="mt-3">
      <button class="btn btn-outline-primary btn-sm" onclick="shareCharacter()"><i class="fa-solid fa-share-nodes me-1"></i>Share Character</button>
    </div>`;
}

// ─── Locations ───

let locationMap: any = null;
let locationMarkers: any[] = [];
let pickMarker: any = null;
let editingLocId: number | null = null;

function initLocationMap() {
  const container = document.getElementById('locMapContainer');
  if (!container || locationMap) return;
  locationMap = L.map('locMapContainer', {
    center: [30, 0], zoom: 2,
    zoomControl: true, attributionControl: false,
  });
  L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
    maxZoom: 19, subdomains: 'abcd',
  }).addTo(locationMap);
  setTimeout(() => locationMap.invalidateSize(), 200);
}

function clearLocationMarkers() {
  if (locationMarkers.length) {
    locationMarkers.forEach(m => locationMap.removeLayer(m));
    locationMarkers = [];
  }
}

async function renderLocations() {
  const sidebar = document.getElementById('locSidebar');
  if (!sidebar) return;
  initLocationMap();
  try {
    const links = await api('GET', `/api/characters/${currentChar.id}/locations`);
    if (!document.getElementById('locSidebar') || document.getElementById('locationList')) return;
    clearLocationMarkers();

    const linkedIds = new Set(links.map((l: any) => l.location_id));
    const withCoords = allLocations.filter((l: any) => l.latitude != null && l.longitude != null);
    const noCoords = allLocations.filter((l: any) => l.latitude == null || l.longitude == null);

    // Add markers for locations with coordinates
    withCoords.forEach((l: any) => {
      const isLinked = linkedIds.has(l.id);
      const linkInfo = links.find((x: any) => x.location_id === l.id);
      const color = isLinked ? '#8b0000' : '#b8963e';
      const marker = L.circleMarker([l.latitude, l.longitude], {
        radius: isLinked ? 10 : 7, fillColor: color, color: '#fff', weight: 2, opacity: 1, fillOpacity: 0.9,
      }).addTo(locationMap);
      marker.bindPopup(`
        <div style="font-family:'Playfair Display',serif;min-width:180px">
          <strong style="font-size:1.1rem">${esc(l.name)}</strong>
          <br><span style="color:#8b7355;font-style:italic">${esc(l.type)}</span>
          ${isLinked ? `<br><span style="color:#8b0000;font-weight:600">${esc(linkInfo.relationship)}</span>` : ''}
          ${l.description ? `<br><small>${esc(l.description).substring(0, 120)}</small>` : ''}
          <br><small style="color:#8b7355">${l.latitude.toFixed(4)}, ${l.longitude.toFixed(4)}</small>
        </div>`);
      marker.on('click', () => {
        sidebar.querySelectorAll('.loc-item').forEach(el => el.classList.remove('loc-active'));
        const item = document.getElementById('loc-sidebar-' + l.id);
        if (item) item.classList.add('loc-active');
      });
      locationMarkers.push(marker);
    });

    // Fit bounds if there are markers
    if (withCoords.length > 0) {
      const group = L.featureGroup(locationMarkers);
      locationMap.fitBounds(group.getBounds().pad(0.15));
    } else {
      locationMap.setView([30, 0], 2);
    }

    // Build sidebar
    let linkedHtml = links.length
      ? `<div class="list-group list-group-flush">${links.map((l: any) => {
          const loc = allLocations.find((x: any) => x.id === l.location_id);
          return `<div class="list-group-item loc-item" id="loc-sidebar-${l.location_id}" style="cursor:pointer;border-left:3px solid #8b0000"
            onclick="flyToLocation(${loc?.latitude ?? 'null'},${loc?.longitude ?? 'null'},'loc-sidebar-${l.location_id}')">
            <div class="fw-bold small">${esc(l.location_name)}</div>
            <div><span class="badge badge-gold" style="font-size:0.65rem">${esc(l.relationship)}</span>
              ${loc ? `<span class="text-muted" style="font-size:0.7rem">${esc(loc.type)}</span>` : ''}</div>
            ${l.notes ? `<div class="text-muted" style="font-size:0.7rem">${esc(l.notes)}</div>` : ''}
            <div class="mt-1"><button class="btn btn-sm btn-outline-danger py-0 px-1" style="font-size:0.65rem" onclick="event.stopPropagation();unlinkLocation(${l.id})"><i class="fa-solid fa-unlink"></i></button></div>
          </div>`;
        }).join('')}</div>`
      : '<div class="text-center text-muted py-4"><i class="fa-solid fa-map-pin fa-lg mb-2 d-block"></i><small>No linked locations</small></div>';

    let allHtml = noCoords.length > 0
      ? `<div class="list-group list-group-flush">${noCoords.map((l: any) =>
          `<div class="list-group-item loc-item" id="loc-${l.id}" style="cursor:pointer;opacity:0.7"
            onclick="showEditLocation(${l.id})">
            <div class="small">${esc(l.name)} <span class="text-muted" style="font-size:0.65rem">(${esc(l.type)})</span></div>
            ${l.description ? `<div class="text-muted" style="font-size:0.65rem">${esc(l.description).substring(0, 60)}</div>` : ''}
          </div>`).join('')}</div>`
      : '';

    sidebar.innerHTML = `
      <div class="p-2 border-bottom" style="background:var(--parchment-dark)">
        <small class="fw-bold text-muted">LINKED (${links.length})</small>
      </div>
      ${linkedHtml}
      <div class="p-2 border-bottom" style="background:var(--parchment-dark)">
        <small class="fw-bold text-muted">ALL LOCATIONS (${allLocations.length})</small>
      </div>
      ${allLocations.map((l: any) => {
        if (linkedIds.has(l.id)) return '';
        const hasCoord = l.latitude != null && l.longitude != null;
        return `<div class="list-group-item loc-item" id="loc-sidebar-${l.id}" style="cursor:pointer;border-left:3px solid ${hasCoord ? '#b8963e' : 'transparent'}"
          onclick="${hasCoord ? `flyToLocation(${l.latitude},${l.longitude},'loc-sidebar-${l.id}')` : `showEditLocation(${l.id})`}">
          <div class="small fw-bold">${esc(l.name)}</div>
          <div><span class="text-muted" style="font-size:0.65rem">${esc(l.type)}</span>
            ${hasCoord ? '<span class="text-muted" style="font-size:0.6rem"> · mapped</span>' : '<span class="text-muted" style="font-size:0.6rem"> · no coords</span>'}</div>
          ${l.description ? `<div class="text-muted" style="font-size:0.65rem">${esc(l.description).substring(0, 60)}</div>` : ''}
        </div>`;
      }).join('')}
      ${allLocations.length === 0 ? '<div class="text-center text-muted py-4"><i class="fa-solid fa-map fa-lg mb-2 d-block"></i><small>No locations yet</small></div>' : ''}`;
  } catch { sidebar.innerHTML = '<div class="text-center text-muted py-4">Could not load locations</div>'; }
}

function getLocSidebar(): HTMLElement { return document.getElementById('locSidebar')!; }

(window as any).flyToLocation = function (lat: number | null, lng: number | null, activeId: string) {
  if (lat != null && lng != null) {
    locationMap.setView([lat, lng], 8, { animate: true });
  }
  getLocSidebar().querySelectorAll('.loc-item').forEach(el => el.classList.remove('loc-active'));
  const item = document.getElementById(activeId);
  if (item) item.classList.add('loc-active');
};

// ─── Link / Unlink ───

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

// ─── Create / Edit ───

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
    <div class="row g-2 mb-3">
      <div class="col-6"><label class="form-label">Latitude</label><input class="form-control" id="newLocLat" type="number" step="any" placeholder="e.g. 51.5"></div>
      <div class="col-6"><label class="form-label">Longitude</label><input class="form-control" id="newLocLng" type="number" step="any" placeholder="e.g. -0.12"></div>
    </div>
    <button class="btn btn-outline-secondary btn-sm w-100 mb-2" onclick="pickFromMap('new')"><i class="fa-solid fa-crosshairs me-1"></i>Pick from Map</button>
    <button class="btn btn-primary w-100" onclick="saveNewLocation()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
};

(window as any).saveNewLocation = async function () {
  const lat = parseFloat((document.getElementById('newLocLat') as HTMLInputElement).value);
  const lng = parseFloat((document.getElementById('newLocLng') as HTMLInputElement).value);
  await api('POST', '/api/locations', {
    name: (document.getElementById('newLocName') as HTMLInputElement).value,
    type: (document.getElementById('newLocType') as HTMLSelectElement).value,
    description: (document.getElementById('newLocDesc') as HTMLTextAreaElement).value,
    latitude: isNaN(lat) ? null : lat,
    longitude: isNaN(lng) ? null : lng,
  });
  hideModal();
  allLocations = await api('GET', '/api/locations');
  renderLocations();
  toast('Location created');
};

(window as any).showEditLocation = async function (locId: number) {
  editingLocId = locId;
  const loc = allLocations.find((l: any) => l.id === locId);
  if (!loc) return;
  showModal('Edit Location', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="editLocName" value="${esc(loc.name)}"></div>
    <div class="mb-3"><label class="form-label">Type</label>
      <select class="form-select" id="editLocType">${['region','city','town','dungeon','tavern','temple','shop','wilderness','other'].map(t =>
        `<option value="${t}" ${t === loc.type ? 'selected' : ''}>${t.charAt(0).toUpperCase() + t.slice(1)}</option>`).join('')}</select></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="editLocDesc" rows="3">${esc(loc.description)}</textarea></div>
    <div class="row g-2 mb-3">
      <div class="col-6"><label class="form-label">Latitude</label><input class="form-control" id="editLocLat" type="number" step="any" value="${loc.latitude ?? ''}" placeholder="Optional"></div>
      <div class="col-6"><label class="form-label">Longitude</label><input class="form-control" id="editLocLng" type="number" step="any" value="${loc.longitude ?? ''}" placeholder="Optional"></div>
    </div>
    <button class="btn btn-outline-secondary btn-sm w-100 mb-2" onclick="pickFromMap('edit')"><i class="fa-solid fa-crosshairs me-1"></i>Pick from Map</button>
    <div class="d-flex gap-2">
      <button class="btn btn-primary flex-grow-1" onclick="saveEditLocation(${locId})"><i class="fa-solid fa-floppy-disk me-1"></i>Save</button>
      <button class="btn btn-outline-danger" onclick="deleteLocation(${locId})"><i class="fa-solid fa-trash"></i></button>
    </div>
  `);
};

(window as any).saveEditLocation = async function (locId: number) {
  const lat = parseFloat((document.getElementById('editLocLat') as HTMLInputElement).value);
  const lng = parseFloat((document.getElementById('editLocLng') as HTMLInputElement).value);
  await api('PUT', `/api/locations/${locId}`, {
    name: (document.getElementById('editLocName') as HTMLInputElement).value,
    type: (document.getElementById('editLocType') as HTMLSelectElement).value,
    description: (document.getElementById('editLocDesc') as HTMLTextAreaElement).value,
    latitude: isNaN(lat) ? null : lat,
    longitude: isNaN(lng) ? null : lng,
  });
  hideModal();
  allLocations = await api('GET', '/api/locations');
  renderLocations();
  toast('Location updated');
};

(window as any).deleteLocation = async function (locId: number) {
  if (!confirm('Delete this location?')) return;
  await api('DELETE', `/api/locations/${locId}`);
  hideModal();
  allLocations = await api('GET', '/api/locations');
  renderLocations();
  toast('Location deleted');
};

(window as any).pickFromMap = function (mode: string) {
  hideModal();
  toast('Click on the map to place a pin', false);
  if (pickMarker) locationMap.removeLayer(pickMarker);
  locationMap.once('click', function (e: any) {
    const lat = e.latlng.lat.toFixed(5);
    const lng = e.latlng.lng.toFixed(5);
    pickMarker = L.marker([lat, lng], {
      icon: L.divIcon({ className: '', html: '<i class="fa-solid fa-map-pin" style="color:#8b0000;font-size:2rem;text-shadow:0 1px 3px rgba(0,0,0,.5)"></i>', iconSize: [24, 24], iconAnchor: [12, 24] }),
    }).addTo(locationMap);
    if (mode === 'new') {
      (document.getElementById('newLocLat') as HTMLInputElement).value = lat;
      (document.getElementById('newLocLng') as HTMLInputElement).value = lng;
      (window as any).showCreateLocation();
    } else if (editingLocId) {
      (document.getElementById('editLocLat') as HTMLInputElement).value = lat;
      (document.getElementById('editLocLng') as HTMLInputElement).value = lng;
      (window as any).showEditLocation(editingLocId);
    }
  });
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

let journalEditor: Editor | null = null;

function destroyJournalEditor() {
  if (journalEditor) { journalEditor.destroy(); journalEditor = null; }
}

function initJournalEditor(content?: string) {
  setTimeout(() => {
    const el = document.getElementById('journalEditor');
    if (!el) return;
    journalEditor = new Editor({
      element: el,
      extensions: [
        StarterKit.configure({ heading: { levels: [1, 2, 3] } }),
        Placeholder.configure({ placeholder: 'Write your character\'s thoughts...' }),
      ],
      content: content || '<p></p>',
    });
    const toolbar = document.getElementById('journalToolbar');
    if (toolbar) {
      const btns = [
        { icon: 'fa-bold', action: () => journalEditor?.chain().focus().toggleBold().run(), test: () => journalEditor?.isActive('bold') },
        { icon: 'fa-italic', action: () => journalEditor?.chain().focus().toggleItalic().run(), test: () => journalEditor?.isActive('italic') },
        { icon: 'fa-heading', action: () => journalEditor?.chain().focus().toggleHeading({ level: 2 }).run(), test: () => journalEditor?.isActive('heading', { level: 2 }) },
        { icon: 'fa-list-ul', action: () => journalEditor?.chain().focus().toggleBulletList().run(), test: () => journalEditor?.isActive('bulletList') },
        { icon: 'fa-list-ol', action: () => journalEditor?.chain().focus().toggleOrderedList().run(), test: () => journalEditor?.isActive('orderedList') },
        { icon: 'fa-quote-right', action: () => journalEditor?.chain().focus().toggleBlockquote().run(), test: () => journalEditor?.isActive('blockquote') },
      ];
      btns.forEach(b => {
        const btn = document.createElement('button');
        btn.type = 'button'; btn.className = 'editor-btn';
        btn.innerHTML = `<i class="fa-solid ${b.icon}"></i>`;
        btn.onclick = (e: MouseEvent) => { e.preventDefault(); b.action(); };
        toolbar.appendChild(btn);
      });
      journalEditor.on('selectionUpdate', () => {
        toolbar.querySelectorAll('.editor-btn').forEach((el: Element, i: number) => {
          el.classList.toggle('active', btns[i]?.test() || false);
        });
      });
    }
    const modal = document.getElementById('genericModal');
    modal?.addEventListener('hidden.bs.modal', destroyJournalEditor, { once: true });
  }, 50);
}

async function renderJournal() {
  const el = document.getElementById('journalSection')!;
  try {
    const entries = await api('GET', `/api/characters/${currentChar.id}/journal`);
    const months = ['January','February','March','April','May','June','July','August','September','October','November','December'];
    const groups: Record<string, any[]> = {};
    entries.forEach((j: any) => {
      const d = new Date(j.entry_date + 'T00:00:00');
      const key = months[d.getMonth()] + ' ' + d.getFullYear();
      if (!groups[key]) groups[key] = [];
      groups[key].push(j);
    });
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center flex-wrap gap-2 mb-3">
        <h5 class="mb-0"><i class="fa-solid fa-book-journal-whills me-2"></i>Character Journal</h5>
        <button class="btn btn-primary btn-sm" onclick="showAddJournal()"><i class="fa-solid fa-plus me-1"></i>Write Entry</button>
      </div>
      <div class="journal-timeline">
        ${Object.keys(groups).length ? Object.entries(groups).map(([month, monthEntries]) => `
          <div class="journal-month-group">
            <div class="journal-month-header">${month} <small class="text-muted">(${(monthEntries as any[]).length} entries)</small></div>
            ${(monthEntries as any[]).reverse().map((j: any) => `
              <div class="journal-entry-card">
                <div class="journal-entry-header" onclick="this.closest('.journal-entry-card').classList.toggle('expanded')">
                  <div class="d-flex justify-content-between align-items-start w-100">
                    <div class="min-w-0">
                      <span class="fw-bold">${esc(j.title) || 'Untitled'}</span>
                      <span class="badge badge-gold ms-2">${j.entry_date}</span>
                    </div>
                    <div class="d-flex gap-1 flex-shrink-0" onclick="event.stopPropagation()">
                      <button class="btn btn-sm btn-outline-primary" onclick="showEditJournal(${j.id})"><i class="fa-solid fa-pen"></i></button>
                      <button class="btn btn-sm btn-outline-danger" onclick="deleteJournal(${j.id})"><i class="fa-solid fa-trash"></i></button>
                    </div>
                  </div>
                  <i class="fa-solid fa-chevron-down journal-expand-icon"></i>
                </div>
                <div class="journal-entry-body">${j.entry}</div>
              </div>
            `).join('')}
          </div>
        `).join('') : '<div class="empty-state"><i class="fa-solid fa-book-open fa-2x mb-2 d-block text-muted"></i><p class="fw-bold">Empty Journal</p><p class="small text-muted">Record your character\'s thoughts and experiences.</p></div>'}
      </div>`;
  } catch { el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Could not load journal. Try again later.</p></div>'; }
}

(window as any).showAddJournal = function () {
  showModal('Journal Entry', `
    <div class="journal-editor-modal">
      <div class="mb-3"><label class="form-label">Date</label><input class="form-control" id="journalDate" type="date" value="${new Date().toISOString().split('T')[0]}"></div>
      <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="journalTitle" placeholder="Day 1: Arrival in Waterdeep"></div>
      <div class="mb-3"><label class="form-label">Entry</label><div class="editor-toolbar" id="journalToolbar"></div><div id="journalEditor" class="journal-editor"></div></div>
      <button class="btn btn-primary w-100" onclick="saveJournal()"><i class="fa-solid fa-save me-1"></i>Save</button>
    </div>
  `);
  initJournalEditor();
};

(window as any).showEditJournal = async function (id: number) {
  const entries = await api('GET', `/api/characters/${currentChar.id}/journal`);
  const j = entries.find((e: any) => e.id === id);
  if (!j) return;
  showModal('Edit Journal Entry', `
    <div class="journal-editor-modal">
      <div class="mb-3"><label class="form-label">Date</label><input class="form-control" id="journalDate" type="date" value="${esc(j.entry_date)}"></div>
      <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="journalTitle" value="${esc(j.title)}"></div>
      <div class="mb-3"><label class="form-label">Entry</label><div class="editor-toolbar" id="journalToolbar"></div><div id="journalEditor" class="journal-editor"></div></div>
      <button class="btn btn-primary w-100" onclick="saveJournal(${id})"><i class="fa-solid fa-save me-1"></i>Update</button>
    </div>
  `);
  initJournalEditor(j.entry);
};

(window as any).saveJournal = async function (editId?: number) {
  const entry = journalEditor?.getHTML() || '';
  const title = (document.getElementById('journalTitle') as HTMLInputElement)?.value || '';
  const entry_date = (document.getElementById('journalDate') as HTMLInputElement)?.value || new Date().toISOString().split('T')[0];
  if (editId) {
    await api('PUT', `/api/journal/${editId}`, { entry_date, title, entry });
  } else {
    await api('POST', `/api/characters/${currentChar.id}/journal`, { entry_date, title, entry });
  }
  destroyJournalEditor();
  hideModal();
  renderJournal();
  toast(editId ? 'Journal entry updated' : 'Journal entry saved');
};

(window as any).deleteJournal = async function (id: number) {
  if (!confirm('Delete this journal entry?')) return;
  await api('DELETE', `/api/journal/${id}`);
  renderJournal();
  toast('Journal entry deleted');
};

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

// ─── Dice Constants ───

const DICE_PRESETS = ['d4', 'd6', 'd8', 'd10', 'd12', 'd20', 'd100'];
const DICE_NOTATION_PRESETS = [
  { label: 'd20', expr: '1d20' },
  { label: 'Advantage', expr: '2d20kh1', icon: 'fa-solid fa-angles-up' },
  { label: 'Disadvantage', expr: '2d20kl1', icon: 'fa-solid fa-angles-down' },
  { label: 'd6', expr: '1d6' },
  { label: '2d6', expr: '2d6' },
  { label: '3d6', expr: '3d6' },
  { label: '4d6kh3', expr: '4d6kh3', sub: 'stats' },
  { label: 'd8', expr: '1d8' },
  { label: 'd10', expr: '1d10' },
  { label: 'd12', expr: '1d12' },
  { label: 'd100', expr: '1d100' },
  { label: 'd4', expr: '1d4' },
];

// Pip positions for d6 faces (1-6)
const D6_PIPS: Record<number, Array<{top: string; left: string}>> = {
  1: [{top:'50%',left:'50%'}],
  2: [{top:'25%',left:'25%'},{top:'75%',left:'75%'}],
  3: [{top:'25%',left:'25%'},{top:'50%',left:'50%'},{top:'75%',left:'75%'}],
  4: [{top:'25%',left:'25%'},{top:'25%',left:'75%'},{top:'75%',left:'25%'},{top:'75%',left:'75%'}],
  5: [{top:'25%',left:'25%'},{top:'25%',left:'75%'},{top:'50%',left:'50%'},{top:'75%',left:'25%'},{top:'75%',left:'75%'}],
  6: [{top:'25%',left:'25%'},{top:'25%',left:'75%'},{top:'50%',left:'25%'},{top:'50%',left:'75%'},{top:'75%',left:'25%'},{top:'75%',left:'75%'}],
};

// ─── 3D Geometry Helpers ───

/** Generate N evenly-distributed points on a sphere (Fibonacci sphere algorithm). */
function fibonacciSphere(n: number, radius: number): Array<{x: number; y: number; z: number}> {
  if (n <= 1) return [{ x: 0, y: 0, z: radius }];
  const points: Array<{x: number; y: number; z: number}> = [];
  const phi = Math.PI * (3 - Math.sqrt(5));
  for (let i = 0; i < n; i++) {
    const y = 1 - (i / (n - 1)) * 2;
    const r = Math.sqrt(1 - y * y);
    const theta = phi * i;
    points.push({ x: r * Math.cos(theta) * radius, y: y * radius, z: r * Math.sin(theta) * radius });
  }
  return points;
}

/** Compute CSS rotation angles to make (nx, ny, nz) face the viewer (+Z). */
function normalToRotation(nx: number, ny: number, nz: number): { rx: number; ry: number } {
  // Angle around Y axis
  const ry = Math.atan2(nx, nz) * (180 / Math.PI);
  // Angle around X axis
  const len = Math.sqrt(nx * nx + ny * ny + nz * nz);
  const rx = Math.asin(-ny / len) * (180 / Math.PI);
  return { rx, ry };
}

/** Get pre-calculated face transforms for a given die type. */
function getFaceTransforms(sides: number): Array<{ rx: number; ry: number; rz: number }> {
  const radius = 36; // half the die size
  if (sides === 6) {
    // Cube: 6 faces at cardinal positions
    return [
      { rx: 0,   ry: 0,   rz: 0 },   // front (face 1)
      { rx: 0,   ry: 180, rz: 0 },   // back  (face 6, opposite to front)
      { rx: 0,   ry: 90,  rz: 0 },   // right (face 3)
      { rx: 0,   ry: -90, rz: 0 },   // left  (face 4)
      { rx: -90, ry: 0,   rz: 0 },   // top   (face 2)
      { rx: 90,  ry: 0,   rz: 0 },   // bottom (face 5)
    ];
  }

  // For other polyhedra, distribute faces evenly on sphere
  const points = fibonacciSphere(sides, radius);

  // For tetrahedron (d4), d4 has special face-value mapping
  if (sides === 4) {
    // Use the 4 points of a tetrahedron for face positions
    const t = Math.sqrt(1 / 3);
    const tetraVerts = [
      { x:  t * radius, y:  t * radius, z:  t * radius },
      { x: -t * radius, y: -t * radius, z:  t * radius },
      { x: -t * radius, y:  t * radius, z: -t * radius },
      { x:  t * radius, y: -t * radius, z: -t * radius },
    ];
    return tetraVerts.map(p => {
      const norm = normalToRotation(p.x, p.y, p.z);
      return { rx: norm.rx, ry: norm.ry, rz: 0 };
    });
  }

  return points.map(p => {
    const norm = normalToRotation(p.x, p.y, p.z);
    return { rx: norm.rx, ry: norm.ry, rz: 0 };
  });
}

/** Compute die rotation string to show a specific face value. */
function rotateDieToShow(sides: number, value: number): string {
  const faceIdx = Math.max(0, Math.min(sides - 1, value - 1));
  const transforms = FACE_TRANSFORMS[sides as keyof typeof FACE_TRANSFORMS]
    || getFaceTransforms(sides);

  if (faceIdx < transforms.length) {
    const t = transforms[faceIdx];
    return `rotateX(${-t.rx}deg) rotateY(${-t.ry}deg)`;
  }
  return '';
}

/** Compute the rolling animation class for a die type. */
function rollingClass(sides: number): string {
  if (sides === 4) return 'rolling-d4';
  if (sides === 6) return 'rolling';
  if (sides === 8) return 'rolling-d8';
  if (sides === 10) return 'rolling-d10';
  if (sides === 12) return 'rolling-d12';
  if (sides === 20) return 'rolling-d20';
  return 'rolling';
}

/** Face shape class for the polyhedron type. */
function faceShapeClass(sides: number): string {
  if (sides === 4 || sides === 8 || sides === 20) return 'tri-face';
  if (sides === 10) return 'kite-face';
  if (sides === 12) return 'pent-face';
  return ''; // cube uses square faces (default)
}

// Cache pre-computed face transforms
const FACE_TRANSFORMS: Record<number, Array<{ rx: number; ry: number; rz: number }>> = {};
[4, 6, 8, 10, 12, 20].forEach(s => { FACE_TRANSFORMS[s] = getFaceTransforms(s); });

// ─── Build 3D Die HTML ───

function build3DDie(value: number, sides: number, dieLabel: string): string {
  const maxSides = sides >= 100 ? 100 : sides;
  const dieClass = 'd' + (maxSides >= 100 ? 100 : maxSides);
  const transforms = FACE_TRANSFORMS[maxSides as keyof typeof FACE_TRANSFORMS]
    || getFaceTransforms(maxSides);
  const shapeCls = faceShapeClass(maxSides);
  const dieRot = rotateDieToShow(maxSides, value);

  const facesHtml = transforms.map((t, i) => {
    const faceValue = i + 1; // face index → value
    const displayVal = maxSides === 100
      ? (faceValue === 10 ? '00' : String(faceValue * 10))
      : String(faceValue);

    // d6 uses pips instead of numerals
    if (maxSides === 6 && faceValue >= 1 && faceValue <= 6) {
      const pips = D6_PIPS[faceValue] || [];
      const pipHtml = pips.map(p =>
        `<span class="pip" style="top:${p.top};left:${p.left};transform:translate(-50%,-50%)"></span>`
      ).join('');
      return `<div class="dice-3d-face ${shapeCls}" style="transform:rotateX(${t.rx}deg) rotateY(${t.ry}deg) translateZ(36px)">${pipHtml}</div>`;
    }

    return `<div class="dice-3d-face ${shapeCls}" style="transform:rotateX(${t.rx}deg) rotateY(${t.ry}deg) translateZ(36px)">${displayVal}</div>`;
  }).join('');

  return `<div class="dice-3d-die ${dieClass}" data-sides="${maxSides}" data-value="${value}" style="transform:${dieRot}">${facesHtml}</div>`;
}

// ─── Die Value Helpers ───

/** Extract the numeric value from a breakdown roll entry (DieRollDetail or plain number). */
function rollValue(r: any): number {
  return typeof r === 'number' ? r : (r.value ?? 0);
}

/** Check if a die roll detail should be displayed normally (useInTotal). */
function rollUsed(r: any): boolean {
  return typeof r === 'number' ? true : (r.useInTotal !== false);
}

/** Get modifier flags for a die roll (e.g., "dropped", "exploded"). */
function rollFlags(r: any): string {
  return typeof r === 'number' ? '' : (r.modifierFlags || '');
}

/** Parse sides from a die label like "d20" or "+3" (returns 0 for modifiers). */
function parseSides(dieLabel: string): number {
  const m = dieLabel.match(/^d(\d+)$/i);
  return m ? parseInt(m[1]) : 0;
}

// ─── Dice Expression / Quick Roll ───

function setDiceExpr(expr: string) {
  const input = document.getElementById('diceExpr') as HTMLInputElement;
  input.value = expr;
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
      // Check for 20/1 on d20 in the FIRST breakdown group's rolls
      const d20Rolls = result.breakdown?.filter((b: any) => b.die === 'd20' && b.rolls) || [];
      for (const bg of d20Rolls) {
        for (const r of bg.rolls) {
          const v = rollValue(r);
          if (v === 20) badge = '<span class="badge bg-success ms-2">Critical Hit!</span>';
          else if (v === 1) badge = '<span class="badge bg-danger ms-2">Critical Fail!</span>';
        }
      }
      resultDiv.style.display = 'block';
      resultDiv.innerHTML = `
        <div class="dice-result-box text-center">
          <div class="roll-expression">${esc(result.expression)} (${isAdv ? 'advantage' : 'disadvantage'})</div>
          <div class="d-flex justify-content-center gap-3 mb-2">
            ${rolls.map((r: any, i: number) => {
              const v = rollValue(r);
              const used = rollUsed(r);
              const style = used ? 'border-color:var(--gold);box-shadow:0 0 0 2px var(--gold)' : 'opacity:0.4';
              return `<span class="die-face${used ? '' : ' die-dropped'}" style="${style}">${v}</span>`;
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

// ─── Render Dice Tab ───

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
          <input class="form-control text-center" id="diceExpr" value="1d20" placeholder="e.g. 2d6+3, 4d6kh3, 1d20!" style="font-size:1.3rem;font-weight:700">
        </div>
      </div>
      <div class="dice-quick-btns mb-3">
        ${DICE_NOTATION_PRESETS.map(p =>
          `<button class="btn btn-sm dice-btn" onclick="setDiceExpr('${esc(p.expr)}')" title="${p.sub ? p.sub : p.expr}">
            ${p.icon ? `<i class="${p.icon} me-1"></i>` : ''}${esc(p.label)}
          </button>`
        ).join('')}
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

// ─── Rolling Animation ───

function animateDiceRoll(breakdown: any[]) {
  const container = document.getElementById('dice3dContainer');
  if (!container) return;

  // Build dice placeholder with rolling animation
  container.innerHTML = breakdown.map((b: any) => {
    if (!b.rolls || b.rolls.length === 0) return '';
    const sides = parseSides(b.die);
    if (sides === 0) return ''; // skip modifiers
    const dieLabel = b.die;
    const rollCls = rollingClass(sides);
    const transforms = FACE_TRANSFORMS[sides as keyof typeof FACE_TRANSFORMS]
      || getFaceTransforms(sides);
    const shapeCls = faceShapeClass(sides);

    return b.rolls.map((r: any) => {
      // Create a rolling die HTML placeholder (all faces, no rotation)
      const facesHtml = transforms.map((t, i) => {
        const displayVal = sides >= 100
          ? (i + 1 === 10 ? '00' : String((i + 1) * 10))
          : String(i + 1);
        if (sides === 6) {
          const pips = D6_PIPS[i + 1] || [];
          const pipHtml = pips.map(p =>
            `<span class="pip" style="top:${p.top};left:${p.left};transform:translate(-50%,-50%)"></span>`
          ).join('');
          return `<div class="dice-3d-face ${shapeCls}" style="transform:rotateX(${t.rx}deg) rotateY(${t.ry}deg) translateZ(36px)">${pipHtml}</div>`;
        }
        return `<div class="dice-3d-face ${shapeCls}" style="transform:rotateX(${t.rx}deg) rotateY(${t.ry}deg) translateZ(36px)">${displayVal}</div>`;
      }).join('');
      const dieHtml = `<div class="dice-3d-die ${b.die} ${rollCls}" data-sides="${sides}" data-value="0">${facesHtml}</div>`;
      return `<div class="dice-3d-wrapper"><span class="dice-3d-label">${dieLabel}</span>${dieHtml}</div>`;
    }).join('');
  }).join('');
}

// ─── Settle Dice (Final Result) ───

function settleDice(breakdown: any[]) {
  const container = document.getElementById('dice3dContainer');
  if (!container) return;

  setTimeout(() => {
    container.innerHTML = breakdown.map((b: any) => {
      if (!b.rolls || b.rolls.length === 0) return '';
      const sides = parseSides(b.die);
      if (sides === 0) return ''; // skip modifiers
      const dieLabel = b.die;

      return b.rolls.map((r: any) => {
        const v = rollValue(r);
        const dieHtml = build3DDie(v, sides, dieLabel);

        // Check for crits on d20
        let extraClass = '';
        if (sides === 20) {
          if (v === 20) extraClass = ' dice-crit-success';
          else if (v === 1) extraClass = ' dice-crit-fail';
        }

        const wrapper = document.createElement('div');
        wrapper.className = 'dice-3d-wrapper' + extraClass;
        wrapper.innerHTML = `<span class="dice-3d-label">${dieLabel}</span>${dieHtml}`;
        return wrapper.outerHTML;
      }).join('');
    }).join('');
  }, 900);
}

// ─── Main Roll Handler ───

async function doRoll() {
  const expr = (document.getElementById('diceExpr') as HTMLInputElement).value;
  if (!expr) return;

  // Show rolling animation immediately
  const resultEl = document.getElementById('diceResult')!;
  resultEl.style.display = 'none';
  const container = document.getElementById('dice3dContainer');
  if (container) {
    // Parse basic die patterns to create placeholder dice
    const m = expr.match(/(\d+)d(\d+)/gi);
    if (m) {
      const parts = m.map(s => {
        const [, count, sides] = s.match(/(\d+)d(\d+)/i)!;
        return { die: 'd' + sides, count: parseInt(count || '1'), sides: parseInt(sides) };
      });
      const breakdown = parts.flatMap(p =>
        Array.from({ length: p.count }, () => ({ die: p.die, rolls: [1], total: 0 }))
      );
      animateDiceRoll(breakdown);
    } else {
      // Fallback: try to extract any d-something
      const m2 = expr.match(/d(\d+)/i);
      if (m2) {
        const sides = parseInt(m2[1]);
        animateDiceRoll([{ die: 'd' + sides, rolls: [1], total: 0 }]);
      }
    }
  }

  try {
    const result = await api('POST', '/api/roll', { expression: expr, character_id: currentChar?.id });
    resultEl.style.display = 'block';

    // Settle dice with actual results
    if (result.breakdown) {
      settleDice(result.breakdown);
    }

    // Build breakdown text (skipping modifier groups)
    let facesHtml = '';
    if (result.breakdown) {
      facesHtml = result.breakdown.map((b: any) => {
        if (!b.rolls || b.rolls.length === 0) return ''; // skip modifier-only groups
        const dieLabel = b.die;
        const sides = parseSides(dieLabel);
        if (sides === 0) return '';
        const rolls = b.rolls.map((r: any) => {
          const v = rollValue(r);
          const used = rollUsed(r);
          const flags = rollFlags(r);
          // Kept/dropped indicator
          const itemClass = used ? 'die-face die-kept' : 'die-face die-dropped';
          const flagText = flags ? ` data-flags="${esc(flags)}"` : '';
          return `<span class="${itemClass}"${flagText}>${v}${flags === 'dropped' ? '✕' : ''}</span>`;
        }).join('');
        return `<div class="die-group"><span class="die-label">${dieLabel}:</span> <span class="die-faces">${rolls}</span></div>`;
      }).filter((h: string) => h).join('');
    }

    // Check for crits on d20
    let critBadge = '';
    if (result.breakdown) {
      for (const b of result.breakdown) {
        if (b.die === 'd20' && b.rolls) {
          for (const r of b.rolls) {
            const v = rollValue(r);
            if (v === 20) critBadge = '<span class="badge bg-success ms-2"><i class="fa-solid fa-bolt me-1"></i>Critical Hit!</span>';
            else if (v === 1) critBadge = '<span class="badge bg-danger ms-2"><i class="fa-solid fa-skull me-1"></i>Critical Fail!</span>';
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
    <div class="text-center mt-2"><button class="btn btn-sm btn-outline-gold" onclick="generateRandomChar()"><i class="fa-solid fa-dice me-1"></i>Random Character</button></div>
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

    let html = `<div class="d-flex justify-content-between align-items-center mb-3">
      <h1 class="h2 mb-0"><i class="fa-solid fa-flag me-2"></i>Party View</h1>
      <div class="d-flex gap-2">
        <button class="btn btn-gold btn-sm" onclick="showCreateCampaign()"><i class="fa-solid fa-plus me-1"></i>New Campaign</button>
        ${currentUser?.role === 'dm' || currentUser?.role === 'admin' ? `<button class="btn btn-outline-primary btn-sm" onclick="showCreateParty()"><i class="fa-solid fa-flag me-1"></i>New Party</button>` : ''}
      </div>
    </div>`;

    // DM/Admin: Party management section
    if (currentUser?.role === 'dm' || currentUser?.role === 'admin') {
      try {
        const parties = await api('GET', '/api/parties');
        if (parties.length) {
          html += `<h5 class="mb-2"><i class="fa-solid fa-flag me-1"></i>Parties</h5>`;
          for (const p of parties) {
            const factions = await api('GET', `/api/parties/${p.id}/factions`).catch(() => []);
            const uploads = await api('GET', `/api/parties/${p.id}/uploads`).catch(() => []);
            const fileCount = uploads.length;
            html += `<div class="card mb-3">
              <div class="card-header d-flex justify-content-between align-items-center py-2">
                <span><strong>${esc(p.name)}</strong> ${p.description ? `<span class="text-muted small ms-2">${esc(p.description)}</span>` : ''}</span>
                <div class="d-flex gap-1">
                  <span class="badge badge-gold">${factions.length} factions</span>
                  ${fileCount ? `<span class="badge bg-info">${fileCount} files</span>` : ''}
                  <button class="btn btn-sm btn-outline-primary" onclick="renameParty(${p.id},'${esc(p.name)}','${esc(p.description)}')"><i class="fa-solid fa-pen"></i></button>
                  <button class="btn btn-sm btn-outline-danger" onclick="deleteParty(${p.id})"><i class="fa-solid fa-trash"></i></button>
                </div>
              </div>
              ${factions.length ? `<div class="card-body py-2">
                <div class="small"><strong>Factions:</strong></div>
                <div class="d-flex flex-wrap gap-1 mt-1">${factions.map((f: any) =>
                  `<span class="badge bg-light text-dark border">${esc(f.name)}${f.type ? ` <span class="text-muted">(${esc(f.type)})</span>` : ''}</span>`
                ).join('')}</div>
              </div>` : ''}
              ${uploads.length ? `<div class="card-footer py-1">
                <div class="small text-muted">${fileCount} file(s) uploaded</div>
              </div>` : ''}
            </div>`;
          }
        }
      } catch {}
    }

    // Campaign-based party groups
    html += groups.map((g:any) => {
      const own = g.id ? isOwner(g.id) : false;
      const dm = g.id ? isDm(g.id) : false;
      const canOpen = (userId: number) => userId === currentUser?.id || currentUser?.role === 'admin' || dm;
      const partyLabel = g.party_name ? esc(g.party_name) : esc(g.name || 'Unnamed Campaign');
      const subLabel = g.party_name ? `<span class="small text-muted ms-2">Campaign: ${esc(g.name)}</span>` : '';
      return `<div class="card mb-3">
        <div class="card-header d-flex justify-content-between align-items-center">
          <div>
            <strong>${partyLabel}</strong>
            ${subLabel}
            ${g.owner_name ? `<span class="ms-2 small text-muted">DM: ${esc(g.owner_name)}</span>` : ''}
          </div>
          <div class="d-flex align-items-center gap-2">
            <span class="badge badge-gold">${g.members.length} members</span>
            ${g.id && (own || dm) ? `
              <button class="btn btn-outline-gold btn-sm" onclick="showCampaignDashboard(${g.id},'${esc(g.name)}')" title="Dashboard"><i class="fa-solid fa-chart-simple"></i></button>
              <button class="btn btn-outline-primary btn-sm" onclick="showManageCampaign(${g.id},'${esc(g.name)}','${esc(g.party_name || '')}')" title="Manage"><i class="fa-solid fa-users-gear"></i></button>
              <button class="btn btn-outline-info btn-sm" onclick="shareParty(${g.id})" title="Share Party"><i class="fa-solid fa-share-nodes"></i></button>
            ` : ''}
            ${g.id && own ? `<button class="btn btn-outline-danger btn-sm" onclick="deleteCampaign(${g.id})" title="Delete"><i class="fa-solid fa-trash"></i></button>` : ''}
            ${g.id && (own || dm) && currentUser?.role === 'admin' ? `<button class="btn btn-outline-gold btn-sm" onclick="sendCampaignHighlights(${g.id})" title="Email Highlights"><i class="fa-solid fa-envelope"></i></button>` : ''}
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
        ${g.id && (own || dm) ? `
        <div class="card-footer py-2">
          <div class="d-flex gap-2 flex-wrap">
            <button class="btn btn-sm btn-outline-gold" onclick="showPartyInventory(${g.id})"><i class="fa-solid fa-box me-1"></i>Party Inventory</button>
            <button class="btn btn-sm btn-outline-primary" onclick="showSessionPlanner(${g.id})"><i class="fa-solid fa-calendar me-1"></i>Session Planner</button>
            <button class="btn btn-sm btn-outline-gold" onclick="showEncounterDifficulty()"><i class="fa-solid fa-crosshairs me-1"></i>Difficulty</button>
            <button class="btn btn-sm btn-outline-gold" onclick="showTreasureGenerator()"><i class="fa-solid fa-coins me-1"></i>Treasure</button>
          </div>
        </div>
        ` : ''}
      </div>`;
    }).join('') || '<div class="empty-state"><i class="fa-solid fa-flag fa-2x mb-2 d-block text-muted"></i>No characters yet. Create a campaign and add members to build your party!</div>';

    el.innerHTML = html;
  } catch (e:any) {
    el.innerHTML = `<div class="empty-state"><i class="fa-solid fa-circle-exclamation fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">Failed: ${esc(e.message)}</p></div>`;
  }
};

(window as any).showCreateCampaign = function () {
  showModal('Create Campaign', `
    <div class="mb-3"><label class="form-label">Campaign Name</label><input class="form-control" id="newCampaignName"></div>
    <div class="mb-3"><label class="form-label">Party Name</label><input class="form-control" id="newPartyName" placeholder="e.g. The Dawnbringers"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newCampaignDesc" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="doCreateCampaign()">Create</button>
  `);
};

(window as any).doCreateCampaign = async function () {
  try {
    const name = (document.getElementById('newCampaignName') as HTMLInputElement).value;
    if (!name) { toast('Name required', true); return; }
    const partyName = (document.getElementById('newPartyName') as HTMLInputElement).value;
    await api('POST', '/api/campaigns', { name, party_name: partyName, description: (document.getElementById('newCampaignDesc') as HTMLTextAreaElement).value });
    hideModal();
    toast('Campaign created');
    (window as any).showParty();
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).showManageCampaign = async function (campaignId: number, name: string, partyName: string = '') {
  const [campaigns, members] = await Promise.all([
    api('GET', '/api/campaigns'),
    api('GET', `/api/campaigns/${campaignId}/members`).catch(() => []),
  ]);
  const c = campaigns.find((x: any) => x.id === campaignId);
  const curPartyName = (c && c.party_name) || partyName;
  const curDesc = (c && c.description) || '';
  const membersHtml = members.length
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
  showModal(`Manage: ${esc(name)}`, `
    <div class="mb-2"><label class="form-label small">Campaign Name</label><input class="form-control" id="editCampaignName" value="${esc(name)}"></div>
    <div class="mb-2"><label class="form-label small">Party Name</label><input class="form-control" id="editPartyName" value="${esc(curPartyName)}" placeholder="e.g. The Dawnbringers"></div>
    <div class="mb-3"><label class="form-label small">Description</label><textarea class="form-control" id="editCampaignDesc" rows="2">${esc(curDesc)}</textarea></div>
    <button class="btn btn-gold w-100 mb-3" onclick="doUpdateCampaign(${campaignId})"><i class="fa-solid fa-floppy-disk me-1"></i>Save Settings</button>
    <hr>
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

(window as any).doUpdateCampaign = async function (campaignId: number) {
  try {
    const name = (document.getElementById('editCampaignName') as HTMLInputElement).value;
    if (!name) { toast('Name required', true); return; }
    const partyName = (document.getElementById('editPartyName') as HTMLInputElement).value;
    const description = (document.getElementById('editCampaignDesc') as HTMLTextAreaElement).value;
    await api('PUT', `/api/campaigns/${campaignId}`, { name, party_name: partyName, description });
    toast('Campaign updated');
    (window as any).showParty();
    hideModal();
  } catch (e: any) {
    toast(e.message, true);
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

// ─── Party Management ───

(window as any).showCreateParty = function () {
  showModal('Create Party', `
    <div class="mb-3"><label class="form-label">Party Name</label><input class="form-control" id="newPartyNameInput"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="newPartyDesc" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="doCreateParty()">Create</button>
  `);
};

(window as any).doCreateParty = async function () {
  const name = (document.getElementById('newPartyNameInput') as HTMLInputElement).value;
  if (!name) { toast('Party name required', true); return; }
  const description = (document.getElementById('newPartyDesc') as HTMLTextAreaElement).value;
  try {
    await api('POST', '/api/parties', { name, description });
    hideModal();
    toast('Party created');
    (window as any).showParty();
  } catch (e: any) { toast(e.message, true); }
};

(window as any).renameParty = function (id: number, name: string, description: string) {
  showModal('Rename Party', `
    <div class="mb-3"><label class="form-label">Party Name</label><input class="form-control" id="editPartyNameInput" value="${esc(name)}"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="editPartyDesc" rows="2">${esc(description)}</textarea></div>
    <button class="btn btn-primary w-100" onclick="doRenameParty(${id})">Save</button>
  `);
};

(window as any).doRenameParty = async function (id: number) {
  const name = (document.getElementById('editPartyNameInput') as HTMLInputElement).value;
  if (!name) { toast('Party name required', true); return; }
  const description = (document.getElementById('editPartyDesc') as HTMLTextAreaElement).value;
  try {
    await api('PUT', `/api/parties/${id}`, { name, description });
    hideModal();
    toast('Party updated');
    (window as any).showParty();
  } catch (e: any) { toast(e.message, true); }
};

(window as any).deleteParty = async function (id: number) {
  if (!confirm('Delete this party?')) return;
  try {
    await api('DELETE', `/api/parties/${id}`);
    toast('Party deleted');
    (window as any).showParty();
  } catch (e: any) { toast(e.message, true); }
};

// ─── Share & Email ───

(window as any).shareCharacter = async function () {
  if (!currentChar) return;
  try {
    const result = await api('POST', '/api/share', {
      entity_type: 'character',
      entity_id: currentChar.id,
    });
    showModal('Share Character', `
      <p>Share this link to let others view <strong>${esc(currentChar.name)}</strong>.</p>
      <div class="input-group mb-3">
        <input class="form-control" id="shareUrl" value="${esc(result.url)}" readonly onclick="this.select()">
        <button class="btn btn-gold" onclick="copyShareUrl()"><i class="fa-solid fa-copy"></i></button>
      </div>
      <div class="d-flex gap-2">
        <button class="btn btn-primary flex-grow-1" onclick="window.open('mailto:?subject=Check out my character ${esc(currentChar.name)}&body=${encodeURIComponent(result.url)}','_blank')"><i class="fa-solid fa-envelope me-1"></i>Email</button>
        <button class="btn btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).copyShareUrl = function () {
  const input = document.getElementById('shareUrl') as HTMLInputElement;
  if (input) {
    input.select();
    navigator.clipboard.writeText(input.value).then(() => toast('Link copied!')).catch(() => {});
  }
};

(window as any).shareParty = async function (campaignId: number) {
  try {
    const result = await api('POST', '/api/share', {
      entity_type: 'party',
      entity_id: campaignId,
    });
    showModal('Share Party', `
      <p>Share this link to let others view your party.</p>
      <div class="input-group mb-3">
        <input class="form-control" id="shareUrl" value="${esc(result.url)}" readonly onclick="this.select()">
        <button class="btn btn-gold" onclick="copyShareUrl()"><i class="fa-solid fa-copy"></i></button>
      </div>
      <div class="d-flex gap-2">
        <button class="btn btn-primary flex-grow-1" onclick="window.open('mailto:?subject=Check out our party&body=${encodeURIComponent(result.url)}','_blank')"><i class="fa-solid fa-envelope me-1"></i>Email</button>
        <button class="btn btn-outline-secondary" onclick="hideModal()">Close</button>
      </div>
    `);
  } catch (e: any) {
    toast(e.message, true);
  }
};

(window as any).sendCampaignHighlights = async function (campaignId: number) {
  try {
    const result = await api('POST', '/api/admin/campaign-highlights', { campaign_id: campaignId });
    const msg = result.errors && result.errors.length
      ? `Sent to ${result.sent} recipients, but ${result.errors.length} failed.`
      : `Campaign highlights sent to ${result.sent} recipient(s)!`;
    toast(msg);
    if (result.errors) console.warn('Email errors:', result.errors);
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

// ─── Portrait Upload ───

(window as any).uploadPortrait = async function () {
  const input = document.getElementById('portraitUpload') as HTMLInputElement;
  if (!input.files || !input.files[0]) { toast('Select an image', true); return; }
  const form = new FormData();
  form.append('image', input.files[0]);
  try {
    const res = await fetch('/api/upload', {
      method: 'POST', headers: { 'X-CSRF-Token': csrfToken }, credentials: 'include', body: form,
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Upload failed');
    await updateField('portrait_url', data.url);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSheet();
    toast('Portrait uploaded');
  } catch (e: any) { toast(e.message, true); }
};

(window as any).clearPortrait = async function () {
  await updateField('portrait_url', '');
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSheet();
  toast('Portrait removed');
};

// ─── Multi-Class ───

(window as any).addClass = function () {
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
};

(window as any).saveClass = async function () {
  await api('POST', `/api/characters/${currentChar.id}/classes`, {
    class: (document.getElementById('mcClass') as HTMLInputElement).value,
    subclass: (document.getElementById('mcSubclass') as HTMLInputElement).value,
    level: +(document.getElementById('mcLevel') as HTMLInputElement).value || 1,
    hit_dice: (document.getElementById('mcHD') as HTMLSelectElement).value,
  });
  hideModal();
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSheet();
  toast('Class added');
};

(window as any).editClass = function (id: number) {
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
};

(window as any).saveEditClass = async function (id: number) {
  await api('PUT', `/api/classes/${id}`, {
    class: (document.getElementById('mcClass') as HTMLInputElement).value,
    subclass: (document.getElementById('mcSubclass') as HTMLInputElement).value,
    level: +(document.getElementById('mcLevel') as HTMLInputElement).value || 1,
    hit_dice: (document.getElementById('mcHD') as HTMLSelectElement).value,
  });
  hideModal();
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSheet();
  toast('Class updated');
};

(window as any).deleteClass = async function (id: number) {
  if (!confirm('Remove this class?')) return;
  await api('DELETE', `/api/classes/${id}`);
  currentChar = await api('GET', `/api/characters/${currentChar.id}`);
  renderSheet();
  toast('Class removed');
};

// ─── Encounter Builder ───

(window as any).showEncounterBuilder = async function () {
  showView('encounter');
  const el = document.getElementById('encounterContent')!;
  el.innerHTML = '<div class="ornament">✧ Loading encounters... ✧</div>';
  try {
    const encounters = await api('GET', '/api/encounters');
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-center mb-3">
        <div>
          <button class="btn btn-gold btn-sm me-2" onclick="showCreateEncounter()"><i class="fa-solid fa-plus me-1"></i>New Encounter</button>
          <button class="btn btn-outline-primary btn-sm" onclick="showEncounterXPCalc()"><i class="fa-solid fa-calculator me-1"></i>XP Calculator</button>
        </div>
      </div>
      <div class="row g-3" id="encounterList">
        ${encounters.length ? encounters.map((e: any) => `
          <div class="col-md-6 col-lg-4">
            <div class="character-card" onclick="showEncounterDetail(${e.id})">
              <div class="char-name">${esc(e.name)}</div>
              <div class="char-detail">${esc(e.environment)} · ${esc(e.difficulty)} · ${e.total_xp} XP</div>
              <div class="char-hp mt-1">${esc(e.description).substring(0, 100)}</div>
            </div>
          </div>`).join('')
          : '<div class="empty-state"><i class="fa-solid fa-crosshairs fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Encounters Yet</p><p class="small text-muted">Build balanced encounters with XP budgeting.</p></div>'}
      </div>`;
  } catch (e: any) { el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(e.message)}</p></div>`; }
};

(window as any).showCreateEncounter = function () {
  showModal('New Encounter', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="encName" placeholder="Goblin Ambush"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="encDesc" rows="2"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Environment</label>
        <select class="form-select" id="encEnv">
          <option value="">Any</option><option value="forest">Forest</option><option value="dungeon">Dungeon</option>
          <option value="mountain">Mountain</option><option value="swamp">Swamp</option><option value="urban">Urban</option>
          <option value="underdark">Underdark</option><option value="coastal">Coastal</option><option value="arctic">Arctic</option>
          <option value="desert">Desert</option><option value="grassland">Grassland</option>
        </select></div>
      <div class="col-6"><label class="form-label">Difficulty</label>
        <select class="form-select" id="encDiff">
          <option value="easy">Easy</option><option value="medium" selected>Medium</option>
          <option value="hard">Hard</option><option value="deadly">Deadly</option>
        </select></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveEncounter()"><i class="fa-solid fa-plus me-1"></i>Create</button>
  `);
};

(window as any).saveEncounter = async function () {
  const name = (document.getElementById('encName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  await api('POST', '/api/encounters', {
    name, description: (document.getElementById('encDesc') as HTMLTextAreaElement).value,
    environment: (document.getElementById('encEnv') as HTMLSelectElement).value,
    difficulty: (document.getElementById('encDiff') as HTMLSelectElement).value,
  });
  hideModal();
  (window as any).showEncounterBuilder();
  toast('Encounter created');
};

(window as any).showEncounterDetail = async function (id: number) {
  showView('singleEncounter');
  const el = document.getElementById('singleEncounterContent')!;
  el.innerHTML = '<div class="ornament">✧ Loading... ✧</div>';
  try {
    const e = await api('GET', `/api/encounters/${id}`);
    const monsters = e.monsters || [];
    const totalCount = monsters.reduce((s: number, m: any) => s + m.count, 0);
    el.innerHTML = `
      <div class="d-flex justify-content-between align-items-start flex-wrap gap-2 mb-2">
        <div>
          <h1 class="h2 mb-0">${esc(e.name)}</h1>
          <p class="text-muted fst-italic mb-0 mt-1">
            <span class="badge badge-gold">${esc(e.difficulty)}</span>
            ${e.environment ? `<span class="badge badge-muted ms-1">${esc(e.environment)}</span>` : ''}
            <span class="badge badge-blood ms-1">${e.total_xp} XP</span>
          </p>
        </div>
    <div class="d-flex gap-2 flex-wrap">
      <button class="btn btn-gold btn-sm" onclick="newChar()"><i class="fa-solid fa-plus me-1"></i>New Character</button>
      <button class="btn btn-outline-primary btn-sm" id="compareBtn" onclick="toggleCompareMode()"><i class="fa-solid fa-arrow-right-arrow-left me-1"></i>Compare</button>
      <button class="btn btn-outline-primary btn-sm" onclick="showImport()"><i class="fa-solid fa-file-import me-1"></i>Import</button>
    </div>
      </div>
      ${e.description ? `<p class="text-muted">${esc(e.description)}</p>` : ''}
      <div class="ornament my-2">✧</div>
      <h5>Monsters <span class="text-muted small">(${totalCount} total)</span></h5>
      <div id="monsterList">
        ${monsters.length ? monsters.map((m: any) => `
          <div class="inv-item">
            <div>
              <span class="fw-bold">${esc(m.name)}</span>
              <span class="badge badge-blood ms-1">x${m.count}</span>
              <span class="badge badge-gold ms-1">CR ${esc(m.cr)}</span>
              <span class="badge badge-muted ms-1">${m.xp} XP</span>
              <span class="text-muted small ms-2">AC ${m.ac} · HP ${m.hp}</span>
            </div>
            <div class="d-flex gap-1">
              <button class="btn btn-sm btn-outline-primary" onclick="editMonster(${e.id}, ${m.id})"><i class="fa-solid fa-pen"></i></button>
              <button class="btn btn-sm btn-outline-danger" onclick="deleteMonster(${e.id}, ${m.id})"><i class="fa-solid fa-trash"></i></button>
            </div>
          </div>`).join('')
          : '<div class="empty-state"><i class="fa-solid fa-skull fa-2x mb-2 d-block text-muted"></i><p class="small text-muted">No monsters yet. Add some!</p></div>'}
      </div>
      <div class="ornament my-2">✧</div>
      <h5>XP Budget</h5>
      <div class="row g-3">
        <div class="col-md-6">
          <div class="combat-stat"><div class="stat-label">Total Monster XP</div><div class="stat-value">${e.total_xp}</div></div>
        </div>
        <div class="col-md-6">
          <div class="combat-stat"><div class="stat-label">Difficulty</div><div class="stat-value text-capitalize">${esc(e.difficulty)}</div></div>
        </div>
      </div>
      ${e.notes ? `<div class="mt-3"><h6>Notes</h6><p class="text-muted small">${esc(e.notes)}</p></div>` : ''}`;
  } catch (err: any) {
    el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(err.message)}</p>
      <button class="btn btn-outline-secondary btn-sm" onclick="(window as any).showEncounterBuilder()">Back</button></div>`;
  }
};

(window as any).editEncounter = function (id: number) { (window as any).showEncounterBuilder(); };

(window as any).deleteEncounter = async function (id: number) {
  if (!confirm('Delete this encounter?')) return;
  await api('DELETE', `/api/encounters/${id}`);
  (window as any).showEncounterBuilder();
  toast('Encounter deleted');
};

(window as any).addMonster = function (eid: number) {
  showModal('Add Monster', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="monName" list="monsterSuggestions">
      <datalist id="monsterSuggestions">
        ${['Goblin','Hobgoblin','Bugbear','Orc','Ogre','Troll','Giant Spider','Skeleton','Zombie','Wolf','Dire Wolf','Bandit','Kobold','Gnoll','Owlbear','Harpy','Basilisk','Chimera','Dragon Wyrmling'].map(n => `<option value="${n}">`).join('')}
      </datalist></div>
    <div class="row g-3 mb-3">
      <div class="col-4"><label class="form-label">Count</label><input class="form-control" id="monCount" type="number" value="1" min="1"></div>
      <div class="col-4"><label class="form-label">CR</label>
        <select class="form-select" id="monCR">
          ${['0','1/8','1/4','1/2','1','2','3','4','5','6','7','8','9','10','11','12','13','14','15','16','17','18','19','20'].map(c => `<option value="${c}">${c}</option>`).join('')}
        </select></div>
      <div class="col-4"><label class="form-label">XP</label><input class="form-control" id="monXP" type="number" value="0"></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-4"><label class="form-label">AC</label><input class="form-control" id="monAC" type="number" value="10"></div>
      <div class="col-4"><label class="form-label">HP</label><input class="form-control" id="monHP" type="number" value="1"></div>
      <div class="col-4"><label class="form-label">Init Mod</label><input class="form-control" id="monInit" type="number" value="0"></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveMonster(${eid})"><i class="fa-solid fa-plus me-1"></i>Add</button>
  `);
};

(window as any).saveMonster = async function (eid: number) {
  await api('POST', `/api/encounters/${eid}/monsters`, {
    name: (document.getElementById('monName') as HTMLInputElement).value,
    count: +(document.getElementById('monCount') as HTMLInputElement).value || 1,
    cr: (document.getElementById('monCR') as HTMLSelectElement).value,
    xp: +(document.getElementById('monXP') as HTMLInputElement).value || 0,
    ac: +(document.getElementById('monAC') as HTMLInputElement).value || 10,
    hp: +(document.getElementById('monHP') as HTMLInputElement).value || 1,
    initiative_mod: +(document.getElementById('monInit') as HTMLInputElement).value || 0,
  });
  hideModal();
  (window as any).showEncounterDetail(eid);
  toast('Monster added');
};

(window as any).editMonster = function (eid: number, mid: number) {
  const e = null; // fetch again or pass data
  const m = null;
  showModal('Edit Monster', `<p class="text-muted">Edit via encounter detail page.</p>`);
};

(window as any).deleteMonster = async function (eid: number, mid: number) {
  if (!confirm('Remove this monster?')) return;
  await api('DELETE', `/api/encounter-monsters/${mid}`);
  (window as any).showEncounterDetail(eid);
  toast('Monster removed');
};

(window as any).showEncounterXPCalc = function () {
  showModal('XP Calculator', `
    <p class="text-muted small">Enter party levels and monster CRs to calculate encounter difficulty.</p>
    <div class="mb-3"><label class="form-label">Party Levels (comma-separated)</label>
      <input class="form-control" id="xpPartyLevels" placeholder="1, 1, 2, 3" value="1,1,1,1"></div>
    <div class="mb-3"><label class="form-label">Monsters (format: CR,count per line)</label>
      <textarea class="form-control" id="xpMonsters" rows="3" placeholder="1/4,3&#10;1,1&#10;0,2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="doXPCalc()"><i class="fa-solid fa-calculator me-1"></i>Calculate</button>
    <div id="xpCalcResult" class="mt-3"></div>
  `);
};

(window as any).doXPCalc = async function () {
  const levels = (document.getElementById('xpPartyLevels') as HTMLInputElement).value.split(',').map(s => parseInt(s.trim())).filter(n => !isNaN(n));
  const lines = (document.getElementById('xpMonsters') as HTMLTextAreaElement).value.split('\n').filter(l => l.trim());
  const monsters = lines.map(l => {
    const [cr, count] = l.split(',').map(s => s.trim());
    return { cr, count: parseInt(count) || 1, name: 'Custom' };
  });
  if (!levels.length || !monsters.length) { toast('Enter party levels and monsters', true); return; }
  try {
    const result = await api('POST', '/api/encounters/calculate-xp', { party_levels: levels, monsters });
    const el = document.getElementById('xpCalcResult')!;
    el.innerHTML = `
      <div class="card mt-2"><div class="card-body">
        <div class="d-flex justify-content-between"><span class="fw-bold">Total XP</span><span>${result.total_xp}</span></div>
        <div class="d-flex justify-content-between"><span class="fw-bold">Adjusted XP</span><span>${result.adjusted_xp}</span></div>
        <div class="d-flex justify-content-between"><span class="fw-bold">Difficulty</span>
          <span class="badge ${result.difficulty === 'deadly' ? 'bg-danger' : result.difficulty === 'hard' ? 'badge-blood' : result.difficulty === 'medium' ? 'badge-gold' : 'bg-success'}">${result.difficulty}</span></div>
        <hr>
        <div class="small text-muted">Party: ${result.party_size} · Monsters: ${result.monster_count} · Mult: ${result.size_multiplier}x</div>
        <div class="small text-muted">Thresholds: Easy ${result.thresholds.easy} / Med ${result.thresholds.medium} / Hard ${result.thresholds.hard} / Deadly ${result.thresholds.deadly}</div>
      </div></div>`;
  } catch (e: any) { toast(e.message, true); }
};

// ─── Calendar ───

// ─── Timeline ───

(window as any).showTimeline = function () {
  showView('timeline');
  const el = document.getElementById('timelineContent')!;
  el.setAttribute('hx-get', '/htmx/timeline');
  el.setAttribute('hx-trigger', 'load');
  el.setAttribute('hx-swap', 'innerHTML');
  el.innerHTML = '<div class="ornament">✧ Loading timeline... ✧</div>';
  htmx.process(el);
};

// ─── Conditions / Ailments ───

(window as any).showAddCondition = function () {
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
};

(window as any).saveCondition = async function () {
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
};

(window as any).tickConditions = async function () {
  if (!currentChar) return;
  const result = await api('POST', '/api/conditions/tick', {
    character_id: currentChar.id, count: 1, duration_type: 'round',
  });
  renderCombat();
  if (result.expired > 0) toast(`${result.expired} condition(s) expired`);
  else toast('Rounds advanced');
};

(window as any).deleteCondition = async function (id: number) {
  await api('DELETE', `/api/conditions/${id}`);
  renderCombat();
};

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

(window as any).showAddFeat = function () {
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
};

let editingFeatId: number | null = null;

(window as any).showEditFeat = function (id: number, name: string, description: string, prerequisites: string, source: string, level: number) {
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
};

(window as any).saveFeat = async function () {
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
};

(window as any).deleteFeat = async function (id: number) {
  if (!confirm('Remove this feat?')) return;
  await api('DELETE', `/api/feats/${id}`);
  renderFeats();
  toast('Feat removed');
};

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
                    <span class="badge badge-muted ms-1">${esc(comp.race)}</span>
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

(window as any).showAddCompanion = function () {
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
};

(window as any).saveCompanion = async function () {
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
};

(window as any).editCompanion = async function (id: number) {
  const comps = await api('GET', `/api/companions?character_id=${currentChar.id}`);
  const comp = comps.find((c: any) => c.id === id);
  if (!comp) return;
  showModal('Edit Companion', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="compName" value="${esc(comp.name)}"></div>
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
};

(window as any).saveEditCompanion = async function (id: number) {
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
    abilities: (document.getElementById('compAbilities') as HTMLTextAreaElement).value,
    notes: (document.getElementById('compNotes') as HTMLTextAreaElement).value,
    is_alive: true,
  });
  hideModal();
  renderCompanions();
  toast('Companion updated');
};

(window as any).deleteCompanion = async function (id: number) {
  if (!confirm('Remove this companion?')) return;
  await api('DELETE', `/api/companions/${id}`);
  renderCompanions();
  toast('Companion removed');
};

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

(window as any).showAddNote = function () {
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
};

(window as any).saveNote = async function () {
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
};

(window as any).editNote = async function (id: number) {
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
};

(window as any).saveEditNote = async function (id: number) {
  await api('PUT', `/api/notes/${id}`, {
    title: (document.getElementById('noteTitle') as HTMLInputElement).value,
    content: (document.getElementById('noteContent') as HTMLTextAreaElement).value,
    visibility: (document.getElementById('noteVis') as HTMLSelectElement).value,
    category: (document.getElementById('noteCat') as HTMLSelectElement).value,
  });
  hideModal();
  renderNotes();
  toast('Note updated');
};

(window as any).deleteNote = async function (id: number) {
  if (!confirm('Delete this note?')) return;
  await api('DELETE', `/api/notes/${id}`);
  renderNotes();
  toast('Note deleted');
};

// ─── Factions View ───

(window as any).showFactions = function () {
  showView('factions');
  const el = document.getElementById('factionsContent')!;
  el.setAttribute('hx-get', '/htmx/factions');
  el.setAttribute('hx-trigger', 'load');
  el.setAttribute('hx-swap', 'innerHTML');
  el.innerHTML = '<div class="ornament">✧ Loading factions... ✧</div>';
  htmx.process(el);
};

(window as any).showShops = async function () {
  showView('shops');
  const el = document.getElementById('shopsContent')!;
  try {
    const data = await api('GET', '/api/shops');
    if (!data.length) {
      el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-store fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Shops</p><p class="small text-muted">No shops have been created yet.</p></div>';
      return;
    }
    el.innerHTML = '<div class="mb-3"><select class="form-select" id="shopSelect">' +
      data.map((s: any) => `<option value="${s.id}">${esc(s.name)}</option>`).join('') +
      '</select></div><div id="shopItems"></div>';
  } catch (e: any) { toast(e.message, true); }
};

(window as any).showOneShots = function () {
  showView('oneshot');
  const el = document.getElementById('oneshotSection')!;
  el.setAttribute('hx-get', '/htmx/oneshot-adventures');
  el.setAttribute('hx-trigger', 'load');
  el.setAttribute('hx-swap', 'innerHTML');
  el.innerHTML = '<div class="ornament">✧ Loading one-shot adventures... ✧</div>';
  htmx.process(el);
};

(window as any).loadFactionReputations = async function () {
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
};

(window as any).editReputation = function (charId: number, factionId: number, standing: number, rank: string, notes: string) {
  showModal('Set Reputation', `
    <div class="mb-3"><label class="form-label">Standing (-100 to 100)</label>
      <input class="form-control" id="repStanding" type="number" value="${standing}" min="-100" max="100"></div>
    <div class="mb-3"><label class="form-label">Rank / Title</label><input class="form-control" id="repRank" value="${esc(rank)}"></div>
    <div class="mb-3"><label class="form-label">Notes</label><textarea class="form-control" id="repNotes" rows="2">${esc(notes)}</textarea></div>
    <button class="btn btn-primary w-100" onclick="saveReputation(${charId}, ${factionId})"><i class="fa-solid fa-save me-1"></i>Save</button>
  `);
};

(window as any).saveReputation = async function (charId: number, factionId: number) {
  await api('POST', '/api/faction-reputation', {
    character_id: charId, faction_id: factionId,
    standing: +(document.getElementById('repStanding') as HTMLInputElement).value || 0,
    rank: (document.getElementById('repRank') as HTMLInputElement).value,
    notes: (document.getElementById('repNotes') as HTMLTextAreaElement).value,
  });
  hideModal();
  (window as any).loadFactionReputations();
  toast('Reputation updated');
};

// ─── Combat Section Update for conditions and concentration ───

function renderCombat() {
  const c = currentChar;
  const el = document.getElementById('combatSection')!;
  const pct = c.hp_max > 0 ? Math.round((c.hp_current / c.hp_max) * 100) : 0;
  el.innerHTML = `
    <div class="row g-3">
      <div class="col-4"><div class="combat-stat" title="Armor Class"><div class="stat-label">AC</div><div class="stat-value">${c.ac}</div></div></div>
      <div class="col-4"><div class="combat-stat" title="Initiative modifier"><div class="stat-label">Initiative</div><div class="stat-value">${c.initiative >= 0 ? '+' : ''}${c.initiative}</div></div></div>
      <div class="col-4"><div class="combat-stat" title="Movement speed"><div class="stat-label">Speed</div><div class="stat-value">${c.speed}</div></div></div>
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
    <div id="conditionsArea" class="mt-3">
      <div class="d-flex justify-content-between align-items-center">
        <h5 class="mt-0 mb-2">Conditions</h5>
        <div class="d-flex gap-1">
          <button class="btn btn-sm btn-outline-primary" onclick="showAddCondition()"><i class="fa-solid fa-plus"></i></button>
          <button class="btn btn-sm btn-outline-secondary" onclick="tickConditions()" title="Advance 1 round"><i class="fa-solid fa-forward"></i></button>
        </div>
      </div>
      <div id="conditionBadges"></div>
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
      <div class="col-6"><label class="form-label small">Successes</label><input type="number" class="form-control form-control-sm" value="${c.death_saves_successes}" oninput="autoSaveField('death_saves_successes',this)" min="0" max="3"></div>
      <div class="col-6"><label class="form-label small">Failures</label><input type="number" class="form-control form-control-sm" value="${c.death_saves_failures}" oninput="autoSaveField('death_saves_failures',this)" min="0" max="3"></div>
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
  // Load condition badges async
  loadConditionBadges();
}

async function loadConditionBadges() {
  if (!currentChar) return;
  try {
    const conds = await api('GET', `/api/conditions/summary?character_id=${currentChar.id}`);
    const el = document.getElementById('conditionBadges');
    if (!el) return;
    if (!conds.length) {
      el.innerHTML = '<div class="text-muted small fst-italic">No active conditions</div>';
      return;
    }
    const iconMap: Record<string, string> = {
      blinded: 'fa-eye-slash', charmed: 'fa-heart', deafened: 'fa-ear-deaf',
      exhaustion: 'fa-battery-quarter', frightened: 'fa-ghost', grappled: 'fa-handcuffs',
      incapacitated: 'fa-bed', invisible: 'fa-ghost', paralyzed: 'fa-snowflake',
      petrified: 'fa-monument', poisoned: 'fa-skull', prone: 'fa-person-falling',
      restrained: 'fa-lock', stunned: 'fa-star', unconscious: 'fa-circle',
      concentration: 'fa-brain',
    };
    const colorMap: Record<string, string> = {
      blinded: '#8b0000', charmed: '#dda0dd', deafened: '#666',
      exhaustion: '#ff8c00', frightened: '#4b0082', grappled: '#8b4513',
      incapacitated: '#555', invisible: '#87ceeb', paralyzed: '#00bfff',
      petrified: '#808080', poisoned: '#32cd32', prone: '#d2b48c',
      restrained: '#ffd700', stunned: '#ff4500', unconscious: '#2f4f4f',
      concentration: '#4169e1',
    };
    el.innerHTML = '<div class="d-flex flex-wrap gap-1 mb-2">' + conds.map((cond: any) => {
      const icon = iconMap[cond.type] || 'fa-circle';
      const color = colorMap[cond.type] || '#b8963e';
      const durStr = cond.duration_type === 'permanent' ? 'perm' : cond.duration + cond.duration_type.substring(0, 1);
      return `<span class="badge" style="background:${color};color:#fff;font-size:0.75rem;padding:0.3rem 0.5rem;border-radius:4px" title="${esc(cond.name)} (${durStr})">
        <i class="fa-solid ${icon} me-1"></i>${esc(cond.name)} ${durStr}
        <a href="#" onclick="deleteCondition(${cond.id});return false" class="text-white text-decoration-none ms-1">×</a>
      </span>`;
    }).join('') + '</div>';
  } catch {}
}

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

(window as any).startRecipe = async function (recipeId: number) {
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
};

(window as any).confirmStartRecipe = async function (recipeId: number, name: string, hours: number, dc: number) {
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
};

(window as any).advanceCrafting = async function (id: number) {
  try {
    await api('PUT', `/api/crafting/${id}`, { progress_hours: 1 });
    renderCrafting();
    toast('Crafting advanced by 1 hour');
  } catch (e: any) { toast(e.message, true); }
};

(window as any).completeCrafting = async function (id: number) {
  try {
    await api('PUT', `/api/crafting/${id}`, { status: 'complete' });
    renderCrafting();
    toast('Crafting completed! Item added to inventory.');
  } catch (e: any) { toast(e.message, true); }
};

(window as any).abandonCrafting = async function (id: number) {
  if (!confirm('Abandon this project?')) return;
  try {
    await api('PUT', `/api/crafting/${id}`, { status: 'abandoned' });
    renderCrafting();
    toast('Project abandoned');
  } catch (e: any) { toast(e.message, true); }
};

(window as any).showStartCrafting = function () {
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
};

(window as any).confirmCustomCraft = async function () {
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
};

// ─── HP Auto-Calc in details ───

(window as any).calcHP = async function () {
  if (!currentChar) return;
  try {
    const result = await api('POST', `/api/characters/${currentChar.id}/calc-hp`);
    currentChar = await api('GET', `/api/characters/${currentChar.id}`);
    renderSheet();
    toast(`HP calculated: ${result.hp_max} HP`);
  } catch (e: any) { toast(e.message, true); }
};

// ─── Random Character Generator ───

(window as any).generateRandomChar = async function () {
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
};

// ─── Character Comparison ───

(window as any).showComparison = async function () {
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
};

// ─── Add character comparison to character list view ───

let compareMode = false;

(window as any).toggleCompareMode = function () {
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
};

// ─── Combat Tracker ───

(window as any).showCombatTracker = async function () {
  showView('combatTracker');
  const el = document.getElementById('combatTrackerContent')!;
  el.innerHTML = '<div class="ornament">✧ Loading combat tracker... ✧</div>';
  try {
    const [entries, campaigns] = await Promise.all([
      api('GET', '/api/combat'),
      api('GET', '/api/campaigns'),
    ]);
    if (!entries.length) {
      el.innerHTML = '<div class="empty-state"><i class="fa-solid fa-swords fa-3x mb-2 d-block text-muted"></i><p class="fw-bold">No Combatants</p><p class="small text-muted">Add combat entries from a character sheet or create them here.</p><button class="btn btn-gold btn-sm mt-2" onclick="showAddCombatEntry()"><i class="fa-solid fa-plus me-1"></i>Add Combatant</button></div>';
      return;
    }
    const sorted = [...entries].sort((a: any, b: any) => b.initiative_roll - a.initiative_roll || b.turn_order - a.turn_order);
    const isActive = (entry: any) => entry.is_active;

    let html = `<div class="d-flex justify-content-between align-items-center mb-3 flex-wrap gap-2">
      <div class="d-flex gap-2">
        <button class="btn btn-gold btn-sm" onclick="showAddCombatEntry()"><i class="fa-solid fa-plus me-1"></i>Add</button>
        <button class="btn btn-outline-primary btn-sm" onclick="rollAllInitiative()"><i class="fa-solid fa-dice me-1"></i>Roll Init</button>
        <button class="btn btn-outline-secondary btn-sm" onclick="advanceCombatTurn()"><i class="fa-solid fa-forward me-1"></i>Next Turn</button>
      </div>
    </div>
    <div class="table-responsive">
      <table class="table table-hover align-middle mb-0" id="combatTrackerTable">
        <thead><tr>
          <th style="width:30px"></th>
          <th>Name</th>
          <th style="width:60px">Init</th>
          <th style="width:80px">AC</th>
          <th style="width:120px">HP</th>
          <th style="width:60px">Status</th>
          <th style="width:140px">Actions</th>
        </tr></thead>
        <tbody id="combatTrackerBody">`;

    for (const entry of sorted) {
      const active = entry.is_active;
      const hpPct = entry.hp_max > 0 ? Math.round((entry.hp_current / entry.hp_max) * 100) : 0;
      const hpColor = hpPct > 50 ? 'var(--bs-success)' : hpPct > 25 ? 'var(--gold)' : 'var(--bs-danger)';
      const rowClass = active ? 'table-active fw-bold' : '';
      const icon = entry.type === 'character' ? 'fa-user' : entry.type === 'monster' ? 'fa-dragon' : 'fa-user-group';
      html += `<tr class="${rowClass}" draggable="true" id="ce-${entry.id}"
        ondragstart="dragCombatEntry(event, ${entry.id})"
        ondrop="dropCombatEntry(event, ${entry.id})"
        ondragover="event.preventDefault()">
        <td class="text-muted" style="cursor:grab"><i class="fa-solid fa-grip-vertical"></i></td>
        <td><i class="fa-solid ${icon} me-1 text-muted"></i>${esc(entry.name)}
          ${entry.type === 'character' ? '<span class="badge badge-blood ms-1" style="font-size:0.6rem">PC</span>' : ''}
        </td>
        <td class="text-center fw-bold">${entry.initiative_roll > 0 ? entry.initiative_roll : '-'}</td>
        <td class="text-center">${entry.ac}</td>
        <td>
          <div class="d-flex align-items-center gap-1">
            <div class="hp-bar flex-grow-1" style="height:6px;min-width:50px">
              <div class="hp-bar-fill" style="width:${hpPct}%;height:100%;background:${hpColor}"></div>
            </div>
            <span class="small" style="font-size:0.7rem;white-space:nowrap">${entry.hp_current}/${entry.hp_max}</span>
          </div>
          <div class="d-flex gap-1 mt-1">
            <input type="number" class="form-control form-control-sm" id="qdamage-${entry.id}" placeholder="dmg" style="width:55px;font-size:0.7rem;height:24px">
            <button class="btn btn-sm btn-danger py-0 px-1" style="font-size:0.65rem;height:24px" onclick="combatTrackerDamage(${entry.id})"><i class="fa-solid fa-minus"></i></button>
            <button class="btn btn-sm btn-success py-0 px-1" style="font-size:0.65rem;height:24px" onclick="combatTrackerHeal(${entry.id})"><i class="fa-solid fa-plus"></i></button>
          </div>
        </td>
        <td class="text-center">
          <button class="btn btn-sm ${active ? 'btn-gold' : 'btn-outline-secondary'} py-0 px-1" style="font-size:0.65rem" onclick="toggleCombatActive(${entry.id})">
            ${active ? '<i class="fa-solid fa-check"></i>' : '<i class="fa-solid fa-pause"></i>'}
          </button>
        </td>
        <td>
          <div class="d-flex gap-1">
            <button class="btn btn-sm btn-outline-danger py-0 px-1" style="font-size:0.65rem" onclick="deleteCombatEntry(${entry.id})"><i class="fa-solid fa-trash"></i></button>
          </div>
        </td>
      </tr>`;
    }
    html += '</tbody></table></div>';
    el.innerHTML = html;
  } catch (e: any) {
    el.innerHTML = `<div class="empty-state"><p class="small text-muted">Error: ${esc(e.message)}</p></div>`;
  }
};

async function findCombatEntry(id: number): Promise<any> {
  const entries = await api('GET', '/api/combat');
  return entries.find((e: any) => e.id === id);
}

(window as any).combatTrackerDamage = async function (id: number) {
  const input = document.getElementById('qdamage-' + id) as HTMLInputElement;
  const dmg = parseInt(input?.value || '0');
  if (!dmg) return;
  try {
    const entry = await findCombatEntry(id);
    if (!entry) { toast('Entry not found', true); return; }
    entry.hp_current = Math.max(0, entry.hp_current - dmg);
    await api('PUT', '/api/combat/' + id, entry);
    (window as any).showCombatTracker();
  } catch (e: any) { toast(e.message, true); }
};

(window as any).combatTrackerHeal = async function (id: number) {
  const input = document.getElementById('qdamage-' + id) as HTMLInputElement;
  const heal = parseInt(input?.value || '0');
  if (!heal) return;
  try {
    const entry = await findCombatEntry(id);
    if (!entry) { toast('Entry not found', true); return; }
    entry.hp_current = Math.min(entry.hp_max, entry.hp_current + heal);
    await api('PUT', '/api/combat/' + id, entry);
    (window as any).showCombatTracker();
  } catch (e: any) { toast(e.message, true); }
};

(window as any).toggleCombatActive = async function (id: number) {
  try {
    const entry = await findCombatEntry(id);
    if (!entry) { toast('Entry not found', true); return; }
    entry.is_active = !entry.is_active;
    await api('PUT', '/api/combat/' + id, entry);
    (window as any).showCombatTracker();
  } catch (e: any) { toast(e.message, true); }
};

(window as any).deleteCombatEntry = async function (id: number) {
  if (!confirm('Remove this combatant?')) return;
  try {
    await api('DELETE', '/api/combat/' + id);
    (window as any).showCombatTracker();
  } catch (e: any) { toast(e.message, true); }
};

(window as any).rollAllInitiative = async function () {
  try {
    const entries = await api('GET', '/api/combat');
    for (const e of entries) {
      const roll = Math.floor(Math.random() * 20) + 1 + (e.initiative_mod || 0);
      e.initiative_roll = roll;
      try { await api('PUT', '/api/combat/' + e.id, e); } catch {}
    }
    (window as any).showCombatTracker();
    toast('Initiative rolled for all combatants');
  } catch (e: any) { toast(e.message, true); }
};

(window as any).advanceCombatTurn = async function () {
  try {
    const result = await api('POST', '/api/combat/next-turn');
    (window as any).showCombatTracker();
    toast(result.current_entry ? `Turn: ${result.current_entry.name}` : 'Turn advanced');
  } catch (e: any) { toast(e.message, true); }
};

(window as any).showAddCombatEntry = function () {
  showModal('Add Combatant', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="ceName"></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Type</label>
        <select class="form-select" id="ceType"><option value="character">Character</option><option value="monster">Monster</option><option value="npc">NPC</option></select></div>
      <div class="col-6"><label class="form-label">AC</label><input class="form-control" id="ceAC" type="number" value="10"></div>
    </div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">HP Max</label><input class="form-control" id="ceHPMax" type="number" value="10"></div>
      <div class="col-6"><label class="form-label">Init Mod</label><input class="form-control" id="ceInitMod" type="number" value="0"></div>
    </div>
    <button class="btn btn-primary w-100" onclick="saveNewCombatEntry()"><i class="fa-solid fa-plus me-1"></i>Add</button>
  `);
};

(window as any).saveNewCombatEntry = async function () {
  await api('POST', '/api/combat', {
    name: (document.getElementById('ceName') as HTMLInputElement).value,
    type: (document.getElementById('ceType') as HTMLSelectElement).value,
    ac: +(document.getElementById('ceAC') as HTMLInputElement).value || 10,
    hp_max: +(document.getElementById('ceHPMax') as HTMLInputElement).value || 10,
    hp_current: +(document.getElementById('ceHPMax') as HTMLInputElement).value || 10,
    initiative_mod: +(document.getElementById('ceInitMod') as HTMLInputElement).value || 0,
  });
  hideModal();
  (window as any).showCombatTracker();
  toast('Combatant added');
};

let draggedCombatId: number | null = null;

(window as any).dragCombatEntry = function (ev: any, id: number) {
  draggedCombatId = id;
  ev.dataTransfer.effectAllowed = 'move';
};

(window as any).dropCombatEntry = async function (ev: any, targetId: number) {
  ev.preventDefault();
  if (draggedCombatId === null || draggedCombatId === targetId) return;
  try {
    const entries: any[] = await api('GET', '/api/combat');
    const dragged = entries.find((e: any) => e.id === draggedCombatId);
    const target = entries.find((e: any) => e.id === targetId);
    if (!dragged || !target) return;
    const tempOrder = dragged.turn_order;
    dragged.turn_order = target.turn_order;
    target.turn_order = tempOrder;
    await api('PUT', '/api/combat/' + dragged.id, dragged);
    await api('PUT', '/api/combat/' + target.id, target);
    draggedCombatId = null;
    (window as any).showCombatTracker();
    toast('Reordered');
  } catch (e: any) { toast(e.message, true); }
};

// ─── Shops ───

// ─── Wiki ───

(window as any).showWiki = async function (campaignId?: number) {
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
};

(window as any).toggleWikiSidebar = function () {
  const offcanvas = document.getElementById('wikiOffcanvas');
  if (offcanvas) bootstrap.Offcanvas.getOrCreateInstance(offcanvas).toggle();
};

function buildWikiChildren(parentId: number, childMap: Record<number, any[]>, depth: number): string {
  const children = childMap[parentId] || [];
  if (!children.length) return '';
  const pad = depth * 16;
  return children.map((c: any) =>
    `<a href="#" class="list-group-item list-group-item-action py-1 ps-${3 + depth}" style="padding-left:${pad + 16}px!important;font-size:0.9rem" onclick="loadWikiPage(${c.id});if(window.innerWidth<768){const o=document.getElementById('wikiOffcanvas');if(o){bootstrap.Offcanvas.getInstance(o)?.hide()}}">↳ ${esc(c.title)}</a>
    ${buildWikiChildren(c.id, childMap, depth + 1)}`
  ).join('');
}

(window as any).loadWikiPage = async function (pageId: number) {
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
};

(window as any).showAddWikiPage = function (campaignId: number) {
  showModal('New Wiki Page', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="wikiTitle"></div>
    <div class="mb-3"><label class="form-label">Content (Markdown)</label><textarea class="form-control" id="wikiContent" rows="8" placeholder="Write in Markdown..."></textarea></div>
    <div class="mb-3"><label class="form-label">Visibility</label>
      <select class="form-select" id="wikiVis"><option value="public">Public</option><option value="dm-only">DM Only</option></select></div>
    <button class="btn btn-primary w-100" onclick="saveWikiPage(${campaignId})">Create Page</button>
  `);
};

(window as any).saveWikiPage = async function (campaignId: number) {
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
};

(window as any).showEditWikiPage = function (id: number, title: string, content: string, visibility: string) {
  showModal('Edit Wiki Page', `
    <div class="mb-3"><label class="form-label">Title</label><input class="form-control" id="wikiTitle" value="${esc(title)}"></div>
    <div class="mb-3"><label class="form-label">Content (Markdown)</label><textarea class="form-control" id="wikiContent" rows="8">${esc(content)}</textarea></div>
    <div class="mb-3"><label class="form-label">Visibility</label>
      <select class="form-select" id="wikiVis"><option value="public" ${visibility === 'public' ? 'selected' : ''}>Public</option><option value="dm-only" ${visibility === 'dm-only' ? 'selected' : ''}>DM Only</option></select></div>
    <button class="btn btn-primary w-100" onclick="saveEditWikiPage(${id})">Save</button>
  `);
};

(window as any).saveEditWikiPage = async function (id: number) {
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
};

(window as any).deleteWikiPage = async function (id: number) {
  if (!confirm('Delete this wiki page?')) return;
  try {
    await api('DELETE', `/api/wiki/${id}`);
    const cid = await api('GET', '/api/campaigns').then((cs: any[]) => cs[0]?.id);
    (window as any).showWiki(cid);
    toast('Wiki page deleted');
  } catch (e: any) { toast(e.message, true); }
};

// ─── Campaign Graph ───

(window as any).showCampaignGraph = async function (campaignId: number) {
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
};

// ─── One-Shot Tree UI (SortableJS Drag-Reorder) ───

(window as any).initOneShotTree = function (adventureId: number) {
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
};

// ─── One-Shot Items ───

(window as any).showOneShotItemForm = function (adventureId: number) {
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
};

(window as any).saveOneShotItem = async function (adventureId: number) {
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
};

(window as any).editOneShotItem = async function (itemId: number) {
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
};

(window as any).updateOneShotItem = async function (itemId: number) {
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
};

(window as any).deleteOneShotItem = async function (itemId: number) {
  if (!confirm('Delete this item?')) return;
  await api('DELETE', `/api/oneshot-items/${itemId}`);
  toast('Item deleted');
  const itemsCard = document.querySelector('[hx-get*="/items"]');
  if (itemsCard) htmx.trigger(itemsCard, 'load');
};

// ─── One-Shot Shops ───

(window as any).showOneShotShopForm = function (adventureId: number) {
  showModal('Add Shop', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="shopName"></div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="shopDesc" rows="2"></textarea></div>
    <div class="row g-3 mb-3">
      <div class="col-6"><label class="form-label">Sell Markup %</label><input class="form-control" id="shopMarkup" type="number" value="100"></div>
      <div class="col-6"><label class="form-label">Buy Markup %</label><input class="form-control" id="shopBuyMarkup" type="number" value="50"></div>
    </div>
    <button class="btn btn-primary w-100" onclick="createOneShotShop(${adventureId})">Create</button>
  `);
};

(window as any).createOneShotShop = async function (adventureId: number) {
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
};

(window as any).deleteOneShotShop = async function (shopId: number) {
  if (!confirm('Delete this shop?')) return;
  await api('DELETE', `/api/oneshot-adventures/0/shops/${shopId}`);
  toast('Shop deleted');
  const shopsCard = document.querySelector('[hx-get*="/shops"]');
  if (shopsCard) htmx.trigger(shopsCard, 'load');
};

// ─── One-Shot Monsters ───

(window as any).showAddMonsterForm = function (adventureId: number) {
  showModal('Add Monster', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="monsterName"></div>
    <div class="row g-3 mb-3">
      <div class="col-3"><label class="form-label">AC</label><input class="form-control" id="monsterAC" type="number" value="10"></div>
      <div class="col-3"><label class="form-label">HP</label><input class="form-control" id="monsterHP" type="number" value="10"></div>
      <div class="col-3"><label class="form-label">CR</label><input class="form-control" id="monsterCR" placeholder="1/2"></div>
      <div class="col-3"><label class="form-label">Source</label><input class="form-control" id="monsterSource" value="custom"></div>
    </div>
    <div class="row g-3 mb-3">
      ${['str','dex','con','int','wis','cha'].map(s => `<div class="col-2"><label class="form-label">${s.toUpperCase()}</label><input class="form-control" id="monster${s.toUpperCase()}" type="number" value="10"></div>`).join('')}
    </div>
    <div class="mb-3"><label class="form-label">Description</label><textarea class="form-control" id="monsterDesc" rows="2"></textarea></div>
    <div class="mb-3"><label class="form-label">Special Abilities</label><textarea class="form-control" id="monsterAbilities" rows="3"></textarea></div>
    <div class="mb-3"><label class="form-label">Actions</label><textarea class="form-control" id="monsterActions" rows="3"></textarea></div>
    <button class="btn btn-primary w-100" onclick="saveAdventureMonster(${adventureId})">Add Monster</button>
  `);
};

(window as any).saveAdventureMonster = async function (adventureId: number) {
  const name = (document.getElementById('monsterName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  await api('POST', `/api/oneshot-acts/0/monsters`, {
    name,
    adventure_id: adventureId,
    ac: parseInt((document.getElementById('monsterAC') as HTMLInputElement).value) || 10,
    hp: parseInt((document.getElementById('monsterHP') as HTMLInputElement).value) || 10,
    cr: (document.getElementById('monsterCR') as HTMLInputElement).value || '0',
    source: (document.getElementById('monsterSource') as HTMLInputElement).value || 'custom',
    str: parseInt((document.getElementById('monsterSTR') as HTMLInputElement).value) || 10,
    dex: parseInt((document.getElementById('monsterDEX') as HTMLInputElement).value) || 10,
    con: parseInt((document.getElementById('monsterCON') as HTMLInputElement).value) || 10,
    int_: parseInt((document.getElementById('monsterINT') as HTMLInputElement).value) || 10,
    wis: parseInt((document.getElementById('monsterWIS') as HTMLInputElement).value) || 10,
    cha: parseInt((document.getElementById('monsterCHA') as HTMLInputElement).value) || 10,
    special_abilities: (document.getElementById('monsterAbilities') as HTMLTextAreaElement).value,
    actions: (document.getElementById('monsterActions') as HTMLTextAreaElement).value,
    is_full: 1,
  });
  hideModal();
  toast('Monster added');
  const monstersCard = document.querySelector('[hx-get*="/monsters"]');
  if (monstersCard) htmx.trigger(monstersCard, 'load');
};

(window as any).deleteOneShotMonster = async function (monsterId: number) {
  if (!confirm('Delete this monster?')) return;
  await api('DELETE', `/api/oneshot-monsters/${monsterId}`);
  toast('Monster deleted');
  const monstersCard = document.querySelector('[hx-get*="/monsters"]');
  if (monstersCard) htmx.trigger(monstersCard, 'load');
};

// Monster Library
(window as any).showMonsterLibrary = function (adventureId: number) {
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
};

async function loadMonsterLibrary(adventureId: number) {
  const list = document.getElementById('libraryList');
  if (!list) return;
  try {
    const monsters = await api('GET', '/api/monster-library');
    if (!monsters.length) {
      list.innerHTML = '<div class="text-muted small fst-italic py-2">No monsters in library yet.</div>';
      return;
    }
    (window as any)._libraryMonsters = monsters;
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

(window as any).filterLibraryMonsters = function () {
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
};

(window as any).showAddLibraryMonster = function (adventureId: number) {
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
};

(window as any).createLibraryMonster = async function (adventureId: number) {
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
};

(window as any).quickAddLibraryMonster = async function (adventureId: number, libraryId: number) {
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
};

(window as any).deleteLibraryMonster = async function (libraryId: number, adventureId: number) {
  if (!confirm('Delete from library?')) return;
  await api('DELETE', `/api/monster-library/${libraryId}`);
  toast('Library entry deleted');
  if (adventureId) loadMonsterLibrary(adventureId);
};

// Act-level monster display
(window as any).showActMonsters = async function (actId: number) {
  try {
    const monsters = await api('GET', `/api/oneshot-acts/${actId}/monsters`);
    showModal('Act Monsters', `
      <div class="mb-2"><button class="btn btn-sm btn-outline-primary" onclick="showAddActMonster(${actId})"><i class="fa-solid fa-plus me-1"></i>Add Monster</button></div>
      ${monsters.length ? monsters.map((m: any) => `
        <div class="inv-item">
          <div><strong>${esc(m.name)}</strong> <span class="badge bg-danger">CR ${esc(m.cr)}</span> <span class="text-muted small">AC ${m.ac} · HP ${m.hp}</span></div>
          <button class="btn btn-sm btn-outline-danger" onclick="deleteOneShotMonster(${m.id})"><i class="fa-solid fa-trash"></i></button>
        </div>
      `).join('') : '<div class="text-muted small fst-italic">No monsters in this act.</div>'}
    `);
  } catch (e: any) { toast(e.message, true); }
};

(window as any).showAddActMonster = function (actId: number) {
  // Reuse the same monster form but post to act endpoint
  showModal('Add Monster to Act', `
    <div class="mb-3"><label class="form-label">Name</label><input class="form-control" id="actMonsterName"></div>
    <div class="row g-3 mb-3">
      <div class="col-3"><label class="form-label">AC</label><input class="form-control" id="actMonsterAC" type="number" value="10"></div>
      <div class="col-3"><label class="form-label">HP</label><input class="form-control" id="actMonsterHP" type="number" value="10"></div>
      <div class="col-3"><label class="form-label">CR</label><input class="form-control" id="actMonsterCR" placeholder="1/2"></div>
      <div class="col-3"><label class="form-label">Source</label><input class="form-control" id="actMonsterSource" value="custom"></div>
    </div>
    <div class="row g-3 mb-3">
      ${['str','dex','con','int','wis','cha'].map(s => `<div class="col-2"><label class="form-label">${s.toUpperCase()}</label><input class="form-control" id="actMonster${s.toUpperCase()}" type="number" value="10"></div>`).join('')}
    </div>
    <button class="btn btn-primary w-100" onclick="saveActMonster(${actId})">Add</button>
  `);
};

(window as any).saveActMonster = async function (actId: number) {
  const name = (document.getElementById('actMonsterName') as HTMLInputElement).value;
  if (!name) { toast('Name required', true); return; }
  await api('POST', `/api/oneshot-acts/${actId}/monsters`, {
    name,
    ac: parseInt((document.getElementById('actMonsterAC') as HTMLInputElement).value) || 10,
    hp: parseInt((document.getElementById('actMonsterHP') as HTMLInputElement).value) || 10,
    cr: (document.getElementById('actMonsterCR') as HTMLInputElement).value || '0',
    source: (document.getElementById('actMonsterSource') as HTMLInputElement).value || 'custom',
    str: parseInt((document.getElementById('actMonsterSTR') as HTMLInputElement).value) || 10,
    dex: parseInt((document.getElementById('actMonsterDEX') as HTMLInputElement).value) || 10,
    con: parseInt((document.getElementById('actMonsterCON') as HTMLInputElement).value) || 10,
    int_: parseInt((document.getElementById('actMonsterINT') as HTMLInputElement).value) || 10,
    wis: parseInt((document.getElementById('actMonsterWIS') as HTMLInputElement).value) || 10,
    cha: parseInt((document.getElementById('actMonsterCHA') as HTMLInputElement).value) || 10,
    is_full: 1,
  });
  hideModal();
  toast('Monster added');
  const monstersCard = document.querySelector('[hx-get*="/monsters"]');
  if (monstersCard) htmx.trigger(monstersCard, 'load');
};

// ─── One-Shot Linked Player Characters ───

(window as any).showLinkPCForm = function (adventureId: number) {
  showModal('Link Character', `
    <p class="text-muted small mb-3">Search for a character to link to this one-shot.</p>
    <div class="mb-3"><input class="form-control" id="pcSearchInput" placeholder="Search characters..." oninput="searchCharactersForLink(${adventureId})"></div>
    <div id="pcLinkResults" class="mb-3" style="max-height:300px;overflow-y:auto"></div>
  `);
};

(window as any).searchCharactersForLink = async function (adventureId: number) {
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
};

(window as any).linkPCToOneShot = async function (adventureId: number, charId: number) {
  await api('POST', `/api/oneshot-adventures/${adventureId}/characters`, { character_id: charId });
  hideModal();
  toast('Character linked');
  const pcsCard = document.querySelector('[hx-get*="/pcs"]');
  if (pcsCard) htmx.trigger(pcsCard, 'load');
};

(window as any).unlinkPCFromOneShot = async function (adventureId: number, charId: number) {
  if (!confirm('Unlink this character?')) return;
  await api('DELETE', `/api/oneshot-adventures/${adventureId}/characters/${charId}`);
  toast('Character unlinked');
  const pcsCard = document.querySelector('[hx-get*="/pcs"]');
  if (pcsCard) htmx.trigger(pcsCard, 'load');
};

// ─── NPC↔Item Links ───

(window as any).showLinkNPCToItem = function (adventureId: number, itemId: number) {
  showModal('Link NPC to Item', `
    <p class="text-muted small mb-3">Find an NPC in this adventure to link:</p>
    <div class="mb-3"><input class="form-control" id="npcLinkSearch" placeholder="Search NPCs..." oninput="searchNPCsForLink(${adventureId}, ${itemId})"></div>
    <div id="npcLinkResults" class="mb-3" style="max-height:300px;overflow-y:auto"></div>
  `);
  (window as any).searchNPCsForLink(adventureId, itemId);
};

(window as any).searchNPCsForLink = async function (adventureId: number, itemId: number) {
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
};

(window as any).linkNPCToItem = async function (adventureId: number, npcId: number, itemId: number) {
  await api('POST', `/api/oneshot-adventures/${adventureId}/npc-item-links`, { npc_id: npcId, item_id: itemId });
  hideModal();
  toast('NPC linked to item');
  const itemsCard = document.querySelector('[hx-get*="/items"]');
  if (itemsCard) htmx.trigger(itemsCard, 'load');
};

(window as any).unlinkNPCFromItem = async function (linkId: number) {
  if (!confirm('Remove link?')) return;
  await api('DELETE', `/api/npc-item-links/${linkId}`);
  toast('Link removed');
  const itemsCard = document.querySelector('[hx-get*="/items"]');
  if (itemsCard) htmx.trigger(itemsCard, 'load');
};

// ─── Polymorphic File Uploads ───

(window as any).showUploadModal = function (ownerType: string, ownerId: number) {
  showModal('Upload File', `
    <div class="mb-3">
      <label class="form-label">Select file</label>
      <input type="file" class="form-control" id="uploadFileInput">
    </div>
    <button class="btn btn-primary w-100" onclick="doUpload('${ownerType}', ${ownerId})"><i class="fa-solid fa-upload me-1"></i>Upload</button>
  `);
};

(window as any).doUpload = async function (ownerType: string, ownerId: number) {
  const input = document.getElementById('uploadFileInput') as HTMLInputElement;
  if (!input?.files?.length) { toast('Select a file', true); return; }
  const form = new FormData();
  form.append('file', input.files[0]);
  form.append('owner_type', ownerType);
  form.append('owner_id', String(ownerId));
  try {
    const res = await fetch('/api/upload', { method: 'POST', body: form,
      headers: { 'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || csrfToken }
    });
    if (!res.ok) throw new Error((await res.json()).error || 'Upload failed');
    hideModal();
    toast('File uploaded');
  } catch (e: any) { toast(e.message, true); }
};

// ─── Show combat nav for admin ───
// (handled in init by checking role)

// ═══════════════════════════════════════════
// Campaign Completeness Enhancements
// ═══════════════════════════════════════════

// ─── Campaign Dashboard ───

(window as any).showCampaignDashboard = async function (campaignId: number, campaignName: string) {
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
};

// ─── Party Inventory & Treasury ───

(window as any).showPartyInventory = async function (campaignId: number) {
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
};

(window as any).addPartyItem = async function (campaignId: number) {
  showModal('Add Party Item', `
    <div class="mb-2"><label class="form-label">Item Name</label><input class="form-control" id="piName"></div>
    <div class="mb-2"><label class="form-label">Quantity</label><input class="form-control" id="piQty" type="number" value="1"></div>
    <div class="mb-2"><label class="form-label">Notes</label><textarea class="form-control" id="piNotes" rows="2"></textarea></div>
    <button class="btn btn-primary w-100" onclick="savePartyItem(${campaignId})">Add</button>
  `);
};

(window as any).savePartyItem = async function (campaignId: number) {
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
};

(window as any).deletePartyItem = async function (campaignId: number, itemId: number) {
  if (!confirm('Remove this item?')) return;
  await api('DELETE', `/api/party-items/${itemId}`);
  toast('Item removed');
  (window as any).showPartyInventory(campaignId);
};

// ─── Session Planner ───

(window as any).showSessionPlanner = async function (campaignId: number) {
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
};

(window as any).showSessionPlanForm = function (campaignId: number, plan?: any) {
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
};

(window as any).saveSessionPlan = async function (campaignId: number, planId?: number) {
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
};

(window as any).deleteSessionPlan = async function (planId: number, campaignId: number) {
  if (!confirm('Delete this session plan?')) return;
  await api('DELETE', `/api/session-plans/${planId}`);
  toast('Session plan deleted');
  (window as any).showSessionPlanner(campaignId);
};

// ─── Encounter Difficulty Calculator ───

const CR_XP: Record<string, number> = {
  '0': 10, '1/8': 25, '1/4': 50, '1/2': 100, '1': 200, '2': 450, '3': 700,
  '4': 1100, '5': 1800, '6': 2300, '7': 2900, '8': 3900, '9': 5000, '10': 5900,
  '11': 7200, '12': 8400, '13': 10000, '14': 11500, '15': 13000, '16': 15000,
  '17': 18000, '18': 20000, '19': 22000, '20': 25000, '21': 33000, '22': 41000,
  '23': 50000, '24': 62000, '25': 75000, '30': 155000,
};

(window as any).showEncounterDifficulty = function () {
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
};

(window as any).addMonsterRow = function () {
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
};

(window as any).calcEncounterDifficulty = function () {
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
};

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

function rollDice(dice: string): number {
  const m = dice.match(/^(\d+)d(\d+)$/);
  if (!m) return 0;
  const count = parseInt(m[1]);
  const sides = parseInt(m[2]);
  let total = 0;
  for (let i = 0; i < count; i++) {
    total += Math.floor(Math.random() * sides) + 1;
  }
  return total;
}

(window as any).showTreasureGenerator = function () {
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
};

(window as any).generateTreasure = function () {
  const lvl = parseInt((document.getElementById('tgLevel') as HTMLSelectElement).value) || 5;
  const diff = (document.getElementById('tgDifficulty') as HTMLSelectElement).value;
  const resultEl = document.getElementById('tgResult');
  if (!resultEl) return;

  const table = TREASURE_TABLES[diff];
  const lines: string[] = [];
  let totalGp = 0;

  for (const entry of table) {
    const rolled = rollDice(entry.dice);
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
};

// ─── Add Dashboard button to campaign cards ───
// (patched into showParty directly, but helper here for when called from manage modal)

(window as any).openCampaignDashboard = function (campaignId: number, name: string) {
  (window as any).showCampaignDashboard(campaignId, name);
};

init();
