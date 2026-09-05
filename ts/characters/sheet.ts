/**
 * Character sheet rendering — tab bar, section switching, steppers, auto-save.
 * Extracted from app.ts (address-tech-debt-and-ux).
 */
import { expose } from '../lib/expose';
import { currentChar, currentTab, setCurrentTab } from '../lib/state';
import { esc, capitalize, toast, openCompendiumPicker } from '../lib/dom';
import { api, getApiToken } from '../lib/api';
import type { Character } from '../lib/api-types';
import { markDirty, isDirty, isSaving, saveCharacter } from '../lib/save';
import { renderDiceTab } from '../dice';
import { renderStats, renderXPBar } from './stats';
import { renderCombat } from './combat';
import { wealthTotalGp } from './resources';

declare const htmx: { process: (el: Element) => void };

import { sections } from '../lib/tabs';
export { sections };

const htmxTabs = ['spells', 'features', 'feats', 'companions', 'crafting', 'notes'];

type CharRecord = Character & Record<string, unknown>;

export function renderSheet(): void {
  if (!currentChar) return;
  const c = currentChar as CharRecord;
  const classes = c['classes'] as Array<{ class: string; subclass?: string; level: number }> | undefined;
  const multi = classes && classes.length > 0;
  const classStr = multi
    ? classes.map((cc) => `${cc.class}${cc.subclass ? ' (' + cc.subclass + ')' : ''} ${cc.level}`).join(' / ')
    : `${c['class'] as string}${c['subclass'] ? ' (' + (c['subclass'] as string) + ')' : ''}`;
  const sheetName = document.getElementById('sheetName');
  if (sheetName) {
    sheetName.innerHTML = c['portrait_url']
      ? `<img src="${esc(c['portrait_url'] as string)}" class="character-portrait me-2" alt="">${esc(c['name'] as string)}`
      : esc(c['name'] as string);
  }
  const subtitle = document.getElementById('sheetSubtitle');
  if (subtitle) subtitle.textContent = `${c['race'] as string} ${classStr} · Level ${c['level'] as number}`;

  const tabBar = document.getElementById('tabBar');
  if (tabBar) {
    tabBar.innerHTML = sections.map(s => `
      <li class="nav-item"><button class="nav-link ${s === currentTab ? 'active' : ''}" onclick="switchTab('${s}')">${capitalize(s)}</button></li>
    `).join('');
  }

  sections.forEach(s => {
    const el = document.getElementById(s + 'Section');
    if (el) el.style.display = s === currentTab ? 'block' : 'none';
  });

  renderStats();
  renderCombat();
  (window.renderCrafting as (() => void) | undefined)?.();
  (window.renderDetails as (() => void) | undefined)?.();
  renderDiceTab();
  applySheetReadonly();
  ensureSheetAccordion();
  ensureSheetQuickActions();
  (window.updateSaveBtnState as (() => void) | undefined)?.();
}

// Read-only mode for characters without edit rights (linked characters, campaign members)
function sheetCanEdit(): boolean {
  return window.canEditCharacter !== false;
}

function applySheetReadonly(): void {
  const el = document.getElementById('sheetView');
  if (!el) return;
  const ro = !sheetCanEdit();
  el.classList.toggle('readonly', ro);
  if (ro) {
    el.querySelectorAll('input, textarea, select').forEach((i) => { (i as HTMLInputElement).disabled = true; });
  }
}

export function switchTab(tab: string): void {
  if (isDirty() && !isSaving()) {
    void saveCharacter();
  }
  if (tab === 'party') { (window.showView as ((v: string) => void) | undefined)?.('party'); return; }
  setCurrentTab(tab);
  renderSheet();
  if (htmxTabs.includes(tab) && currentChar) {
    const el = document.getElementById(tab + 'Section');
    if (el) {
      el.setAttribute('hx-get', `/htmx/${tab}?character_id=${(currentChar as Character).id}`);
      el.setAttribute('hx-trigger', 'load');
      el.setAttribute('hx-swap', 'innerHTML');
      el.innerHTML = '<div class="ornament">✧ Loading... ✧</div>';
      htmx.process(el);
    }
  }
  if (currentChar) {
    switch (tab) {
      case 'inventory': (window.renderInventory as (() => void) | undefined)?.(); break;
      case 'resources': (window.renderResources as (() => void) | undefined)?.(); break;
      case 'journal': (window as unknown as Record<string, (() => void) | undefined>)['renderJournal']?.(); break;
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

export function autoSaveField(field: string, el: HTMLElement): void {
  const input = el as HTMLInputElement;
  const isCheckbox = input.type === 'checkbox';
  const isTextarea = el.tagName === 'TEXTAREA';
  const raw: unknown = isCheckbox ? input.checked : (el as HTMLInputElement).value;
  const num = parseFloat(String(raw));
  const finalVal: unknown = !isNaN(num) && !isCheckbox && !isTextarea ? num : raw;
  if (!currentChar) return;
  (currentChar as CharRecord)[field] = finalVal;
  markDirty();
}

export function stepperField(field: string, delta: number, min?: number, max?: number): void {
  if (!currentChar) return;
  const rec = currentChar as CharRecord;
  let val = (rec[field] as number | undefined) ?? 0;
  val += delta;
  if (min !== undefined) val = Math.max(min, val);
  if (max !== undefined) val = Math.min(max, val);
  rec[field] = val;
  markDirty();
  renderSheet();
}

export function editStepperValue(field: string, el: HTMLElement): void {
  if (!currentChar) return;
  const rec = currentChar as CharRecord;
  const current = (rec[field] as number | undefined) ?? 0;
  el.innerHTML = `<input type="number" class="form-control stepper-inline-input" value="${current}">`;
  const input = el.querySelector('input');
  if (!input) return;
  input.focus();
  input.select();
  const save = (): void => {
    const parsed = parseInt(input.value);
    if (!isNaN(parsed)) {
      rec[field] = parsed;
      markDirty();
    }
    renderSheet();
  };
  input.addEventListener('blur', save);
  input.addEventListener('keydown', (e: KeyboardEvent) => {
    if (e.key === 'Enter') { e.preventDefault(); save(); }
    if (e.key === 'Escape') { renderSheet(); }
  });
}

export async function coinStepper(coin: string, delta: number): Promise<void> {
  if (!currentChar) return;
  const rec = currentChar as CharRecord;
  const currency = (rec['currency'] as Record<string, number> | undefined) || {};
  const current = currency[coin] || 0;
  const newVal = Math.max(0, current + delta);
  currency[coin] = newVal;
  rec['currency'] = currency;
  const updates: Record<string, number> = {};
  ['cp','sp','ep','gp','pp'].forEach(c => { updates[c] = currency[c] || 0; });
  try {
    await api<void>('PUT', `/api/characters/${(currentChar as Character).id}/currency`, updates);
    toast(`${coin.toUpperCase()} ${delta > 0 ? '+' : ''}${delta}`);
  } catch (e) { toast((e as Error).message, true); }
  renderSheet();
  if (currentTab === 'resources') { (window.renderResources as (() => void) | undefined)?.(); }
}

export function updateXPBar(): void {
  const container = document.getElementById('xpBarContainer');
  if (container && currentChar) container.innerHTML = renderXPBar(currentChar as unknown as Record<string, unknown>);
}

export function updateField(field: string, value: unknown): void {
  if (!currentChar) return;
  (currentChar as CharRecord)[field] = value;
  markDirty();
}

// ─── Compendium links for race/class/background (link-compendium-equipment-shops-npcs) ───

const IDENTITY_LINK = {
  race: { field: 'race', linkField: 'compendium_race_id', form: 'compendium_race_id', type: 'race' },
  class: { field: 'class', linkField: 'compendium_class_id', form: 'compendium_class_id', type: 'class' },
  background: { field: 'background', linkField: 'compendium_background_id', form: 'compendium_background_id', type: 'background' },
} as const;

type IdentityKey = keyof typeof IDENTITY_LINK;

export function linkCharIdentity(which: string): void {
  if (!currentChar) return;
  const def = (IDENTITY_LINK as Record<string, typeof IDENTITY_LINK[IdentityKey]>)[which];
  if (!def) return;
  openCompendiumPicker({
    title: `Link ${capitalize(which)} from Compendium`,
    placeholder: `Search ${def.type}s...`,
    search: (q) => api<unknown[]>('GET', `/api/compendium/search?q=${encodeURIComponent(q)}&type=${def.type}`),
    render: (e: Record<string, unknown>) => `<div><span class="fw-bold">${esc(e['name'] as string)}</span>${e['source'] ? `<span class="text-muted small ms-1">${esc(e['source'] as string)}</span>` : ''}</div>`,
    onPick: (e: Record<string, unknown>) => {
      void (async () => {
        try {
          const fd = new FormData();
          fd.append(def.form, String(e['id']));
          const headers: Record<string, string> = {};
          const apiToken = getApiToken();
          if (apiToken) headers['Authorization'] = `Bearer ${apiToken}`;
          const res = await fetch(`/api/characters/${(currentChar as Character).id}/${which}/link`, { method: 'POST', body: fd, headers, credentials: 'include' });
          if (!res.ok) throw new Error(((await res.json().catch(() => ({}))) as { error?: string }).error || 'Link failed');
          (currentChar as CharRecord)[def.field] = e['name'] as string;
          (currentChar as CharRecord)[def.linkField] = e['id'] as number | string;
          markDirty();
          toast(`${capitalize(which)} linked from compendium`);
        } catch (err) { toast((err as Error).message, true); }
      })();
    },
  });
}

export async function unlinkCharIdentity(which: string): Promise<void> {
  if (!currentChar) return;
  const def = (IDENTITY_LINK as Record<string, typeof IDENTITY_LINK[IdentityKey]>)[which];
  if (!def) return;
  try {
    await api<void>('DELETE', `/api/characters/${(currentChar as Character).id}/${which}/link`);
    (currentChar as CharRecord)[def.linkField] = null;
    markDirty();
    toast(`${capitalize(which)} unlinked (text kept)`);
  } catch (e) { toast((e as Error).message, true); }
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
  const c = currentChar as CharRecord;
  const classStr = [(c['race'] as string), (c['subclass'] as string), (c['class'] as string)].filter(Boolean).join(' ');
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
      sum.addEventListener('click', (ev) => { ev.preventDefault(); (window.switchTab as ((t: string) => void) | undefined)?.(s); });
      acc.appendChild(sum);
      el.parentElement.insertBefore(acc, el);
      acc.appendChild(el);
    }
    acc.classList.toggle('sheet-acc-open', currentTab === s);
    (acc as HTMLDetailsElement).open = currentTab === s;
    const countEl = document.getElementById(s + 'SectionCount') as HTMLElement | null;
    if (countEl) {
      let txt = '';
      if (s === 'stats') txt = `Lvl ${c['level'] as number} · ${classStr}`;
      else if (s === 'combat') txt = `HP ${c['hp_current'] as number}/${c['hp_max'] as number} · AC ${c['ac'] as number}`;
      else if (s === 'spells') { const sp = ((c['spells'] as Array<Record<string, unknown>> | undefined) || []).filter((x) => x['prepared'] || x['always_prepared']); txt = `${sp.length} prepared`; }
      else if (s === 'inventory') txt = `${((c['inventory'] as unknown[]) || []).length} items`;
      else if (s === 'resources') { const total = wealthTotalGp(); txt = total > 0 ? `${Math.round(total * 100) / 100} gp` : ''; }
      else if (s === 'features') txt = `${((c['features'] as unknown[]) || []).length}`;
      else if (s === 'feats') txt = `${((c['feats'] as unknown[]) || []).length}`;
      else if (s === 'companions') txt = `${((c['companions'] as unknown[]) || []).length}`;
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
