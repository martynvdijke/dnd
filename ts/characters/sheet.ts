// @ts-nocheck
/**
 * Character sheet rendering — tab bar, section switching, steppers, auto-save.
 * Extracted from app.ts (address-tech-debt-and-ux).
 */
import { expose } from '../lib/expose';
import { currentChar, currentTab, setCurrentTab } from '../lib/state';
import { esc, capitalize, toast } from '../lib/dom';
import { api, getApiToken } from '../lib/api';
import { markDirty, isDirty, isSaving, saveCharacter } from '../lib/save';
import { renderDiceTab } from '../dice';
import { renderStats, renderXPBar } from './stats';
import { renderCombat } from './combat';
import { wealthTotalGp } from './resources';

declare const htmx: any;

import { sections } from '../lib/tabs';
export { sections };

const htmxTabs = ['spells', 'features', 'feats', 'companions', 'crafting', 'notes'];

export function renderSheet() {
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
    const el = document.getElementById(s + 'Section');
    if (el) el.style.display = s === currentTab ? 'block' : 'none';
  });

  renderStats();
  renderCombat();
  // graph/analytics moved to the party view (reorganize-sheet-tabs); their section divs no longer exist
  (window as any).renderCrafting?.();
  (window as any).renderDetails?.();
  renderDiceTab();
  applySheetReadonly();
  ensureSheetAccordion();
  ensureSheetQuickActions();
  (window as any).updateSaveBtnState?.();
}

// Read-only mode for characters without edit rights (linked characters, campaign members)
function sheetCanEdit(): boolean {
  return (window as any).canEditCharacter !== false;
}

function applySheetReadonly() {
  const el = document.getElementById('sheetView');
  if (!el) return;
  const ro = !sheetCanEdit();
  el.classList.toggle('readonly', ro);
  if (ro) {
    el.querySelectorAll('input, textarea, select').forEach((i: any) => { i.disabled = true; });
  }
}

export function switchTab(tab: string) {
  // Save-on-tab-switch: persist any unsaved changes before rendering the next tab
  if (isDirty() && !isSaving()) {
    saveCharacter();
  }
  if (tab === 'party') { (window as any).showView?.('party'); return; }
  setCurrentTab(tab);
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
      case 'inventory': (window as any).renderInventory?.(); break;
      case 'resources': (window as any).renderResources?.(); break;
      case 'journal': (window as any).renderJournal?.(); break;
    }
    applySheetReadonly();
  }
}

// ─── Helper: Render a stepper control ───

export function renderStepper(field: string, value: number, delta: number, min?: number, max?: number, label?: string, size?: string): string {
  const sizeClass = size === 'lg' ? 'stepper-lg' : size === 'sm' ? 'currency-stepper' : '';
  const ariaInc = label ? `Increase ${label}` : `Increase ${field}`;
  const ariaDec = label ? `Decrease ${label}` : `Decrease ${field}`;
  return `<span class="stepper ${sizeClass}">
    <button class="stepper-btn" onclick="stepperField('${field}', -${Math.abs(delta)}, ${min ?? 'undefined'}, ${max ?? 'undefined'})" aria-label="${ariaDec}">−</button>
    <span class="stepper-value" onclick="editStepperValue('${field}', this)">${value}</span>
    <button class="stepper-btn" onclick="stepperField('${field}', ${Math.abs(delta)}, ${min ?? 'undefined'}, ${max ?? 'undefined'})" aria-label="${ariaInc}">+</button>
  </span>`;
}

// ─── Auto-save (thin wrappers → centralized ts/lib/save.ts) ───

export function autoSaveField(field: string, el: HTMLElement) {
  const input = el as HTMLInputElement;
  const isCheckbox = input.type === 'checkbox';
  const isTextarea = el.tagName === 'TEXTAREA';
  const raw = isCheckbox ? input.checked : (el as any).value;
  const num = parseFloat(String(raw));
  const finalVal = !isNaN(num) && !isCheckbox && !isTextarea ? num : raw;
  if (!currentChar) return;
  currentChar[field] = finalVal;
  markDirty();
}

export function stepperField(field: string, delta: number, min?: number, max?: number) {
  if (!currentChar) return;
  let val = currentChar[field] ?? 0;
  val += delta;
  if (min !== undefined) val = Math.max(min, val);
  if (max !== undefined) val = Math.min(max, val);
  currentChar[field] = val;
  markDirty();
  renderSheet();
}

export function editStepperValue(field: string, el: HTMLElement) {
  if (!currentChar) return;
  const current = currentChar[field] ?? 0;
  el.innerHTML = `<input type="number" class="form-control stepper-inline-input" value="${current}">`;
  const input = el.querySelector('input')!;
  input.focus();
  input.select();
  const save = () => {
    const parsed = parseInt(input.value);
    if (!isNaN(parsed)) {
      currentChar[field] = parsed;
      markDirty();
    }
    renderSheet();
  };
  input.addEventListener('blur', save);
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') { e.preventDefault(); save(); }
    if (e.key === 'Escape') { renderSheet(); }
  });
}

export async function coinStepper(coin: string, delta: number) {
  if (!currentChar) return;
  const currency = currentChar.currency || {};
  const current = currency[coin] || 0;
  const newVal = Math.max(0, current + delta);
  currency[coin] = newVal;
  currentChar.currency = currency;
  const updates: Record<string, number> = {};
  ['cp','sp','ep','gp','pp'].forEach(c => { updates[c] = currency[c] || 0; });
  try {
    await api('PUT', `/api/characters/${currentChar.id}/currency`, updates);
    toast(`${coin.toUpperCase()} ${delta > 0 ? '+' : ''}${delta}`);
  } catch (e: any) { toast(e.message, true); }
  renderSheet();
  if (currentTab === 'resources') { (window as any).renderResources?.(); }
}

export function updateXPBar() {
  const container = document.getElementById('xpBarContainer');
  if (container && currentChar) container.innerHTML = renderXPBar(currentChar);
}

export function updateField(field: string, value: any) {
  if (!currentChar) return;
  currentChar[field] = value;
  markDirty();
}

// ─── Compendium links for race/class/background (link-compendium-equipment-shops-npcs) ───

const IDENTITY_LINK = {
  race: { field: 'race', linkField: 'compendium_race_id', form: 'compendium_race_id', type: 'race' },
  class: { field: 'class', linkField: 'compendium_class_id', form: 'compendium_class_id', type: 'class' },
  background: { field: 'background', linkField: 'compendium_background_id', form: 'compendium_background_id', type: 'background' },
} as const;

export function linkCharIdentity(which: string) {
  if (!currentChar) return;
  const def = (IDENTITY_LINK as any)[which];
  if (!def) return;
  openCompendiumPicker({
    title: `Link ${capitalize(which)} from Compendium`,
    placeholder: `Search ${def.type}s...`,
    search: (q) => api('GET', `/api/compendium/search?q=${encodeURIComponent(q)}&type=${def.type}`),
    render: (e: any) => `<div><span class="fw-bold">${esc(e.name)}</span>${e.source ? `<span class="text-muted small ms-1">${esc(e.source)}</span>` : ''}</div>`,
    onPick: async (e: any) => {
      try {
        const fd = new FormData();
        fd.append(def.form, String(e.id));
        const headers: Record<string, string> = {};
        const apiToken = getApiToken();
        if (apiToken) headers['Authorization'] = `Bearer ${apiToken}`;
        const res = await fetch(`/api/characters/${currentChar.id}/${which}/link`, { method: 'POST', body: fd, headers, credentials: 'include' });
        if (!res.ok) throw new Error(((await res.json().catch(() => ({}))) as any).error || 'Link failed');
        currentChar[def.field] = e.name;
        currentChar[def.linkField] = e.id;
        markDirty();
        toast(`${capitalize(which)} linked from compendium`);
      } catch (err: any) { toast(err.message, true); }
    },
  });
}

export async function unlinkCharIdentity(which: string) {
  if (!currentChar) return;
  const def = (IDENTITY_LINK as any)[which];
  if (!def) return;
  try {
    await api('DELETE', `/api/characters/${currentChar.id}/${which}/link`);
    currentChar[def.linkField] = null;
    markDirty();
    toast(`${capitalize(which)} unlinked (text kept)`);
  } catch (e: any) { toast(e.message, true); }
}

// Window registrations (centralized)
expose('renderSheet', renderSheet);
expose('switchTab', switchTab);
expose('renderStepper', renderStepper);
expose('autoSaveField', autoSaveField);
expose('stepperField', stepperField);
expose('editStepperValue', editStepperValue);
expose('linkCharIdentity', linkCharIdentity);
expose('unlinkCharIdentity', unlinkCharIdentity);
// ─── Player UX: accordion sections + sticky quick actions (player-ux-overhaul) ───

function ensureSheetAccordion(): void {
  if (!currentChar) return;
  const c: any = currentChar;
  const classStr = [c.race, c.subclass, c.class].filter(Boolean).join(' ');
  const titles: Record<string, string> = { stats: 'Stats', combat: 'Combat', spells: 'Spells', inventory: 'Inventory', resources: 'Resources', features: 'Features', feats: 'Feats', companions: 'Companions', crafting: 'Crafting', locations: 'Locations', npcs: 'NPCs', sessions: 'Sessions', quests: 'Quests', journal: 'Journal', notes: 'Notes', graph: 'Graph', analytics: 'Analytics', details: 'Details', dice: 'Dice Roller' };
  for (const s of sections) {
    const el = document.getElementById(s + 'Section') as HTMLElement | null;
    if (!el || !el.parentElement) continue;
    let acc = document.getElementById('sheet-acc-' + s) as HTMLElement | null;
    if (!acc || !acc.contains(el)) {
      if (acc) acc.remove();
      acc = document.createElement('details');
      acc.className = 'sheet-section-acc';
      acc.id = 'sheet-acc-' + s;
      const sum = document.createElement('summary');
      sum.className = 'sheet-acc-summary';
      sum.innerHTML = `<i class="fa-solid fa-chevron-right me-2 sheet-acc-chevron"></i><span class="sheet-acc-title">${titles[s] || s}</span><span class="badge sheet-acc-count" id="${s}SectionCount"></span>`;
      sum.addEventListener('click', (ev) => { ev.preventDefault(); (window as any).switchTab?.(s); });
      acc.appendChild(sum);
      el.parentElement.insertBefore(acc, el);
      acc.appendChild(el); // DOM move preserves listeners + htmx attributes
    }
    acc.classList.toggle('sheet-acc-open', currentTab === s);
    (acc as any).open = currentTab === s;
    const countEl = document.getElementById(s + 'SectionCount') as HTMLElement | null;
    if (countEl) {
      let txt = '';
      if (s === 'stats') txt = `Lvl ${c.level} · ${classStr}`;
      else if (s === 'combat') txt = `HP ${c.hp_current}/${c.hp_max} · AC ${c.ac}`;
      else if (s === 'spells') { const sp = (c.spells || []).filter((x: any) => x.prepared || x.always_prepared); txt = `${sp.length} prepared`; }
      else if (s === 'inventory') txt = `${(c.inventory || []).length} items`;
      else if (s === 'resources') { const total = wealthTotalGp(); txt = total > 0 ? `${Math.round(total * 100) / 100} gp` : ''; }
      else if (s === 'features') txt = `${(c.features || []).length}`;
      else if (s === 'feats') txt = `${(c.feats || []).length}`;
      else if (s === 'companions') txt = `${(c.companions || []).length}`;
      countEl.textContent = txt;
      countEl.style.display = txt ? '' : 'none';
    }
  }
}

function ensureSheetQuickActions(): void {
  if (!currentChar) return;
  const sheetView = document.getElementById('sheetView');
  if (!sheetView) return;
  let bar = document.getElementById('sheetQuickActions') as HTMLElement | null;
  if (!bar) {
    bar = document.createElement('div');
    bar.id = 'sheetQuickActions';
    bar.className = 'sheet-quick-actions no-print';
    const header = sheetView.querySelector('h2') || sheetView.firstElementChild;
    sheetView.insertBefore(bar, header ? header.nextSibling : sheetView.firstChild);
  }
  const sess = document.body.classList.contains('session-mode');
  bar.innerHTML = sess
    ? `<button class="btn btn-sm btn-danger" onclick="applyDamage()"><i class="fa-solid fa-heart-crack me-1"></i>Damage</button><button class="btn btn-sm btn-success" onclick="applyHeal()"><i class="fa-solid fa-heart-pulse me-1"></i>Heal</button><button class="btn btn-sm btn-outline-secondary" onclick="showAddCondition()"><i class="fa-solid fa-skull me-1"></i>Condition</button>`
    : `<button class="btn btn-sm btn-outline-primary" onclick="rollAllInitiative()" title="Roll Initiative"><i class="fa-solid fa-dice-d20 me-1"></i>Initiative</button><button class="btn btn-sm btn-outline-success" onclick="doRest('short')" title="Short rest"><i class="fa-solid fa-mug-hot me-1"></i>Short Rest</button><button class="btn btn-sm btn-outline-primary" onclick="doRest('long')" title="Long rest"><i class="fa-solid fa-bed me-1"></i>Long Rest</button><button class="btn btn-sm btn-gold" onclick="doLevelUp()" title="Level up"><i class="fa-solid fa-arrow-up me-1"></i>Level Up</button>`;
  bar.style.display = sheetCanEdit() ? '' : 'none';
}

// Re-render the quick-action bar whenever session mode (body class) changes.
new MutationObserver(() => ensureSheetQuickActions()).observe(document.body, { attributes: true, attributeFilter: ['class'] });

expose('coinStepper', coinStepper);
expose('updateField', updateField);
expose('updateXPBar', updateXPBar);
